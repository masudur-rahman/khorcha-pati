package web

import (
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/stretchr/testify/assert"
)

func TestSortAdminUsers(t *testing.T) {
	byCreatedAt := func(u models.Profile) int64 { return u.CreatedAt }

	tests := []struct {
		name    string
		users   []models.Profile
		desc    bool
		wantIDs []int64
	}{
		{
			name:    "descending puts newest first",
			users:   []models.Profile{{ID: 1, CreatedAt: 100}, {ID: 2, CreatedAt: 300}, {ID: 3, CreatedAt: 200}},
			desc:    true,
			wantIDs: []int64{2, 3, 1},
		},
		{
			name:    "ascending puts oldest first",
			users:   []models.Profile{{ID: 1, CreatedAt: 100}, {ID: 2, CreatedAt: 300}, {ID: 3, CreatedAt: 200}},
			desc:    false,
			wantIDs: []int64{1, 3, 2},
		},
		{
			name:    "stable order on ties",
			users:   []models.Profile{{ID: 1, CreatedAt: 50}, {ID: 2, CreatedAt: 50}, {ID: 3, CreatedAt: 50}},
			desc:    true,
			wantIDs: []int64{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortAdminUsers(tt.users, byCreatedAt, tt.desc)

			got := make([]int64, len(tt.users))
			for i, u := range tt.users {
				got[i] = u.ID
			}
			assert.Equal(t, tt.wantIDs, got)
		})
	}
}
