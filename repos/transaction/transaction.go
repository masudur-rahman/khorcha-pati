package transaction

import (
	"context"
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
		db:     db.Table(models.Transaction{}.TableName()),
		logger: logger,
	}
}

func (t *SQLTransactionRepository) WithUnitOfWork(uow styx.UnitOfWork) repos.TransactionRepository {
	return &SQLTransactionRepository{
		db:     uow.SQL.Table(models.Transaction{}.TableName()),
		logger: t.logger,
	}
}

func (t *SQLTransactionRepository) AddTransaction(txn models.Transaction) error {
	t.logger.Infow("inserting new transaction")
	if txn.Timestamp == 0 {
		txn.Timestamp = time.Now().Unix()
	}
	ctx := context.Background()
	// req tag on deleted_at ensures the zero value is written, marking the transaction as active.
	_, err := t.db.InsertOne(ctx, txn)
	return err
}

// GetTransactionByID returns a transaction by its primary key.
func (t *SQLTransactionRepository) GetTransactionByID(id int64) (*models.Transaction, error) {
	t.logger.Infow("get transaction by ID", "id", id)
	ctx := context.Background()
	var txn models.Transaction
	found, err := t.db.ID(id).FindOne(ctx, &txn)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("transaction %d not found", id)
	}
	return &txn, nil
}

// UpdateTransaction updates a transaction record by ID.
func (t *SQLTransactionRepository) UpdateTransaction(id int64, txn *models.Transaction) error {
	t.logger.Infow("updating transaction", "id", id)
	ctx := context.Background()
	// req tag on deleted_at ensures it is included in the update even when zero.
	return t.db.ID(id).UpdateOne(ctx, txn)
}

// GetLastActiveTransaction returns the most recent non-deleted transaction for a user.
func (t *SQLTransactionRepository) GetLastActiveTransaction(userID int64) (*models.Transaction, error) {
	t.logger.Infow("get last active transaction", "userID", userID)
	ctx := context.Background()
	var txns []models.Transaction
	// req tag on deleted_at auto-includes WHERE deleted_at=0 from the zero-value struct field.
	if err := t.db.FindMany(ctx, &txns, models.Transaction{UserID: userID}); err != nil {
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
	ctx := context.Background()
	// deletedAt is non-zero (unix timestamp), so it is included in the UPDATE naturally.
	// req tag ensures deleted_at is always written in UpdateOne, even if it were zero.
	return t.db.ID(txnID).UpdateOne(ctx, &models.Transaction{DeletedAt: deletedAt})
}

func (t *SQLTransactionRepository) ListTransactions(filter models.Transaction) ([]models.Transaction, error) {
	t.logger.Infow("list transactions")
	ctx := context.Background()
	txns := make([]models.Transaction, 0)
	// req tag on deleted_at auto-includes WHERE deleted_at=0 from the filter's zero value.
	err := t.db.FindMany(ctx, &txns, filter)
	return txns, err
}

func (t *SQLTransactionRepository) ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error) {
	t.logger.Infow("list transactions by category")
	ctx := context.Background()
	txns := make([]models.Transaction, 0)
	// req tag on deleted_at auto-includes WHERE deleted_at=0 from the filter's zero value.
	err := t.db.Where(fmt.Sprintf("sub_category_id LIKE %s%%", catID)).
		FindMany(ctx, &txns, models.Transaction{UserID: userID})
	return txns, err
}

func (t *SQLTransactionRepository) ListTransactionsByTime(userID int64, txnType models.TransactionType, startTime, endTime int64) ([]models.Transaction, error) {
	t.logger.Infow("list transactions by time")
	ctx := context.Background()
	txns := make([]models.Transaction, 0)
	// req tag on deleted_at auto-includes WHERE deleted_at=0 from the filter's zero value.
	err := t.db.Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		FindMany(ctx, &txns, models.Transaction{UserID: userID, Type: txnType})
	return txns, err
}

func (t *SQLTransactionRepository) GetTxnCategoryName(catID string) (string, error) {
	ctx := context.Background()
	cat := models.TxnCategory{}
	has, err := t.db.Table(models.TxnCategory{}.TableName()).ID(catID).FindOne(ctx, &cat)
	if err != nil {
		return "", err
	} else if !has {
		return "", errors.New("not found")
	}

	return cat.Name, nil
}

func (t *SQLTransactionRepository) ListTxnCategories() ([]models.TxnCategory, error) {
	t.logger.Infow("list transaction category")
	ctx := context.Background()
	cats := make([]models.TxnCategory, 0)
	err := t.db.Table(models.TxnCategory{}.TableName()).FindMany(ctx, &cats)
	return cats, err
}

func (t *SQLTransactionRepository) GetTxnSubcategoryName(subcatID string) (string, error) {
	ctx := context.Background()
	subcat := models.TxnSubcategory{}
	has, err := t.db.Table(models.TxnSubcategory{}.TableName()).ID(subcatID).FindOne(ctx, &subcat)
	if err != nil {
		return "", err
	} else if !has {
		return "", errors.New("not found")
	}

	return subcat.Name, nil
}

func (t *SQLTransactionRepository) ListTxnSubcategories(catID string) ([]models.TxnSubcategory, error) {
	t.logger.Infow("list transaction category")
	ctx := context.Background()
	subcats := make([]models.TxnSubcategory, 0)
	err := t.db.Table(models.TxnSubcategory{}.TableName()).FindMany(ctx, &subcats, models.TxnSubcategory{CatID: catID})
	return subcats, err
}

func (t *SQLTransactionRepository) UpdateTxnCategories() error {
	ctx := context.Background()
	quiet := t.db.ShowSQL(false)

	catDB := quiet.Table(models.TxnCategory{}.TableName())
	catt := models.TxnCategory{}
	var catUpdated, catInserted int
	for _, cat := range models.TxnCategories {
		if has, err := catDB.ID(cat.ID).FindOne(ctx, &catt); err != nil {
			return err
		} else if has {
			if catt.Name != cat.Name {
				if err = catDB.ID(catt.ID).UpdateOne(ctx, cat); err != nil {
					return err
				}
				catUpdated++
			}
			continue
		}

		if _, err := catDB.InsertOne(ctx, cat); err != nil {
			return err
		}
		catInserted++
	}

	subcatDB := quiet.Table(models.TxnSubcategory{}.TableName())
	subcatt := models.TxnSubcategory{}
	var subcatUpdated, subcatInserted int
	for _, subcat := range models.TxnSubcategories {
		if has, err := subcatDB.ID(subcat.ID).FindOne(ctx, &subcatt); err != nil {
			return err
		} else if has {
			if subcatt.Name != subcat.Name || subcatt.CatID != subcat.CatID {
				if err = subcatDB.ID(subcatt.ID).UpdateOne(ctx, subcat); err != nil {
					return err
				}
				subcatUpdated++
			}
			continue
		}
		if _, err := subcatDB.InsertOne(ctx, subcat); err != nil {
			return err
		}
		subcatInserted++
	}

	t.logger.Infow("synced transaction categories",
		"categories", len(models.TxnCategories),
		"catInserted", catInserted, "catUpdated", catUpdated,
		"subcategories", len(models.TxnSubcategories),
		"subcatInserted", subcatInserted, "subcatUpdated", subcatUpdated,
	)
	return nil
}
