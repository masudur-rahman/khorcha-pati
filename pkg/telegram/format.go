package telegram

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"
)

// FormatSummary converts a SummaryGroups object into a hierarchical Markdown text.
func FormatSummary(sg gqtypes.SummaryGroups, title string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📊 *%s*\n", title))
	sb.WriteString(models.Separator + "\n\n")

	formatMap := func(m map[string]gqtypes.FieldCost, header, emoji string) {
		if len(m) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("%s *%s*\n", emoji, header))
		
		// Sort keys for consistent output
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			v := m[k]
			prefix := " ├ "
			if i == len(keys)-1 {
				prefix = " └ "
			}
			name := v.Name
			if name == "" {
				name = k
			}
			sb.WriteString(fmt.Sprintf("%s%s: *%.2f*\n", prefix, name, v.Amount))
		}
		sb.WriteString("\n")
	}

	formatMap(sg.Type, "By Type", "💰")
	formatMap(sg.Category, "By Category", "🏷")
	formatMap(sg.Subcategory, "By Subcategory", "🗂")

	return strings.TrimSpace(sb.String())
}

// FormatTransactionList converts a slice of transactions into a concise Markdown list.
func FormatTransactionList(txns []models.Transaction, page, pageSize int) string {
	if len(txns) == 0 {
		return "No transactions found."
	}

	// Ensure transactions are sorted descending (newest first)
	sort.Slice(txns, func(i, j int) bool {
		return txns[i].Timestamp > txns[j].Timestamp
	})

	// Slice for pagination
	start := (page - 1) * pageSize
	if start >= len(txns) {
		return "No more items."
	}
	end := start + pageSize
	if end > len(txns) {
		end = len(txns)
	}
	pageTxns := txns[start:end]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📄 *Transactions (Page %d)*\n", page))
	sb.WriteString(models.Separator + "\n\n")

	for i, txn := range pageTxns {
		emoji := "💸"
		if txn.Type == models.IncomeTransaction {
			emoji = "💰"
		} else if txn.Type == models.TransferTransaction {
			emoji = "↔️"
		}

		subName := models.SubCatNameMap[txn.SubcategoryID]
		if subName == "" {
			subName = txn.SubcategoryID
		}

		dateStr := time.Unix(txn.Timestamp, 0).Format("02 Jan")
		
		sb.WriteString(fmt.Sprintf("%d. %s *%.2f* | %s\n", start+i+1, emoji, txn.Amount, subName))
		sb.WriteString(fmt.Sprintf("   `[%s]`", dateStr))
		if txn.Remarks != "" {
			sb.WriteString(fmt.Sprintf(" — %s", txn.Remarks))
		}
		sb.WriteString("\n\n")
	}

	return strings.TrimSpace(sb.String())
}

// EscapeMarkdownV2 escapes special characters for Telegram MarkdownV2.
func EscapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
	)
	return replacer.Replace(text)
}
