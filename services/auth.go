package services

import "github.com/masudur-rahman/expense-tracker-bot/modules/auth"

// AuthService handles OTP, QR, and JWT-based authentication.
type AuthService interface {
	// OTP flow
	RequestOTP(identifier string) error
	VerifyOTP(identifier, code string) (*auth.TokenPair, error)

	// QR flow
	InitQRSession() (sessionID, deepLink string, err error)
	ConfirmQRSession(sessionID string, telegramID int64) error
	PollQRSession(sessionID string) (*auth.TokenPair, string, error)

	// Token management
	RefreshTokens(refreshToken string) (*auth.TokenPair, error)
	Logout(refreshToken string) error
}
