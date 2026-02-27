package handlers

import (
	"fmt"
	"log"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

type UserCallbackOptions struct {
	NickName string `json:"id"`
	FullName string `json:"name"`
	Email    string `json:"email"`
}

func handleUserCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	msg, err := ctx.Bot().Reply(ctx.Message(), `Reply to this Message with the following data

<nick name> <full name> <email(optional)>
i.e.: john "John Doe" john@doe.com
`, &telebot.SendOptions{
		ReplyTo: ctx.Message(),
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
		log.Println(err)
		return ctx.Send(err.Error())
	}

	return ctx.Send(fmt.Sprintf("✅ Contact *%v* added!", uop.FullName), telebot.ModeMarkdown)
}
