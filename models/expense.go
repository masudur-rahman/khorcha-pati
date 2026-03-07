package models

import "time"

type Expense struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (Expense) TableName() string {
	return "expense"
}
