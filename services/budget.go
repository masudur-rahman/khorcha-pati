package services

import "github.com/masudur-rahman/expense-tracker-bot/models"

// BudgetService defines business logic for budgeting.
type BudgetService interface {
	SetBudget(userID int64, categoryID string, amount float64, alertAt int64) error
	GetBudgetStatus(userID int64, categoryID string) (*models.BudgetStatus, error)
	ListBudgetStatuses(userID int64) ([]models.BudgetStatus, error)
	DeleteBudget(userID int64, categoryID string) error
	CheckBudgetAlerts(userID int64, subcategoryID string) ([]models.BudgetAlert, error)
}
