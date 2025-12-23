package main

import (
	"epay-bot/bot"
	"epay-bot/db"
	"epay-bot/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		token = "YOUR_BOT_TOKEN_HERE"
		log.Println("警告: 未设置 TELEGRAM_BOT_TOKEN，正在使用默认占位符")
	}

	// Initialize DB
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("无法创建数据目录: %v", err)
	}
	database, err := db.NewDB("data/epay.db")
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	defer database.Close()

	// Initialize Service
	epayService := service.NewEpayService()

	// Initialize Bot
	b, err := bot.NewBot(token, database, epayService)
	if err != nil {
		log.Fatalf("无法创建机器人: %v", err)
	}

	// Start Bot
	go b.Start()

	// Start periodic cleanup
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			log.Println("正在清理旧记录...")
			if err := database.CleanOldRecords(730); err != nil {
				log.Printf("清理旧记录失败: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	b.Stop()
	log.Println("机器人已停止")
}
