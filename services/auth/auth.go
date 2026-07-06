package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	authmod "github.com/masudur-rahman/khorcha-pati/modules/auth"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
	"github.com/masudur-rahman/khorcha-pati/repos"
	"github.com/masudur-rahman/khorcha-pati/services"

	"github.com/google/uuid"
)

const (
	otpTTL = 5 * time.Minute
	qrTTL  = 5 * time.Minute

	qrStatusPending  = "pending"
	qrStatusApproved = "approved"
	qrStatusDenied   = "denied"
)

type authService struct {
	userRepo      repos.UserRepository
	authRepo      repos.AuthRepository
	messenger     authmod.Messenger
	jwtSecret     string
	refreshSecret string
	botUsername   string
	baseURL       string
	logger        logr.Logger
	refreshMu     sync.Mutex
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo repos.UserRepository,
	authRepo repos.AuthRepository,
	messenger authmod.Messenger,
	jwtSecret, refreshSecret, botUsername, baseURL string,
	logger logr.Logger,
) services.AuthService {
	return &authService{
		userRepo:      userRepo,
		authRepo:      authRepo,
		messenger:     messenger,
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		botUsername:   botUsername,
		baseURL:       strings.TrimRight(baseURL, "/"),
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
		return nil, models.StatusError{Status: 400, Message: "Verification code has expired or was not requested."}
	}

	var session otpSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, fmt.Errorf("parse otp session: %w", err)
	}

	if session.OTP != code {
		return nil, models.StatusError{Status: 400, Message: "The verification code you entered is incorrect."}
	}

	return s.issueTokenPair(user)
}

func (s *authService) InitQRSession() (string, string, error) {
	sessionID := strings.ReplaceAll(uuid.New().String(), "-", "")
	session := qrSession{Status: qrStatusPending}
	data, _ := json.Marshal(session)

	if err := cache.SetCache(qrKey(sessionID), string(data), qrTTL); err != nil {
		return "", "", fmt.Errorf("cache qr session: %w", err)
	}

	deepLink := fmt.Sprintf("%s/api/v1/auth/qr/redirect?session=%s", s.baseURL, sessionID)
	s.logger.Infow("QR session initialized", "sessionID", sessionID, "deepLink", deepLink)
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

func (s *authService) DenyQRSession(sessionID string) error {
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

	session.Status = qrStatusDenied
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

	if session.Status == qrStatusDenied {
		return nil, qrStatusDenied, nil
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

func (s *authService) CreateMagicLink(userID int64) (string, error) {
	token := strings.ReplaceAll(uuid.New().String(), "-", "")
	if err := cache.SetCache(magicKey(token), fmt.Sprintf("%d", userID), 5*time.Minute); err != nil {
		return "", fmt.Errorf("cache magic link: %w", err)
	}
	return token, nil
}

func (s *authService) VerifyMagicLink(token string) (*authmod.TokenPair, error) {
	raw, ok := cache.GetCache(magicKey(token))
	if !ok {
		return nil, models.StatusError{Status: 400, Message: "magic link expired or invalid"}
	}
	_ = cache.DeleteCache(magicKey(token))

	var userID int64
	if _, err := fmt.Sscanf(raw, "%d", &userID); err != nil {
		return nil, fmt.Errorf("parse userID from cache: %w", err)
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokenPair(user)
}

func (s *authService) RefreshTokens(refreshToken string) (*authmod.TokenPair, error) {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	claims, err := authmod.ParseRefreshToken(refreshToken, s.refreshSecret)
	if err != nil {
		s.logger.Warnw("token_refresh_failed", "reason", "parse_error", "error", err.Error())
		return nil, models.StatusError{Status: 401, Message: "invalid refresh token"}
	}

	dbToken, err := s.authRepo.GetRefreshTokenByUUID(claims.TokenUUID)
	if err != nil {
		if models.IsErrNotFound(err) {
			// Reuse detection: token not found means it was already rotated
			// CHECK CACHE: If we recently successfully refreshed this token, return that result.
			if cached, ok := cache.GetCache("refreshed:" + claims.TokenUUID); ok {
				var pair authmod.TokenPair
				if err := json.Unmarshal([]byte(cached), &pair); err == nil {
					s.logger.Infow("token_refresh_concurrent_success", "uuid", claims.TokenUUID)
					return &pair, nil
				}
			}

			s.logger.Warnw("token_refresh_failed", "reason", "token_not_found", "uuid", claims.TokenUUID, "userID", claims.UserID)
			_ = s.authRepo.RevokeAllUserTokens(claims.UserID)
			return nil, models.StatusError{Status: 401, Message: "refresh token reuse detected"}
		}
		s.logger.Errorw("token_refresh_failed", "reason", "db_error", "error", err.Error())
		return nil, fmt.Errorf("lookup refresh token: %w", err)
	}

	if dbToken.Revoked == 1 {
		// CHECK CACHE: If we recently successfully refreshed this token, return that result.
		if cached, ok := cache.GetCache("refreshed:" + claims.TokenUUID); ok {
			var pair authmod.TokenPair
			if err := json.Unmarshal([]byte(cached), &pair); err == nil {
				s.logger.Infow("token_refresh_concurrent_success", "uuid", claims.TokenUUID)
				return &pair, nil
			}
		}

		s.logger.Warnw("token_refresh_failed", "reason", "token_revoked", "uuid", claims.TokenUUID, "userID", claims.UserID)
		_ = s.authRepo.RevokeAllUserTokens(claims.UserID)
		return nil, models.StatusError{Status: 401, Message: "refresh token revoked"}
	}

	if err := s.authRepo.RevokeRefreshToken(claims.TokenUUID); err != nil {
		s.logger.Errorw("token_refresh_failed", "reason", "revoke_error", "error", err.Error())
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		s.logger.Errorw("token_refresh_failed", "reason", "user_not_found", "userID", claims.UserID)
		return nil, err
	}

	s.logger.Infow("token_refreshed", "userID", user.ID)
	pair, err := s.issueTokenPair(user)
	if err == nil {
		// Cache the successful result for 15 seconds to handle concurrent requests
		data, _ := json.Marshal(pair)
		_ = cache.SetCache("refreshed:"+claims.TokenUUID, string(data), 15*time.Second)
	}
	return pair, err
}

func (s *authService) Logout(refreshToken string) error {
	claims, err := authmod.ParseRefreshToken(refreshToken, s.refreshSecret)
	if err != nil {
		return models.StatusError{Status: 401, Message: "invalid refresh token"}
	}
	return s.authRepo.RevokeRefreshToken(claims.TokenUUID)
}
