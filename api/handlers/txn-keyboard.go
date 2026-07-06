package handlers

import (
	"fmt"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

func generateAmountTypeInlineButton(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	amounts := []float64{50, 100, 500}
	inlineButtons := make([]telebot.InlineButton, 0, 3)
	for _, amount := range amounts {
		callbackOpts.Transaction.Amount = amount
		btn := generateInlineButton(callbackOpts, fmt.Sprintf("%v", amount))
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateSrcDstTypeInlineButton(ctx telebot.Context, callbackOpts CallbackOptions, src bool) ([]telebot.InlineButton, error) {
	svc := all.GetServices()
	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return nil, err
	}

	acs, err := svc.Wallet.ListWallets(user.ID)
	if err != nil {
		return nil, err
	}

	inlineButtons := make([]telebot.InlineButton, 0, len(acs))

	var srcOrDst *string
	if src {
		srcOrDst = &callbackOpts.Transaction.SrcID
	} else {
		srcOrDst = &callbackOpts.Transaction.DstID
	}

	for _, ac := range acs {
		*srcOrDst = ac.ShortName
		btn := generateInlineButton(callbackOpts, ac.Name)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateTransactionCategoryTypeInlineButton(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	svc := all.GetServices()
	cats, err := svc.Txn.ListTxnCategories()
	if err != nil {
		return nil, err
	}

	txnType := callbackOpts.Transaction.Type
	inlineButtons := make([]telebot.InlineButton, 0, len(cats))
	for _, cat := range cats {
		if txnType != "" && !models.ContainsType(models.CategoryTypes[cat.ID], txnType) {
			continue
		}
		callbackOpts.Transaction.CategoryID = cat.ID
		btn := generateInlineButton(callbackOpts, cat.Name)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateJustTxnCategoryTypeInlineButton(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	svc := all.GetServices()
	cats, err := svc.Txn.ListTxnCategories()
	if err != nil {
		return nil, err
	}

	inlineButtons := make([]telebot.InlineButton, 0, len(cats))
	for _, cat := range cats {
		callbackOpts.Category.CategoryID = cat.ID
		btn := generateInlineButton(callbackOpts, cat.Name)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateTransactionSubcategoryTypeInlineButton(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	svc := all.GetServices()
	subcats, err := svc.Txn.ListTxnSubcategories(callbackOpts.Transaction.CategoryID)
	if err != nil {
		return nil, err
	}

	txnType := callbackOpts.Transaction.Type
	inlineButtons := make([]telebot.InlineButton, 0, len(subcats))
	for _, subcat := range subcats {
		if txnType != "" && !models.ContainsType(models.SubcategoryTypes[subcat.ID], txnType) {
			continue
		}
		callbackOpts.Transaction.SubcategoryID = subcat.ID
		btn := generateInlineButton(callbackOpts, subcat.Name)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateJustTxnSubcategoryTypeInlineButton(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	svc := all.GetServices()
	subcats, err := svc.Txn.ListTxnSubcategories(callbackOpts.Category.CategoryID)
	if err != nil {
		return nil, err
	}

	inlineButtons := make([]telebot.InlineButton, 0, len(subcats))
	for _, subcat := range subcats {
		callbackOpts.Category.SubcategoryID = subcat.ID
		btn := generateInlineButton(callbackOpts, subcat.Name)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateTransactionUserTypeInlineButton(ctx telebot.Context, callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	svc := all.GetServices()
	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return nil, err
	}

	contact, err := svc.Contact.ListContacts(user.ID)
	if err != nil {
		return nil, err
	}

	inlineButtons := make([]telebot.InlineButton, 0, len(contact))
	for _, user := range contact {
		callbackOpts.Transaction.ContactName = user.NickName
		btn := generateInlineButton(callbackOpts, user.FullName)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}

func generateTransactionRemarksTypeInlineButton(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	btn := generateInlineButton(callbackOpts, "Done")
	inlineButtons := []telebot.InlineButton{btn}

	return inlineButtons, nil
}

func generateInlineButton[str ~string](obj any, btnText str) telebot.InlineButton {
	return telebot.InlineButton{
		Text: string(btnText),
		Data: cache.StoreData(obj),
	}
}

func generateInlineKeyboard(inlineButtons []telebot.InlineButton) [][]telebot.InlineButton {
	var keyboard [][]telebot.InlineButton
	var tmpInlnBtns []telebot.InlineButton
	for _, btn := range inlineButtons {
		tmpInlnBtns = append(tmpInlnBtns, btn)
		if len(tmpInlnBtns) == 3 {
			keyboard = append(keyboard, tmpInlnBtns)
			tmpInlnBtns = nil
		}
	}
	if len(tmpInlnBtns) > 0 {
		keyboard = append(keyboard, tmpInlnBtns)
	}

	return keyboard
}

func commonSendOptions(ctx telebot.Context, inlineButtons []telebot.InlineButton) *telebot.SendOptions {
	return &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyTo:   ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	}
}
