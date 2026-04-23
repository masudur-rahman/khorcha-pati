package handlers

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

// HandleUndo reverses the most recent active transaction for the calling user.
func HandleUndo(c telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(c.Sender().ID)
	if err != nil {
		return c.Send(models.ErrCommonResponse(err))
	}
	txn, err := all.GetServices().Txn.Undo(user.ID)
	if err != nil {
		return c.Send("⚠️ " + models.ErrCommonResponse(err))
	}
	return c.Send("✅ *Transaction Undone*\n\n"+txn.Summary(), telebot.ModeMarkdown)
}
