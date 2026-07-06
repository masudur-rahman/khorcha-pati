package mocks

import (
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"

	"github.com/stretchr/testify/mock"
)

// AuthRepo is a mock for repos.AuthRepository.
type AuthRepo struct {
	mock.Mock
}

var _ repos.AuthRepository = &AuthRepo{}

func (m *AuthRepo) CreateRefreshToken(token *models.RefreshToken) error {
	return m.Called(token).Error(0)
}

func (m *AuthRepo) GetRefreshTokenByUUID(uuid string) (*models.RefreshToken, error) {
	args := m.Called(uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *AuthRepo) RevokeRefreshToken(uuid string) error {
	return m.Called(uuid).Error(0)
}

func (m *AuthRepo) RevokeAllUserTokens(userID int64) error {
	return m.Called(userID).Error(0)
}

func (m *AuthRepo) DeleteExpiredTokens() error {
	return m.Called().Error(0)
}
