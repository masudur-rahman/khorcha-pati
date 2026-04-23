package wallets

import (
	"context"
	"errors"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"
	"github.com/masudur-rahman/styx/dberr"
	isql "github.com/masudur-rahman/styx/sql"
)

type SQLWalletRepository struct {
	db     isql.Engine
	logger logr.Logger
}

func NewSQLWalletRepository(db isql.Engine, logger logr.Logger) *SQLWalletRepository {
	return &SQLWalletRepository{
		db:     db.Table(models.Wallet{}.TableName()),
		logger: logger,
	}
}

func (a *SQLWalletRepository) WithUnitOfWork(uow styx.UnitOfWork) repos.WalletRepository {
	return &SQLWalletRepository{
		db:     uow.SQL.Table(models.Wallet{}.TableName()),
		logger: a.logger,
	}
}

func (a *SQLWalletRepository) GetWalletByShortName(userID int64, shortName string) (*models.Wallet, error) {
	a.logger.Infow("get wallet by short name", "shortName", shortName)
	ctx := context.Background()
	var w models.Wallet
	found, err := a.db.FindOne(ctx, &w, models.Wallet{ShortName: shortName, UserID: userID})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrAccountNotFound{AccID: shortName}
	}
	return &w, nil
}

func (a *SQLWalletRepository) ListWallets(userID int64) ([]models.Wallet, error) {
	a.logger.Infow("list wallets")
	ctx := context.Background()
	wallets := make([]models.Wallet, 0)
	err := a.db.FindMany(ctx, &wallets, models.Wallet{UserID: userID})
	return wallets, err
}

func (a *SQLWalletRepository) ListWalletsByType(userID int64, typ models.WalletType) ([]models.Wallet, error) {
	a.logger.Infow("list wallets by type", "type", typ)
	ctx := context.Background()
	wallets := make([]models.Wallet, 0)
	err := a.db.FindMany(ctx, &wallets, models.Wallet{UserID: userID, Type: typ})
	return wallets, err
}

func (a *SQLWalletRepository) AddNewWallet(wallet *models.Wallet) error {
	a.logger.Infow("add new wallet", "name", wallet.Name)
	ctx := context.Background()
	_, err := a.GetWalletByShortName(wallet.UserID, wallet.ShortName)
	if err == nil {
		return models.ErrAccountAlreadyExist{ShortName: wallet.ShortName}
	} else if !models.IsErrNotFound(err) {
		return err
	}
	_, err = a.db.MustCols("version", "balance", "last_txn_amount").InsertOne(ctx, wallet)
	return err
}

func (a *SQLWalletRepository) UpdateWalletBalance(userID int64, shortName string, txnAmount float64) error {
	a.logger.Infow("updating wallet balance", "wallet", shortName)
	ctx := context.Background()
	w, err := a.GetWalletByShortName(userID, shortName)
	if err != nil {
		return err
	}
	oldVersion := w.Version
	w.Balance += txnAmount
	w.LastTxnAmount = txnAmount
	w.LastTxnTimestamp = time.Now().Unix()
	w.Version = oldVersion + 1
	err = a.db.ID(w.ID).Where("version = ?", oldVersion).MustCols("balance", "last_txn_amount", "version").UpdateOne(ctx, w)
	if errors.Is(err, dberr.DataNotFound) {
		return models.ErrOptimisticLock
	}
	return err
}

func (a *SQLWalletRepository) DeleteWallet(userID int64, shortName string) error {
	a.logger.Infow("deleting wallet", "wallet", shortName)
	ctx := context.Background()
	return a.db.DeleteOne(ctx, models.Wallet{ShortName: shortName, UserID: userID})
}
