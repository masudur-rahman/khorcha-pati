package access

import (
	"context"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"
	accessrepo "github.com/masudur-rahman/khorcha-pati/repos/access"
	"github.com/masudur-rahman/khorcha-pati/services"

	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupAccessService(t *testing.T) *accessService {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(context.Background(), models.Setting{}, models.AllowedUser{}))

	repo := accessrepo.NewSQLAccessRepository(db, zap.NewNop().Sugar())
	return NewAccessService(repo, zap.NewNop().Sugar())
}

func seed(restricted bool, users []string, owner string) services.AccessSeed {
	return services.AccessSeed{
		Restricted:   restricted,
		AllowedUsers: users,
		ReplyText:    "use the live bot",
		Owner:        owner,
	}
}

func TestEnsureSeeded_FirstBootAppliesConfig(t *testing.T) {
	svc := setupAccessService(t)

	require.NoError(t, svc.EnsureSeeded(seed(true, []string{"alice", "@bob", "123456"}, "owner")))

	assert.True(t, svc.IsRestricted())
	assert.Equal(t, "use the live bot", svc.RestrictedReplyText())
	assert.True(t, svc.IsUserAllowed("Alice", 0), "username match is case-insensitive")
	assert.True(t, svc.IsUserAllowed("bob", 0), "@ prefix stripped")
	assert.True(t, svc.IsUserAllowed("", 123456), "numeric entry matches telegram id")
	assert.True(t, svc.IsUserAllowed("OWNER", 0), "owner implicitly allowed")
	assert.False(t, svc.IsUserAllowed("stranger", 999))
}

// Restarts must never resurrect config state over admin edits.
func TestEnsureSeeded_SecondBootIgnoresConfig(t *testing.T) {
	svc := setupAccessService(t)
	require.NoError(t, svc.EnsureSeeded(seed(true, []string{"alice"}, "owner")))

	// Admin actions via UI: revoke alice, disable restriction.
	entries := svc.ListAllowedUsers()
	require.Len(t, entries, 1)
	require.NoError(t, svc.Revoke(entries[0].ID))
	require.NoError(t, svc.SetRestricted(false))

	// Restart with the same config — must not re-seed.
	require.NoError(t, svc.EnsureSeeded(seed(true, []string{"alice"}, "owner")))

	assert.False(t, svc.IsRestricted())
	assert.False(t, svc.IsUserAllowed("alice", 0))
	assert.Empty(t, svc.ListAllowedUsers())
}

func TestAllowRevoke(t *testing.T) {
	svc := setupAccessService(t)
	require.NoError(t, svc.EnsureSeeded(seed(true, nil, "owner")))

	entry, err := svc.Allow("carol", 777)
	require.NoError(t, err)
	assert.True(t, svc.IsUserAllowed("carol", 0))
	assert.True(t, svc.IsUserAllowed("", 777))

	_, err = svc.Allow("carol", 777)
	assert.Error(t, err, "duplicate allow rejected")

	require.NoError(t, svc.Revoke(entry.ID))
	assert.False(t, svc.IsUserAllowed("carol", 777))
}

// A username-only entry gets pinned to the telegram id on first sighting, so
// a later username change can't break or leak access.
func TestNoteSeen_BackfillsTelegramID(t *testing.T) {
	svc := setupAccessService(t)
	require.NoError(t, svc.EnsureSeeded(seed(true, []string{"dave"}, "owner")))

	svc.NoteSeen("Dave", 4242)

	assert.True(t, svc.IsUserAllowed("", 4242))
	// Persisted, not just cached.
	require.NoError(t, svc.reload())
	assert.True(t, svc.IsUserAllowed("", 4242))
}
