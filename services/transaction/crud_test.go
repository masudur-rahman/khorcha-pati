package transaction

import (
	"fmt"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionByID_success(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	txnRepo.On("GetTransactionByID", int64(10)).Return(&models.Transaction{
		ID: 10, UserID: testUserID, Amount: 100, SubcategoryID: "food-groc",
	}, nil)

	txn, err := svc.GetTransactionByID(testUserID, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), txn.ID)
}

func TestGetTransactionByID_wrongUser(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	txnRepo.On("GetTransactionByID", int64(10)).Return(&models.Transaction{
		ID: 10, UserID: 999, Amount: 100,
	}, nil)

	txn, err := svc.GetTransactionByID(testUserID, 10)

	assert.Error(t, err)
	assert.Nil(t, txn)
}

func TestGetTransactionByID_deleted(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	txnRepo.On("GetTransactionByID", int64(10)).Return(&models.Transaction{
		ID: 10, UserID: testUserID, DeletedAt: 12345,
	}, nil)

	txn, err := svc.GetTransactionByID(testUserID, 10)

	assert.Error(t, err)
	assert.Nil(t, txn)
}

func TestGetTransactionByID_notFound(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	txnRepo.On("GetTransactionByID", int64(10)).Return(nil, fmt.Errorf("transaction 10 not found"))

	txn, err := svc.GetTransactionByID(testUserID, 10)

	assert.Error(t, err)
	assert.Nil(t, txn)
}

func TestDeleteTransaction_success(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	txnRepo.On("GetTransactionByID", int64(10)).Return(&models.Transaction{
		ID: 10, UserID: testUserID, Amount: 200, SubcategoryID: "food-groc",
		Type: models.ExpenseTransaction, SrcID: "cash",
	}, nil)
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 200.0).Return(nil)
	txnRepo.On("SoftDeleteTransaction", int64(10), mock.AnythingOfType("int64")).Return(nil)

	err := svc.DeleteTransaction(testUserID, 10)

	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}

func TestDeleteTransaction_wrongUser(t *testing.T) {
	t.Parallel()
	svc, _, _, txnRepo := newTestService()

	txnRepo.On("GetTransactionByID", int64(10)).Return(&models.Transaction{
		ID: 10, UserID: 999, Amount: 100,
	}, nil)

	err := svc.DeleteTransaction(testUserID, 10)

	assert.Error(t, err)
}

func TestUpdateTransaction_success(t *testing.T) {
	t.Parallel()
	svc, walletRepo, _, txnRepo := newTestService()

	old := &models.Transaction{
		ID: 10, UserID: testUserID, Amount: 100, SubcategoryID: "food-groc",
		Type: models.ExpenseTransaction, SrcID: "cash", CreatedAt: 1000,
	}
	txnRepo.On("GetTransactionByID", int64(10)).Return(old, nil)

	// Reverse old: +100 to cash
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", 100.0).Return(nil)
	// Apply new: -250 from cash
	walletRepo.On("UpdateWalletBalance", testUserID, "cash", -250.0).Return(nil)
	txnRepo.On("UpdateTransaction", int64(10), mock.AnythingOfType("*models.Transaction")).Return(nil)

	updated := models.Transaction{
		Amount: 250, SubcategoryID: "food-groc",
		Type: models.ExpenseTransaction, SrcID: "cash",
	}
	err := svc.UpdateTransaction(testUserID, 10, updated)

	assert.NoError(t, err)
	walletRepo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}
