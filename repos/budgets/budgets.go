package budgets

import (
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	isql "github.com/masudur-rahman/styx/sql"
)

// SQLBudgetRepository implements BudgetRepository with styx ORM.
type SQLBudgetRepository struct {
	db     isql.Engine
	logger logr.Logger
}

// NewSQLBudgetRepository creates a new budget repository.
func NewSQLBudgetRepository(db isql.Engine, logger logr.Logger) *SQLBudgetRepository {
	return &SQLBudgetRepository{
		db:     db.Table(models.Budget{}.TableName()),
		logger: logger,
	}
}

var _ repos.BudgetRepository = &SQLBudgetRepository{}

// GetBudget returns a single budget by user and category.
func (r *SQLBudgetRepository) GetBudget(userID int64, categoryID string) (*models.Budget, error) {
	r.logger.Infow("get budget", "userID", userID, "categoryID", categoryID)
	var b models.Budget
	// Use explicit Where for category_id because styx skips zero-value "" in struct filters
	found, err := r.db.Where("category_id = ?", categoryID).FindOne(&b, models.Budget{UserID: userID})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrBudgetNotFound{UserID: userID, CategoryID: categoryID}
	}
	return &b, nil
}

// ListBudgets returns all budgets for a user.
func (r *SQLBudgetRepository) ListBudgets(userID int64) ([]models.Budget, error) {
	r.logger.Infow("list budgets", "userID", userID)
	result := make([]models.Budget, 0)
	err := r.db.FindMany(&result, models.Budget{UserID: userID})
	return result, err
}

// UpsertBudget inserts or updates a budget for the given user+category.
func (r *SQLBudgetRepository) UpsertBudget(budget *models.Budget) error {
	r.logger.Infow("upsert budget", "userID", budget.UserID, "categoryID", budget.CategoryID)
	existing, err := r.GetBudget(budget.UserID, budget.CategoryID)
	if err != nil && !models.IsErrNotFound(err) {
		return err
	}

	now := time.Now().Unix()
	if existing != nil {
		existing.Amount = budget.Amount
		existing.AlertAt = budget.AlertAt
		existing.UpdatedAt = now
		return r.db.ID(existing.ID).MustCols("amount", "alert_at", "category_id").UpdateOne(existing)
	}

	budget.CreatedAt = now
	budget.UpdatedAt = now
	_, err = r.db.MustCols("alert_at", "category_id").InsertOne(budget)
	return err
}

// DeleteBudget removes a budget for the given user+category.
func (r *SQLBudgetRepository) DeleteBudget(userID int64, categoryID string) error {
	r.logger.Infow("delete budget", "userID", userID, "categoryID", categoryID)
	// Use explicit Where for category_id because styx skips zero-value "" in struct filters
	return r.db.Where("category_id = ?", categoryID).DeleteOne(models.Budget{UserID: userID})
}
