package models

// MonthlyOverview holds aggregate totals for the dashboard hero cards.
type MonthlyOverview struct {
	TotalBalance float64 `json:"totalBalance"`
	MonthIncome  float64 `json:"monthIncome"`
	MonthExpense float64 `json:"monthExpense"`
	BudgetUsage  float64 `json:"budgetUsage"` // percent 0-100
}

// CategorySpend represents spending in a single category.
type CategorySpend struct {
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Amount       float64 `json:"amount"`
	Percent      float64 `json:"percent"`
}

// MonthlyComparison holds income vs expense for a single month.
type MonthlyComparison struct {
	Month   string  `json:"month"` // "2026-01"
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}
