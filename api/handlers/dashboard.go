package handlers

import (
	"fmt"
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

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

	cfg := configs.TrackerConfig.WebDashboard
	if !cfg.Enabled {
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
	dashboardURL := fmt.Sprintf("%s/login?token=%s", strings.TrimRight(targetURL, "/"), token)

	btn := &telebot.ReplyMarkup{}
	webBtn := btn.URL("Open Dashboard", dashboardURL)
	btn.Inline(btn.Row(webBtn))

	return ctx.Send("Open your expense dashboard (one-time login link):", btn)
}
