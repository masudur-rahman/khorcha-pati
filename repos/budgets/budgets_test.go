package budgets

import (
	"context"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testUserID int64 = 99

func setupBudgetRepo(t *testing.T) *SQLBudgetRepository {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(context.Background(), models.Budget{}))

	logger := zap.NewNop().Sugar()
	return NewSQLBudgetRepository(db, logger)
}

func seedBudget(t *testing.T, repo *SQLBudgetRepository, categoryID string, amount float64) {
	t.Helper()
	require.NoError(t, repo.UpsertBudget(&models.Budget{
		UserID:     testUserID,
		CategoryID: categoryID,
		Amount:     amount,
		AlertAt:    80,
	}))
}

func TestUpsertBudget_insert(t *testing.T) {
	repo := setupBudgetRepo(t)

	err := repo.UpsertBudget(&models.Budget{
		UserID:     testUserID,
		CategoryID: "food",
		Amount:     5000,
		AlertAt:    80,
	})

	assert.NoError(t, err)

	b, err := repo.GetBudget(testUserID, "food")
	assert.NoError(t, err)
	assert.Equal(t, float64(5000), b.Amount)
	assert.Equal(t, int64(80), b.AlertAt)
}

func TestUpsertBudget_update(t *testing.T) {
	repo := setupBudgetRepo(t)
	seedBudget(t, repo, "food", 5000)

	err := repo.UpsertBudget(&models.Budget{
		UserID:     testUserID,
		CategoryID: "food",
		Amount:     8000,
		AlertAt:    90,
	})

	assert.NoError(t, err)

	b, err := repo.GetBudget(testUserID, "food")
	assert.NoError(t, err)
	assert.Equal(t, float64(8000), b.Amount)
	assert.Equal(t, int64(90), b.AlertAt)
}

func TestGetBudget_notFound(t *testing.T) {
	repo := setupBudgetRepo(t)

	b, err := repo.GetBudget(testUserID, "food")

	assert.Nil(t, b)
	assert.Error(t, err)
	assert.True(t, models.IsErrNotFound(err))
}

func TestListBudgets_empty(t *testing.T) {
	repo := setupBudgetRepo(t)

	budgets, err := repo.ListBudgets(testUserID)

	assert.NoError(t, err)
	assert.Empty(t, budgets)
}

func TestListBudgets_multiple(t *testing.T) {
	repo := setupBudgetRepo(t)
	seedBudget(t, repo, "food", 5000)
	seedBudget(t, repo, "", 25000)

	budgets, err := repo.ListBudgets(testUserID)

	assert.NoError(t, err)
	assert.Len(t, budgets, 2)
}

func TestDeleteBudget_success(t *testing.T) {
	repo := setupBudgetRepo(t)
	seedBudget(t, repo, "food", 5000)

	err := repo.DeleteBudget(testUserID, "food")

	assert.NoError(t, err)

	_, err = repo.GetBudget(testUserID, "food")
	assert.True(t, models.IsErrNotFound(err))
}
