package wallets

import (
	"fmt"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"
	"github.com/masudur-rahman/khorcha-pati/services"
)

type walletService struct {
	walletRepo repos.WalletRepository
}

var _ services.WalletService = &walletService{}

func NewWalletService(walletRepo repos.WalletRepository) *walletService {
	return &walletService{walletRepo: walletRepo}
}

func (ws *walletService) GetWalletByShortName(userID int64, shortName string) (*models.Wallet, error) {
	return ws.walletRepo.GetWalletByShortName(userID, shortName)
}

func (ws *walletService) ListWallets(userID int64) ([]models.Wallet, error) {
	return ws.walletRepo.ListWallets(userID)
}

func (ws *walletService) ListWalletsByType(userID int64, typ models.WalletType) ([]models.Wallet, error) {
	return ws.walletRepo.ListWalletsByType(userID, typ)
}

func (ws *walletService) CreateWallet(wallet *models.Wallet) error {
	if wallet.UserID == 0 {
		return fmt.Errorf("user-id can't be empty")
	}
	wallet.CreatedAt = time.Now().Unix()
	return ws.walletRepo.AddNewWallet(wallet)
}

func (ws *walletService) UpdateWalletBalance(userID int64, shortName string, amount float64) error {
	return ws.walletRepo.UpdateWalletBalance(userID, shortName, amount)
}

func (ws *walletService) DeleteWallet(userID int64, shortName string) error {
	return ws.walletRepo.DeleteWallet(userID, shortName)
}
