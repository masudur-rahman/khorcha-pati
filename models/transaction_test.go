package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubcategoryTyping_everySubHasAtLeastOneType(t *testing.T) {
	t.Parallel()
	for _, sub := range TxnSubcategories {
		assert.NotEmpty(t, SubcategoryTypes[sub.ID], "subcategory %q has no Types", sub.ID)
	}
}

func TestCategoryTyping_isUnionOfSubcategories(t *testing.T) {
	t.Parallel()
	for _, cat := range TxnCategories {
		catTypes := CategoryTypes[cat.ID]
		for _, sub := range TxnSubcategories {
			if sub.CatID != cat.ID {
				continue
			}
			for _, st := range SubcategoryTypes[sub.ID] {
				assert.True(t, ContainsType(catTypes, st), "category %q missing type %q from sub %q", cat.ID, st, sub.ID)
			}
		}
	}
}

func TestSubcategoryTyping_finSpotChecks(t *testing.T) {
	t.Parallel()
	cases := map[string]TransactionType{
		"fin-sal":      IncomeTransaction,
		"fin-prof":     IncomeTransaction,
		"fin-borrow":   IncomeTransaction,
		"fin-recover":  IncomeTransaction,
		"fin-loan":     IncomeTransaction,
		"fin-repay":    ExpenseTransaction,
		"fin-lend":     ExpenseTransaction,
		"fin-return":   ExpenseTransaction,
		"fin-tax":      ExpenseTransaction,
		"fin-with":     TransferTransaction,
		"fin-deposit":  TransferTransaction,
		"fin-transfer": TransferTransaction,
	}
	for id, want := range cases {
		types, ok := SubcategoryTypes[id]
		assert.True(t, ok, "subcategory %q not registered", id)
		assert.True(t, ContainsType(types, want), "subcategory %q missing type %q (got %v)", id, want, types)
	}
}

func TestSubcategoryTyping_miscMixed(t *testing.T) {
	t.Parallel()
	mixed := []string{"misc-init", "misc-gift", "misc-charity", "misc-adj"}
	for _, id := range mixed {
		types, ok := SubcategoryTypes[id]
		assert.True(t, ok, "subcategory %q not registered", id)
		assert.True(t, ContainsType(types, ExpenseTransaction), "%q should allow Expense", id)
		assert.True(t, ContainsType(types, IncomeTransaction), "%q should allow Income", id)
	}
}

func TestContainsType(t *testing.T) {
	t.Parallel()
	types := []TransactionType{ExpenseTransaction, IncomeTransaction}
	assert.True(t, ContainsType(types, ExpenseTransaction))
	assert.True(t, ContainsType(types, IncomeTransaction))
	assert.False(t, ContainsType(types, TransferTransaction))
	assert.False(t, ContainsType(nil, ExpenseTransaction))
}
