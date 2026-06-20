package models

import "testing"

func TestFormatMoneyStyle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		amount float64
		style  GroupingStyle
		want   string
	}{
		{"bd small", 320, GroupingBangladeshi, "৳320"},
		{"bd thousand", 7630, GroupingBangladeshi, "৳7,630"},
		{"bd lakh", 100000, GroupingBangladeshi, "৳1,00,000"},
		{"bd crore", 1234567, GroupingBangladeshi, "৳12,34,567"},
		{"western thousand", 7630, GroupingWestern, "৳7,630"},
		{"western hundredk", 100000, GroupingWestern, "৳100,000"},
		{"western million", 1234567, GroupingWestern, "৳1,234,567"},
		{"rounds decimals", 319.6, GroupingBangladeshi, "৳320"},
		{"abs negative", -500, GroupingBangladeshi, "৳500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatMoneyStyle(tt.amount, tt.style); got != tt.want {
				t.Errorf("FormatMoneyStyle(%v, %v) = %q, want %q", tt.amount, tt.style, got, tt.want)
			}
		})
	}
}

func TestFormatMoneyValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{"positive no sign", 1000, "৳1,000"},
		{"negative gets minus", -1000, "−৳1,000"},
		{"zero no sign", 0, "৳0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatMoneyValue(tt.amount); got != tt.want {
				t.Errorf("FormatMoneyValue(%v) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}

func TestFormatMoneySigned(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		amount  float64
		txnType TransactionType
		want    string
	}{
		{"income gets plus", 52000, IncomeTransaction, "+৳52,000"},
		{"expense gets minus", 320, ExpenseTransaction, "−৳320"},
		{"transfer no sign", 1000, TransferTransaction, "৳1,000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatMoneySigned(tt.amount, tt.txnType); got != tt.want {
				t.Errorf("FormatMoneySigned(%v, %v) = %q, want %q", tt.amount, tt.txnType, got, tt.want)
			}
		})
	}
}
