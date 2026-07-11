package handlers

import (
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/pkg"
	"github.com/masudur-rahman/khorcha-pati/services/all"

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
	loc := pkg.LoadTimezone(user.Timezone)
	return c.Send("✅ *Transaction Undone*\n\n"+txn.Summary(loc), telebot.ModeMarkdown)
}
