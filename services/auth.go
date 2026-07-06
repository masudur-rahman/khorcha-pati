package services

import "github.com/masudur-rahman/khorcha-pati/modules/auth"

// AuthService handles OTP, QR, and JWT-based authentication.
type AuthService interface {
	// OTP flow
	RequestOTP(identifier string) error
	VerifyOTP(identifier, code string) (*auth.TokenPair, error)

	// QR flow
	InitQRSession() (sessionID, deepLink string, err error)
	ConfirmQRSession(sessionID string, telegramID int64) error
	DenyQRSession(sessionID string) error
	PollQRSession(sessionID string) (*auth.TokenPair, string, error)

	// Magic Link flow
	CreateMagicLink(userID int64) (string, error)
	VerifyMagicLink(token string) (*auth.TokenPair, error)

	// Token management
	RefreshTokens(refreshToken string) (*auth.TokenPair, error)
	Logout(refreshToken string) error
}
