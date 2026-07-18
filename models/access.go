package models

// Setting keys for instance-level runtime configuration. Values are seeded
// from the config file on first boot only; afterwards the DB (edited via the
// admin dashboard) is the sole source of truth.
const (
	SettingAllowedUsersOnly    = "allowed_users_only"
	SettingRestrictedReplyText = "restricted_reply_text"
	// SettingAccessSeeded marks that the one-time config seed already ran, so
	// restarts never resurrect config state over admin edits.
	SettingAccessSeeded = "access_seeded"
)

// Setting is a single key/value row of instance runtime configuration.
type Setting struct {
	ID    int64  `db:"id,pk" json:"id"`
	Key   string `db:"key,uq" json:"key"`
	Value string `db:"value" json:"value"`
}

func (Setting) TableName() string {
	return "setting"
}

// AllowedUser is an allowlist entry for restricted (stage/dev) instances.
// TelegramID 0 means the entry was added by username only and is backfilled
// once that user interacts with the bot.
type AllowedUser struct {
	ID         int64  `db:"id,pk" json:"id"`
	TelegramID int64  `db:"telegram_id" json:"telegramId"`
	Username   string `db:"username" json:"username"`
	CreatedAt  int64  `db:"created_at" json:"createdAt"`
}

func (AllowedUser) TableName() string {
	return "allowed_user"
}
