package transaction

import (
	"strings"
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/models"
)

// knownContacts for debt tests: John/Sarah/Karim/Rifat are on file; Friend/Ali are not,
// so they must land in [Person: ...] remarks.
func debtContacts(name string) bool {
	switch strings.ToLower(name) {
	case "john", "sarah", "karim", "rifat":
		return true
	}
	return false
}

// TestClassifyDebt covers every lending/recovering case listed in the README.
func TestClassifyDebt(t *testing.T) {
	tests := []struct {
		text string
		want string // "" means not debt-related
	}{
		// Lending → fin-lend
		{"gave john 500", models.LendSubID},
		{"lent 1000 to friend", models.LendSubID},
		{"handed 2000 to sarah", models.LendSubID},
		{"paid for friend 1500", models.LendSubID},
		{"covering ali's share 800", models.LendSubID},
		// Recovering → fin-recover
		{"got 500 back from john", models.LendRecoverySubID},
		{"john returned 1000", models.LendRecoverySubID},
		{"received loan repayment 2000", models.LendRecoverySubID},
		{"friend paid me back 1500", models.LendRecoverySubID},
		{"cashback from lending 800", models.LendRecoverySubID},
		// Banglish
		{"john ke 500 disi", models.LendSubID},
		{"500 ferot pelam", models.LendRecoverySubID},
		{"dhar shod 2000", models.BorrowReturnSubID},
		// Ambiguous (contact/person present, no direction) → lend default
		{"john 500", models.LendSubID},
		{"transaction with friend 1000", models.LendSubID},
		{"settlement 2000", ""},
		// Non-debt must stay untouched — including a contact next to a real category
		{"paid 1500 for wifi bill", ""},
		{"got bonus 20k", ""},
		{"lunch 250", ""},
		{"paid john for lunch 300", ""},
		{"sent to sarah for medicine 500", ""},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, ok := classifyDebt(tt.text, debtContacts)
			if tt.want == "" {
				if ok {
					t.Fatalf("classifyDebt(%q) = %q, want not-debt", tt.text, got)
				}
				return
			}
			if !ok {
				t.Fatalf("classifyDebt(%q) = not-debt, want %q", tt.text, tt.want)
			}
			if got != tt.want {
				t.Errorf("classifyDebt(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// TestClassifyDebt_incomeVariants asserts the Banglish "received from person" reads as
// money-in (borrow or recover), never as an expense.
func TestClassifyDebt_incomeVariants(t *testing.T) {
	got, ok := classifyDebt("friend theke 1000 pelam", debtContacts)
	if !ok {
		t.Fatal("expected debt classification")
	}
	if got != models.BorrowSubID && got != models.LendRecoverySubID {
		t.Errorf("got %q, want fin-borrow or fin-recover", got)
	}
}

// TestParseTransaction_debtPerson checks person handling end-to-end: known contact →
// ContactName; unknown person → [Person: ...] remark with empty ContactName.
func TestParseTransaction_debtPerson(t *testing.T) {
	initCache()
	accounts := func(string) bool { return false }

	tests := []struct {
		name          string
		text          string
		wantSub       string
		wantType      models.TransactionType
		wantContact   string
		wantRemarkSub string // substring expected in remarks ("" = none)
	}{
		{"known contact lend", "gave john 500", models.LendSubID, models.ExpenseTransaction, "john", ""},
		{"unknown person recover", "friend paid me back 1500", models.LendRecoverySubID, models.IncomeTransaction, "", "[Person: friend]"},
		{"contact subject recover", "john returned 1000", models.LendRecoverySubID, models.IncomeTransaction, "john", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransaction(tt.text, debtContacts, accounts)
			if err != nil {
				t.Fatalf("ParseTransaction(%q) error = %v", tt.text, err)
			}
			if got.SubcategoryID != tt.wantSub {
				t.Errorf("SubcategoryID = %q, want %q", got.SubcategoryID, tt.wantSub)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.ContactName != tt.wantContact {
				t.Errorf("ContactName = %q, want %q", got.ContactName, tt.wantContact)
			}
			if tt.wantRemarkSub != "" && !strings.Contains(got.Remarks, tt.wantRemarkSub) {
				t.Errorf("Remarks = %q, want to contain %q", got.Remarks, tt.wantRemarkSub)
			}
		})
	}
}
