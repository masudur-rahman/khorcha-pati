package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/services/all"
)

// sessionIDPattern matches UUID-like session identifiers (alphanumeric + hyphens).
var sessionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

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
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type qrInitResponse struct {
	SessionID string `json:"sessionID"`
	DeepLink  string `json:"deepLink"`
}

type qrStatusResponse struct {
	Status       string `json:"status"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type magicLinkRequest struct {
	Token string `json:"token"`
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
	WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken})
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

// HandleQRRedirect handles GET /auth/qr/redirect?session=<id>.
func HandleQRRedirect(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" || !sessionIDPattern.MatchString(sessionID) {
		http.Error(w, "missing or invalid session param", http.StatusBadRequest)
		return
	}

	botUsername := strings.TrimSpace(strings.TrimPrefix(configs.TrackerConfig.Server.BotUsername, "@"))
	param := "login_" + sessionID
	tgLink := fmt.Sprintf("tg://resolve?domain=%s&start=%s", botUsername, param)
	httpLink := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, param)

	fmt.Println(httpLink, "<==>", tgLink)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, qrRedirectPage, tgLink, httpLink, httpLink) //nolint:gosec // sessionID validated against alphanumeric pattern; botUsername comes from config
}

const qrRedirectPage = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Opening Telegram...</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<script>
// Try tg:// first. If app doesn't open within 1.5s, redirect to https://t.me/.
// Use a flag to prevent both from firing.
var opened = false;
window.location.href = %q;
document.addEventListener("visibilitychange", function() {
  if (document.hidden) opened = true;
});
setTimeout(function(){ if (!opened) window.location.href = %q; }, 1500);
</script>
</head><body style="font-family:sans-serif;text-align:center;padding:40px">
<p>Opening Telegram...</p>
<p><a href=%q>Click here if Telegram didn't open</a></p>
</body></html>`

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
		resp.RefreshToken = pair.RefreshToken
	}
	WriteJSON(w, http.StatusOK, resp)
}

// HandleRefresh handles POST /auth/refresh.
func HandleRefresh(w http.ResponseWriter, r *http.Request) {
	// Try JSON body first, then cookie fallback
	var token string
	var req refreshRequest
	if err := ReadJSON(r, &req); err == nil && req.RefreshToken != "" {
		token = req.RefreshToken
	} else {
		var err error
		token, err = RefreshTokenFromCookie(r)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "missing_token", "refresh token required")
			return
		}
	}

	pair, err := all.GetServices().Auth.RefreshTokens(token)
	if err != nil {
		ClearRefreshCookie(w)
		WriteError(w, http.StatusUnauthorized, "refresh_failed", err.Error())
		return
	}

	SetRefreshCookie(w, pair.RefreshToken, refreshCookieMaxAge)
	WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken})
}

// HandleLogout handles POST /auth/logout.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	var token string
	var req refreshRequest
	if err := ReadJSON(r, &req); err == nil && req.RefreshToken != "" {
		token = req.RefreshToken
	} else {
		token, _ = RefreshTokenFromCookie(r)
	}

	if token != "" {
		_ = all.GetServices().Auth.Logout(token)
	}
	ClearRefreshCookie(w)
	WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// HandleVerifyMagicLink handles POST /auth/magic-link.
func HandleVerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req magicLinkRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if req.Token == "" {
		WriteError(w, http.StatusBadRequest, "bad_request", "token is required")
		return
	}

	pair, err := all.GetServices().Auth.VerifyMagicLink(req.Token)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "verify_failed", err.Error())
		return
	}

	SetRefreshCookie(w, pair.RefreshToken, refreshCookieMaxAge)
	WriteJSON(w, http.StatusOK, tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken})
}
