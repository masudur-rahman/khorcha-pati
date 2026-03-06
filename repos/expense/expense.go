package expense

import (
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
		db:     db.Table("expense"),
		logger: logger,
	}
}

func (e *SQLExpenseRepository) GetLastExpense() (*models.Expense, error) {
	return nil, nil
}

func (e *SQLExpenseRepository) ListAllExpenses() ([]*models.Expense, error) {
	e.logger.Infow("listing all expenses")
	filter := models.Expense{}
	expenses := make([]*models.Expense, 0)
	err := e.db.FindMany(&expenses, filter)
	return expenses, err
}

func (e *SQLExpenseRepository) AddNewExpense(expense *models.Expense) error {
	e.logger.Infow("adding new expense")
	if expense.ID == "" {
		expense.ID = xid.New().String()
	}

	id, err := e.db.InsertOne(expense)
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
