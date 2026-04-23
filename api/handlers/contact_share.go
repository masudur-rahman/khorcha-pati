package handlers

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

// HandleContactShare saves the user's phone number from a shared contact.
func HandleContactShare(ctx telebot.Context) error {
	contact := ctx.Message().Contact
	if contact == nil || contact.PhoneNumber == "" {
		return ctx.Send("No phone number received.")
	}

	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if err := all.GetServices().User.UpdateMobileNumber(user.ID, contact.PhoneNumber); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send("Phone number saved. You can now use it to log in to the web dashboard.")
}
