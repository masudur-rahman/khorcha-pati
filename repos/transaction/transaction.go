package transaction

import (
	"errors"
	"fmt"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"
	isql "github.com/masudur-rahman/styx/sql"
)

type SQLTransactionRepository struct {
	db     isql.Engine
	logger logr.Logger
}

func NewSQLTransactionRepository(db isql.Engine, logger logr.Logger) *SQLTransactionRepository {
	return &SQLTransactionRepository{
		db:     db.Table("transaction"),
		logger: logger,
	}
}

func (t *SQLTransactionRepository) WithUnitOfWork(uow styx.UnitOfWork) repos.TransactionRepository {
	return &SQLTransactionRepository{
		db:     uow.SQL.Table("transaction"),
		logger: t.logger,
	}
}

func (t *SQLTransactionRepository) AddTransaction(txn models.Transaction) error {
	t.logger.Infow("inserting new transaction")
	if txn.Timestamp == 0 {
		txn.Timestamp = time.Now().Unix()
	}
	_, err := t.db.InsertOne(txn)
	return err
}

// GetLastActiveTransaction returns the most recent non-deleted transaction for a user.
func (t *SQLTransactionRepository) GetLastActiveTransaction(userID int64) (*models.Transaction, error) {
	t.logger.Infow("get last active transaction", "userID", userID)
	var txns []models.Transaction
	if err := t.db.Where("deleted_at = 0").
		FindMany(&txns, models.Transaction{UserID: userID}); err != nil {
		return nil, err
	}
	if len(txns) == 0 {
		return nil, errors.New("no active transactions found")
	}

	latest := txns[0]
	for _, txn := range txns[1:] {
		if txn.CreatedAt > latest.CreatedAt {
			latest = txn
		}
	}
	return &latest, nil
}

// SoftDeleteTransaction marks a transaction as deleted by setting DeletedAt.
func (t *SQLTransactionRepository) SoftDeleteTransaction(txnID int64, deletedAt int64) error {
	t.logger.Infow("soft-deleting transaction", "txnID", txnID)
	return t.db.ID(txnID).MustCols("deleted_at").UpdateOne(&models.Transaction{DeletedAt: deletedAt})
}

func (t *SQLTransactionRepository) ListTransactions(filter models.Transaction) ([]models.Transaction, error) {
	t.logger.Infow("list transactions")
	txns := make([]models.Transaction, 0)
	err := t.db.Where("deleted_at = 0").FindMany(&txns, filter)
	return txns, err
}

func (t *SQLTransactionRepository) ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error) {
	t.logger.Infow("list transactions by category")
	txns := make([]models.Transaction, 0)
	err := t.db.Where(fmt.Sprintf("sub_category_id LIKE %s%%", catID)).
		Where("deleted_at = 0").
		FindMany(&txns, models.Transaction{UserID: userID})
	return txns, err
}

func (t *SQLTransactionRepository) ListTransactionsByTime(userID int64, txnType models.TransactionType, startTime, endTime int64) ([]models.Transaction, error) {
	t.logger.Infow("list transactions by time")
	txns := make([]models.Transaction, 0)
	err := t.db.Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Where("deleted_at = 0").
		FindMany(&txns, models.Transaction{UserID: userID, Type: txnType})
	return txns, err
}

func (t *SQLTransactionRepository) GetTxnCategoryName(catID string) (string, error) {
	cat := models.TxnCategory{}
	has, err := t.db.Table("txn_category").ID(catID).FindOne(&cat)
	if err != nil {
		return "", err
	} else if !has {
		return "", errors.New("not found")
	}

	return cat.Name, nil
}

func (t *SQLTransactionRepository) ListTxnCategories() ([]models.TxnCategory, error) {
	t.logger.Infow("list transaction category")
	cats := make([]models.TxnCategory, 0)
	err := t.db.Table("txn_category").FindMany(&cats)
	return cats, err
}

func (t *SQLTransactionRepository) GetTxnSubcategoryName(subcatID string) (string, error) {
	subcat := models.TxnSubcategory{}
	has, err := t.db.Table("txn_subcategory").ID(subcatID).FindOne(&subcat)
	if err != nil {
		return "", err
	} else if !has {
		return "", errors.New("not found")
	}

	return subcat.Name, nil
}

func (t *SQLTransactionRepository) ListTxnSubcategories(catID string) ([]models.TxnSubcategory, error) {
	t.logger.Infow("list transaction category")
	subcats := make([]models.TxnSubcategory, 0)
	err := t.db.Table("txn_subcategory").FindMany(&subcats, models.TxnSubcategory{CatID: catID})
	return subcats, err
}

func (t *SQLTransactionRepository) UpdateTxnCategories() error {
	db := t.db.Table("txn_category")
	catt := models.TxnCategory{}
	for _, cat := range models.TxnCategories {
		if has, err := db.ID(cat.ID).FindOne(&catt); err != nil {
			return err
		} else if has {
			if catt.Name != cat.Name {
				if err = db.ID(catt.ID).UpdateOne(cat); err != nil {
					return err
				}
			}
			continue
		}

		if _, err := db.InsertOne(cat); err != nil {
			return err
		}
	}

	db = t.db.Table("txn_subcategory")
	subcatt := models.TxnSubcategory{}
	for _, subcat := range models.TxnSubcategories {
		if has, err := db.ID(subcat.ID).FindOne(&subcatt); err != nil {
			return err
		} else if has {
			if subcatt.Name != subcat.Name || subcatt.CatID != subcat.CatID {
				if err = db.ID(subcatt.ID).UpdateOne(subcat); err != nil {
					return err
				}
			}
			continue
		}
		if _, err := db.InsertOne(subcat); err != nil {
			return err
		}
	}
	return nil
}
