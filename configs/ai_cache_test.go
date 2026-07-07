package configs

import (
	"context"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"

	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAICacheTest wires an in-memory SQLite engine and memory cache into the package globals.
func setupAICacheTest(t *testing.T) {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(context.Background(), models.AICache{}))

	dbMu.Lock()
	sqlDB = db
	dbMu.Unlock()
	cache.Init(cache.Config{Type: cache.CacheMap})

	t.Cleanup(func() {
		dbMu.Lock()
		sqlDB = nil
		dbMu.Unlock()
	})
}

func TestCreateAICache_persistsAndCaches(t *testing.T) {
	setupAICacheTest(t)

	entry, err := CreateAICache("lunch at cafe", "food-rest", "Expense", 1.0)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)
	assert.Equal(t, "food-rest", entry.SubcategoryID)

	cached, ok := cache.GetCache("lunch at cafe")
	assert.True(t, ok)
	assert.Contains(t, cached, "food-rest")
	assert.Contains(t, cached, "Expense")
}

func TestCreateAICache_duplicate(t *testing.T) {
	setupAICacheTest(t)

	_, err := CreateAICache("salary credit", "fin-sal", "Income", 1.0)
	require.NoError(t, err)

	_, err = CreateAICache("salary credit", "fin-sal", "Income", 1.0)
	assert.ErrorIs(t, err, ErrAICacheDuplicate)
}

func TestUpdateAICacheClassification_updatesDBAndCache(t *testing.T) {
	setupAICacheTest(t)

	entry, err := CreateAICache("monthly rent", "food-rest", "Expense", 1.0)
	require.NoError(t, err)

	updated, err := UpdateAICacheClassification(entry.ID, "house-rent", "Expense", 0.9)
	require.NoError(t, err)
	assert.Equal(t, "house-rent", updated.SubcategoryID)
	assert.Equal(t, 0.9, updated.Confidence)

	cached, ok := cache.GetCache("monthly rent")
	require.True(t, ok)
	assert.Contains(t, cached, "house-rent")
}

func TestUpdateAICacheClassification_notFound(t *testing.T) {
	setupAICacheTest(t)

	_, err := UpdateAICacheClassification(999, "food-rest", "Expense", 1.0)
	assert.ErrorIs(t, err, ErrAICacheNotFound)
}

func TestDeleteAICache_removesDBAndCache(t *testing.T) {
	setupAICacheTest(t)

	entry, err := CreateAICache("taxi fare", "trans-taxi", "Expense", 1.0)
	require.NoError(t, err)

	require.NoError(t, DeleteAICache(entry.ID))

	_, ok := cache.GetCache("taxi fare")
	assert.False(t, ok)

	rows, err := ListAICache("taxi", 0)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestDeleteAICache_notFound(t *testing.T) {
	setupAICacheTest(t)
	assert.ErrorIs(t, DeleteAICache(404), ErrAICacheNotFound)
}

func TestListAICache_filterAndLimit(t *testing.T) {
	setupAICacheTest(t)

	_, err := CreateAICache("grocery shopping", "food-groc", "Expense", 1.0)
	require.NoError(t, err)
	_, err = CreateAICache("fuel refill", "trans-fuel", "Expense", 1.0)
	require.NoError(t, err)

	all, err := ListAICache("", 0)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	filtered, err := ListAICache("grocery", 0)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	assert.Equal(t, "grocery shopping", filtered[0].InputText)
}
