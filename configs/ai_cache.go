package configs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
)

// Sentinel errors for AI-cache admin operations.
var (
	ErrAICacheNotFound  = errors.New("ai cache entry not found")
	ErrAICacheDuplicate = errors.New("ai cache entry with this input text already exists")
)

const defaultAICacheListLimit = 200

// AI-cache import conflict modes (keyed on input_text).
const (
	ImportSkip             = "skip"       // keep existing entry untouched
	ImportOverwrite        = "overwrite"  // incoming replaces existing classification
	ImportHigherConfidence = "confidence" // replace only when incoming confidence is higher
)

// AICacheImportSummary reports the outcome of an import run.
type AICacheImportSummary struct {
	Imported    int `json:"imported"`
	Overwritten int `json:"overwritten"`
	Skipped     int `json:"skipped"`
	Invalid     int `json:"invalid"`
}

// IsValidImportMode reports whether mode is a supported conflict mode.
func IsValidImportMode(mode string) bool {
	return mode == ImportSkip || mode == ImportOverwrite || mode == ImportHigherConfidence
}

// setAICacheMemory writes an entry into the in-memory classifier cache under its input text.
func setAICacheMemory(entry models.AICache) {
	resultJSON, _ := json.Marshal(map[string]any{
		"intent":         entry.Intent,
		"subcategory_id": entry.SubcategoryID,
		"confidence":     entry.Confidence,
	})
	_ = cache.SetCache(entry.InputText, string(resultJSON), -1)
}

// ListAICache returns AI cache entries filtered by an optional input-text substring,
// newest first, capped at limit (a default is applied when limit <= 0).
func ListAICache(q string, limit int) ([]models.AICache, error) {
	if sqlDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if limit <= 0 {
		limit = defaultAICacheListLimit
	}

	eng := GetUnitOfWork().SQL.Table(models.AICache{}.TableName())
	if q = strings.TrimSpace(q); q != "" {
		eng = eng.Where("input_text LIKE ?", "%"+q+"%")
	}

	var rows []models.AICache
	if err := eng.OrderBy("created_at", "DESC").Limit(int64(limit)).FindMany(context.Background(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ExportAICache returns every AI-cache entry, newest first (no limit) — the source
// for the admin export/sync.
func ExportAICache() ([]models.AICache, error) {
	if sqlDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var rows []models.AICache
	if err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).
		OrderBy("created_at", "DESC").FindMany(context.Background(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ImportAICacheEntries upserts pre-validated entries using the given conflict mode
// (keyed on input_text). It reuses Create/Update so the in-memory cache stays in
// sync, and reports per-entry outcomes.
func ImportAICacheEntries(entries []models.AICache, mode string) (AICacheImportSummary, error) {
	var s AICacheImportSummary
	if sqlDB == nil {
		return s, fmt.Errorf("database not initialized")
	}
	ctx := context.Background()

	for _, e := range entries {
		e.InputText = strings.TrimSpace(e.InputText)
		if e.InputText == "" {
			s.Invalid++
			continue
		}

		existing, found, err := getAICacheByInputText(ctx, e.InputText)
		if err != nil {
			return s, err
		}
		if !found {
			if _, err := CreateAICache(e.InputText, e.SubcategoryID, e.Intent, e.Confidence); err != nil {
				return s, err
			}
			s.Imported++
			continue
		}

		switch mode {
		case ImportOverwrite:
			if _, err := UpdateAICacheClassification(existing.ID, e.SubcategoryID, e.Intent, e.Confidence); err != nil {
				return s, err
			}
			s.Overwritten++
		case ImportHigherConfidence:
			if e.Confidence > existing.Confidence {
				if _, err := UpdateAICacheClassification(existing.ID, e.SubcategoryID, e.Intent, e.Confidence); err != nil {
					return s, err
				}
				s.Overwritten++
			} else {
				s.Skipped++
			}
		default: // ImportSkip
			s.Skipped++
		}
	}
	return s, nil
}

// getAICacheByInputText loads an entry by its unique input text.
func getAICacheByInputText(ctx context.Context, inputText string) (models.AICache, bool, error) {
	var entry models.AICache
	found, err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).
		FindOne(ctx, &entry, models.AICache{InputText: inputText})
	return entry, found, err
}

// CreateAICache validates input-text uniqueness, persists the entry, and updates the in-memory cache.
func CreateAICache(inputText, subID, intent string, confidence float64) (models.AICache, error) {
	if sqlDB == nil {
		return models.AICache{}, fmt.Errorf("database not initialized")
	}
	ctx := context.Background()

	var existing models.AICache
	found, err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).
		FindOne(ctx, &existing, models.AICache{InputText: inputText})
	if err != nil {
		return models.AICache{}, err
	}
	if found {
		return models.AICache{}, ErrAICacheDuplicate
	}

	entry := models.AICache{
		InputText:     inputText,
		SubcategoryID: subID,
		Intent:        intent,
		Confidence:    confidence,
		CreatedAt:     time.Now().Unix(),
	}
	// Pass a pointer so styx assigns the generated primary key back into entry.
	if _, err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).InsertOne(ctx, &entry); err != nil {
		return models.AICache{}, err
	}
	setAICacheMemory(entry)
	return entry, nil
}

// UpdateAICacheClassification updates an entry's classification fields (input text is immutable)
// and refreshes the in-memory cache under the same key.
func UpdateAICacheClassification(id int64, subID, intent string, confidence float64) (models.AICache, error) {
	if sqlDB == nil {
		return models.AICache{}, fmt.Errorf("database not initialized")
	}
	ctx := context.Background()

	entry, err := findAICacheByID(ctx, id)
	if err != nil {
		return models.AICache{}, err
	}

	update := &models.AICache{SubcategoryID: subID, Intent: intent, Confidence: confidence}
	if err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).ID(id).UpdateOne(ctx, update); err != nil {
		return models.AICache{}, err
	}

	entry.SubcategoryID, entry.Intent, entry.Confidence = subID, intent, confidence
	setAICacheMemory(entry)
	return entry, nil
}

// DeleteAICache removes an entry from both the database and the in-memory cache.
func DeleteAICache(id int64) error {
	if sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}
	ctx := context.Background()

	entry, err := findAICacheByID(ctx, id)
	if err != nil {
		return err
	}
	if err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).ID(id).DeleteOne(ctx); err != nil {
		return err
	}
	_ = cache.DeleteCache(entry.InputText)
	return nil
}

// findAICacheByID loads a single entry, returning ErrAICacheNotFound when it is absent.
func findAICacheByID(ctx context.Context, id int64) (models.AICache, error) {
	var entry models.AICache
	found, err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).ID(id).FindOne(ctx, &entry)
	if err != nil {
		return models.AICache{}, err
	}
	if !found {
		return models.AICache{}, ErrAICacheNotFound
	}
	return entry, nil
}
