package telegram

import (
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"

	"github.com/stretchr/testify/assert"
)

func TestFormatSummary_Connectors(t *testing.T) {
	sg := gqtypes.SummaryGroups{
		Category: map[string]gqtypes.FieldCost{
			"food":  {Name: "Food", Amount: 100},
			"shop":  {Name: "Shopping", Amount: 200},
			"trans": {Name: "Transport", Amount: 300},
		},
	}

	formatted := FormatSummary(sg, "Test Summary")

	// Verify hierarchical connectors (alphabetical order: food, shop, trans)
	assert.Contains(t, formatted, "📊 *Test Summary*")
	assert.Contains(t, formatted, " ├ Food: *৳100*")
	assert.Contains(t, formatted, " ├ Shopping: *৳200*")
	assert.Contains(t, formatted, " └ Transport: *৳300*")
}

func TestFormatTransactionList_PaginationAndSorting(t *testing.T) {
	txns := []models.Transaction{
		{Amount: 10, Timestamp: 1000, SubcategoryID: "food-groc", Type: models.ExpenseTransaction},
		{Amount: 20, Timestamp: 2000, SubcategoryID: "food-groc", Type: models.ExpenseTransaction},
		{Amount: 30, Timestamp: 3000, SubcategoryID: "food-groc", Type: models.ExpenseTransaction},
	}

	// Page 1, Size 2 (Should show txns with timestamp 3000 and 2000)
	formatted := FormatTransactionList(txns, 1, 2)

	assert.Contains(t, formatted, "Page 1")
	assert.Contains(t, formatted, "1. 💸 *−৳30*")
	assert.Contains(t, formatted, "2. 💸 *−৳20*")
	assert.NotContains(t, formatted, "3. 💸 *−৳10*")

	// Page 2, Size 2 (Should show txn with timestamp 1000)
	formatted2 := FormatTransactionList(txns, 2, 2)
	assert.Contains(t, formatted2, "Page 2")
	assert.Contains(t, formatted2, "3. 💸 *−৳10*")
}

func TestEscapeMarkdownV2(t *testing.T) {
	input := "Hello_World! 1.0 + 2.0 = 3.0"
	escaped := EscapeMarkdownV2(input)

	// Characters like _, !, ., +, = should be escaped
	assert.Contains(t, escaped, "\\_")
	assert.Contains(t, escaped, "\\!")
	assert.Contains(t, escaped, "\\.")
	assert.Contains(t, escaped, "\\+")
	assert.Contains(t, escaped, "\\=")
}
