package access

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"
	"github.com/masudur-rahman/khorcha-pati/services"
)

// accessService keeps settings and the allowlist in memory (write-through on
// mutation) so per-message checks never hit the DB.
type accessService struct {
	repo   repos.AccessRepository
	logger logr.Logger

	mu         sync.RWMutex
	restricted bool
	replyText  string
	owner      string
	allowed    []models.AllowedUser
}

func NewAccessService(repo repos.AccessRepository, logger logr.Logger) *accessService {
	return &accessService{repo: repo, logger: logger}
}

// EnsureSeeded applies the config bootstrap exactly once (guarded by the
// seeded marker), then loads DB state into memory. Owner applies every boot.
func (s *accessService) EnsureSeeded(seed services.AccessSeed) error {
	s.mu.Lock()
	s.owner = strings.TrimPrefix(seed.Owner, "@")
	s.mu.Unlock()

	_, seeded, err := s.repo.GetSetting(models.SettingAccessSeeded)
	if err != nil {
		return fmt.Errorf("read seed marker: %w", err)
	}
	if !seeded {
		if err := s.seedFromConfig(seed); err != nil {
			return err
		}
	}
	return s.reload()
}

func (s *accessService) seedFromConfig(seed services.AccessSeed) error {
	for _, raw := range seed.AllowedUsers {
		entry := parseAllowedEntry(raw)
		if entry.Username == "" && entry.TelegramID == 0 {
			continue
		}
		if err := s.repo.AddAllowedUser(&entry); err != nil {
			return fmt.Errorf("seed allowed user %q: %w", raw, err)
		}
	}
	if err := s.repo.SetSetting(models.SettingAllowedUsersOnly, strconv.FormatBool(seed.Restricted)); err != nil {
		return err
	}
	if err := s.repo.SetSetting(models.SettingRestrictedReplyText, seed.ReplyText); err != nil {
		return err
	}
	return s.repo.SetSetting(models.SettingAccessSeeded, "true")
}

// parseAllowedEntry turns a config entry into an allowlist row: numeric
// entries are Telegram IDs, anything else is a username (optional leading @).
func parseAllowedEntry(raw string) models.AllowedUser {
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "@"))
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return models.AllowedUser{TelegramID: id}
	}
	return models.AllowedUser{Username: raw}
}

func (s *accessService) reload() error {
	restrictedVal, _, err := s.repo.GetSetting(models.SettingAllowedUsersOnly)
	if err != nil {
		return err
	}
	replyText, _, err := s.repo.GetSetting(models.SettingRestrictedReplyText)
	if err != nil {
		return err
	}
	allowed, err := s.repo.ListAllowedUsers()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.restricted = restrictedVal == "true"
	s.replyText = replyText
	s.allowed = allowed
	return nil
}

func (s *accessService) IsRestricted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.restricted
}

func (s *accessService) RestrictedReplyText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.replyText
}

func (s *accessService) IsUserAllowed(username string, telegramID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.owner != "" && strings.EqualFold(username, s.owner) {
		return true
	}
	for _, e := range s.allowed {
		if e.TelegramID != 0 && e.TelegramID == telegramID {
			return true
		}
		if e.Username != "" && strings.EqualFold(e.Username, username) {
			return true
		}
	}
	return false
}

// NoteSeen pins a username-matched entry to the user's Telegram ID so a later
// username change can't break (or leak) their access.
func (s *accessService) NoteSeen(username string, telegramID int64) {
	if username == "" || telegramID == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.allowed {
		if s.allowed[i].TelegramID == 0 && strings.EqualFold(s.allowed[i].Username, username) {
			s.allowed[i].TelegramID = telegramID
			if err := s.repo.UpdateAllowedUser(&s.allowed[i]); err != nil {
				s.logger.Errorw("backfill allowed user telegram id", "username", username, "error", err.Error())
			}
			return
		}
	}
}

func (s *accessService) SetRestricted(v bool) error {
	if err := s.repo.SetSetting(models.SettingAllowedUsersOnly, strconv.FormatBool(v)); err != nil {
		return err
	}
	s.mu.Lock()
	s.restricted = v
	s.mu.Unlock()
	return nil
}

func (s *accessService) SetReplyText(text string) error {
	if err := s.repo.SetSetting(models.SettingRestrictedReplyText, text); err != nil {
		return err
	}
	s.mu.Lock()
	s.replyText = text
	s.mu.Unlock()
	return nil
}

func (s *accessService) ListAllowedUsers() []models.AllowedUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.AllowedUser, len(s.allowed))
	copy(out, s.allowed)
	return out
}

func (s *accessService) Allow(username string, telegramID int64) (*models.AllowedUser, error) {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if username == "" && telegramID == 0 {
		return nil, fmt.Errorf("username or telegram id required")
	}
	if s.IsUserAllowed(username, telegramID) {
		return nil, fmt.Errorf("user already allowed")
	}

	entry := models.AllowedUser{Username: username, TelegramID: telegramID}
	if err := s.repo.AddAllowedUser(&entry); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.allowed = append(s.allowed, entry)
	s.mu.Unlock()
	return &entry, nil
}

func (s *accessService) Revoke(id int64) error {
	if err := s.repo.RemoveAllowedUser(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.allowed {
		if s.allowed[i].ID == id {
			s.allowed = append(s.allowed[:i], s.allowed[i+1:]...)
			break
		}
	}
	return nil
}
