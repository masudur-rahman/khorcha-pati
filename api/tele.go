package api

import (
	"log"
	"os"
	"time"

	"github.com/masudur-rahman/khorcha-pati/api/handlers"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

func TeleBotRoutes() (*telebot.Bot, error) {
	settings := telebot.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}

	bot.Use(rejectBots())
	bot.Use(rejectDisabledUsers())
	bot.Use(AutoKeyboardReset())

	bot.Handle("/start", handlers.StartTrackingExpenses)
	bot.Handle("/", handlers.StartTrackingExpenses)

	bot.Handle(telebot.OnCallback, handlers.Callback)
	bot.Handle(telebot.OnText, handlers.TransactionTextCallback)

	bot.Handle("/new", handlers.New)
	bot.Handle("/newtxn", handlers.NewTransaction)

	bot.Handle("/contacts", handlers.ListContacts)
	bot.Handle("/balance", handlers.ListWallets)

	bot.Handle("/list", handlers.ListTransactions)
	bot.Handle("/expense", handlers.ListExpenses)

	bot.Handle("/allsummary", handlers.TransactionSummaryCallback)
	bot.Handle("/summary", handlers.TransactionSummary)
	bot.Handle("/report", handlers.TransactionReportCallback)

	bot.Handle("/cat", handlers.TransactionCategoryCallback)
	bot.Handle("/budget", handlers.BudgetCommand)

	bot.Handle("/help", handlers.Help)
	bot.Handle("/dashboard", handlers.Dashboard)
	bot.Handle("/phone", handlers.SharePhone)

	bot.Handle("/sync", handlers.SyncSQLiteDatabase)
	bot.Handle("/undo", handlers.HandleUndo)

	bot.Handle(telebot.OnContact, handlers.HandleContactShare)

	bot.Handle("/admin", handlers.Admin, adminOnly())

	if err = setBotCommands(bot); err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	return bot, nil
}

// setBotCommands registers user-visible commands in the Telegram command menu.
func setBotCommands(bot *telebot.Bot) error {
	return bot.SetCommands([]telebot.Command{
		{Text: "new", Description: "Add new wallet or contact"},
		{Text: "newtxn", Description: "Add new transaction (interactive)"},
		{Text: "undo", Description: "Undo last transaction"},
		{Text: "balance", Description: "Show wallet balances"},
		{Text: "contacts", Description: "List contacts and balances"},
		{Text: "report", Description: "Generate PDF transaction report"},
		{Text: "summary", Description: "Monthly transaction summary"},
		{Text: "allsummary", Description: "Detailed summary by type/category"},
		{Text: "budget", Description: "View and manage budgets"},
		{Text: "dashboard", Description: "Open web dashboard"},
		{Text: "phone", Description: "Share phone number for dashboard login"},
		{Text: "list", Description: "List recent transactions"},
		{Text: "expense", Description: "List expenses"},
		{Text: "cat", Description: "Browse transaction categories"},
		{Text: "help", Description: "Show usage help"},
	})
}

func adminOnly() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
			if err != nil || !user.IsAdmin {
				if models.IsErrNotFound(err) {
					return ctx.Send("You are not registered.")
				}
				return ctx.Send("Invalid command.")
			}
			return next(ctx)
		}
	}
}

// rejectDisabledUsers blocks bot interaction for users whose IsActive=false.
// Unregistered senders pass through (registration flows must remain reachable).
func rejectDisabledUsers() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
			if err == nil && !user.IsActive {
				return ctx.Send("Account disabled. Contact admin.")
			}
			return next(ctx)
		}
	}
}

func rejectBots() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			//if ctx.Sender().Username != configs.TrackerConfig.Telegram.User {
			//	return ctx.Send(fmt.Sprintf("Ohho!!! Looks like you're not the admin of this bot.\n\nIf you wish to know how to use this bot, go to https://github.com/masudur-rahman/khorcha-pati ."))
			//}
			if ctx.Sender().IsBot {
				return ctx.Send("Bot not allowed")
			}

			return next(ctx)
		}
	}
}

// AutoKeyboardReset is a middleware that ensures any sticky ForceReply or ReplyKeyboardMarkup
// is removed when a new message is sent without an explicit markup.
func AutoKeyboardReset() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			return next(&keyboardResetContext{Context: ctx})
		}
	}
}

type keyboardResetContext struct {
	telebot.Context
}

func (c *keyboardResetContext) Send(what any, opts ...any) error {
	return c.Context.Send(what, c.ensureMarkup(opts...)...)
}

func (c *keyboardResetContext) Reply(what any, opts ...any) error {
	return c.Context.Reply(what, c.ensureMarkup(opts...)...)
}

func (c *keyboardResetContext) ensureMarkup(opts ...any) []any {
	hasMarkup := false
	for _, opt := range opts {
		switch v := opt.(type) {
		case *telebot.SendOptions:
			if v.ReplyMarkup != nil {
				hasMarkup = true
			} else {
				// Inject removal if not present
				v.ReplyMarkup = &telebot.ReplyMarkup{RemoveKeyboard: true}
				hasMarkup = true
			}
		case *telebot.ReplyMarkup:
			hasMarkup = true
		}
	}

	if !hasMarkup {
		opts = append(opts, &telebot.ReplyMarkup{RemoveKeyboard: true})
	}
	return opts
}
