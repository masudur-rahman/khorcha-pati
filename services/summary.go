package services

import "github.com/masudur-rahman/khorcha-pati/models"

// SummaryService provides aggregated data for dashboard charts.
type SummaryService interface {
	GetMonthlyOverview(userID int64, year, month int) (*models.MonthlyOverview, error)
	GetExpenseByCategory(userID int64, year, month int) ([]models.CategorySpend, error)
	GetIncomeVsExpense(userID int64, months int) ([]models.MonthlyComparison, error)
}
