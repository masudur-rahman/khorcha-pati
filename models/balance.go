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
	ID               int64 `db:"id,pk autoincr"`
	UserID           int64 `db:",uqs"`
	Type             WalletType
	ShortName        string `db:",uqs"`
	Name             string
	Balance          float64
	LastTxnAmount    float64
	LastTxnTimestamp int64
	Version          int64 `db:"version"`
}

func (Wallet) TableName() string {
	return "wallet"
}

type Event struct {
	ID        int64 `db:"id,pk autoincr"`
	UserID    int64
	Message   string
	Timestamp int64
}

func (Event) TableName() string {
	return "event"
}
