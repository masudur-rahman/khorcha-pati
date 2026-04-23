package expense

import (
	"context"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"

	isql "github.com/masudur-rahman/styx/sql"

	"github.com/rs/xid"
)

type SQLExpenseRepository struct {
	db     isql.Engine
	logger logr.Logger
}

func NewSQLExpenseRepository(db isql.Engine, logger logr.Logger) *SQLExpenseRepository {
	return &SQLExpenseRepository{
		db:     db.Table(models.Expense{}.TableName()),
		logger: logger,
	}
}

func (e *SQLExpenseRepository) GetLastExpense() (*models.Expense, error) {
	return nil, nil
}

func (e *SQLExpenseRepository) ListAllExpenses() ([]*models.Expense, error) {
	e.logger.Infow("listing all expenses")
	ctx := context.Background()
	expenses := make([]*models.Expense, 0)
	err := e.db.FindMany(ctx, &expenses, models.Expense{})
	return expenses, err
}

func (e *SQLExpenseRepository) AddNewExpense(expense *models.Expense) error {
	e.logger.Infow("adding new expense")
	if expense.ID == "" {
		expense.ID = xid.New().String()
	}
	ctx := context.Background()
	id, err := e.db.MustCols("amount").InsertOne(ctx, expense)
	if err != nil {
		return err
	}
	e.logger.Infow("expense created", "id", id)
	return nil
}

func (e *SQLExpenseRepository) DeleteExpense(id string) error {
	return nil
}

func (e *SQLExpenseRepository) EditExpense(expense *models.Expense) error {
	return nil
}
