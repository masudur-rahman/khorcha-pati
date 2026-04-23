package models

import "fmt"

// ErrOptimisticLock is returned when a concurrent wallet modification is detected.
var ErrOptimisticLock = fmt.Errorf("wallet was modified concurrently, please retry")

type WalletType string

const (
	CashAccount WalletType = "Cash"
	BankAccount WalletType = "Bank"
)

type Wallet struct {
	ID               int64      `db:"id,pk autoincr" json:"id"`
	UserID           int64      `db:",uqs" json:"userId"`
	Type             WalletType `json:"type"`
	ShortName        string     `db:",uqs" json:"shortName"`
	Name             string     `json:"name"`
	Balance          float64    `db:"balance" json:"balance"`
	LastTxnAmount    float64    `db:"last_txn_amount" json:"lastTxnAmount"`
	LastTxnTimestamp int64      `db:"last_txn_timestamp" json:"lastTxnTimestamp"`
	Version          int64      `db:"version" json:"version"`
}

func (Wallet) TableName() string {
	return "wallet"
}

type Event struct {
	ID        int64  `db:"id,pk autoincr" json:"id"`
	UserID    int64  `json:"userId"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func (Event) TableName() string {
	return "event"
}
