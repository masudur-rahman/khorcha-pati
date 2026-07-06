package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authmod "github.com/masudur-rahman/khorcha-pati/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key"

func TestJWTAuth_validToken(t *testing.T) {
	token, err := authmod.GenerateAccessToken(1, "alice", testJWTSecret, false)
	require.NoError(t, err)

	var gotClaims *authmod.AccessClaims
	handler := JWTAuth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := UserFromContext(r.Context())
		assert.True(t, ok)
		gotClaims = claims
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(1), gotClaims.UserID)
	assert.Equal(t, "alice", gotClaims.Username)
}

func TestJWTAuth_missingHeader(t *testing.T) {
	handler := JWTAuth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_wrongSecret(t *testing.T) {
	token, err := authmod.GenerateAccessToken(1, "alice", "other-secret", false)
	require.NoError(t, err)

	handler := JWTAuth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_malformedToken(t *testing.T) {
	handler := JWTAuth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
