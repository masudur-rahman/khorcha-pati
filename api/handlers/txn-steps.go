package handlers

import (
	"fmt"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"gopkg.in/telebot.v3"
)

func loanOrBorrowTypeTransaction(callbackOpts CallbackOptions) bool {
	return callbackOpts.Transaction.SubcategoryID == models.LoanRepaymentSubID ||
		callbackOpts.Transaction.SubcategoryID == models.BorrowSubID ||
		callbackOpts.Transaction.SubcategoryID == models.LoanReceivedSubID ||
		callbackOpts.Transaction.SubcategoryID == models.LendSubID ||
		callbackOpts.Transaction.SubcategoryID == models.LendRecoverySubID ||
		callbackOpts.Transaction.SubcategoryID == models.BorrowReturnSubID
}

func sendTransactionAmountTypeQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepAmount
	inlineButtons, err := generateAmountTypeInlineButton(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	msg, err := ctx.Bot().Reply(ctx.Message(),
		fmt.Sprintf("%vSelect an amount or Reply with an amount to this Message", callbackOpts.LastSelectedValue),
		commonSendOptions(ctx, inlineButtons),
	)
	if err != nil {
		return err
	}

	callbackData[msg.ID] = callbackOpts
	return nil
}

func sendTransactionSrcTypeQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepSrcID
	inlineButtons, err := generateSrcDstTypeInlineButton(ctx, callbackOpts, true)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("%vSelect Source Wallet:", callbackOpts.LastSelectedValue), commonSendOptions(ctx, inlineButtons))
}

func sendTransactionDstTypeQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepDstID
	inlineButtons, err := generateSrcDstTypeInlineButton(ctx, callbackOpts, false)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("%vSelect Destination Wallet:", callbackOpts.LastSelectedValue), commonSendOptions(ctx, inlineButtons))
}

func sendTransactionCategoryQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepCategory
	inlineButtons, err := generateTransactionCategoryTypeInlineButton(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("%vSelect Transaction category:", callbackOpts.LastSelectedValue), commonSendOptions(ctx, inlineButtons))
}

func sendTransactionSubcategoryQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepSubcategory
	inlineButtons, err := generateTransactionSubcategoryTypeInlineButton(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("%vSelect Transaction subcategory:", callbackOpts.LastSelectedValue), commonSendOptions(ctx, inlineButtons))
}

func sendTransactionUserQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepUser
	inlineButtons, err := generateTransactionUserTypeInlineButton(ctx, callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send(fmt.Sprintf("%vSelect the contact for this transaction:", callbackOpts.LastSelectedValue), commonSendOptions(ctx, inlineButtons))
}

func sendTransactionRemarksQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Transaction.NextStep = StepRemarks
	inlineButtons, err := generateTransactionRemarksTypeInlineButton(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	msg, err := ctx.Bot().Reply(ctx.Message(), fmt.Sprintf("%vComplete the transaction by Pressing Done or Reply with Remarks to this message:", callbackOpts.LastSelectedValue),
		commonSendOptions(ctx, inlineButtons))
	if err != nil {
		return err
	}

	callbackData[msg.ID] = callbackOpts
	return nil
}

func processTransaction(ctx telebot.Context, txn TransactionCallbackOptions) (models.Transaction, error) {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return models.Transaction{}, err
	}

	params := models.Transaction{
		UserID:        user.ID,
		Amount:        txn.Amount,
		SubcategoryID: txn.SubcategoryID,
		Type:          txn.Type,
		SrcID:         txn.SrcID,
		DstID:         txn.DstID,
		ContactName:   txn.ContactName,
		Remarks:       txn.Remarks,
	}
	err = all.GetServices().Txn.AddTransaction(params)
	return params, err
}
