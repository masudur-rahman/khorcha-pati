package expense

import (
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/models/gqtypes"
	"github.com/masudur-rahman/khorcha-pati/repos"
	"github.com/masudur-rahman/khorcha-pati/services"
)

type expenseService struct {
	expenseRepo repos.ExpenseRepository
}

var _ services.ExpenseService = &expenseService{}

func NewExpenseService(expenseRepo repos.ExpenseRepository) *expenseService {
	return &expenseService{expenseRepo: expenseRepo}
}

func (es *expenseService) AddExpense(params gqtypes.Expense) error {
	expense := &models.Expense{
		Amount:      params.Amount,
		Description: params.Description,
		Date:        time.Now(),
	}

	return es.expenseRepo.AddNewExpense(expense)
}

func (es *expenseService) ListExpenses() ([]*models.Expense, error) {
	expenses, err := es.expenseRepo.ListAllExpenses()
	return expenses, err
}
