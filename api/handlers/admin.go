package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

// Admin handles /admin commands with subcommands: stats, users, user <username>.
func Admin(ctx telebot.Context) error {
	payload := strings.TrimSpace(ctx.Message().Payload)
	parts := strings.Fields(payload)

	subcmd := "stats"
	if len(parts) > 0 {
		subcmd = parts[0]
	}

	switch subcmd {
	case "stats":
		return adminStats(ctx)
	case "users":
		return adminUsers(ctx)
	case "user":
		if len(parts) < 2 {
			return ctx.Send("Usage: `/admin user <username>`", telebot.ModeMarkdown)
		}
		return adminUserDetail(ctx, parts[1])
	case "disable":
		if len(parts) < 2 {
			return ctx.Send("Usage: `/admin disable <username>`", telebot.ModeMarkdown)
		}
		return adminSetActive(ctx, parts[1], false)
	case "enable":
		if len(parts) < 2 {
			return ctx.Send("Usage: `/admin enable <username>`", telebot.ModeMarkdown)
		}
		return adminSetActive(ctx, parts[1], true)
	default:
		return ctx.Send("Unknown subcommand. Available: `stats`, `users`, `user <username>`, `disable <username>`, `enable <username>`", telebot.ModeMarkdown)
	}
}

func adminSetActive(ctx telebot.Context, username string, active bool) error {
	svc := all.GetServices()
	user, err := svc.User.GetUserByUsername(username)
	if err != nil {
		return ctx.Send(fmt.Sprintf("User `%s` not found.", username), telebot.ModeMarkdown)
	}
	if user.TelegramID == ctx.Sender().ID {
		return ctx.Send("You cannot change your own active status.")
	}
	if err := svc.User.SetActive(user.ID, active); err != nil {
		return ctx.Send(fmt.Sprintf("Failed to update user: %v", err))
	}
	state := "disabled"
	if active {
		state = "enabled"
	}
	return ctx.Send(fmt.Sprintf("User `@%s` %s.", username, state), telebot.ModeMarkdown)
}

func adminStats(ctx telebot.Context) error {
	svc := all.GetServices()
	users, err := svc.User.ListUsers()
	if err != nil {
		return ctx.Send(fmt.Sprintf("Failed to fetch users: %v", err))
	}

	txnCount, walletCount := countAllResources()

	dbType := string(configs.TrackerConfig.Database.Type)
	if dbType == "" {
		dbType = "sqlite"
	}

	msg := fmt.Sprintf(
		"*Admin Stats*\n\n"+
			"Users: `%d`\n"+
			"Transactions: `%d`\n"+
			"Wallets: `%d`\n"+
			"Database: `%s`",
		len(users), txnCount, walletCount, dbType,
	)
	return ctx.Send(msg, telebot.ModeMarkdown)
}

func adminUsers(ctx telebot.Context) error {
	users, err := all.GetServices().User.ListUsers()
	if err != nil {
		return ctx.Send(fmt.Sprintf("Failed to fetch users: %v", err))
	}

	if len(users) == 0 {
		return ctx.Send("No registered users.")
	}

	var sb strings.Builder
	sb.WriteString("*Registered Users*\n\n")
	for i, u := range users {
		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		admin := ""
		if u.IsAdmin {
			admin = " (admin)"
		}
		sb.WriteString(fmt.Sprintf("%d. `@%s` — %s%s\n", i+1, u.Username, name, admin))
	}
	return ctx.Send(sb.String(), telebot.ModeMarkdown)
}

func adminUserDetail(ctx telebot.Context, username string) error {
	svc := all.GetServices()
	user, err := svc.User.GetUserByUsername(username)
	if err != nil {
		return ctx.Send(fmt.Sprintf("User `%s` not found.", username), telebot.ModeMarkdown)
	}

	wallets, _ := svc.Wallet.ListWallets(user.ID)
	txns, _ := svc.Txn.ListTransactions(user.ID)

	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	msg := fmt.Sprintf(
		"*User Detail*\n\n"+
			"Username: `@%s`\n"+
			"Name: %s\n"+
			"Telegram ID: `%d`\n"+
			"Wallets: `%d`\n"+
			"Transactions: `%d`\n"+
			"Admin: `%v`",
		user.Username, name, user.TelegramID,
		len(wallets), len(txns), user.IsAdmin,
	)
	return ctx.Send(msg, telebot.ModeMarkdown)
}

func countAllResources() (txnCount, walletCount int64) {
	db := configs.GetUnitOfWork().SQL
	bgCtx := context.Background()

	var txns []models.Transaction
	if err := db.Table(models.Transaction{}.TableName()).FindMany(bgCtx, &txns); err == nil {
		txnCount = int64(len(txns))
	}

	var ws []models.Wallet
	if err := db.Table(models.Wallet{}.TableName()).FindMany(bgCtx, &ws); err == nil {
		walletCount = int64(len(ws))
	}

	return txnCount, walletCount
}
