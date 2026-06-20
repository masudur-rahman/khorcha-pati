package models

import (
	"math"
	"strconv"
	"strings"
)

// CurrencySymbol is the Bangladeshi Taka sign used across all money output.
const CurrencySymbol = "৳"

// minusSign is the typographic minus (U+2212), not the ASCII hyphen.
const minusSign = "−"

// GroupingStyle selects digit grouping for money amounts.
type GroupingStyle int

const (
	// GroupingBangladeshi groups as lakh/crore: 12,34,567. Default.
	GroupingBangladeshi GroupingStyle = iota
	// GroupingWestern groups in thousands: 1,234,567.
	GroupingWestern
)

// DefaultGrouping is the process-wide default grouping style. Region-aware
// selection can override this per call later via FormatMoneyStyle.
var DefaultGrouping = GroupingBangladeshi

// FormatMoney returns an unsigned Taka amount with the default grouping and no
// decimals, e.g. "৳1,00,000".
func FormatMoney(amount float64) string {
	return FormatMoneyStyle(amount, DefaultGrouping)
}

// FormatMoneyStyle returns an unsigned Taka amount with the given grouping
// style and no decimals.
func FormatMoneyStyle(amount float64, style GroupingStyle) string {
	return CurrencySymbol + GroupAmount(amount, style)
}

// GroupAmount returns the grouped, decimal-less digits of an amount without the
// currency symbol — for renderers whose font lacks the ৳ glyph, or callers that
// add the symbol themselves.
func GroupAmount(amount float64, style GroupingStyle) string {
	return groupDigits(int64(math.Round(math.Abs(amount))), style)
}

// FormatMoneyValue returns a Taka amount signed by its numeric value: a minus
// for negatives, no prefix for non-negatives (for balances and net figures).
func FormatMoneyValue(amount float64) string {
	if amount < 0 {
		return minusSign + FormatMoney(amount)
	}
	return FormatMoney(amount)
}

// FormatMoneySigned returns a Taka amount prefixed by a sign per transaction
// type: '+' income, '−' expense, none for transfer, e.g. "−৳320".
func FormatMoneySigned(amount float64, txnType TransactionType) string {
	switch txnType {
	case IncomeTransaction:
		return "+" + FormatMoney(amount)
	case ExpenseTransaction:
		return minusSign + FormatMoney(amount)
	default: // transfer
		return FormatMoney(amount)
	}
}

// groupDigits inserts thousands separators into a non-negative integer per the
// given grouping style.
func groupDigits(n int64, style GroupingStyle) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	head, tail := s[:len(s)-3], s[len(s)-3:]
	groupSize := 2 // Bangladeshi lakh/crore
	if style == GroupingWestern {
		groupSize = 3
	}
	return groupEvery(head, groupSize) + "," + tail
}

// groupEvery inserts commas every size digits from the right of s.
func groupEvery(s string, size int) string {
	var parts []string
	for len(s) > size {
		parts = append([]string{s[len(s)-size:]}, parts...)
		s = s[:len(s)-size]
	}
	return strings.Join(append([]string{s}, parts...), ",")
}
