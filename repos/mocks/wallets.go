package mocks

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"

	"github.com/stretchr/testify/mock"
)

// WalletRepo is a mock for repos.WalletRepository.
type WalletRepo struct {
	mock.Mock
}

var _ repos.WalletRepository = &WalletRepo{}

func (m *WalletRepo) WithUnitOfWork(_ styx.UnitOfWork) repos.WalletRepository {
	return m
}

func (m *WalletRepo) GetWalletByShortName(userID int64, shortName string) (*models.Wallet, error) {
	args := m.Called(userID, shortName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *WalletRepo) ListWallets(userID int64) ([]models.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Wallet), args.Error(1)
}

func (m *WalletRepo) ListWalletsByType(userID int64, typ models.WalletType) ([]models.Wallet, error) {
	args := m.Called(userID, typ)
	return args.Get(0).([]models.Wallet), args.Error(1)
}

func (m *WalletRepo) AddNewWallet(wallet *models.Wallet) error {
	return m.Called(wallet).Error(0)
}

func (m *WalletRepo) UpdateWalletBalance(userID int64, shortName string, amount float64) error {
	return m.Called(userID, shortName, amount).Error(0)
}

func (m *WalletRepo) DeleteWallet(userID int64, shortName string) error {
	return m.Called(userID, shortName).Error(0)
}
