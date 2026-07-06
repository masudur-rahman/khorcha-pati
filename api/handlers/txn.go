package handlers

import (
	"fmt"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/pkg"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

const (
	StepCategoryID    NextStep = "cat-id"
	StepSubcategoryID NextStep = "subcat-id"
)

type TxnCategoryCallbackOptions struct {
	NextStep      NextStep `json:"nextStep"`
	CategoryID    string   `json:"categoryID"`
	SubcategoryID string   `json:"subcategoryID"`
}

func handleTransactionWithFlagsCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	msg, err := ctx.Bot().Reply(ctx.Message(), `Reply to this Message with the following data


<amount> -t=<type> -s=<subcat> -f=<src> -d=<dst> -u=<user> -r=<remarks>
i.e.: 6666 -t=Expense -s=food-rest -f=cash -r="Coffee with no one"
`, &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			ForceReply: true,
		},
	})
	if err != nil {
		return err
	}

	callbackData[msg.ID] = callbackOpts
	return nil
}

func TransactionCategoryCallback(ctx telebot.Context) error {
	callbackOpts := CallbackOptions{
		Type: TxnCategoryTypeCallback,
		Category: TxnCategoryCallbackOptions{
			NextStep: StepCategoryID,
		},
	}
	return sendJustTxnCategoryQuery(ctx, callbackOpts)
}

func handleTransactionCategoryCallback(ctx telebot.Context, callbackOptions CallbackOptions) error {
	cat := callbackOptions.Category
	switch cat.NextStep {
	case StepCategoryID:
		return sendJustTxnSubcategoryQuery(ctx, callbackOptions)
	case StepSubcategoryID:
		return sendTransactionCategoryInformation(ctx, callbackOptions.Category)
	default:
		return ctx.Send("⚠️ Invalid step.")
	}
}

func sendJustTxnCategoryQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	inlineButtons, err := generateJustTxnCategoryTypeInlineButton(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send("Select Transaction category!", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func sendJustTxnSubcategoryQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Category.NextStep = StepSubcategoryID
	inlineButtons, err := generateJustTxnSubcategoryTypeInlineButton(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send("Select Transaction subcategory!", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func sendTransactionCategoryInformation(ctx telebot.Context, cop TxnCategoryCallbackOptions) error {
	txn := all.GetServices().Txn
	cat, err := txn.GetTxnCategoryName(cop.CategoryID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	subcat, err := txn.GetTxnSubcategoryName(cop.SubcategoryID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	msg := fmt.Sprintf("🏷 *Category Info*\n%s\n📂 *Category:* %s (`%s`)\n📁 *Subcategory:* %s (`%s`)",
		models.Separator, cat, cop.CategoryID, subcat, cop.SubcategoryID)
	return ctx.Send(msg, telebot.ModeMarkdown)
}

func ListTransactionSubcategories(ctx telebot.Context) error {
	cat := pkg.SplitString(ctx.Text(), ' ')
	if len(cat) != 2 {
		return ctx.Send("⚠️ Usage: /subcategories <category-id>")
	}

	subcats, err := all.GetServices().Txn.ListTxnSubcategories(cat[1])
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send("Choose one: ", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: func() [][]telebot.InlineButton {
				var keyboard [][]telebot.InlineButton
				var inlineBtn []telebot.InlineButton
				for _, subcat := range subcats {
					inlineBtn = append(inlineBtn, telebot.InlineButton{Text: subcat.Name, Data: subcat.ID})
					if len(inlineBtn) == 3 {
						keyboard = append(keyboard, inlineBtn)
						inlineBtn = nil
					}
				}
				if len(inlineBtn) > 0 {
					keyboard = append(keyboard, inlineBtn)
				}
				return keyboard
			}(),
			ResizeKeyboard: true,
		},
	})
}
