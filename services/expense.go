package services

import (
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/models/gqtypes"
)

type ExpenseService interface {
	AddExpense(params gqtypes.Expense) error
	ListExpenses() ([]*models.Expense, error)
}
