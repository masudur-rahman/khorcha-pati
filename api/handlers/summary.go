package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
	pkgtg "github.com/masudur-rahman/expense-tracker-bot/pkg/telegram"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"gopkg.in/telebot.v3"
)

type SummaryGroupBy string

type SummaryDuration string

const (
	StepGroupBy  NextStep = "group-by"
	StepDuration NextStep = "duration"

	GroupByTxnType        SummaryGroupBy = "Transaction Type"
	GroupByTxnCategory    SummaryGroupBy = "Category"
	GroupByTxnSubCategory SummaryGroupBy = "Subcategory"

	DurationOneWeek   SummaryDuration = "1 Week"
	DurationThisMonth SummaryDuration = "This Month"
	DurationOneMonth  SummaryDuration = "One Month"
	DurationHalfYear  SummaryDuration = "Last 6 Months"
	DurationThisYear  SummaryDuration = "This Year"
	DurationOneYear   SummaryDuration = "One Year"
)

type SummaryCallbackOptions struct {
	NextStep NextStep        `json:"nextStep"`
	GroupBy  SummaryGroupBy  `json:"groupBy,omitempty"`
	Duration SummaryDuration `json:"duration,omitempty"`
}

func TransactionSummary(ctx telebot.Context) error {
	user, err := all.GetServices().User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	txns, err := all.GetServices().Txn.ListTransactionsByTime(user.ID, "", pkg.StartOfMonth().Unix(), time.Now().Unix())
	if err != nil {
		return err
	}

	summary := generateSummary(txns)
	formatted := pkgtg.FormatSummary(summary, "Transaction Summary (This Month)")
	for _, chunk := range pkgtg.SplitMessage(formatted) {
		if err := ctx.Send(chunk, telebot.ModeMarkdown); err != nil {
			return err
		}
	}
	return nil
}

func generateSummary(txns []models.Transaction) gqtypes.SummaryGroups {
	summary := gqtypes.SummaryGroups{
		Type: map[string]gqtypes.FieldCost{},
	}

	for _, txn := range txns {
		fc := summary.Type[string(txn.Type)]
		fc.Amount += txn.Amount
		summary.Type[string(txn.Type)] = fc
	}

	return summary
}

func TransactionSummaryCallback(ctx telebot.Context) error {
	callbackOpts := CallbackOptions{
		Type: SummaryTypeCallback,
		Summary: SummaryCallbackOptions{
			NextStep: StepGroupBy,
		},
	}

	groupBies := []SummaryGroupBy{GroupByTxnType, GroupByTxnCategory, GroupByTxnSubCategory}
	inlineButtons := make([]telebot.InlineButton, 0, 3)
	for _, groupBy := range groupBies {
		callbackOpts.Summary.GroupBy = groupBy
		btn := generateInlineButton(callbackOpts, string(groupBy))
		inlineButtons = append(inlineButtons, btn)
	}

	return ctx.Send("Select Summarization Group", &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
		},
	})
}

func handleSummaryCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	summary := callbackOpts.Summary
	switch summary.NextStep {
	case StepGroupBy:
		return sendSummaryDurationQuery(ctx, callbackOpts)
	case StepDuration:
		sg, err := processSummary(ctx, summary)
		if err != nil {
			return ctx.Send(models.ErrCommonResponse(err))
		}

		title := fmt.Sprintf("Summary: %s (by %s)", summary.Duration, summary.GroupBy)
		formatted := pkgtg.FormatSummary(sg, title)

		return ctx.Send(formatted, telebot.ModeMarkdown)
	default:
		return ctx.Send("⚠️ Invalid step.")
	}
}

func processSummary(ctx telebot.Context, smop SummaryCallbackOptions) (gqtypes.SummaryGroups, error) {
	now, startTime := time.Now(), calculateStartTime(smop.Duration)

	svc := all.GetServices()
	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return gqtypes.SummaryGroups{}, err
	}

	txns, err := svc.Txn.ListTransactionsByTime(user.ID, "", startTime.Unix(), now.Unix())
	if err != nil {
		return gqtypes.SummaryGroups{}, err
	}

	summary := gqtypes.SummaryGroups{
		Type:        map[string]gqtypes.FieldCost{},
		Category:    map[string]gqtypes.FieldCost{},
		Subcategory: map[string]gqtypes.FieldCost{},
	}
	for _, txn := range txns {
		if smop.GroupBy == GroupByTxnType {
			fc := summary.Type[string(txn.Type)]
			fc.Amount += txn.Amount
			summary.Type[string(txn.Type)] = fc
		} else if smop.GroupBy == GroupByTxnSubCategory {
			fc := summary.Subcategory[txn.SubcategoryID]
			fc.Amount += txn.Amount
			summary.Subcategory[txn.SubcategoryID] = fc
		} else if smop.GroupBy == GroupByTxnCategory {
			cat := strings.Split(txn.SubcategoryID, "-")[0]
			fc := summary.Category[cat]
			fc.Amount += txn.Amount
			summary.Category[cat] = fc
		}
	}

	for k, fc := range summary.Category {
		fc.Name, err = svc.Txn.GetTxnCategoryName(k)
		if err != nil {
			return gqtypes.SummaryGroups{}, err
		}

		summary.Category[k] = fc
	}

	for k, fc := range summary.Subcategory {
		fc.Name, err = svc.Txn.GetTxnSubcategoryName(k)
		if err != nil {
			return gqtypes.SummaryGroups{}, err
		}

		summary.Subcategory[k] = fc
	}

	return summary, nil
}

func sendSummaryDurationQuery(ctx telebot.Context, callbackOpts CallbackOptions) error {
	callbackOpts.Summary.NextStep = StepDuration
	inlineButtons := generateSummaryDurationInlineButton(callbackOpts)

	return ctx.Send("Select Duration for Summary", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func generateSummaryDurationInlineButton(callbackOpts CallbackOptions) []telebot.InlineButton {
	durations := []SummaryDuration{DurationOneWeek, DurationThisMonth, DurationOneMonth, DurationHalfYear, DurationThisYear, DurationOneYear}
	inlineButtons := make([]telebot.InlineButton, 0, 3)
	for _, duration := range durations {
		callbackOpts.Summary.Duration = duration
		btn := generateInlineButton(callbackOpts, duration)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons
}
