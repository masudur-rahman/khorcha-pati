package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNormalizeIntent(t *testing.T) {
	tests := []struct {
		name   string
		intent string
		subID  string
		want   string
	}{
		{"valid income", "income", "fin-sal", "income"},
		{"valid expense", "expense", "food-rest", "expense"},
		{"valid transfer", "transfer", "fin-transfer", "transfer"},
		{"case and space normalized", "  Income ", "fin-sal", "income"},
		{"garbage derives income from single-type sub", "Record a birthday gift income", "fin-sal", "income"},
		{"garbage derives transfer from single-type sub", "moving money around", "fin-transfer", "transfer"},
		{"garbage derives expense from single-type sub", "Record a gift expense for a niece", "food-rest", "expense"},
		{"garbage on multi-type sub defaults to expense", "Record a birthday gift expense", "misc-gift", "expense"},
		{"garbage on unknown sub defaults to expense", "some reasoning text", "no-such-sub", "expense"},
		{"empty intent defaults to expense", "", "misc-gift", "expense"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeIntent(tt.intent, tt.subID); got != tt.want {
				t.Errorf("normalizeIntent(%q, %q) = %q, want %q", tt.intent, tt.subID, got, tt.want)
			}
		})
	}
}

func TestTxnCategoryClassifier(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("no AI API key set, skipping")
	}
	type args struct {
		ctx       context.Context
		userInput string
		ai        []Classifier
	}
	tests := []struct {
		name         string
		args         args
		wantSubCatID string
		wantErr      bool
	}{
		{
			name: "gemini",
			args: args{
				ctx:       context.Background(),
				userInput: "apple",
				ai:        []Classifier{ClassifierGemini},
			},
			wantSubCatID: "food-fruit",
			wantErr:      false,
		},
		{
			name: "nvdia",
			args: args{
				ctx:       context.Background(),
				userInput: "apple",
				//ai:        []Classifier{ClassifierGemini},
			},
			wantSubCatID: "food-fruit",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result, err := TxnCategoryClassifier(tt.args.ctx, tt.args.userInput, tt.args.ai...)
			if (err != nil) != tt.wantErr {
				if strings.Contains(err.Error(), "API error") || strings.Contains(err.Error(), "rate limit") {
					t.Logf("TxnCategoryClassifier() failed due to AI API issue: %v", err)
					return
				}
				t.Errorf("TxnCategoryClassifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// If API key is missing, it returns the input text in subcategory
			if result != nil && result.Subcategory == tt.args.userInput {
				t.Logf("API key missing or request failed, returned input text: %v", result.Subcategory)
				return
			}
			if result != nil && result.Subcategory != tt.wantSubCatID {
				t.Errorf("TxnCategoryClassifier() gotSubCatID = %v, want %v", result.Subcategory, tt.wantSubCatID)
			}

			fmt.Println("Time Taken: ", time.Since(start).String())
		})
	}
}
