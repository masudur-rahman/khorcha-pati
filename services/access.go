package services

import (
	"github.com/masudur-rahman/khorcha-pati/models"
)

// AccessSeed carries the config-file bootstrap values for access control.
// It is applied once (first boot); afterwards the DB is the source of truth.
// Owner is re-applied every boot and is always allowed, without an allowlist row.
type AccessSeed struct {
	Restricted   bool
	AllowedUsers []string
	ReplyText    string
	Owner        string
}

// AccessService gates restricted (stage/dev) instances to allowed users only.
type AccessService interface {
	EnsureSeeded(seed AccessSeed) error
	IsRestricted() bool
	RestrictedReplyText() string
	IsUserAllowed(username string, telegramID int64) bool
	// NoteSeen backfills the Telegram ID of a username-only allowlist entry.
	NoteSeen(username string, telegramID int64)
	SetRestricted(v bool) error
	SetReplyText(text string) error
	ListAllowedUsers() []models.AllowedUser
	Allow(username string, telegramID int64) (*models.AllowedUser, error)
	Revoke(id int64) error
}
