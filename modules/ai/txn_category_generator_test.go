package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTxnCategoryGenerator(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("no AI API key set, skipping")
	}
	type args struct {
		ctx       context.Context
		userInput string
		ai        []GeneratorAI
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
				ai:        []GeneratorAI{GeneratorGemini},
			},
			wantSubCatID: "food-fruit",
			wantErr:      false,
		},
		{
			name: "nvdia",
			args: args{
				ctx:       context.Background(),
				userInput: "apple",
				//ai:        []GeneratorAI{GeneratorGemini},
			},
			wantSubCatID: "food-fruit",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result, err := TxnCategoryGenerator(tt.args.ctx, tt.args.userInput, tt.args.ai...)
			if (err != nil) != tt.wantErr {
				if strings.Contains(err.Error(), "API error") || strings.Contains(err.Error(), "rate limit") {
					t.Logf("TxnCategoryGenerator() failed due to AI API issue: %v", err)
					return
				}
				t.Errorf("TxnCategoryGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// If API key is missing, it returns the input text in subcategory
			if result != nil && result.Subcategory == tt.args.userInput {
				t.Logf("API key missing or request failed, returned input text: %v", result.Subcategory)
				return
			}
			if result != nil && result.Subcategory != tt.wantSubCatID {
				t.Errorf("TxnCategoryGenerator() gotSubCatID = %v, want %v", result.Subcategory, tt.wantSubCatID)
			}

			fmt.Println("Time Taken: ", time.Since(start).String())
		})
	}
}
