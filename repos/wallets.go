package repos

import (
	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/masudur-rahman/styx"
)

type WalletRepository interface {
	WithUnitOfWork(uow styx.UnitOfWork) WalletRepository
	GetWalletByID(userID, id int64) (*models.Wallet, error)
	GetWalletByShortName(userID int64, shortName string) (*models.Wallet, error)
	ListWallets(userID int64) ([]models.Wallet, error)
	ListWalletsByType(userID int64, typ models.WalletType) ([]models.Wallet, error)
	AddNewWallet(wallet *models.Wallet) error
	UpdateWallet(wallet *models.Wallet) error
	UpdateWalletBalance(userID int64, shortName string, txnAmount float64) error
	DeleteWallet(userID int64, shortName string) error
}
