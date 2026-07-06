package transaction

import (
	"fmt"
	"strings"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos/mocks"

	"github.com/masudur-rahman/styx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testUserID int64 = 42

func newTestService() (*txnService, *mocks.WalletRepo, *mocks.ContactRepo, *mocks.TransactionRepo) {
	walletRepo := &mocks.WalletRepo{}
	contactRepo := &mocks.ContactRepo{}
	txnRepo := &mocks.TransactionRepo{}
	evRepo := &mocks.EventRepo{}

	svc := NewTxnService(styx.UnitOfWork{}, walletRepo, contactRepo, txnRepo, evRepo)
	return svc, walletRepo, contactRepo, txnRepo
}

func TestAddTransaction_expense(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
	}

	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -100.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err := svc.AddTransaction(txn)

	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}

func TestAddTransaction_income(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        500,
		SubcategoryID: "fin-sal",
		Type:          models.IncomeTransaction,
		DstID:         "bank",
	}

	walletRepo.On("UpdateWalletBalance", testUserID, "bank", 500.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err := svc.AddTransaction(txn)

	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
}

func TestAddTransaction_transfer(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        200,
		SubcategoryID: "fin-transfer",
		Type:          models.TransferTransaction,
		SrcID:         "cash",
		DstID:         "bank",
	}

	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -200.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "bank", 200.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err := svc.AddTransaction(txn)

	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
}

func TestAddTransaction_missingUserID(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestService()

	txn := models.Transaction{Amount: 100, SubcategoryID: "food-groc"}

	err := svc.AddTransaction(txn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userid is required")
}

func TestAddTransaction_missingSubcategory(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestService()

	txn := models.Transaction{UserID: testUserID, Amount: 100}

	err := svc.AddTransaction(txn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subcategory is required")
}

func TestAddTransaction_expenseWithLend(t *testing.T) {
	t.Parallel()
	svc, walletRepo, contactRepo, txnRepo := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        300,
		SubcategoryID: models.LendSubID,
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		ContactName:   "alice",
	}

	contactRepo.On("GetContactByName", testUserID, "alice").
		Return(&models.Contacts{ID: 1, NickName: "alice"}, nil)
	contactRepo.On("UpdateContactBalance", int64(1), 300.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -300.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err := svc.AddTransaction(txn)

	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
}

func TestAddTransaction_incomeWithBorrow(t *testing.T) {
	t.Parallel()
	svc, walletRepo, contactRepo, txnRepo := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        150,
		SubcategoryID: models.BorrowSubID,
		Type:          models.IncomeTransaction,
		DstID:         "cash",
		ContactName:   "bob",
	}

	contactRepo.On("GetContactByName", testUserID, "bob").
		Return(&models.Contacts{ID: 2, NickName: "bob"}, nil)
	contactRepo.On("UpdateContactBalance", int64(2), -150.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 150.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err := svc.AddTransaction(txn)

	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
}

func TestAddTransaction_invalidSubcategoryForExpense(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: models.BorrowSubID,
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
	}

	err := svc.AddTransaction(txn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed for transaction type")
}

func TestAddTransaction_invalidSubcategoryForIncome(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: models.LendSubID,
		Type:          models.IncomeTransaction,
		DstID:         "cash",
	}

	err := svc.AddTransaction(txn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed for transaction type")
}

func TestAddTransaction_walletUpdateFails(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, _ := newTestService()

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
	}

	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -100.0).
		Return(models.ErrOptimisticLock)

	err := svc.AddTransaction(txn)

	assert.ErrorIs(t, err, models.ErrOptimisticLock)
}

func TestUndo_success(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	existing := &models.Transaction{
		ID:            1,
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
	}

	txnRepo.On("GetLastActiveTransaction", testUserID).Return(existing, nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 100.0).Return(nil)
	txnRepo.On("SoftDeleteTransaction", int64(1), mock.AnythingOfType("int64")).Return(nil)

	result, err := svc.Undo(testUserID)

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	walletRepo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}

func TestUndo_noTransactions(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	txnRepo.On("GetLastActiveTransaction", testUserID).
		Return(nil, fmt.Errorf("not found"))

	result, err := svc.Undo(testUserID)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to undo")
}

func TestUndo_incomeReversal(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	existing := &models.Transaction{
		ID:            2,
		UserID:        testUserID,
		Amount:        500,
		SubcategoryID: "fin-sal",
		Type:          models.IncomeTransaction,
		DstID:         "bank",
	}

	txnRepo.On("GetLastActiveTransaction", testUserID).Return(existing, nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "bank", -500.0).Return(nil)
	txnRepo.On("SoftDeleteTransaction", int64(2), mock.AnythingOfType("int64")).Return(nil)

	result, err := svc.Undo(testUserID)

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
}

func TestUndo_transferReversal(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	existing := &models.Transaction{
		ID:            3,
		UserID:        testUserID,
		Amount:        200,
		SubcategoryID: "fin-transfer",
		Type:          models.TransferTransaction,
		SrcID:         "cash",
		DstID:         "bank",
	}

	txnRepo.On("GetLastActiveTransaction", testUserID).Return(existing, nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 200.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "bank", -200.0).Return(nil)
	txnRepo.On("SoftDeleteTransaction", int64(3), mock.AnythingOfType("int64")).Return(nil)

	result, err := svc.Undo(testUserID)

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
}

func TestUndo_lendReversal(t *testing.T) {
	t.Parallel()
	svc, walletRepo, contactRepo, txnRepo := newTestService()

	existing := &models.Transaction{
		ID:            4,
		UserID:        testUserID,
		Amount:        300,
		SubcategoryID: models.LendSubID,
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		ContactName:   "alice",
	}

	txnRepo.On("GetLastActiveTransaction", testUserID).Return(existing, nil)
	contactRepo.On("GetContactByName", testUserID, "alice").
		Return(&models.Contacts{ID: 1}, nil)
	contactRepo.On("UpdateContactBalance", int64(1), -300.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 300.0).Return(nil)
	txnRepo.On("SoftDeleteTransaction", int64(4), mock.AnythingOfType("int64")).Return(nil)

	result, err := svc.Undo(testUserID)

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	contactRepo.AssertExpectations(t)
}

func TestListTransactions_success(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	expected := []models.Transaction{
		{ID: 1, UserID: testUserID, Amount: 100},
		{ID: 2, UserID: testUserID, Amount: 200},
	}

	txnRepo.On("ListTransactions", models.Transaction{UserID: testUserID}).
		Return(expected, nil)

	result, err := svc.ListTransactions(testUserID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestListTransactionsByType_success(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	expected := []models.Transaction{
		{ID: 1, UserID: testUserID, Type: models.ExpenseTransaction},
	}

	txnRepo.On("ListTransactions", models.Transaction{
		UserID: testUserID,
		Type:   models.ExpenseTransaction,
	}).Return(expected, nil)

	result, err := svc.ListTransactionsByType(testUserID, models.ExpenseTransaction)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestUserJourney_IncomeExpenseUndo(t *testing.T) {
	svc, walletRepo, _, txnRepo := newTestService()

	// 1. Add Income: Salary 5000
	income := models.Transaction{
		UserID: testUserID, Amount: 5000, Type: models.IncomeTransaction,
		DstID: "bank", SubcategoryID: "fin-sal",
	}
	walletRepo.On("UpdateWalletBalance", testUserID, "bank", 5000.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil).Once()

	err := svc.AddTransaction(income)
	assert.NoError(t, err)

	// 2. Add Expense: Pizza 500
	expense := models.Transaction{
		UserID: testUserID, Amount: 500, Type: models.ExpenseTransaction,
		SrcID: "bank", SubcategoryID: "food-rest",
	}
	walletRepo.On("UpdateWalletBalance", testUserID, "bank", -500.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil).Once()

	err = svc.AddTransaction(expense)
	assert.NoError(t, err)

	// 3. Undo Expense
	txnRepo.On("GetLastActiveTransaction", testUserID).Return(&expense, nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "bank", 500.0).Return(nil)
	txnRepo.On("SoftDeleteTransaction", mock.Anything, mock.Anything).Return(nil)

	undone, err := svc.Undo(testUserID)
	assert.NoError(t, err)
	assert.Equal(t, expense.Amount, undone.Amount)

	walletRepo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}

func TestAddTransaction_unknownContactGoesToRemarks(t *testing.T) {
	t.Parallel()
	svc, walletRepo, contactRepo, txnRepo := newTestService()

	// Lending to "meem" who is not a saved contact.
	txn := models.Transaction{
		UserID: testUserID, Amount: 300, Type: models.ExpenseTransaction,
		SrcID: "cash", SubcategoryID: models.LendSubID, ContactName: "meem",
	}
	contactRepo.On("GetContactByName", testUserID, "meem").
		Return((*models.Contacts)(nil), models.ErrContactNotFound{UserID: testUserID, NickName: "meem"})
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -300.0).Return(nil)
	// The person is noted in remarks and ContactName is cleared, so no balance update.
	txnRepo.On("AddTransaction", mock.MatchedBy(func(tx models.Transaction) bool {
		return tx.ContactName == "" && strings.Contains(tx.Remarks, "[Person: meem]")
	})).Return(nil)

	err := svc.AddTransaction(txn)

	assert.NoError(t, err)
	contactRepo.AssertNotCalled(t, "UpdateContactBalance", mock.Anything, mock.Anything)
	walletRepo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}

func TestUserJourney_LendingCycle(t *testing.T) {
	svc, walletRepo, contactRepo, txnRepo := newTestService()

	// 1. Lend 1000 to Masud
	lend := models.Transaction{
		UserID: testUserID, Amount: 1000, Type: models.ExpenseTransaction,
		SrcID: "cash", SubcategoryID: models.LendSubID, ContactName: "masud",
	}
	contactRepo.On("GetContactByName", testUserID, "masud").Return(&models.Contacts{ID: 1}, nil)
	contactRepo.On("UpdateContactBalance", int64(1), 1000.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -1000.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err := svc.AddTransaction(lend)
	assert.NoError(t, err)

	// 2. Recover 1000 from Masud
	collect := models.Transaction{
		UserID: testUserID, Amount: 1000, Type: models.IncomeTransaction,
		DstID: "cash", SubcategoryID: models.LendRecoverySubID, ContactName: "masud",
	}
	contactRepo.On("GetContactByName", testUserID, "masud").Return(&models.Contacts{ID: 1}, nil)
	contactRepo.On("UpdateContactBalance", int64(1), -1000.0).Return(nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 1000.0).Return(nil)
	txnRepo.On("AddTransaction", mock.AnythingOfType("models.Transaction")).Return(nil)

	err = svc.AddTransaction(collect)
	assert.NoError(t, err)

	contactRepo.AssertExpectations(t)
	walletRepo.AssertExpectations(t)
}
