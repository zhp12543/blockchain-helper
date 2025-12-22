package oklink

import "github.com/shopspring/decimal"

// Transfer 存储单笔转账信息
type Transfer struct {
	Sender   string
	Receiver string
	Amount   decimal.Decimal
	Token    string
}

// TxDetails 存储交易详情
type TxDetails struct {
	Hash      string
	Status    string
	Block     string
	Time      string
	Sender    string
	Receiver  string
	Amount    decimal.Decimal
	Token     string // 新增 Token 字段
	Fee       decimal.Decimal
	Transfers []Transfer
}

// LinkMap 用于存储页面上的链接信息
type LinkMap struct {
	Text string `json:"text"`
	Href string `json:"href"`
}
