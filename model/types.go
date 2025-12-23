package model

import (
	"encoding/json"
	"time"
)

// EpayResponse represents the standard response from the epay API
type EpayResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []T    `json:"data"`
}

// Order represents an order from the epay API
type Order struct {
	TradeNo    string      `json:"trade_no"`
	OutTradeNo string      `json:"out_trade_no"`
	Type       string      `json:"type"`
	Pid        json.Number `json:"pid"`
	Addtime    string      `json:"addtime"`
	Endtime    string      `json:"endtime"`
	Name       string      `json:"name"`
	Money      string      `json:"money"`
	Status     interface{} `json:"status"`
}

// Settlement represents a settlement from the epay API
type Settlement struct {
	ID        json.Number `json:"id"`
	Pid       json.Number `json:"pid"`
	Account   string      `json:"account"`
	Money     string      `json:"money"`
	Realmoney string      `json:"realmoney"`
	Addtime   string      `json:"addtime"`
	Endtime   string      `json:"endtime"`
	Status    interface{} `json:"status"`
}

// MerchantInfo represents the merchant configuration for a user
type MerchantInfo struct {
	ChatID int64
	Domain string
	Pid    string
	Key    string
}

// PollingStatus represents the polling status for a user
type PollingStatus struct {
	ChatID   int64
	Active   bool
	LastPoll time.Time
}
