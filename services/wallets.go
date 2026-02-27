package services

import "github.com/masudur-rahman/expense-tracker-bot/models"

type WalletService interface {
	GetWalletByShortName(userID int64, shortName string) (*models.Wallet, error)
	ListWallets(userID int64) ([]models.Wallet, error)
	ListWalletsByType(userID int64, typ models.WalletType) ([]models.Wallet, error)
	CreateWallet(wallet *models.Wallet) error
	UpdateWalletBalance(userID int64, shortName string, amount float64) error
	DeleteWallet(userID int64, shortName string) error
}
