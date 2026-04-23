package auth

import (
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"
	"github.com/masudur-rahman/expense-tracker-bot/repos/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInitQRSession_success(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	sessionID, deepLink, err := svc.InitQRSession()

	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)
	assert.Contains(t, deepLink, "https://example.com/api/v1/auth/qr/redirect?session=")
}

func TestConfirmQRSession_success(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	sessionID, _, err := svc.InitQRSession()
	require.NoError(t, err)

	userRepo.On("GetUser", models.Profile{TelegramID: int64(100)}).
		Return(testUser, nil)

	err = svc.ConfirmQRSession(sessionID, 100)

	assert.NoError(t, err)
}

func TestConfirmQRSession_expired(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	err := svc.ConfirmQRSession("nonexistent", 100)

	assert.Error(t, err)
}

func TestPollQRSession_pending(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	sessionID, _, err := svc.InitQRSession()
	require.NoError(t, err)

	pair, status, err := svc.PollQRSession(sessionID)

	assert.NoError(t, err)
	assert.Nil(t, pair)
	assert.Equal(t, "pending", status)
}

func TestPollQRSession_approved(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	sessionID, _, err := svc.InitQRSession()
	require.NoError(t, err)

	userRepo.On("GetUser", models.Profile{TelegramID: int64(100)}).
		Return(testUser, nil)
	require.NoError(t, svc.ConfirmQRSession(sessionID, 100))

	userRepo.On("GetUserByID", int64(1)).Return(testUser, nil)
	authRepo.On("CreateRefreshToken", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	pair, status, err := svc.PollQRSession(sessionID)

	assert.NoError(t, err)
	assert.Equal(t, "approved", status)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestPollQRSession_expired(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	// Clear cache to ensure miss
	cache.Init(cache.Config{Type: cache.CacheMap})

	pair, status, err := svc.PollQRSession("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, pair)
	assert.Equal(t, "expired", status)
}
