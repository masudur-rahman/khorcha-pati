package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/models"

	"golang.org/x/time/rate"
)

const aiRateLimit = 5 // max requests per second

var (
	limiter = rate.NewLimiter(rate.Limit(aiRateLimit), aiRateLimit)
)

type Classifier string

const (
	ClassifierGemini     Classifier = "gemini"
	ClassifierOpenRouter Classifier = "open-router"
	// ClassifierPool spreads requests across all configured providers with sticky rotation
	// and rate-limit failover. See provider_pool.go.
	ClassifierPool Classifier = "pool"
)

type ClassificationResult struct {
	Intent      string  `json:"intent"`
	Category    string  `json:"category_id"`
	Subcategory string  `json:"subcategory_id"`
	Confidence  float64 `json:"confidence"`
}

func TxnCategoryClassifier(ctx context.Context, userInput string, ai ...Classifier) (result *ClassificationResult, err error) {
	return TxnCategoryClassifierForType(ctx, userInput, "", ai...)
}

// TxnCategoryClassifierForType narrows the taxonomy passed to the AI to subcategories matching txnType.
// Pass an empty txnType to use the full taxonomy.
func TxnCategoryClassifierForType(ctx context.Context, userInput string, txnType models.TransactionType, ai ...Classifier) (result *ClassificationResult, err error) {
	classifier := Classifier(configs.TrackerConfig.System.AIClassifier)
	if len(ai) > 0 {
		classifier = ai[0]
	}

	if err = limiter.Wait(ctx); err != nil {
		return nil, err
	}

	taxonomy := models.TxnSubcategories
	if txnType != "" {
		filtered := make([]models.TxnSubcategory, 0, len(taxonomy))
		for _, sub := range taxonomy {
			if models.ContainsType(models.SubcategoryTypes[sub.ID], txnType) {
				filtered = append(filtered, sub)
			}
		}
		taxonomy = filtered
	}

	taxonomyJSON, err := json.MarshalIndent(taxonomy, "", "  ")
	if err != nil {
		return nil, err
	}

	var aiUsed bool
	if classifier == ClassifierPool {
		result, aiUsed, err = classifyWithPool(ctx, userInput, string(taxonomyJSON))
	} else {
		result, aiUsed, err = runClassifier(ctx, classifier, userInput, string(taxonomyJSON))
	}
	if err != nil {
		return nil, err
	}
	if !aiUsed {
		// No AI provider configured — return the raw input untouched, as before.
		return result, nil
	}

	fmt.Printf("Matched: %s > %s (Intent: %s, Confidence: %v)\n", result.Category, result.Subcategory, result.Intent, result.Confidence)
	if _, valid := models.SubCatNameMap[result.Subcategory]; !valid && result.Subcategory != "" {
		fmt.Printf("AI returned invalid subcategory ID %q, falling back to misc-misc\n", result.Subcategory)
		result.Subcategory = "misc-misc"
	}
	if result.Subcategory == "" {
		return nil, errors.New("transaction category can't be determined")
	}
	result.Intent = normalizeIntent(result.Intent, result.Subcategory)
	return result, nil
}

// validIntents are the only intent values the parser and cache understand.
var validIntents = map[string]bool{"income": true, "expense": true, "transfer": true}

// normalizeIntent coerces an AI-provided intent into the {income, expense, transfer} set.
// Free-form (non-Gemini) providers can return arbitrary text in the intent field; when it is
// not a valid intent, derive one from the subcategory's allowed types (a single allowed type
// wins) and fall back to expense, so a malformed intent never reaches the cache or the parser.
func normalizeIntent(intent, subID string) string {
	got := strings.ToLower(strings.TrimSpace(intent))
	if validIntents[got] {
		return got
	}
	fmt.Printf("AI returned invalid intent %q, deriving from subcategory %q\n", intent, subID)
	if types := models.SubcategoryTypes[subID]; len(types) == 1 {
		return intentForType(types[0])
	}
	return "expense"
}

// intentForType maps a transaction type to its intent string.
func intentForType(t models.TransactionType) string {
	switch t {
	case models.IncomeTransaction:
		return "income"
	case models.TransferTransaction:
		return "transfer"
	default:
		return "expense"
	}
}

// runClassifier calls a single AI provider. The bool reports whether an AI actually ran; it is
// false when the provider has no configured key (raw input is returned as a fallback).
func runClassifier(ctx context.Context, classifier Classifier, userInput, taxonomyJSON string) (*ClassificationResult, bool, error) {
	switch classifier {
	case ClassifierGemini:
		apiKey := configs.TrackerConfig.System.GeminiKey
		if apiKey == "" {
			return &ClassificationResult{Subcategory: userInput}, false, nil
		}
		result, err := TxnSubcategoryClassifier(ctx, apiKey, userInput, taxonomyJSON)
		return result, true, err
	case ClassifierOpenRouter:
		apiKey := configs.TrackerConfig.System.OpenRouterKey
		if apiKey == "" {
			return &ClassificationResult{Subcategory: userInput}, false, nil
		}
		result, err := NewClient(apiKey).TxnSubcategoryClassifier(ctx, userInput, taxonomyJSON)
		return result, true, err
	default:
		return &ClassificationResult{Subcategory: userInput}, false, nil
	}
}

// classifyWithPool runs the request through the provider pool: it tries the active provider and,
// on a rate-limit error, fails over to the next provider in the sequence.
func classifyWithPool(ctx context.Context, userInput, taxonomyJSON string) (*ClassificationResult, bool, error) {
	pool := getPool()
	seq := pool.sequence()
	if len(seq) == 0 {
		return &ClassificationResult{Subcategory: userInput}, false, nil
	}

	var (
		result *ClassificationResult
		used   bool
		err    error
	)
	for i, gen := range seq {
		result, used, err = runClassifier(ctx, gen, userInput, taxonomyJSON)
		if err == nil {
			return result, used, nil
		}
		if isRateLimited(err) {
			fmt.Printf("AI provider %q rate-limited, failing over\n", gen)
			pool.markRateLimited()
			if i < len(seq)-1 {
				continue
			}
		}
		return nil, used, err
	}
	return nil, used, err
}
