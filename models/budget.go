package models

// Budget represents a monthly spending limit for a category or overall.
type Budget struct {
	ID         int64   `db:"id,pk autoincr"`
	UserID     int64   `db:",uqs"`
	CategoryID string  `db:",uqs"` // TxnCategory.ID or "" for overall
	Amount     float64 `db:"amount"`
	AlertAt    int64   `db:"alert_at"` // percent threshold (default 80)
	CreatedAt  int64   `db:"created_at"`
	UpdatedAt  int64   `db:"updated_at"`
}

// TableName returns the database table name.
func (Budget) TableName() string {
	return "budget"
}

// BudgetStatus is a computed view of a budget with current month spending.
type BudgetStatus struct {
	Budget
	CategoryName string
	Spent        float64
	Remaining    float64
	Percent      float64
}

// BudgetAlert is returned when spending crosses a budget threshold.
type BudgetAlert struct {
	CategoryName string
	Spent        float64
	Limit        float64
	Percent      float64
	Exceeded     bool
}
