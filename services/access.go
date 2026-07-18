package services

import (
	"github.com/masudur-rahman/khorcha-pati/models"
)

// AccessSeed carries the config-file bootstrap values for access control,
// applied additively on every boot: allowlist entries already in the table
// (active or revoked) are skipped, and settings keys are insert-if-absent —
// admin edits and revocations always survive restarts.
// Owner is in-memory only and always allowed, without an allowlist row.
type AccessSeed struct {
	Restricted   bool
	AllowedUsers []string
	ReplyText    string
	Owner        string
}

// AccessService gates restricted (stage/dev) instances to allowed users only.
type AccessService interface {
	Seed(seed AccessSeed) error
	IsRestricted() bool
	RestrictedReplyText() string
	IsUserAllowed(username string, telegramID int64) bool
	// NoteSeen backfills the Telegram ID of a username-only allowlist entry.
	NoteSeen(username string, telegramID int64)
	SetRestricted(v bool) error
	SetReplyText(text string) error
	// ListAllowedUsers returns active entries; includeRevoked adds tombstones.
	ListAllowedUsers(includeRevoked bool) []models.AllowedUser
	// Allow adds a new entry, or restores a matching revoked one.
	Allow(username string, telegramID int64) (*models.AllowedUser, error)
	// Revoke tombstones an entry — the row is kept so the user stays visible
	// and config seeding can never resurrect their access.
	Revoke(id int64) error
	Restore(id int64) error
}
