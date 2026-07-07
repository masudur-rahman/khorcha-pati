package handlers

import (
	"fmt"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

// Dashboard sends an inline button linking to the web dashboard.
func Dashboard(ctx telebot.Context) error {
	svc := all.GetServices()
	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send("Please /start the bot first to create your profile.")
	}

	if user.Username == "" && user.MobileNumber == "" {
		return ctx.Send("Set a Telegram username or share your phone number first so the dashboard can identify you.")
	}

	cfg := configs.TrackerConfig.Server
	if !cfg.DashboardEnabled {
		return ctx.Send("Web dashboard is not enabled.")
	}

	token, err := svc.Auth.CreateMagicLink(user.ID)
	if err != nil {
		return ctx.Send("Failed to generate magic link. Please try again later.")
	}

	targetURL := cfg.DashboardURL
	if targetURL == "" {
		targetURL = cfg.BaseURL
	}
	baseURL := strings.TrimRight(targetURL, "/")
	dashboardURL := fmt.Sprintf("%s/login?token=%s", baseURL, token)

	btn := &telebot.ReplyMarkup{}
	webBtn := btn.URL("Open Dashboard", dashboardURL)
	btn.Inline(btn.Row(webBtn))

	// URL shown as inline code (not a bare link) so it stays visible but non-tappable —
	// users tap the button, which carries the one-time login token.
	msg := fmt.Sprintf("🖥 *Your dashboard*\n`%s`\n\nTap the button below to open it — you'll be signed in automatically.", baseURL)
	return ctx.Send(msg, btn, telebot.ModeMarkdown)
}
