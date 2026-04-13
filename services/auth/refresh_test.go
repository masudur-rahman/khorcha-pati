package auth

import (
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	authmod "github.com/masudur-rahman/expense-tracker-bot/modules/auth"
	"github.com/masudur-rahman/expense-tracker-bot/repos/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokens_success(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	// Issue an initial token pair
	refreshStr, refreshUUID, err := authmod.GenerateRefreshToken(1, "test-refresh-secret")
	require.NoError(t, err)

	authRepo.On("GetRefreshTokenByUUID", refreshUUID).Return(&models.RefreshToken{
		UserID:    1,
		TokenUUID: refreshUUID,
		Revoked:   0,
	}, nil)
	authRepo.On("RevokeRefreshToken", refreshUUID).Return(nil)
	userRepo.On("GetUserByID", int64(1)).Return(testUser, nil)
	authRepo.On("CreateRefreshToken", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	pair, err := svc.RefreshTokens(refreshStr)

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestRefreshTokens_reuseDetection(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	refreshStr, refreshUUID, err := authmod.GenerateRefreshToken(1, "test-refresh-secret")
	require.NoError(t, err)

	// Token not found in DB — reuse detected
	authRepo.On("GetRefreshTokenByUUID", refreshUUID).
		Return(nil, models.ErrRefreshTokenNotFound{UUID: refreshUUID})
	authRepo.On("RevokeAllUserTokens", int64(1)).Return(nil)

	pair, err := svc.RefreshTokens(refreshStr)

	assert.Error(t, err)
	assert.Nil(t, pair)
	authRepo.AssertCalled(t, "RevokeAllUserTokens", int64(1))
}

func TestRefreshTokens_revoked(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	refreshStr, refreshUUID, err := authmod.GenerateRefreshToken(1, "test-refresh-secret")
	require.NoError(t, err)

	authRepo.On("GetRefreshTokenByUUID", refreshUUID).Return(&models.RefreshToken{
		UserID:    1,
		TokenUUID: refreshUUID,
		Revoked:   1,
	}, nil)
	authRepo.On("RevokeAllUserTokens", int64(1)).Return(nil)

	pair, err := svc.RefreshTokens(refreshStr)

	assert.Error(t, err)
	assert.Nil(t, pair)
	authRepo.AssertCalled(t, "RevokeAllUserTokens", int64(1))
}

func TestRefreshTokens_invalidToken(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	pair, err := svc.RefreshTokens("garbage-token")

	assert.Error(t, err)
	assert.Nil(t, pair)
}

func TestLogout_success(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	refreshStr, refreshUUID, err := authmod.GenerateRefreshToken(1, "test-refresh-secret")
	require.NoError(t, err)

	authRepo.On("RevokeRefreshToken", refreshUUID).Return(nil)

	err = svc.Logout(refreshStr)

	assert.NoError(t, err)
	authRepo.AssertCalled(t, "RevokeRefreshToken", refreshUUID)
}

func TestLogout_invalidToken(t *testing.T) {
	userRepo := &mocks.UserRepo{}
	authRepo := &mocks.AuthRepo{}
	msgr := &mocks.Messenger{}
	svc := newTestAuthService(userRepo, authRepo, msgr)

	err := svc.Logout("garbage-token")

	assert.Error(t, err)
}
