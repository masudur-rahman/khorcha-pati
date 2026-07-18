package repos

import (
	"github.com/masudur-rahman/khorcha-pati/models"
)

// AccessRepository persists instance settings and the allowed-users list.
type AccessRepository interface {
	GetSetting(key string) (string, bool, error)
	SetSetting(key, value string) error

	ListAllowedUsers() ([]models.AllowedUser, error)
	AddAllowedUser(entry *models.AllowedUser) error
	UpdateAllowedUser(entry *models.AllowedUser) error
	RemoveAllowedUser(id int64) error
}
