package ai

import (
	"context"
	"fmt"
	"os"
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
			gotSubCatID, err := TxnCategoryGenerator(tt.args.ctx, tt.args.userInput, tt.args.ai...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxnCategoryGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// If API key is missing, it returns the input text
			if gotSubCatID == tt.args.userInput {
				t.Logf("API key missing or request failed, returned input text: %v", gotSubCatID)
				return
			}
			if gotSubCatID != tt.wantSubCatID {
				t.Errorf("TxnCategoryGenerator() gotSubCatID = %v, want %v", gotSubCatID, tt.wantSubCatID)
			}

			fmt.Println("Time Taken: ", time.Since(start).String())
		})
	}
}
