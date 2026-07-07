package wallets

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"
	"github.com/masudur-rahman/khorcha-pati/services"

	"github.com/masudur-rahman/styx"
)

type walletService struct {
	uow        styx.UnitOfWork
	walletRepo repos.WalletRepository
	txnRepo    repos.TransactionRepository
}

var _ services.WalletService = &walletService{}

func NewWalletService(uow styx.UnitOfWork, walletRepo repos.WalletRepository, txnRepo repos.TransactionRepository) *walletService {
	return &walletService{
		uow:        uow,
		walletRepo: walletRepo,
		txnRepo:    txnRepo,
	}
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

func (ws *walletService) CreateWallet(wallet *models.Wallet) (err error) {
	if wallet.UserID == 0 {
		return fmt.Errorf("user-id can't be empty")
	}
	wallet.CreatedAt = time.Now().Unix()

	initialBalance := wallet.Balance
	if initialBalance != 0 {
		wallet.Balance = 0
	}

	uow, err := ws.uow.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = uow.Rollback()
			return
		}
		err = uow.Commit()
	}()

	if err = ws.walletRepo.WithUnitOfWork(uow).AddNewWallet(wallet); err != nil {
		return err
	}

	if initialBalance != 0 {
		var txnType models.TransactionType
		var srcID, dstID string

		if initialBalance > 0 {
			txnType = models.IncomeTransaction
			dstID = wallet.ShortName
		} else {
			txnType = models.ExpenseTransaction
			srcID = wallet.ShortName
		}

		txn := models.Transaction{
			UserID:        wallet.UserID,
			Amount:        math.Abs(initialBalance),
			SubcategoryID: "misc-init",
			Type:          txnType,
			SrcID:         srcID,
			DstID:         dstID,
			Timestamp:     time.Now().Unix(),
			Remarks:       "Initial Amount",
			CreatedAt:     time.Now().Unix(),
		}

		if err = ws.txnRepo.WithUnitOfWork(uow).AddTransaction(txn); err != nil {
			return err
		}

		if err = ws.walletRepo.WithUnitOfWork(uow).UpdateWalletBalance(wallet.UserID, wallet.ShortName, initialBalance); err != nil {
			return err
		}
	}

	return nil
}

func (ws *walletService) UpdateWalletBalance(userID int64, shortName string, amount float64) error {
	return ws.walletRepo.UpdateWalletBalance(userID, shortName, amount)
}

func (ws *walletService) DeleteWallet(userID int64, shortName string) error {
	return ws.walletRepo.DeleteWallet(userID, shortName)
}
