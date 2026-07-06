package auth

import (
	"fmt"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	authmod "github.com/masudur-rahman/khorcha-pati/modules/auth"
)

type otpSession struct {
	OTP        string `json:"otp"`
	Identifier string `json:"identifier"`
}

type qrSession struct {
	Status string `json:"status"`
	UserID int64  `json:"userID,omitempty"`
}

func otpKey(userID int64) string { return fmt.Sprintf("otp:%d", userID) }

func qrKey(sessionID string) string { return fmt.Sprintf("qr:%s", sessionID) }

func magicKey(token string) string { return fmt.Sprintf("magic:%s", token) }

func (s *authService) lookupUser(identifier string) (*models.Profile, error) {
	user, err := s.userRepo.GetUser(models.Profile{Username: identifier})
	if err == nil {
		return user, nil
	}
	if !models.IsErrNotFound(err) {
		return nil, err
	}
	return s.userRepo.GetUser(models.Profile{MobileNumber: identifier})
}

func (s *authService) issueTokenPair(user *models.Profile) (*authmod.TokenPair, error) {
	if !user.IsActive {
		return nil, models.StatusError{Status: 403, Message: "account disabled"}
	}
	accessToken, err := authmod.GenerateAccessToken(user.ID, user.Username, s.jwtSecret, user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshUUID, err := authmod.GenerateRefreshToken(user.ID, s.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	dbToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenUUID: refreshUUID,
		ExpiresAt: time.Now().Add(authmod.RefreshTokenTTL).Unix(),
		CreatedAt: time.Now().Unix(),
	}
	if err := s.authRepo.CreateRefreshToken(dbToken); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &authmod.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		RefreshUUID:  refreshUUID,
	}, nil
}
