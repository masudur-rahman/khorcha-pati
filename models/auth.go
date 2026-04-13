package models

// RefreshToken stores issued refresh tokens for reuse detection and revocation.
type RefreshToken struct {
	ID        int64  `db:"id,pk autoincr"`
	UserID    int64  `db:"user_id"`
	TokenUUID string `db:"token_uuid,uq"`
	ExpiresAt int64  `db:"expires_at"`
	Revoked   int64  `db:"revoked"` // 0 = active, 1 = revoked
	CreatedAt int64  `db:"created_at"`
}

// TableName returns the database table name.
func (RefreshToken) TableName() string {
	return "refresh_token"
}
