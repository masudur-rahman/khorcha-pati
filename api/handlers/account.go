package handlers

import (
	"fmt"
	"log"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

const (
	StepAccountType NextStep = "type"
	StepAccountInfo NextStep = "info"
)

type AccountCallbackOptions struct {
	NextStep  NextStep           `json:"nextStep"`
	Type      models.AccountType `json:"type"`
	ShortName string             `json:"shortName"`
	Name      string             `json:"name"`
}

func handleAccountCallback(ctx telebot.Context, callbackOptions CallbackOptions) error {
	// Skip Cash type Account entirely
	callbackOptions.Account = AccountCallbackOptions{
		NextStep: StepAccountInfo,
		Type:     models.BankAccount,
	}

	ac := callbackOptions.Account
	switch ac.NextStep {
	case StepAccountType, "":
		return sendAccountTypeQuery(ctx, callbackOptions)
	case StepAccountInfo:
		return sendAccountInfoQuery(ctx, callbackOptions)
	default:
		return ctx.Send("Invalid Step")
	}
}

func sendAccountTypeQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Account.NextStep = StepAccountInfo
	inlineButtons := generateAccountTypeInlineButton(callbackOpts)

	return ctx.Send("Select Type of Account", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func sendAccountInfoQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	msg, err := ctx.Bot().Reply(ctx.Message(), fmt.Sprintf(`Reply to this Message with the following data

<short name> <account name>
i.e.: %v
`, accountExample(callbackOpts.Account.Type)), &telebot.SendOptions{
		ReplyTo: ctx.Message(),
	})
	if err != nil {
		return err
	}

	callbackData[msg.ID] = callbackOpts
	return nil
}

func accountExample(typ models.AccountType) string {
	if typ == models.CashAccount {
		return "cash \"Cash in Hand\""
	}
	return "brac \"BRAC Bank\""
}

func processAccountCreation(ctx telebot.Context, aop AccountCallbackOptions) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	acc := &models.Account{
		UserID:    user.ID,
		Type:      aop.Type,
		ShortName: aop.ShortName,
		Name:      aop.Name,
	}
	if err := all.GetServices().Account.CreateAccount(acc); err != nil {
		log.Println(err)
		return ctx.Send(err.Error())
	}

	return ctx.Send(fmt.Sprintf("New Account [%v] Added !", acc.Name))
}

func generateAccountTypeInlineButton(callbackOpts CallbackOptions) []telebot.InlineButton {
	types := []models.AccountType{models.CashAccount, models.BankAccount}
	inlineButtons := make([]telebot.InlineButton, 0, 3)
	for _, typ := range types {
		callbackOpts.Account.Type = typ
		btn := generateInlineButton(callbackOpts, typ)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons
}
