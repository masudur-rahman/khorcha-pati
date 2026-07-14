package web

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePageLimit(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantPage  int64
		wantLimit int64
	}{
		{"no params defaults page 1 no limit", "", 1, 0},
		{"page and limit", "?page=3&limit=20", 3, 20},
		{"limit only keeps default page", "?limit=5", 1, 5},
		{"zero and negative ignored", "?page=0&limit=-4", 1, 0},
		{"garbage ignored", "?page=abc&limit=xyz", 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/v1/transactions"+tt.query, nil)
			page, limit := parsePageLimit(r)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantLimit, limit)
		})
	}
}

func TestSlicePage(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	t.Run("no limit returns all", func(t *testing.T) {
		got, total := slicePage(items, 1, 0)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, items, got)
	})
	t.Run("first page", func(t *testing.T) {
		got, total := slicePage(items, 1, 2)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, []int{1, 2}, got)
	})
	t.Run("last partial page", func(t *testing.T) {
		got, total := slicePage(items, 3, 2)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, []int{5}, got)
	})
	t.Run("page past end is empty", func(t *testing.T) {
		got, total := slicePage(items, 9, 2)
		assert.Equal(t, int64(5), total)
		assert.Empty(t, got)
	})
}
