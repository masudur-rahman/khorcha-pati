package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/go-chi/chi/v5"
)

type aiCacheResponse struct {
	ID            int64   `json:"id"`
	InputText     string  `json:"inputText"`
	SubcategoryID string  `json:"subcategoryId"`
	Intent        string  `json:"intent"`
	Confidence    float64 `json:"confidence"`
	CreatedAt     int64   `json:"createdAt"`
}

type aiCacheRequest struct {
	InputText     string   `json:"inputText"`
	SubcategoryID string   `json:"subcategoryId"`
	Intent        string   `json:"intent"`
	Confidence    *float64 `json:"confidence"`
}

// aiCacheExportEntry is the portable shape for export/import (no server-local id/createdAt).
type aiCacheExportEntry struct {
	InputText     string  `json:"inputText"`
	SubcategoryID string  `json:"subcategoryId"`
	Intent        string  `json:"intent"`
	Confidence    float64 `json:"confidence"`
}

type aiCacheImportRequest struct {
	Mode    string               `json:"mode"`
	Entries []aiCacheExportEntry `json:"entries"`
}

func toAICacheResponse(e models.AICache) aiCacheResponse {
	return aiCacheResponse{
		ID:            e.ID,
		InputText:     e.InputText,
		SubcategoryID: e.SubcategoryID,
		Intent:        e.Intent,
		Confidence:    e.Confidence,
		CreatedAt:     e.CreatedAt,
	}
}

// HandleAdminListAICache returns AI-cache entries, optionally filtered by ?q= and capped by ?limit=.
func HandleAdminListAICache(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePageLimit(r)
	rows, err := configs.ListAICache(r.URL.Query().Get("q"), 0)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "ai_cache_error", err.Error())
		return
	}
	out := make([]aiCacheResponse, 0, len(rows))
	for _, e := range rows {
		out = append(out, toAICacheResponse(e))
	}
	// Back-compat: envelope only when a limit is requested; else the legacy array.
	if limit > 0 {
		items, total := slicePage(out, page, limit)
		writePaged(w, http.StatusOK, items, page, limit, total)
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// HandleAdminExportAICache returns every AI-cache entry in the portable shape so
// the admin can sync learnings between servers.
func HandleAdminExportAICache(w http.ResponseWriter, r *http.Request) {
	rows, err := configs.ExportAICache()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "ai_cache_error", err.Error())
		return
	}
	out := make([]aiCacheExportEntry, 0, len(rows))
	for _, e := range rows {
		out = append(out, aiCacheExportEntry{
			InputText:     e.InputText,
			SubcategoryID: e.SubcategoryID,
			Intent:        e.Intent,
			Confidence:    e.Confidence,
		})
	}
	WriteJSON(w, http.StatusOK, out)
}

// HandleAdminImportAICache upserts a batch of entries using the requested conflict
// mode, validating each classification and reporting per-entry outcomes.
func HandleAdminImportAICache(w http.ResponseWriter, r *http.Request) {
	var body aiCacheImportRequest
	if err := ReadJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if !configs.IsValidImportMode(body.Mode) {
		WriteError(w, http.StatusBadRequest, "invalid_mode", "mode must be skip, overwrite or confidence")
		return
	}

	valid := make([]models.AICache, 0, len(body.Entries))
	invalid := 0
	for _, e := range body.Entries {
		e.InputText = strings.TrimSpace(e.InputText)
		e.Intent = strings.ToLower(strings.TrimSpace(e.Intent))
		if e.InputText == "" || validateClassification(e.SubcategoryID, e.Intent) != nil {
			invalid++
			continue
		}
		conf := e.Confidence
		if conf <= 0 {
			conf = 1.0
		} else if conf > 1 {
			conf = 1.0
		}
		valid = append(valid, models.AICache{
			InputText:     e.InputText,
			SubcategoryID: e.SubcategoryID,
			Intent:        e.Intent,
			Confidence:    conf,
		})
	}

	summary, err := configs.ImportAICacheEntries(valid, body.Mode)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "import_failed", err.Error())
		return
	}
	summary.Invalid += invalid
	WriteJSON(w, http.StatusOK, summary)
}

// HandleAdminCreateAICache adds a curated AI-cache entry. Confidence defaults to 1.0 (100%).
func HandleAdminCreateAICache(w http.ResponseWriter, r *http.Request) {
	var body aiCacheRequest
	if err := ReadJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	body.InputText = strings.TrimSpace(body.InputText)
	if body.InputText == "" {
		WriteError(w, http.StatusBadRequest, "invalid_input", "input text is required")
		return
	}
	if err := validateClassification(body.SubcategoryID, body.Intent); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_classification", err.Error())
		return
	}

	entry, err := configs.CreateAICache(body.InputText, body.SubcategoryID, body.Intent, normalizeConfidence(body.Confidence))
	if err != nil {
		if errors.Is(err, configs.ErrAICacheDuplicate) {
			WriteError(w, http.StatusConflict, "duplicate", err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, toAICacheResponse(entry))
}

// HandleAdminUpdateAICache edits an entry's classification. Input text stays immutable.
func HandleAdminUpdateAICache(w http.ResponseWriter, r *http.Request) {
	id, ok := aiCacheIDParam(w, r)
	if !ok {
		return
	}
	var body aiCacheRequest
	if err := ReadJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if err := validateClassification(body.SubcategoryID, body.Intent); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_classification", err.Error())
		return
	}

	entry, err := configs.UpdateAICacheClassification(id, body.SubcategoryID, body.Intent, normalizeConfidence(body.Confidence))
	if err != nil {
		writeAICacheError(w, err, "update_failed")
		return
	}
	WriteJSON(w, http.StatusOK, toAICacheResponse(entry))
}

// HandleAdminDeleteAICache removes an entry from the DB and the in-memory cache.
func HandleAdminDeleteAICache(w http.ResponseWriter, r *http.Request) {
	id, ok := aiCacheIDParam(w, r)
	if !ok {
		return
	}
	if err := configs.DeleteAICache(id); err != nil {
		writeAICacheError(w, err, "delete_failed")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"id": id})
}

// validateClassification ensures the subcategory exists and the intent is one of its allowed types.
func validateClassification(subID, intent string) error {
	if _, ok := models.SubcategoryByID[subID]; !ok {
		return errors.New("unknown subcategory: " + subID)
	}
	for _, t := range models.SubcategoryTypes[subID] {
		if strings.EqualFold(string(t), intent) {
			return nil
		}
	}
	return errors.New("intent " + intent + " is not valid for subcategory " + subID)
}

// normalizeConfidence defaults a missing confidence to 1.0 and clamps it to [0, 1].
func normalizeConfidence(c *float64) float64 {
	if c == nil {
		return 1.0
	}
	switch {
	case *c < 0:
		return 0
	case *c > 1:
		return 1
	default:
		return *c
	}
}

// aiCacheIDParam parses the {id} path param, writing a 400 and returning false on failure.
func aiCacheIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "id must be a number")
		return 0, false
	}
	return id, true
}

// writeAICacheError maps not-found to 404 and everything else to 500.
func writeAICacheError(w http.ResponseWriter, err error, code string) {
	if errors.Is(err, configs.ErrAICacheNotFound) {
		WriteError(w, http.StatusNotFound, "not_found", "ai cache entry not found")
		return
	}
	WriteError(w, http.StatusInternalServerError, code, err.Error())
}
