package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateClassification(t *testing.T) {
	tests := []struct {
		name    string
		subID   string
		intent  string
		wantErr bool
	}{
		{"valid expense", "food-groc", "Expense", false},
		{"valid expense lowercase", "food-groc", "expense", false},
		{"valid income", "fin-sal", "Income", false},
		{"valid income lowercase", "fin-sal", "income", false},
		{"unknown subcategory", "not-a-real-id", "Expense", true},
		{"intent not allowed for subcategory", "fin-sal", "Expense", true},
		{"empty intent", "food-groc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClassification(tt.subID, tt.intent)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeConfidence(t *testing.T) {
	ptr := func(f float64) *float64 { return &f }
	tests := []struct {
		name string
		in   *float64
		want float64
	}{
		{"nil defaults to full", nil, 1.0},
		{"in range kept", ptr(0.75), 0.75},
		{"below zero clamped", ptr(-0.5), 0},
		{"above one clamped", ptr(1.5), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeConfidence(tt.in))
		})
	}
}
