package web

import (
	"net/http"

	"github.com/masudur-rahman/expense-tracker-bot/services/all"
)

// 7 days in seconds.
const refreshCookieMaxAge = 7 * 24 * 60 * 60

type otpRequest struct {
	Identifier string `json:"identifier"`
}

type otpVerifyRequest struct {
	Identifier string `json:"identifier"`
	Code       string `json:"code"`
}

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
}

type qrInitResponse struct {
	SessionID string `json:"sessionID"`
	DeepLink  string `json:"deepLink"`
}

type qrStatusResponse struct {
	Status      string `json:"status"`
	AccessToken string `json:"accessToken,omitempty"`
}

// HandleRequestOTP handles POST /auth/request-otp.
func HandleRequestOTP(w http.ResponseWriter, r *http.Request) {
	var req otpRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if req.Identifier == "" {
		WriteError(w, http.StatusBadRequest, "bad_request", "identifier is required")
		return
	}

	if err := all.GetServices().Auth.RequestOTP(req.Identifier); err != nil {
		WriteError(w, http.StatusInternalServerError, "otp_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "otp sent"})
}

// HandleVerifyOTP handles POST /auth/verify-otp.
func HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req otpVerifyRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if req.Identifier == "" || req.Code == "" {
		WriteError(w, http.StatusBadRequest, "bad_request", "identifier and code are required")
		return
	}

	pair, err := all.GetServices().Auth.VerifyOTP(req.Identifier, req.Code)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "verify_failed", err.Error())
		return
	}

	SetRefreshCookie(w, pair.RefreshToken, refreshCookieMaxAge)
	WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: pair.AccessToken})
}

// HandleQRInit handles POST /auth/qr/init.
func HandleQRInit(w http.ResponseWriter, r *http.Request) {
	sessionID, deepLink, err := all.GetServices().Auth.InitQRSession()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "qr_init_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, qrInitResponse{SessionID: sessionID, DeepLink: deepLink})
}

// HandleQRPoll handles GET /auth/qr/status?session=<id>.
func HandleQRPoll(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "bad_request", "session query param required")
		return
	}

	pair, status, err := all.GetServices().Auth.PollQRSession(sessionID)
	if err != nil && status != "expired" {
		WriteError(w, http.StatusInternalServerError, "poll_failed", err.Error())
		return
	}

	resp := qrStatusResponse{Status: status}
	if pair != nil {
		SetRefreshCookie(w, pair.RefreshToken, refreshCookieMaxAge)
		resp.AccessToken = pair.AccessToken
	}
	WriteJSON(w, http.StatusOK, resp)
}

// HandleRefresh handles POST /auth/refresh.
func HandleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := RefreshTokenFromCookie(r)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "missing_token", "refresh token cookie required")
		return
	}

	pair, err := all.GetServices().Auth.RefreshTokens(refreshToken)
	if err != nil {
		ClearRefreshCookie(w)
		WriteError(w, http.StatusUnauthorized, "refresh_failed", err.Error())
		return
	}

	SetRefreshCookie(w, pair.RefreshToken, refreshCookieMaxAge)
	WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: pair.AccessToken})
}

// HandleLogout handles POST /auth/logout.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := RefreshTokenFromCookie(r)
	if err != nil {
		ClearRefreshCookie(w)
		WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		return
	}

	_ = all.GetServices().Auth.Logout(refreshToken)
	ClearRefreshCookie(w)
	WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
