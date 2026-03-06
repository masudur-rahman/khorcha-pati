package wallets

import (
	"fmt"
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos/mocks"

	"github.com/stretchr/testify/assert"
)

const testUserID int64 = 42

func TestCreateWallet_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	wallet := &models.Wallet{
		UserID:    testUserID,
		ShortName: "cash",
		Name:      "Cash",
		Type:      models.CashAccount,
	}

	repo.On("AddNewWallet", wallet).Return(nil)

	err := svc.CreateWallet(wallet)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreateWallet_missingUserID(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	wallet := &models.Wallet{ShortName: "cash", Name: "Cash"}

	err := svc.CreateWallet(wallet)

	assert.EqualError(t, err, "user-id can't be empty")
}

func TestCreateWallet_duplicateError(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	wallet := &models.Wallet{
		UserID:    testUserID,
		ShortName: "cash",
		Name:      "Cash",
	}

	repo.On("AddNewWallet", wallet).Return(fmt.Errorf("duplicate key"))

	err := svc.CreateWallet(wallet)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestGetWalletByShortName_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	expected := &models.Wallet{
		ID:        1,
		UserID:    testUserID,
		ShortName: "cash",
		Balance:   1000,
	}

	repo.On("GetWalletByShortName", testUserID, "cash").Return(expected, nil)

	result, err := svc.GetWalletByShortName(testUserID, "cash")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetWalletByShortName_notFound(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	repo.On("GetWalletByShortName", testUserID, "nope").
		Return(nil, fmt.Errorf("not found"))

	result, err := svc.GetWalletByShortName(testUserID, "nope")

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestListWallets_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	expected := []models.Wallet{
		{ID: 1, ShortName: "cash", Balance: 500},
		{ID: 2, ShortName: "bank", Balance: 2000},
	}

	repo.On("ListWallets", testUserID).Return(expected, nil)

	result, err := svc.ListWallets(testUserID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUpdateWalletBalance_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	repo.On("UpdateWalletBalance", testUserID, "cash", 100.0).Return(nil)

	err := svc.UpdateWalletBalance(testUserID, "cash", 100)

	assert.NoError(t, err)
}

func TestUpdateWalletBalance_optimisticLock(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	repo.On("UpdateWalletBalance", testUserID, "cash", 100.0).
		Return(models.ErrOptimisticLock)

	err := svc.UpdateWalletBalance(testUserID, "cash", 100)

	assert.ErrorIs(t, err, models.ErrOptimisticLock)
}

func TestDeleteWallet_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.WalletRepo{}
	svc := NewWalletService(repo)

	repo.On("DeleteWallet", testUserID, "cash").Return(nil)

	err := svc.DeleteWallet(testUserID, "cash")

	assert.NoError(t, err)
}
