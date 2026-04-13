package web

import (
	"encoding/json"
	"net/http"
)

const refreshCookieName = "refresh_token"

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON sends a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError sends a JSON error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, errorResponse{Code: code, Message: message})
}

// ReadJSON decodes a JSON request body into the target.
func ReadJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// SetRefreshCookie sets the refresh token as an httpOnly secure cookie.
func SetRefreshCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearRefreshCookie removes the refresh token cookie.
func ClearRefreshCookie(w http.ResponseWriter) {
	SetRefreshCookie(w, "", -1)
}

// RefreshTokenFromCookie extracts the refresh token from the request cookie.
func RefreshTokenFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
