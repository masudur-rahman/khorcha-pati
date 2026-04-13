package web

import (
	"context"
	"net/http"
	"strings"

	authmod "github.com/masudur-rahman/expense-tracker-bot/modules/auth"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// JWTAuth returns middleware that validates Bearer access tokens.
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				WriteError(w, http.StatusUnauthorized, "missing_token", "authorization header required")
				return
			}

			claims, err := authmod.ParseAccessToken(token, jwtSecret)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext extracts AccessClaims stored by JWTAuth middleware.
func UserFromContext(ctx context.Context) (*authmod.AccessClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*authmod.AccessClaims)
	return claims, ok
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}
