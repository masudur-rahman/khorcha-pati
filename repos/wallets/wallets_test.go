package wallets

import (
	"context"
	"testing"

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
	repo *SQLWalletRepository
	db   isql.Engine
}

func setupWalletRepo(t *testing.T) testEnv {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(context.Background(), models.Wallet{}))

	logger := zap.NewNop().Sugar()
	return testEnv{
		repo: NewSQLWalletRepository(db, logger),
		db:   db,
	}
}

func seedWallet(t *testing.T, env testEnv) {
	t.Helper()
	w := &models.Wallet{
		UserID:    testUserID,
		ShortName: "cash",
		Name:      "Cash Wallet",
		Type:      models.CashAccount,
		Balance:   1000,
		Version:   1, // Start at version 1 to verify optimistic-lock increments correctly
	}
	require.NoError(t, env.repo.AddNewWallet(w))
}

func TestAddNewWallet_success(t *testing.T) {

	env := setupWalletRepo(t)

	w := &models.Wallet{
		UserID:    testUserID,
		ShortName: "cash",
		Name:      "Cash",
		Type:      models.CashAccount,
	}
	err := env.repo.AddNewWallet(w)

	assert.NoError(t, err)
}

func TestAddNewWallet_duplicate(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	dup := &models.Wallet{
		UserID:    testUserID,
		ShortName: "cash",
		Name:      "Another Cash",
		Type:      models.CashAccount,
	}
	err := env.repo.AddNewWallet(dup)

	assert.Error(t, err)
	assert.IsType(t, models.ErrAccountAlreadyExist{}, err)
}

func TestGetWalletByShortName_success(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	w, err := env.repo.GetWalletByShortName(testUserID, "cash")

	assert.NoError(t, err)
	assert.Equal(t, "cash", w.ShortName)
	assert.Equal(t, float64(1000), w.Balance)
}

func TestGetWalletByShortName_notFound(t *testing.T) {

	env := setupWalletRepo(t)

	w, err := env.repo.GetWalletByShortName(testUserID, "nope")

	assert.Nil(t, w)
	assert.Error(t, err)
	assert.True(t, models.IsErrNotFound(err))
}

func TestListWallets_success(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	bank := &models.Wallet{
		UserID:    testUserID,
		ShortName: "bank",
		Name:      "Bank",
		Type:      models.BankAccount,
		Version:   1,
	}
	require.NoError(t, env.repo.AddNewWallet(bank))

	wallets, err := env.repo.ListWallets(testUserID)

	assert.NoError(t, err)
	assert.Len(t, wallets, 2)
}

func TestListWalletsByType_success(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	bank := &models.Wallet{
		UserID:    testUserID,
		ShortName: "bank",
		Name:      "Bank",
		Type:      models.BankAccount,
		Version:   1,
	}
	require.NoError(t, env.repo.AddNewWallet(bank))

	cashWallets, err := env.repo.ListWalletsByType(testUserID, models.CashAccount)

	assert.NoError(t, err)
	assert.Len(t, cashWallets, 1)
	assert.Equal(t, "cash", cashWallets[0].ShortName)
}

func TestUpdateWalletBalance_success(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	err := env.repo.UpdateWalletBalance(testUserID, "cash", -200)

	assert.NoError(t, err)

	w, err := env.repo.GetWalletByShortName(testUserID, "cash")
	require.NoError(t, err)
	assert.Equal(t, float64(800), w.Balance)
	assert.Equal(t, int64(2), w.Version) // 1 -> 2
}

func TestUpdateWalletBalance_consecutive(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	require.NoError(t, env.repo.UpdateWalletBalance(testUserID, "cash", 50))
	require.NoError(t, env.repo.UpdateWalletBalance(testUserID, "cash", -100))

	w, err := env.repo.GetWalletByShortName(testUserID, "cash")
	require.NoError(t, err)
	assert.Equal(t, float64(950), w.Balance)
	assert.Equal(t, int64(3), w.Version) // 1 -> 2 -> 3
}

func TestUpdateWalletBalance_notFound(t *testing.T) {

	env := setupWalletRepo(t)

	err := env.repo.UpdateWalletBalance(testUserID, "nope", 100)

	assert.Error(t, err)
	assert.True(t, models.IsErrNotFound(err))
}

func TestUpdateWalletBalance_optimisticLockMechanism(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	w, err := env.repo.GetWalletByShortName(testUserID, "cash")
	require.NoError(t, err)

	ctx := context.Background()

	// Stale version should affect 0 rows
	result, err := env.db.Exec(
		ctx,
		`UPDATE "wallet" SET balance = ?, version = ? WHERE id = ? AND version = ?`,
		500.0, 99, w.ID, 999, // version 999 doesn't exist
	)
	require.NoError(t, err)
	rows, _ := result.RowsAffected()
	assert.Equal(t, int64(0), rows, "stale version should match 0 rows")

	// Correct version succeeds
	result, err = env.db.Exec(
		ctx,
		`UPDATE "wallet" SET balance = ?, version = ? WHERE id = ? AND version = ?`,
		500.0, 2, w.ID, 1, // version 1 matches
	)
	require.NoError(t, err)
	rows, _ = result.RowsAffected()
	assert.Equal(t, int64(1), rows, "correct version should match 1 row")
}

func TestDeleteWallet_success(t *testing.T) {

	env := setupWalletRepo(t)
	seedWallet(t, env)

	err := env.repo.DeleteWallet(testUserID, "cash")

	assert.NoError(t, err)

	_, err = env.repo.GetWalletByShortName(testUserID, "cash")
	assert.True(t, models.IsErrNotFound(err))
}
