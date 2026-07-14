package web

import (
	"net/http"
	"strconv"
)

// parsePageLimit reads the standard 1-based page + limit query params. A limit
// <= 0 means "no pagination" (return everything); page defaults to 1.
func parsePageLimit(r *http.Request) (page, limit int64) {
	page = 1
	if v, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64); err == nil && v > 0 {
		limit = v
	}
	return page, limit
}

// paginationMeta is the standard pagination block shared by every list endpoint.
type paginationMeta struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Total int64 `json:"total"`
}

// writePaged writes the standard { data, pagination } envelope. When limit <= 0
// the response reports a single page covering the whole set.
func writePaged(w http.ResponseWriter, status int, data any, page, limit, total int64) {
	if limit <= 0 {
		page = 1
		limit = total
	}
	WriteJSON(w, status, map[string]any{
		"data":       data,
		"pagination": paginationMeta{Page: page, Limit: limit, Total: total},
	})
}

// slicePage paginates an already-fetched slice in memory — for the small lists
// that aren't worth DB-level paging. Returns the page slice + total count. A
// limit <= 0 returns everything.
func slicePage[T any](items []T, page, limit int64) ([]T, int64) {
	total := int64(len(items))
	if limit <= 0 {
		return items, total
	}
	start := (page - 1) * limit
	if start < 0 || start >= total {
		return []T{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total
}
