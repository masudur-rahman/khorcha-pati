package models

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
	Version          int `db:"version"`
}

type Event struct {
	ID        int64 `db:"id,pk autoincr"`
	UserID    int64
	Message   string
	Timestamp int64
}
