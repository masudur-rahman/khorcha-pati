package handlers

import (
	"fmt"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"
)

// formatBudgetStatuses builds the status display message.
func formatBudgetStatuses(statuses []models.BudgetStatus) string {
	var sb strings.Builder
	sb.WriteString("*Monthly Budgets*\n")
	sb.WriteString(models.Separator + "\n")

	for _, s := range statuses {
		icon := budgetCategoryIcon(s.CategoryID)
		warning := ""
		if s.Percent >= 100 {
			warning = " *EXCEEDED*"
		} else if s.Percent >= float64(s.AlertAt) {
			warning = " ⚠️"
		}
		sb.WriteString(fmt.Sprintf(
			"%s %s: ৳%s / ৳%s (%.0f%%)%s\n",
			icon, s.CategoryName,
			FormatBDT(s.Spent), FormatBDT(s.Amount),
			s.Percent, warning,
		))
	}

	sb.WriteString(models.Separator + "\n")
	sb.WriteString("_Use /budget set to add or update budgets_")
	return sb.String()
}

// budgetCategoryIcon returns an emoji for a category ID.
func budgetCategoryIcon(categoryID string) string {
	icons := map[string]string{
		"":     "💰",
		"food": "🍔", "trans": "🚗", "shop": "🛒",
		"fin": "🏦", "house": "🏠", "health": "💊",
		"pc": "💇", "fam": "👨‍👩‍👧", "edu": "📚",
		"ent": "🎬", "trv": "✈️", "fest": "🎉",
		"misc": "📦",
	}
	if icon, ok := icons[categoryID]; ok {
		return icon
	}
	return "📌"
}

// resolveBudgetCategoryName returns display name for a category ID.
func resolveBudgetCategoryName(categoryID string) string {
	if categoryID == "" {
		return "Overall"
	}
	name, err := all.GetServices().Txn.GetTxnCategoryName(categoryID)
	if err != nil {
		return categoryID
	}
	return name
}

// FormatBudgetAlerts returns alert text for a transaction, or empty string.
func FormatBudgetAlerts(userID int64, txnType models.TransactionType, subcategoryID string) string {
	if txnType != models.ExpenseTransaction {
		return ""
	}

	alerts, err := all.GetServices().Budget.CheckBudgetAlerts(userID, subcategoryID)
	if err != nil || len(alerts) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, a := range alerts {
		if a.Exceeded {
			sb.WriteString(fmt.Sprintf(
				"\n🚨 *Budget exceeded!* %s: ৳%s/৳%s (%.0f%%)",
				a.CategoryName, FormatBDT(a.Spent), FormatBDT(a.BudgetAmount), a.Percent,
			))
		} else {
			sb.WriteString(fmt.Sprintf(
				"\n⚠️ *Budget warning!* %s: ৳%s/৳%s (%.0f%%)",
				a.CategoryName, FormatBDT(a.Spent), FormatBDT(a.BudgetAmount), a.Percent,
			))
		}
	}
	return sb.String()
}
