package repos

import "github.com/masudur-rahman/khorcha-pati/models"

// AuthRepository defines data access for refresh tokens.
type AuthRepository interface {
	CreateRefreshToken(token *models.RefreshToken) error
	GetRefreshTokenByUUID(uuid string) (*models.RefreshToken, error)
	RevokeRefreshToken(uuid string) error
	RevokeAllUserTokens(userID int64) error
	DeleteExpiredTokens() error
}
