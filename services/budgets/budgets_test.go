package budgets

import (
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testUserID int64 = 42

func newTestService() (*budgetService, *mocks.BudgetRepo, *mocks.TransactionRepo) {
	budgetRepo := &mocks.BudgetRepo{}
	txnRepo := &mocks.TransactionRepo{}
	svc := NewBudgetService(budgetRepo, txnRepo)
	return svc, budgetRepo, txnRepo
}

func TestSetBudget_success(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, _ := newTestService()

	budgetRepo.On("UpsertBudget", mock.AnythingOfType("*models.Budget")).Return(nil)

	err := svc.SetBudget(testUserID, "food", 5000, 80)

	assert.NoError(t, err)
	budgetRepo.AssertExpectations(t)
}

func TestSetBudget_invalidAmount(t *testing.T) {
	t.Parallel()
	svc, _, _ := newTestService()

	err := svc.SetBudget(testUserID, "food", 0, 80)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "positive")
}

func TestSetBudget_defaultAlertAt(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, _ := newTestService()

	budgetRepo.On("UpsertBudget", mock.MatchedBy(func(b *models.Budget) bool {
		return b.AlertAt == defaultAlertAt
	})).Return(nil)

	err := svc.SetBudget(testUserID, "food", 5000, 0)

	assert.NoError(t, err)
	budgetRepo.AssertExpectations(t)
}

func TestDeleteBudget_success(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, _ := newTestService()

	budgetRepo.On("DeleteBudget", testUserID, "food").Return(nil)

	err := svc.DeleteBudget(testUserID, "food")

	assert.NoError(t, err)
	budgetRepo.AssertExpectations(t)
}

func TestGetBudgetStatus_success(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, txnRepo := newTestService()

	budgetRepo.On("GetBudget", testUserID, "food").Return(&models.Budget{
		UserID:     testUserID,
		CategoryID: "food",
		Amount:     5000,
		AlertAt:    80,
	}, nil)

	txnRepo.On("ListTransactionsByTime", testUserID, models.ExpenseTransaction, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{
			{Amount: 2000, SubcategoryID: "food-groc"},
			{Amount: 1000, SubcategoryID: "food-rest"},
			{Amount: 500, SubcategoryID: "trans-taxi"},
		}, nil)

	txnRepo.On("GetTxnCategoryName", "food").Return("Food", nil)

	status, err := svc.GetBudgetStatus(testUserID, "food")

	assert.NoError(t, err)
	assert.Equal(t, float64(3000), status.Spent)
	assert.Equal(t, float64(2000), status.Remaining)
	assert.Equal(t, float64(60), status.Percent)
	assert.Equal(t, "Food", status.CategoryName)
}

func TestCheckBudgetAlerts_noAlerts(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, txnRepo := newTestService()

	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget{
		{CategoryID: "food", Amount: 5000, AlertAt: 80},
	}, nil)

	txnRepo.On("ListTransactionsByTime", testUserID, models.ExpenseTransaction, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{
			{Amount: 1000, SubcategoryID: "food-groc"},
		}, nil)

	txnRepo.On("GetTxnCategoryName", "food").Return("Food", nil)

	alerts, err := svc.CheckBudgetAlerts(testUserID, "food-groc")

	assert.NoError(t, err)
	assert.Empty(t, alerts)
}

func TestCheckBudgetAlerts_warning(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, txnRepo := newTestService()

	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget{
		{CategoryID: "food", Amount: 5000, AlertAt: 80},
	}, nil)

	txnRepo.On("ListTransactionsByTime", testUserID, models.ExpenseTransaction, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{
			{Amount: 4200, SubcategoryID: "food-groc"},
		}, nil)

	txnRepo.On("GetTxnCategoryName", "food").Return("Food", nil)

	alerts, err := svc.CheckBudgetAlerts(testUserID, "food-groc")

	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.False(t, alerts[0].Exceeded)
	assert.Equal(t, float64(84), alerts[0].Percent)
}

func TestCheckBudgetAlerts_exceeded(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, txnRepo := newTestService()

	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget{
		{CategoryID: "food", Amount: 5000, AlertAt: 80},
	}, nil)

	txnRepo.On("ListTransactionsByTime", testUserID, models.ExpenseTransaction, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{
			{Amount: 5500, SubcategoryID: "food-rest"},
		}, nil)

	txnRepo.On("GetTxnCategoryName", "food").Return("Food", nil)

	alerts, err := svc.CheckBudgetAlerts(testUserID, "food-rest")

	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.True(t, alerts[0].Exceeded)
	assert.InDelta(t, 110, alerts[0].Percent, 0.01)
}

func TestCheckBudgetAlerts_categoryAndOverall(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, txnRepo := newTestService()

	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget{
		{CategoryID: "food", Amount: 5000, AlertAt: 80},
		{CategoryID: "", Amount: 20000, AlertAt: 80},
	}, nil)

	txnRepo.On("ListTransactionsByTime", testUserID, models.ExpenseTransaction, mock.AnythingOfType("int64"), mock.AnythingOfType("int64")).
		Return([]models.Transaction{
			{Amount: 4500, SubcategoryID: "food-groc"},
			{Amount: 12000, SubcategoryID: "house-rent"},
		}, nil)

	txnRepo.On("GetTxnCategoryName", "food").Return("Food", nil)

	alerts, err := svc.CheckBudgetAlerts(testUserID, "food-groc")

	assert.NoError(t, err)
	assert.Len(t, alerts, 2)
}

func TestListBudgetStatuses_empty(t *testing.T) {
	t.Parallel()
	svc, budgetRepo, _ := newTestService()

	budgetRepo.On("ListBudgets", testUserID).Return([]models.Budget(nil), nil)

	statuses, err := svc.ListBudgetStatuses(testUserID)

	assert.NoError(t, err)
	assert.Nil(t, statuses)
}
