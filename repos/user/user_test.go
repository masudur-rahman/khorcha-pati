package user

import (
	"context"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	isql "github.com/masudur-rahman/styx/sql"
	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupUserRepo(t *testing.T) (*SQLUserRepository, isql.Engine) {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(context.Background(), models.Profile{}))

	return NewSQLUserRepository(db, zap.NewNop().Sugar()), db
}

// Regression: an active user must be findable by struct-filter lookups. The
// `is_active,req` tag previously forced `is_active = false` into every WHERE,
// so active users (is_active = true) were never matched.
func TestGetUser_ActiveUserIsFound(t *testing.T) {
	repo, _ := setupUserRepo(t)

	admin := &models.Profile{TelegramID: 555001, Username: "admin", IsAdmin: true, IsActive: true}
	require.NoError(t, repo.AddNewUser(admin))

	t.Run("by telegram id", func(t *testing.T) {
		got, err := repo.GetUser(models.Profile{TelegramID: 555001})
		require.NoError(t, err)
		assert.Equal(t, "admin", got.Username)
		assert.True(t, got.IsActive)
	})

	t.Run("by username", func(t *testing.T) {
		got, err := repo.GetUserByUsername("admin")
		require.NoError(t, err)
		assert.Equal(t, int64(555001), got.TelegramID)
	})
}

// Disabled users remain findable (blocking is done by explicit IsActive checks
// in the auth/middleware layer, not by the lookup filter).
func TestGetUser_DisabledUserIsFound(t *testing.T) {
	repo, _ := setupUserRepo(t)

	require.NoError(t, repo.AddNewUser(&models.Profile{TelegramID: 555002, Username: "disabled", IsActive: false}))

	got, err := repo.GetUser(models.Profile{TelegramID: 555002})
	require.NoError(t, err)
	assert.Equal(t, "disabled", got.Username)
	assert.False(t, got.IsActive)
}
