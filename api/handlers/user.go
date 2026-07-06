package handlers

import (
	"fmt"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

type UserCallbackOptions struct {
	NickName string `json:"id"`
	FullName string `json:"name"`
	Email    string `json:"email"`
}

func handleUserCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	prompt := `Reply to this Message with the following data

<nick name> <full name> <email(optional)>
i.e.: john John Doe john@doe.com
`
	msg, err := ctx.Bot().Reply(ctx.Message(), prompt, &telebot.SendOptions{
		ReplyTo:   ctx.Message(),
		ParseMode: telebot.ModeMarkdown,
		ReplyMarkup: &telebot.ReplyMarkup{
			ForceReply:  true,
			Placeholder: "john John Doe",
		},
	})
	if err != nil {
		return err
	}

	callbackData[msg.ID] = callbackOpts
	return nil
}

func processUserCreation(ctx telebot.Context, uop UserCallbackOptions) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if err := all.GetServices().Contact.CreateContact(&models.Contacts{
		UserID:   user.ID,
		NickName: uop.NickName,
		FullName: uop.FullName,
		Email:    uop.Email,
	}); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("✅ Contact *%v* added!", uop.FullName), telebot.ModeMarkdown)
}
