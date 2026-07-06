package handlers

import (
	"fmt"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

const (
	qrApprovePrefix = "qr_approve_"
	qrDenyPrefix    = "qr_deny_"
)

// HandleQRLogin auto-confirms QR session and logs user into dashboard.
// Falls back to magic link if QR session expired.
func HandleQRLogin(sessionID string, ctx telebot.Context) error {
	svc := all.GetServices()
	if svc.Auth == nil {
		return ctx.Send("Web dashboard is not enabled.")
	}

	err := svc.Auth.ConfirmQRSession(sessionID, ctx.Sender().ID)
	if err == nil {
		return ctx.Send("✅ Dashboard login approved. Return to your browser — it will log you in automatically.")
	}

	// "already used" = duplicate /start trigger (tg:// + https://t.me/). Silently ignore.
	if strings.Contains(err.Error(), "already used") {
		return nil
	}

	// QR expired → fallback to magic link
	return sendDashboardLink(ctx)
}

// HandleQRCallback handles approve/deny callbacks for QR login.
func HandleQRCallback(ctx telebot.Context) error {
	data := ctx.Callback().Data
	svc := all.GetServices()

	switch {
	case strings.HasPrefix(data, qrApprovePrefix):
		sessionID := strings.TrimPrefix(data, qrApprovePrefix)
		if err := svc.Auth.ConfirmQRSession(sessionID, ctx.Sender().ID); err == nil {
			return ctx.Edit("✅ Dashboard login approved. Return to your browser.")
		}
		return sendDashboardLinkEdit(ctx)

	case strings.HasPrefix(data, qrDenyPrefix):
		sessionID := strings.TrimPrefix(data, qrDenyPrefix)
		_ = svc.Auth.DenyQRSession(sessionID)
		return ctx.Edit("🚫 Login denied.")

	default:
		return ctx.Send("⚠️ Unknown QR action.")
	}
}

// sendDashboardLink sends a fresh magic link as a new message.
func sendDashboardLink(ctx telebot.Context) error {
	url, err := buildDashboardURL(ctx)
	if err != nil {
		return ctx.Send(err.Error())
	}

	btn := &telebot.ReplyMarkup{}
	webBtn := btn.URL("Open Dashboard", url)
	btn.Inline(btn.Row(webBtn))

	return ctx.Send("✅ You're all set! Open your dashboard:", btn)
}

// sendDashboardLinkEdit edits existing message with a magic link.
func sendDashboardLinkEdit(ctx telebot.Context) error {
	url, err := buildDashboardURL(ctx)
	if err != nil {
		return ctx.Edit(err.Error())
	}

	btn := &telebot.ReplyMarkup{}
	webBtn := btn.URL("Open Dashboard", url)
	btn.Inline(btn.Row(webBtn))

	return ctx.Edit("⚠️ QR session expired. Open your dashboard:", btn)
}

// buildDashboardURL generates a one-time magic link URL for dashboard login.
func buildDashboardURL(ctx telebot.Context) (string, error) {
	svc := all.GetServices()

	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return "", fmt.Errorf("❌ profile not found, please /start the bot first")
	}

	token, err := svc.Auth.CreateMagicLink(user.ID)
	if err != nil {
		return "", fmt.Errorf("❌ failed to generate login link, please try again")
	}

	cfg := configs.TrackerConfig.Server
	targetURL := cfg.DashboardURL
	if targetURL == "" {
		targetURL = cfg.BaseURL
	}

	return fmt.Sprintf("%s/login?token=%s", strings.TrimRight(targetURL, "/"), token), nil
}
