package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"

	isql "github.com/masudur-rahman/styx/sql"
	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testUserID int64 = 99

type testEnv struct {
	repo *SQLTransactionRepository
	db   isql.Engine
}

func setupTxnRepo(t *testing.T) testEnv {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(
		context.Background(),
		models.Transaction{},
		models.TxnCategory{},
		models.TxnSubcategory{},
	))

	logger := zap.NewNop().Sugar()
	return testEnv{
		repo: NewSQLTransactionRepository(db, logger),
		db:   db,
	}
}

// seedTransaction inserts a txn into the test DB.
func seedTransaction(t *testing.T, env testEnv, txn models.Transaction) {
	t.Helper()
	require.NoError(t, env.repo.AddTransaction(txn))
}

func TestAddTransaction_success(t *testing.T) {

	env := setupTxnRepo(t)

	txn := models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     time.Now().Unix(),
	}

	err := env.repo.AddTransaction(txn)

	assert.NoError(t, err)
}

func TestAddTransaction_setsTimestamp(t *testing.T) {

	env := setupTxnRepo(t)

	before := time.Now().Unix()
	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        50,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     time.Now().Unix(),
	})

	txns, err := env.repo.ListTransactions(models.Transaction{UserID: testUserID})
	require.NoError(t, err)
	require.Len(t, txns, 1)
	assert.GreaterOrEqual(t, txns[0].Timestamp, before)
}

func TestGetLastActiveTransaction_success(t *testing.T) {

	env := setupTxnRepo(t)

	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     1000,
	})
	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        200,
		SubcategoryID: "food-rest",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     2000,
	})

	last, err := env.repo.GetLastActiveTransaction(testUserID)

	assert.NoError(t, err)
	assert.Equal(t, float64(200), last.Amount)
	assert.Equal(t, int64(2000), last.CreatedAt)
}

func TestGetLastActiveTransaction_noTransactions(t *testing.T) {

	env := setupTxnRepo(t)

	last, err := env.repo.GetLastActiveTransaction(testUserID)

	assert.Nil(t, last)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active transactions")
}

func TestGetLastActiveTransaction_skipsDeleted(t *testing.T) {

	env := setupTxnRepo(t)

	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     1000,
	})
	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        500,
		SubcategoryID: "food-rest",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     3000,
	})

	// Soft-delete the second (latest) transaction
	last, err := env.repo.GetLastActiveTransaction(testUserID)
	require.NoError(t, err)
	require.NoError(t, env.repo.SoftDeleteTransaction(last.ID, time.Now().Unix()))

	// Now the first transaction should be returned
	last, err = env.repo.GetLastActiveTransaction(testUserID)
	assert.NoError(t, err)
	assert.Equal(t, float64(100), last.Amount)
}

func TestSoftDeleteTransaction_success(t *testing.T) {

	env := setupTxnRepo(t)

	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     time.Now().Unix(),
	})

	last, err := env.repo.GetLastActiveTransaction(testUserID)
	require.NoError(t, err)

	now := time.Now().Unix()
	err = env.repo.SoftDeleteTransaction(last.ID, now)
	assert.NoError(t, err)

	// Should no longer appear in active transactions
	_, err = env.repo.GetLastActiveTransaction(testUserID)
	assert.Error(t, err)
}

func TestListTransactions_filtersDeleted(t *testing.T) {

	env := setupTxnRepo(t)

	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     time.Now().Unix(),
	})
	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        200,
		SubcategoryID: "food-rest",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		CreatedAt:     time.Now().Unix() + 1,
	})

	// Delete the latest
	last, _ := env.repo.GetLastActiveTransaction(testUserID)
	require.NoError(t, env.repo.SoftDeleteTransaction(last.ID, time.Now().Unix()))

	txns, err := env.repo.ListTransactions(models.Transaction{UserID: testUserID})
	assert.NoError(t, err)
	assert.Len(t, txns, 1)
}

func TestListTransactionsByTime_success(t *testing.T) {

	env := setupTxnRepo(t)

	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        100,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		Timestamp:     1000,
		CreatedAt:     1000,
	})
	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        200,
		SubcategoryID: "food-rest",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		Timestamp:     2000,
		CreatedAt:     2000,
	})
	seedTransaction(t, env, models.Transaction{
		UserID:        testUserID,
		Amount:        300,
		SubcategoryID: "food-groc",
		Type:          models.ExpenseTransaction,
		SrcID:         "cash",
		Timestamp:     3000,
		CreatedAt:     3000,
	})

	txns, err := env.repo.ListTransactionsByTime(testUserID, models.ExpenseTransaction, 1500, 2500)

	assert.NoError(t, err)
	require.Len(t, txns, 1)
	assert.Equal(t, float64(200), txns[0].Amount)
}

func TestUpdateTxnCategories_success(t *testing.T) {

	env := setupTxnRepo(t)

	err := env.repo.UpdateTxnCategories()

	assert.NoError(t, err)

	cats, err := env.repo.ListTxnCategories()
	assert.NoError(t, err)
	assert.Equal(t, len(models.TxnCategories), len(cats))

	subcats, err := env.repo.ListTxnSubcategories("food")
	assert.NoError(t, err)
	assert.Greater(t, len(subcats), 0)
}

func TestGetTxnCategoryName_success(t *testing.T) {

	env := setupTxnRepo(t)
	require.NoError(t, env.repo.UpdateTxnCategories())

	name, err := env.repo.GetTxnCategoryName("food")

	assert.NoError(t, err)
	assert.Equal(t, "Food", name)
}

func TestGetTxnCategoryName_notFound(t *testing.T) {

	env := setupTxnRepo(t)

	_, err := env.repo.GetTxnCategoryName("nonexistent")

	assert.Error(t, err)
}

func TestGetTxnSubcategoryName_success(t *testing.T) {

	env := setupTxnRepo(t)
	require.NoError(t, env.repo.UpdateTxnCategories())

	name, err := env.repo.GetTxnSubcategoryName("food-groc")

	assert.NoError(t, err)
	assert.Equal(t, "Grocery", name)
}

func TestUpdateTxnCategories_idempotent(t *testing.T) {

	env := setupTxnRepo(t)

	require.NoError(t, env.repo.UpdateTxnCategories())
	require.NoError(t, env.repo.UpdateTxnCategories())

	cats, err := env.repo.ListTxnCategories()
	assert.NoError(t, err)
	assert.Equal(t, len(models.TxnCategories), len(cats))
}
