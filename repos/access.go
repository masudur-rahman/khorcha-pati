package repos

import (
	"github.com/masudur-rahman/khorcha-pati/models"
)

// AccessRepository persists instance settings and the allowed-users list.
type AccessRepository interface {
	GetSetting(key string) (string, bool, error)
	SetSetting(key, value string) error
	// SetSettingIfAbsent writes the value only when the key has no row yet,
	// so restarts never overwrite admin edits.
	SetSettingIfAbsent(key, value string) error

	// ListAllowedUsers returns all rows, including revoked tombstones.
	ListAllowedUsers() ([]models.AllowedUser, error)
	AddAllowedUser(entry *models.AllowedUser) error
	UpdateAllowedUser(entry *models.AllowedUser) error
}
