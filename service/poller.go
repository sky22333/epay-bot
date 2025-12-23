package service

import (
	"epay-bot/db"
	"epay-bot/model"
	"fmt"
	"log"
	"sync"
	"time"
)

type Notifier interface {
	NotifyOrder(chatID int64, order model.Order) error
	NotifySettlement(chatID int64, settlement model.Settlement) error
}

type PollerManager struct {
	db       *db.DB
	epay     *EpayService
	notifier Notifier
	jobs     map[int64]*pollJob
	mu       sync.RWMutex
	stopCh   chan struct{}
}

type pollJob struct {
	chatID   int64
	interval time.Duration
	stop     chan struct{}
}

func NewPollerManager(database *db.DB, epay *EpayService, notifier Notifier) *PollerManager {
	return &PollerManager{
		db:       database,
		epay:     epay,
		notifier: notifier,
		jobs:     make(map[int64]*pollJob),
		stopCh:   make(chan struct{}),
	}
}

func (pm *PollerManager) Start() {
	// Load all active polling users from DB
	activeChats, err := pm.db.GetActivePollingChats()
	if err != nil {
		log.Printf("Failed to load active polling chats: %v", err)
		return
	}

	for _, chatID := range activeChats {
		pm.StartPolling(chatID)
	}
}

func (pm *PollerManager) Stop() {
	close(pm.stopCh)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, job := range pm.jobs {
		close(job.stop)
	}
	pm.jobs = make(map[int64]*pollJob)
}

func (pm *PollerManager) StartPolling(chatID int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.jobs[chatID]; exists {
		return
	}

	job := &pollJob{
		chatID:   chatID,
		interval: 2 * time.Second,
		stop:     make(chan struct{}),
	}
	pm.jobs[chatID] = job

	go pm.runJob(job)
}

func (pm *PollerManager) StopPolling(chatID int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if job, exists := pm.jobs[chatID]; exists {
		close(job.stop)
		delete(pm.jobs, chatID)
	}
}

func (pm *PollerManager) runJob(job *pollJob) {
	log.Printf("Polling started for chat %d", job.chatID)
	defer log.Printf("Polling stopped for chat %d", job.chatID)

	ticker := time.NewTicker(job.interval)
	defer ticker.Stop()

	consecutiveErrors := 0
	maxErrors := 10

	for {
		select {
		case <-job.stop:
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
			// Adjust ticker if interval changed
			// Note: time.Ticker doesn't support dynamic update easily in loop,
			// usually we just reset it or use time.After/Sleep.
			// For simplicity and adaptiveness, let's use time.Sleep loop instead of ticker for next iteration
		}

		// Perform check
		info, err := pm.db.GetMerchantInfo(job.chatID)
		if err != nil || info == nil {
			log.Printf("Merchant info missing for chat %d", job.chatID)
			// Wait a bit
			time.Sleep(60 * time.Second)
			continue
		}

		pm.db.UpdateLastPollTime(job.chatID)

		// 检查订单
		var errOrder, errSettle error
		orders, err := pm.epay.GetOrders(info.Domain, info.Pid, info.Key)
		if err != nil {
			log.Printf("Error getting orders for %d: %v", job.chatID, err)
			errOrder = err
		} else if len(orders) > 0 {
			for _, order := range orders {
				// 状态 1 表示成功
				status := fmt.Sprintf("%v", order.Status)
				if status == "1" {
					notified, err := pm.db.IsOrderNotified(order.TradeNo, job.chatID)
					if err != nil {
						log.Printf("警告: 检查订单是否已通知时数据库出错 (ChatID: %d, Order: %s): %v", job.chatID, order.TradeNo, err)
						continue
					}
					if !notified {
						pm.notifier.NotifyOrder(job.chatID, order)
						if err := pm.db.MarkOrderNotified(order.TradeNo, job.chatID); err != nil {
							log.Printf("警告: 标记订单为已通知失败 (ChatID: %d, Order: %s): %v", job.chatID, order.TradeNo, err)
						}
					}
				}
			}
		}

		// 检查结算
		settlements, err := pm.epay.GetSettlements(info.Domain, info.Pid, info.Key)
		if err != nil {
			log.Printf("Error getting settlements for %d: %v", job.chatID, err)
			errSettle = err
		} else if len(settlements) > 0 {
			for _, settle := range settlements {
				status := fmt.Sprintf("%v", settle.Status)
				if status == "1" {
					notified, err := pm.db.IsSettlementNotified(settle.ID.String(), job.chatID)
					if err != nil {
						log.Printf("警告: 检查结算是否已通知时数据库出错 (ChatID: %d, SettleID: %s): %v", job.chatID, settle.ID, err)
						continue
					}
					if !notified {
						if err := pm.notifier.NotifySettlement(job.chatID, settle); err != nil {
							continue
						}
						if err := pm.db.MarkSettlementNotified(settle.ID.String(), job.chatID); err != nil {
							log.Printf("警告: 标记结算为已通知失败 (ChatID: %d, SettleID: %s): %v", job.chatID, settle.ID, err)
						}
					}
				}
			}
		}

		// 固定间隔逻辑与错误退避
		if errOrder != nil || errSettle != nil {
			consecutiveErrors++
			if consecutiveErrors >= maxErrors {
				job.interval = 30 * time.Second
			}
		} else {
			// 成功
			consecutiveErrors = 0
			job.interval = 2 * time.Second
		}

		// Wait for next tick
		ticker.Reset(job.interval)
	}
}


