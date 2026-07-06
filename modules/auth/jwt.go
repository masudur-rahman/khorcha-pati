package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// AccessTokenTTL is the lifetime of an access token.
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL is the lifetime of a refresh token.
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// AccessClaims holds the JWT claims for an access token.
type AccessClaims struct {
	jwt.RegisteredClaims
	UserID   int64  `json:"uid"`
	Username string `json:"uname"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
}

// RefreshClaims holds the JWT claims for a refresh token.
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID    int64  `json:"uid"`
	TokenUUID string `json:"tuuid"`
}

// TokenPair holds an access/refresh token pair.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	RefreshUUID  string
}

// GenerateAccessToken creates a signed short-lived JWT.
func GenerateAccessToken(userID int64, username, secret string, isAdmin bool) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			Issuer:    "khorcha-pati",
		},
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a signed long-lived JWT with a unique UUID.
func GenerateRefreshToken(userID int64, secret string) (string, string, error) {
	now := time.Now()
	tokenUUID := uuid.New().String()
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
			Issuer:    "khorcha-pati",
		},
		UserID:    userID,
		TokenUUID: tokenUUID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}
	return signed, tokenUUID, nil
}

// ParseAccessToken validates and extracts claims from an access token.
func ParseAccessToken(tokenStr, secret string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, keyFunc(secret))
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}
	return claims, nil
}

// ParseRefreshToken validates and extracts claims from a refresh token.
func ParseRefreshToken(tokenStr, secret string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, keyFunc(secret))
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	claims, ok := token.Claims.(*RefreshClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}
	return claims, nil
}

func keyFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}
}
