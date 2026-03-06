package handlers

import (
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

// HandleUndo reverses the most recent active transaction for the calling user.
func HandleUndo(c telebot.Context) error {
	txn, err := all.GetServices().Txn.Undo(int64(c.Sender().ID))
	if err != nil {
		return c.Send("⚠️ " + err.Error())
	}
	return c.Send("✅ *Transaction Undone*\n\n"+txn.Summary(), telebot.ModeMarkdown)
}
