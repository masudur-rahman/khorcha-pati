package auth

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

const testUserID int64 = 42

type testEnv struct {
	repo *sqlAuthRepo
	db   isql.Engine
}

func setupAuthRepo(t *testing.T) testEnv {
	t.Helper()
	conn, err := lib.GetSQLiteConnection(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	db := sqlite.NewSQLite(conn)
	require.NoError(t, db.Sync(context.Background(), models.RefreshToken{}))

	logger := zap.NewNop().Sugar()
	repo := NewSQLAuthRepository(db, logger).(*sqlAuthRepo)
	return testEnv{repo: repo, db: db}
}

func seedToken(t *testing.T, env testEnv, uuid string, userID int64, revoked int64, expiresAt int64) {
	t.Helper()
	tok := &models.RefreshToken{
		UserID:    userID,
		TokenUUID: uuid,
		ExpiresAt: expiresAt,
		Revoked:   revoked,
		CreatedAt: time.Now().Unix(),
	}
	require.NoError(t, env.repo.CreateRefreshToken(tok))
}

func TestCreateRefreshToken_success(t *testing.T) {
	env := setupAuthRepo(t)

	tok := &models.RefreshToken{
		UserID:    testUserID,
		TokenUUID: "uuid-create",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		CreatedAt: time.Now().Unix(),
	}
	err := env.repo.CreateRefreshToken(tok)

	assert.NoError(t, err)
}

func TestGetRefreshTokenByUUID_success(t *testing.T) {
	env := setupAuthRepo(t)
	seedToken(t, env, "uuid-get", testUserID, 0, time.Now().Add(time.Hour).Unix())

	tok, err := env.repo.GetRefreshTokenByUUID("uuid-get")

	assert.NoError(t, err)
	assert.Equal(t, "uuid-get", tok.TokenUUID)
	assert.Equal(t, testUserID, tok.UserID)
}

func TestGetRefreshTokenByUUID_notFound(t *testing.T) {
	env := setupAuthRepo(t)

	tok, err := env.repo.GetRefreshTokenByUUID("nonexistent")

	assert.Nil(t, tok)
	assert.Error(t, err)
	assert.True(t, models.IsErrNotFound(err))
}

func TestRevokeRefreshToken_success(t *testing.T) {
	env := setupAuthRepo(t)
	seedToken(t, env, "uuid-revoke", testUserID, 0, time.Now().Add(time.Hour).Unix())

	err := env.repo.RevokeRefreshToken("uuid-revoke")

	assert.NoError(t, err)

	tok, err := env.repo.GetRefreshTokenByUUID("uuid-revoke")
	require.NoError(t, err)
	assert.Equal(t, int64(1), tok.Revoked)
}

func TestRevokeRefreshToken_notFound(t *testing.T) {
	env := setupAuthRepo(t)

	err := env.repo.RevokeRefreshToken("nonexistent")

	assert.Error(t, err)
	assert.True(t, models.IsErrNotFound(err))
}

func TestRevokeAllUserTokens_success(t *testing.T) {
	env := setupAuthRepo(t)
	future := time.Now().Add(time.Hour).Unix()
	seedToken(t, env, "uuid-all-1", testUserID, 0, future)
	seedToken(t, env, "uuid-all-2", testUserID, 0, future)
	seedToken(t, env, "uuid-all-3", testUserID, 1, future) // already revoked

	err := env.repo.RevokeAllUserTokens(testUserID)

	assert.NoError(t, err)

	tok1, _ := env.repo.GetRefreshTokenByUUID("uuid-all-1")
	tok2, _ := env.repo.GetRefreshTokenByUUID("uuid-all-2")
	tok3, _ := env.repo.GetRefreshTokenByUUID("uuid-all-3")
	assert.Equal(t, int64(1), tok1.Revoked)
	assert.Equal(t, int64(1), tok2.Revoked)
	assert.Equal(t, int64(1), tok3.Revoked)
}

func TestDeleteExpiredTokens_success(t *testing.T) {
	env := setupAuthRepo(t)
	past := time.Now().Add(-time.Hour).Unix()
	future := time.Now().Add(time.Hour).Unix()
	seedToken(t, env, "uuid-expired", testUserID, 0, past)
	seedToken(t, env, "uuid-valid", testUserID, 0, future)

	err := env.repo.DeleteExpiredTokens()

	assert.NoError(t, err)

	_, err = env.repo.GetRefreshTokenByUUID("uuid-expired")
	assert.True(t, models.IsErrNotFound(err))

	tok, err := env.repo.GetRefreshTokenByUUID("uuid-valid")
	assert.NoError(t, err)
	assert.Equal(t, "uuid-valid", tok.TokenUUID)
}
