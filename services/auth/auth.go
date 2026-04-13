package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	authmod "github.com/masudur-rahman/expense-tracker-bot/modules/auth"
	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"
	"github.com/masudur-rahman/expense-tracker-bot/repos"
	"github.com/masudur-rahman/expense-tracker-bot/services"

	"github.com/google/uuid"
)

const (
	otpTTL = 5 * time.Minute
	qrTTL  = 5 * time.Minute

	qrStatusPending  = "pending"
	qrStatusApproved = "approved"
)

type authService struct {
	userRepo      repos.UserRepository
	authRepo      repos.AuthRepository
	messenger     authmod.Messenger
	jwtSecret     string
	refreshSecret string
	botUsername    string
	logger        logr.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo repos.UserRepository,
	authRepo repos.AuthRepository,
	messenger authmod.Messenger,
	jwtSecret, refreshSecret, botUsername string,
	logger logr.Logger,
) services.AuthService {
	return &authService{
		userRepo:      userRepo,
		authRepo:      authRepo,
		messenger:     messenger,
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		botUsername:    botUsername,
		logger:        logger,
	}
}

func (s *authService) RequestOTP(identifier string) error {
	user, err := s.lookupUser(identifier)
	if err != nil {
		return err
	}

	otp, err := authmod.GenerateOTP()
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	session := otpSession{OTP: otp, Identifier: identifier}
	data, _ := json.Marshal(session)
	if err := cache.SetCache(otpKey(user.ID), string(data), otpTTL); err != nil {
		return fmt.Errorf("cache otp: %w", err)
	}

	msg := fmt.Sprintf("Your login code: %s\nExpires in 5 minutes.", otp)
	return s.messenger.SendMessage(user.TelegramID, msg)
}

func (s *authService) VerifyOTP(identifier, code string) (*authmod.TokenPair, error) {
	user, err := s.lookupUser(identifier)
	if err != nil {
		return nil, err
	}

	raw, ok := cache.GetCache(otpKey(user.ID))
	if !ok {
		return nil, models.StatusError{Status: 400, Message: "otp expired or not requested"}
	}

	var session otpSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, fmt.Errorf("parse otp session: %w", err)
	}

	if session.OTP != code {
		return nil, models.StatusError{Status: 400, Message: "invalid otp code"}
	}

	return s.issueTokenPair(user)
}

func (s *authService) InitQRSession() (string, string, error) {
	sessionID := uuid.New().String()
	session := qrSession{Status: qrStatusPending}
	data, _ := json.Marshal(session)

	if err := cache.SetCache(qrKey(sessionID), string(data), qrTTL); err != nil {
		return "", "", fmt.Errorf("cache qr session: %w", err)
	}

	deepLink := fmt.Sprintf("https://t.me/%s?start=login_%s", s.botUsername, sessionID)
	return sessionID, deepLink, nil
}

func (s *authService) ConfirmQRSession(sessionID string, telegramID int64) error {
	raw, ok := cache.GetCache(qrKey(sessionID))
	if !ok {
		return models.StatusError{Status: 400, Message: "qr session expired"}
	}

	var session qrSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return fmt.Errorf("parse qr session: %w", err)
	}

	if session.Status != qrStatusPending {
		return models.StatusError{Status: 400, Message: "qr session already used"}
	}

	user, err := s.userRepo.GetUser(models.Profile{TelegramID: telegramID})
	if err != nil {
		return err
	}

	session.Status = qrStatusApproved
	session.UserID = user.ID
	data, _ := json.Marshal(session)
	return cache.SetCache(qrKey(sessionID), string(data), qrTTL)
}

func (s *authService) PollQRSession(sessionID string) (*authmod.TokenPair, string, error) {
	raw, ok := cache.GetCache(qrKey(sessionID))
	if !ok {
		return nil, "expired", models.StatusError{Status: 410, Message: "qr session expired"}
	}

	var session qrSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, "", fmt.Errorf("parse qr session: %w", err)
	}

	if session.Status == qrStatusPending {
		return nil, qrStatusPending, nil
	}

	user, err := s.userRepo.GetUserByID(session.UserID)
	if err != nil {
		return nil, "", err
	}

	pair, err := s.issueTokenPair(user)
	if err != nil {
		return nil, "", err
	}
	return pair, qrStatusApproved, nil
}

func (s *authService) RefreshTokens(refreshToken string) (*authmod.TokenPair, error) {
	claims, err := authmod.ParseRefreshToken(refreshToken, s.refreshSecret)
	if err != nil {
		return nil, models.StatusError{Status: 401, Message: "invalid refresh token"}
	}

	dbToken, err := s.authRepo.GetRefreshTokenByUUID(claims.TokenUUID)
	if err != nil {
		// Reuse detection: token not found means it was already rotated
		_ = s.authRepo.RevokeAllUserTokens(claims.UserID)
		return nil, models.StatusError{Status: 401, Message: "refresh token reuse detected"}
	}

	if dbToken.Revoked == 1 {
		_ = s.authRepo.RevokeAllUserTokens(claims.UserID)
		return nil, models.StatusError{Status: 401, Message: "refresh token revoked"}
	}

	if err := s.authRepo.RevokeRefreshToken(claims.TokenUUID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return s.issueTokenPair(user)
}

func (s *authService) Logout(refreshToken string) error {
	claims, err := authmod.ParseRefreshToken(refreshToken, s.refreshSecret)
	if err != nil {
		return models.StatusError{Status: 401, Message: "invalid refresh token"}
	}
	return s.authRepo.RevokeRefreshToken(claims.TokenUUID)
}
