package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

// Formatting helpers are in budget-format.go

const (
	StepBudgetCategory NextStep = "budget-cat"
	StepBudgetAmount   NextStep = "budget-amt"
)

// BudgetCallbackOptions holds state for the budget inline-button flow.
type BudgetCallbackOptions struct {
	NextStep   NextStep `json:"ns"`
	Action     string   `json:"act"`
	CategoryID string   `json:"cid"`
}

// BudgetCommand handles /budget — shows current budget statuses.
func BudgetCommand(ctx telebot.Context) error {
	args := strings.Fields(ctx.Text())
	if len(args) >= 2 {
		switch args[1] {
		case "set":
			return budgetSetStart(ctx)
		case "delete":
			return budgetDeleteStart(ctx)
		}
	}

	return showBudgetStatuses(ctx)
}

// showBudgetStatuses displays all budgets with current month spending.
func showBudgetStatuses(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	statuses, err := all.GetServices().Budget.ListBudgetStatuses(user.ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if len(statuses) == 0 {
		return ctx.Send("No budgets set yet.\nUse `/budget set` to add one.", telebot.ModeMarkdown)
	}

	return ctx.Send(formatBudgetStatuses(statuses), telebot.ModeMarkdown)
}

// budgetSetStart initiates the "set budget" flow — shows category buttons.
func budgetSetStart(ctx telebot.Context) error {
	callbackOpts := CallbackOptions{
		Type: BudgetTypeCallback,
		Budget: BudgetCallbackOptions{
			NextStep: StepBudgetCategory,
			Action:   "set",
		},
	}

	inlineButtons, err := generateBudgetCategoryButtons(callbackOpts)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	return ctx.Send("Select a category to set budget:", commonSendOptions(ctx, inlineButtons))
}

// budgetDeleteStart shows existing budgets as buttons for deletion.
func budgetDeleteStart(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	statuses, err := all.GetServices().Budget.ListBudgetStatuses(user.ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if len(statuses) == 0 {
		return ctx.Send("No budgets to delete. Use /budget set to add one.")
	}

	callbackOpts := CallbackOptions{
		Type: BudgetTypeCallback,
		Budget: BudgetCallbackOptions{
			NextStep: StepBudgetCategory,
			Action:   "delete",
		},
	}

	inlineButtons := make([]telebot.InlineButton, 0, len(statuses))
	for _, s := range statuses {
		callbackOpts.Budget.CategoryID = s.CategoryID
		label := fmt.Sprintf("%s %s (৳%s)", budgetCategoryIcon(s.CategoryID), s.CategoryName, FormatBDT(s.Amount))
		btn := generateInlineButton(callbackOpts, label)
		inlineButtons = append(inlineButtons, btn)
	}

	return ctx.Send("Select budget to delete:", commonSendOptions(ctx, inlineButtons))
}

// handleBudgetCallback routes budget inline button callbacks.
func handleBudgetCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	switch callbackOpts.Budget.Action {
	case "set":
		return handleBudgetSetCallback(ctx, callbackOpts)
	case "delete":
		return handleBudgetDeleteCallback(ctx, callbackOpts)
	default:
		return ctx.Send("⚠️ Invalid budget action.")
	}
}

// handleBudgetSetCallback handles the set-budget flow steps.
func handleBudgetSetCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	switch callbackOpts.Budget.NextStep {
	case StepBudgetCategory:
		return sendBudgetAmountQuery(ctx, callbackOpts)
	case StepBudgetAmount:
		return processBudgetSet(ctx, callbackOpts)
	default:
		return ctx.Send("⚠️ Invalid budget step.")
	}
}

// sendBudgetAmountQuery asks user to reply with a budget amount.
func sendBudgetAmountQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Budget.NextStep = StepBudgetAmount

	name := resolveBudgetCategoryName(callbackOpts.Budget.CategoryID)
	prompt := fmt.Sprintf("Enter monthly budget amount for *%s*:", name)

	msg, err := ctx.Bot().Reply(ctx.Message(), prompt, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
		ReplyTo:   ctx.Message(),
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

// processBudgetSet saves the budget.
func processBudgetSet(ctx telebot.Context, callbackOpts CallbackOptions) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(ctx.Text()), 64)
	if err != nil || amount <= 0 {
		return ctx.Reply("⚠️ Please enter a valid positive number.")
	}

	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if err = all.GetServices().Budget.SetBudget(user.ID, callbackOpts.Budget.CategoryID, amount, 0); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	name := resolveBudgetCategoryName(callbackOpts.Budget.CategoryID)
	return ctx.Send(fmt.Sprintf("✅ Budget set: *%s* ৳%s/month (alert at 80%%)", name, FormatBDT(amount)), telebot.ModeMarkdown)
}

// handleBudgetDeleteCallback handles budget deletion on category selection.
func handleBudgetDeleteCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	if err = all.GetServices().Budget.DeleteBudget(user.ID, callbackOpts.Budget.CategoryID); err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	name := resolveBudgetCategoryName(callbackOpts.Budget.CategoryID)
	return ctx.Send(fmt.Sprintf("✅ Budget removed: *%s*", name), telebot.ModeMarkdown)
}

// handleBudgetTypeTextCallback handles text replies in the budget flow.
func handleBudgetTypeTextCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	switch callbackOpts.Budget.NextStep {
	case StepBudgetAmount:
		return processBudgetSet(ctx, callbackOpts)
	default:
		return ctx.Reply("⚠️ Invalid budget step.")
	}
}

// generateBudgetCategoryButtons builds [Overall] + 13 category buttons.
func generateBudgetCategoryButtons(callbackOpts CallbackOptions) ([]telebot.InlineButton, error) {
	cats, err := all.GetServices().Txn.ListTxnCategories()
	if err != nil {
		return nil, err
	}

	inlineButtons := make([]telebot.InlineButton, 0, len(cats)+1)

	// Overall budget button first
	callbackOpts.Budget.CategoryID = ""
	btn := generateInlineButton(callbackOpts, "💰 Overall")
	inlineButtons = append(inlineButtons, btn)

	for _, cat := range cats {
		callbackOpts.Budget.CategoryID = cat.ID
		icon := budgetCategoryIcon(cat.ID)
		btn = generateInlineButton(callbackOpts, fmt.Sprintf("%s %s", icon, cat.Name))
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons, nil
}
