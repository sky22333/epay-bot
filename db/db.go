package db

import (
	"database/sql"
	"epay-bot/model"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func NewDB(path string) (*DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=2000", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 使用单连接模式，避免 database is locked 造成重复通知
	db.SetMaxOpenConns(1)

	d := &DB{db}
	if err := d.init(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DB) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS notified_orders (
            trade_no TEXT,
            chat_id INTEGER,
            notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (trade_no, chat_id)
        )`,
		`CREATE TABLE IF NOT EXISTS notified_settlements (
            settlement_id TEXT,
            chat_id INTEGER,
            notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (settlement_id, chat_id)
        )`,
		`CREATE TABLE IF NOT EXISTS merchant_info (
            chat_id INTEGER PRIMARY KEY,
            domain TEXT,
            pid TEXT,
            key TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS polling_status (
            chat_id INTEGER PRIMARY KEY,
            active INTEGER DEFAULT 0,
            last_poll TIMESTAMP
        )`,
	}

	for _, query := range queries {
		if _, err := d.Exec(query); err != nil {
			return fmt.Errorf("init db error: %w", err)
		}
	}
	return nil
}

func (d *DB) IsOrderNotified(tradeNo string, chatID int64) (bool, error) {
	var exists int
	err := d.QueryRow("SELECT 1 FROM notified_orders WHERE trade_no = ? AND chat_id = ?", tradeNo, chatID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DB) MarkOrderNotified(tradeNo string, chatID int64) error {
	_, err := d.Exec("INSERT OR REPLACE INTO notified_orders (trade_no, chat_id) VALUES (?, ?)", tradeNo, chatID)
	return err
}

func (d *DB) IsSettlementNotified(settlementID string, chatID int64) (bool, error) {
	var exists int
	err := d.QueryRow("SELECT 1 FROM notified_settlements WHERE settlement_id = ? AND chat_id = ?", settlementID, chatID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DB) MarkSettlementNotified(settlementID string, chatID int64) error {
	_, err := d.Exec("INSERT OR REPLACE INTO notified_settlements (settlement_id, chat_id) VALUES (?, ?)", settlementID, chatID)
	return err
}

func (d *DB) SaveMerchantInfo(info model.MerchantInfo) error {
	_, err := d.Exec("INSERT OR REPLACE INTO merchant_info (chat_id, domain, pid, key) VALUES (?, ?, ?, ?)",
		info.ChatID, info.Domain, info.Pid, info.Key)
	return err
}

func (d *DB) GetMerchantInfo(chatID int64) (*model.MerchantInfo, error) {
	var info model.MerchantInfo
	err := d.QueryRow("SELECT domain, pid, key FROM merchant_info WHERE chat_id = ?", chatID).
		Scan(&info.Domain, &info.Pid, &info.Key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info.ChatID = chatID
	return &info, nil
}

func (d *DB) GetAllMerchantInfo() ([]model.MerchantInfo, error) {
	rows, err := d.Query("SELECT chat_id, domain, pid, key FROM merchant_info")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infos []model.MerchantInfo
	for rows.Next() {
		var info model.MerchantInfo
		if err := rows.Scan(&info.ChatID, &info.Domain, &info.Pid, &info.Key); err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (d *DB) SetPollingStatus(chatID int64, active bool) error {
	val := 0
	if active {
		val = 1
	}
	_, err := d.Exec("INSERT OR REPLACE INTO polling_status (chat_id, active, last_poll) VALUES (?, ?, CURRENT_TIMESTAMP)", chatID, val)
	return err
}

func (d *DB) GetPollingStatus(chatID int64) (bool, error) {
	var active int
	err := d.QueryRow("SELECT active FROM polling_status WHERE chat_id = ?", chatID).Scan(&active)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return active == 1, nil
}

func (d *DB) UpdateLastPollTime(chatID int64) error {
	_, err := d.Exec("UPDATE polling_status SET last_poll = CURRENT_TIMESTAMP WHERE chat_id = ?", chatID)
	return err
}

func (d *DB) GetActivePollingChats() ([]int64, error) {
	rows, err := d.Query("SELECT chat_id FROM polling_status WHERE active = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		chats = append(chats, chatID)
	}
	return chats, nil
}

func (d *DB) CleanOldRecords(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	_, err := d.Exec("DELETE FROM notified_orders WHERE notified_at < ?", cutoff)
	if err != nil {
		return err
	}
	_, err = d.Exec("DELETE FROM notified_settlements WHERE notified_at < ?", cutoff)
	return err
}
