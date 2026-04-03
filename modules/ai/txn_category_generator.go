package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/models"

	"golang.org/x/time/rate"
)

const aiRateLimit = 5 // max requests per second

var (
	limiter = rate.NewLimiter(rate.Limit(aiRateLimit), aiRateLimit)
)

type GeneratorAI string

const (
	GeneratorGemini     GeneratorAI = "gemini"
	GeneratorOpenRouter GeneratorAI = "open-router"
)

type ClassificationResult struct {
	Intent      string  `json:"intent"`
	Category    string  `json:"category_id"`
	Subcategory string  `json:"subcategory_id"`
	Confidence  float64 `json:"confidence"`
}

func TxnCategoryGenerator(ctx context.Context, userInput string, ai ...GeneratorAI) (result *ClassificationResult, err error) {
	generator := GeneratorAI(configs.TrackerConfig.System.AIGenerator)
	if len(ai) > 0 {
		generator = ai[0]
	}

	if err = limiter.Wait(ctx); err != nil {
		return nil, err
	}

	taxonomyJSON, err := json.MarshalIndent(models.TxnSubcategories, "", "  ")
	if err != nil {
		return nil, err
	}

	switch generator {
	case GeneratorGemini:
		apiKey := configs.TrackerConfig.System.GeminiKey
		if apiKey == "" {
			return &ClassificationResult{Subcategory: userInput}, nil
		}
		result, err = TxnSubcategoryClassifier(ctx, apiKey, userInput, string(taxonomyJSON))
		if err != nil {
			return nil, err
		}
	case GeneratorOpenRouter:
		apiKey := configs.TrackerConfig.System.OpenRouterKey
		if apiKey == "" {
			return &ClassificationResult{Subcategory: userInput}, nil
		}
		client := NewClient(apiKey)
		result, err = client.TxnSubcategoryClassifier(ctx, userInput, string(taxonomyJSON))
		if err != nil {
			return nil, err
		}
	default:
		return &ClassificationResult{Subcategory: userInput}, nil
	}

	fmt.Printf("Matched: %s > %s (Intent: %s, Confidence: %v)\n", result.Category, result.Subcategory, result.Intent, result.Confidence)
	if _, valid := models.SubCatNameMap[result.Subcategory]; !valid && result.Subcategory != "" {
		fmt.Printf("AI returned invalid subcategory ID %q, falling back to misc-misc\n", result.Subcategory)
		result.Subcategory = "misc-misc"
	}
	if result.Subcategory == "" {
		return nil, errors.New("transaction category can't be determined")
	}
	return result, nil
}
