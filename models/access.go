package models

// Setting keys for instance-level runtime configuration. Config values are
// seeded insert-if-absent — an existing row (i.e. an admin edit) is never
// overwritten on restart.
const (
	SettingAllowedUsersOnly    = "allowed_users_only"
	SettingRestrictedReplyText = "restricted_reply_text"
)

// Setting is a single key/value row of instance runtime configuration.
type Setting struct {
	ID    int64  `db:"id,pk" json:"id"`
	Key   string `db:"key,uq" json:"key"`
	Value string `db:"value" json:"value"`
}

func (Setting) TableName() string {
	return "settings"
}

// AllowedUser is an allowlist entry for restricted (stage/dev) instances.
// TelegramID 0 means the entry was added by username only and is backfilled
// once that user interacts with the bot. Revoked rows are tombstones: the
// user stays visible in the admin panel and config seeding can never
// resurrect them (a matching row — active or revoked — is skipped).
type AllowedUser struct {
	ID         int64  `db:"id,pk" json:"id"`
	TelegramID int64  `db:"telegram_id" json:"telegramId"`
	Username   string `db:"username" json:"username"`
	Revoked    bool   `db:"revoked,req" json:"revoked"`
	RevokedAt  int64  `db:"revoked_at" json:"revokedAt"`
	CreatedAt  int64  `db:"created_at" json:"createdAt"`
}

func (AllowedUser) TableName() string {
	return "allowed_users"
}
