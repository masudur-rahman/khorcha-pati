package handlers

import (
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

// SharePhone handles /phone — asks for the user's number via a one-tap contact button.
func SharePhone(ctx telebot.Context) error {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
	btn := menu.Contact("📱 Share my phone number")
	menu.Reply(menu.Row(btn))

	return ctx.Send("Tap the button to share your number for dashboard login.", menu)
}

// HandleContactShare saves the user's phone number from a shared contact.
func HandleContactShare(ctx telebot.Context) error {
	removeKeyboard := &telebot.ReplyMarkup{RemoveKeyboard: true}

	contact := ctx.Message().Contact
	if contact == nil || contact.PhoneNumber == "" {
		return ctx.Send("No phone number received.", removeKeyboard)
	}
	// Only the sender's own contact card counts — forwarding someone else's
	// contact must not attach their number to this account.
	if contact.UserID != ctx.Sender().ID {
		return ctx.Send("Not your contact — share your own number via /phone.", removeKeyboard)
	}

	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err), removeKeyboard)
	}

	if err := all.GetServices().User.UpdateMobileNumber(user.ID, contact.PhoneNumber); err != nil {
		return ctx.Send(models.ErrCommonResponse(err), removeKeyboard)
	}

	return ctx.Send("✅ Number saved — you can use it to log in.", removeKeyboard)
}
