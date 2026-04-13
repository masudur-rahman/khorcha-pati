package api

import (
	"github.com/masudur-rahman/expense-tracker-bot/modules/auth"

	"gopkg.in/telebot.v3"
)

type botMessenger struct {
	bot *telebot.Bot
}

// NewBotMessenger wraps a telebot.Bot as an auth.Messenger.
func NewBotMessenger(bot *telebot.Bot) auth.Messenger {
	return &botMessenger{bot: bot}
}

func (m *botMessenger) SendMessage(telegramID int64, text string) error {
	_, err := m.bot.Send(telebot.ChatID(telegramID), text)
	return err
}
