package handlers

import (
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

const (
	qrApprovePrefix = "qr_approve_"
	qrDenyPrefix    = "qr_deny_"
)

// HandleQRLogin sends an approval prompt for a QR login session.
func HandleQRLogin(sessionID string, ctx telebot.Context) error {
	authSvc := all.GetServices().Auth
	if authSvc == nil {
		return ctx.Send("Web dashboard is not enabled.")
	}

	keyboard := [][]telebot.InlineButton{
		{
			{Text: "✅ Approve", Data: qrApprovePrefix + sessionID},
			{Text: "❌ Deny", Data: qrDenyPrefix + sessionID},
		},
	}

	return ctx.Send("🔐 *Dashboard login request*\n\nSomeone is trying to log in to your Expense Tracker dashboard. Was this you?",
		telebot.ModeMarkdown, &telebot.ReplyMarkup{InlineKeyboard: keyboard})
}

// HandleQRCallback handles approve/deny callbacks for QR login.
func HandleQRCallback(ctx telebot.Context) error {
	data := ctx.Callback().Data
	authSvc := all.GetServices().Auth

	switch {
	case strings.HasPrefix(data, qrApprovePrefix):
		sessionID := strings.TrimPrefix(data, qrApprovePrefix)
		if err := authSvc.ConfirmQRSession(sessionID, ctx.Sender().ID); err != nil {
			return ctx.Edit("❌ Login failed: " + err.Error())
		}
		return ctx.Edit("✅ Dashboard login approved. You can return to your browser.")

	case strings.HasPrefix(data, qrDenyPrefix):
		sessionID := strings.TrimPrefix(data, qrDenyPrefix)
		if err := authSvc.DenyQRSession(sessionID); err != nil {
			return ctx.Edit("❌ Failed to deny login: " + err.Error())
		}
		return ctx.Edit("🚫 Login denied.")

	default:
		return ctx.Send("⚠️ Unknown QR action.")
	}
}
