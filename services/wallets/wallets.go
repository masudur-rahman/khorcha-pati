package wallets

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/pkg/validator"
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

	if !validator.IsValidShortName(wallet.ShortName) {
		return models.StatusError{
			Status:  400,
			Message: "wallet short name cannot have spaces or special characters",
		}
	}

	if wallet.Name != "" && !validator.IsValidWalletName(wallet.Name) {
		return models.StatusError{
			Status:  400,
			Message: "wallet name cannot have leading/trailing spaces or special characters",
		}
	}

	wallets, err := ws.walletRepo.ListWallets(wallet.UserID)
	if err != nil {
		return err
	}
	for _, existing := range wallets {
		if strings.EqualFold(existing.Name, wallet.Name) {
			return models.StatusError{
				Status:  409,
				Message: fmt.Sprintf("wallet already exists with name: %s", wallet.Name),
			}
		}
		if strings.EqualFold(existing.ShortName, wallet.ShortName) {
			return models.StatusError{
				Status:  409,
				Message: fmt.Sprintf("wallet already exists with short-name: %s", wallet.ShortName),
			}
		}
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

func (ws *walletService) GetWalletByID(userID, id int64) (*models.Wallet, error) {
	return ws.walletRepo.GetWalletByID(userID, id)
}

func (ws *walletService) UpdateWallet(userID, id int64, name, shortName string) (err error) {
	if name == "" || shortName == "" {
		return models.StatusError{
			Status:  400,
			Message: "wallet name and short name cannot be empty",
		}
	}

	if !validator.IsValidShortName(shortName) {
		return models.StatusError{
			Status:  400,
			Message: "wallet short name cannot have spaces or special characters",
		}
	}

	if !validator.IsValidWalletName(name) {
		return models.StatusError{
			Status:  400,
			Message: "wallet name cannot have leading/trailing spaces or special characters",
		}
	}

	w, err := ws.walletRepo.GetWalletByID(userID, id)
	if err != nil {
		return err
	}

	oldShortName := w.ShortName

	wallets, err := ws.walletRepo.ListWallets(userID)
	if err != nil {
		return err
	}
	for _, existing := range wallets {
		if existing.ID != id {
			if strings.EqualFold(existing.Name, name) {
				return models.StatusError{
					Status:  409,
					Message: fmt.Sprintf("wallet already exists with name: %s", name),
				}
			}
			if strings.EqualFold(existing.ShortName, shortName) {
				return models.StatusError{
					Status:  409,
					Message: fmt.Sprintf("wallet already exists with short-name: %s", shortName),
				}
			}
		}
	}

	w.Name = name
	w.ShortName = shortName

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

	if err = ws.walletRepo.WithUnitOfWork(uow).UpdateWallet(w); err != nil {
		return err
	}

	if oldShortName != shortName {
		if err = ws.txnRepo.WithUnitOfWork(uow).UpdateTransactionsWallet(userID, oldShortName, shortName); err != nil {
			return err
		}
	}

	return nil
}

func (ws *walletService) UpdateWalletBalance(userID int64, shortName string, amount float64) error {
	return ws.walletRepo.UpdateWalletBalance(userID, shortName, amount)
}

func (ws *walletService) DeleteWallet(userID int64, shortName string) error {
	// Check if transactions exist referencing this wallet as SrcID
	txns, err := ws.txnRepo.ListTransactions(models.Transaction{UserID: userID, SrcID: shortName})
	if err != nil {
		return err
	}
	if len(txns) > 0 {
		return models.StatusError{
			Status:  400,
			Message: fmt.Sprintf("cannot delete wallet '%s': it has active transactions", shortName),
		}
	}

	// Check if transactions exist referencing this wallet as DstID
	txns, err = ws.txnRepo.ListTransactions(models.Transaction{UserID: userID, DstID: shortName})
	if err != nil {
		return err
	}
	if len(txns) > 0 {
		return models.StatusError{
			Status:  400,
			Message: fmt.Sprintf("cannot delete wallet '%s': it has active transactions", shortName),
		}
	}

	return ws.walletRepo.DeleteWallet(userID, shortName)
}
