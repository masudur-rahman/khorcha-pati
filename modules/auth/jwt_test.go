package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-jwt-secret"

const testRefreshSecret = "test-refresh-secret"

func TestGenerateAccessToken_roundtrip(t *testing.T) {
	t.Parallel()
	token, err := GenerateAccessToken(42, "masud", testSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseAccessToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "masud", claims.Username)
	assert.Equal(t, "42", claims.Subject)
	assert.Equal(t, "expense-tracker", claims.Issuer)
}

func TestGenerateRefreshToken_roundtrip(t *testing.T) {
	t.Parallel()
	token, uuid, err := GenerateRefreshToken(42, testRefreshSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, uuid)

	claims, err := ParseRefreshToken(token, testRefreshSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, uuid, claims.TokenUUID)
}

func TestGenerateRefreshToken_uniqueUUIDs(t *testing.T) {
	t.Parallel()
	_, uuid1, err := GenerateRefreshToken(1, testRefreshSecret)
	require.NoError(t, err)
	_, uuid2, err := GenerateRefreshToken(1, testRefreshSecret)
	require.NoError(t, err)
	assert.NotEqual(t, uuid1, uuid2)
}

func TestParseAccessToken_wrongSecret(t *testing.T) {
	t.Parallel()
	token, err := GenerateAccessToken(1, "user", testSecret)
	require.NoError(t, err)

	_, err = ParseAccessToken(token, "wrong-secret")
	assert.Error(t, err)
}

func TestParseRefreshToken_wrongSecret(t *testing.T) {
	t.Parallel()
	token, _, err := GenerateRefreshToken(1, testRefreshSecret)
	require.NoError(t, err)

	_, err = ParseRefreshToken(token, "wrong-secret")
	assert.Error(t, err)
}

func TestParseAccessToken_expired(t *testing.T) {
	t.Parallel()
	now := time.Now().Add(-1 * time.Hour)
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			Issuer:    "expense-tracker",
		},
		UserID:   1,
		Username: "user",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = ParseAccessToken(signed, testSecret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestParseAccessToken_malformed(t *testing.T) {
	t.Parallel()
	_, err := ParseAccessToken("not-a-jwt", testSecret)
	assert.Error(t, err)
}
