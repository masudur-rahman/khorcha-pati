package auth

import (
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"
	"github.com/masudur-rahman/expense-tracker-bot/repos/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestAuthService(
	userRepo *mocks.UserRepo,
	authRepo *mocks.AuthRepo,
	messenger *mocks.Messenger,
) *authService {
	cache.Init(cache.Config{Type: cache.CacheMap})
	return &authService{
		userRepo:      userRepo,
		authRepo:      authRepo,
		messenger:     messenger,
		jwtSecret:     "test-jwt-secret",
		refreshSecret: "test-refresh-secret",
		botUsername:   "testbot",
		baseURL:       "https://example.com",
		logger:        zap.NewNop().Sugar(),
	}
}

var testUser = &models.Profile{
	ID:         1,
	TelegramID: 100,
	Username:   "alice",
}

func TestRequestOTP_success(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	userRepo.On("GetUser", models.Profile{Username: "alice"}).
		Return(testUser, nil)
	msgr.On("SendMessage", int64(100), mock.MatchedBy(func(s string) bool {
		return len(s) > 0
	})).Return(nil)

	err := svc.RequestOTP("alice")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	msgr.AssertExpectations(t)
}

func TestRequestOTP_userNotFound(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	userRepo.On("GetUser", models.Profile{Username: "nobody"}).
		Return(nil, models.ErrUserNotFound{Username: "nobody"})
	userRepo.On("GetUser", models.Profile{MobileNumber: "nobody"}).
		Return(nil, models.ErrUserNotFound{})

	err := svc.RequestOTP("nobody")

	assert.Error(t, err)
}

func TestVerifyOTP_success(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	// First request OTP to populate cache
	userRepo.On("GetUser", models.Profile{Username: "alice"}).
		Return(testUser, nil)
	msgr.On("SendMessage", int64(100), mock.Anything).Return(nil)
	require.NoError(t, svc.RequestOTP("alice"))

	// Extract OTP from cache
	raw, ok := cache.GetCache(otpKey(testUser.ID))
	require.True(t, ok)
	var session otpSession
	require.NoError(t, jsonUnmarshal(raw, &session))

	authRepo.On("CreateRefreshToken", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	pair, err := svc.VerifyOTP("alice", session.OTP)

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestVerifyOTP_wrongCode(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	userRepo.On("GetUser", models.Profile{Username: "alice"}).
		Return(testUser, nil)
	msgr.On("SendMessage", int64(100), mock.Anything).Return(nil)
	require.NoError(t, svc.RequestOTP("alice"))

	pair, err := svc.VerifyOTP("alice", "000000")

	assert.Error(t, err)
	assert.Nil(t, pair)
}

func TestVerifyOTP_noSession(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	userRepo.On("GetUser", models.Profile{Username: "alice"}).
		Return(testUser, nil)

	pair, err := svc.VerifyOTP("alice", "123456")

	assert.Error(t, err)
	assert.Nil(t, pair)
}
