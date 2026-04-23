package models

// Budget represents a monthly spending limit for a category or overall.
type Budget struct {
	ID         int64   `db:"id,pk autoincr" json:"id"`
	UserID     int64   `db:",uqs" json:"userId"`
	CategoryID string  `db:",uqs" json:"categoryId"` // TxnCategory.ID or "" for overall
	Amount     float64 `db:"amount" json:"amount"`
	AlertAt    int64   `db:"alert_at" json:"alertAt"` // percent threshold (default 80)
	CreatedAt  int64   `db:"created_at" json:"createdAt"`
	UpdatedAt  int64   `db:"updated_at" json:"updatedAt"`
}

// TableName returns the database table name.
func (Budget) TableName() string {
	return "budget"
}

// BudgetStatus is a computed view of a budget with current month spending.
type BudgetStatus struct {
	Budget
	CategoryName string  `json:"categoryName"`
	Spent        float64 `json:"spent"`
	Remaining    float64 `json:"remaining"`
	Percent      float64 `json:"percent"`
}

// BudgetAlert is returned when spending crosses a budget threshold.
type BudgetAlert struct {
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Spent        float64 `json:"spent"`
	BudgetAmount float64 `json:"budgetAmount"`
	Percent      float64 `json:"percent"`
	Exceeded     bool    `json:"exceeded"`
}
