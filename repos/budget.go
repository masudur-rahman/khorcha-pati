package repos

import "github.com/masudur-rahman/expense-tracker-bot/models"

// BudgetRepository defines data access for budgets.
type BudgetRepository interface {
	GetBudget(userID int64, categoryID string) (*models.Budget, error)
	ListBudgets(userID int64) ([]models.Budget, error)
	UpsertBudget(budget *models.Budget) error
	DeleteBudget(userID int64, categoryID string) error
}
