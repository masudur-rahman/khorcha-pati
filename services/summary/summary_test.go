package summary

import (
	"testing"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testUserID int64 = 42

func newTestService() (*summaryService, *mocks.TransactionRepo, *mocks.WalletRepo, *mocks.BudgetRepo) {
	txnRepo := &mocks.TransactionRepo{}
	walletRepo := &mocks.WalletRepo{}
	budgetRepo := &mocks.BudgetRepo{}
	svc := NewSummaryService(txnRepo, walletRepo, budgetRepo).(*summaryService)
	return svc, txnRepo, walletRepo, budgetRepo
}

func TestGetMonthlyOverview_success(t *testing.T) {
	t.Parallel()
	svc, txnRepo, walletRepo, budgetRepo := newTestService()

	txns := []models.Transaction{
		{Type: models.IncomeTransaction, Amount: 5000},
		{Type: models.ExpenseTransaction, Amount: 1500, SubcategoryID: "food-groc"},
		{Type: models.ExpenseTransaction, Amount: 500, SubcategoryID: "transport-fuel"},
	}
	txnRepo.On("ListTransactionsByTime", testUserID, models.TransactionType(""), mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return(txns, nil)
	walletRepo.On("ListWallets", testUserID).Return([]models.Wallet{
		{Balance: 10000}, {Balance: 5000},
	}, nil)
	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget{
		{CategoryID: "food", Amount: 3000},
	}, nil)

	overview, err := svc.GetMonthlyOverview(testUserID, 2026, 4)

	require.NoError(t, err)
	assert.Equal(t, 15000.0, overview.TotalBalance)
	assert.Equal(t, 5000.0, overview.MonthIncome)
	assert.Equal(t, 2000.0, overview.MonthExpense)
	assert.InDelta(t, 50.0, overview.BudgetUsage, 0.1) // 1500/3000 * 100
}

func TestGetExpenseByCategory_success(t *testing.T) {
	t.Parallel()
	svc, txnRepo, _, _ := newTestService()

	txns := []models.Transaction{
		{Type: models.ExpenseTransaction, Amount: 100, SubcategoryID: "food-groc"},
		{Type: models.ExpenseTransaction, Amount: 200, SubcategoryID: "food-rest"},
		{Type: models.ExpenseTransaction, Amount: 300, SubcategoryID: "transport-fuel"},
	}
	txnRepo.On("ListTransactionsByTime", testUserID, models.ExpenseTransaction, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return(txns, nil)

	cats, err := svc.GetExpenseByCategory(testUserID, 2026, 4)

	require.NoError(t, err)
	assert.Len(t, cats, 2) // food + transport

	catMap := make(map[string]float64)
	for _, c := range cats {
		catMap[c.CategoryID] = c.Amount
	}
	assert.Equal(t, 300.0, catMap["food"])
	assert.Equal(t, 300.0, catMap["transport"])
}

func TestGetIncomeVsExpense_success(t *testing.T) {
	t.Parallel()
	svc, txnRepo, _, _ := newTestService()

	txnRepo.On("ListTransactionsByTime", testUserID, models.TransactionType(""), mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{
			{Type: models.IncomeTransaction, Amount: 1000},
			{Type: models.ExpenseTransaction, Amount: 400},
		}, nil)

	result, err := svc.GetIncomeVsExpense(testUserID, 3)

	require.NoError(t, err)
	assert.Len(t, result, 3)

	now := time.Now()
	expected := now.Format("2006-01")
	assert.Equal(t, expected, result[2].Month)
	assert.Equal(t, 1000.0, result[2].Income)
	assert.Equal(t, 400.0, result[2].Expense)
}

func TestGetMonthlyOverview_noBudgets(t *testing.T) {
	t.Parallel()
	svc, txnRepo, walletRepo, budgetRepo := newTestService()

	txnRepo.On("ListTransactionsByTime", testUserID, models.TransactionType(""), mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{}, nil)
	walletRepo.On("ListWallets", testUserID).Return([]models.Wallet{}, nil)
	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget(nil), nil)

	overview, err := svc.GetMonthlyOverview(testUserID, 2026, 4)

	require.NoError(t, err)
	assert.Equal(t, 0.0, overview.BudgetUsage)
}
