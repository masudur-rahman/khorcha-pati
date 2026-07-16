package models

import "strings"

// The last 8 digits identify a line regardless of country code / leading zero,
// so lookups match +8801712345678, 8801712345678, and 01712345678 alike —
// including countries with 8-digit subscriber numbers. Collisions are resolved
// by full-number verification in repos.FindUserByIdentifier.
const phoneSuffixDigits = 8

// NormalizePhoneNumber strips formatting (+, spaces, dashes), leaving digits only.
func NormalizePhoneNumber(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// PhoneSuffix returns the canonical lookup key for a phone number: its last
// 8 digits. Stored in Profile.MobileSuffix so the DB matches numbers exactly
// whether or not a country code was typed.
func PhoneSuffix(s string) string {
	n := NormalizePhoneNumber(s)
	if len(n) > phoneSuffixDigits {
		return n[len(n)-phoneSuffixDigits:]
	}
	return n
}

// PhoneNumbersMatch verifies a suffix-key candidate: after normalization the
// longer number must end with the shorter one, so a same-suffix number from
// another country never passes.
func PhoneNumbersMatch(a, b string) bool {
	na, nb := NormalizePhoneNumber(a), NormalizePhoneNumber(b)
	if na == "" || nb == "" {
		return false
	}
	return strings.HasSuffix(na, nb) || strings.HasSuffix(nb, na)
}

type Contacts struct {
	ID               int64   `db:"id,pk" json:"id"`
	UserID           int64   `db:",uqs" json:"userId"`
	NickName         string  `db:",uqs" json:"nickName"`
	FullName         string  `json:"fullName"`
	Email            string  `db:"email,uqs" json:"email"`
	ContactInfo      string  `json:"contactInfo"`
	NetBalance       float64 `db:"net_balance" json:"netBalance"`
	LastTxnTimestamp int64   `db:"last_txn_timestamp" json:"lastTxnTimestamp"`
	CreatedAt        int64   `db:"created_at" json:"createdAt"`
}

func (Contacts) TableName() string {
	return "contacts"
}

// Dashboard theme preferences stored on the profile. Empty means unset —
// the dashboard falls back to the device color scheme.
const (
	ThemeLight = "light"
	ThemeDark  = "dark"
)

type Profile struct {
	ID           int64  `db:"id,pk" json:"id"`
	TelegramID   int64  `db:",uq" json:"telegramId"`
	Username     string `db:",uq" json:"username"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Timezone     string `db:"timezone" json:"timezone"`
	MobileNumber string `db:"mobile_number" json:"mobileNumber"`
	// MobileSuffix is the last-8-digits lookup key for MobileNumber,
	// maintained by the user repo on every write.
	MobileSuffix string `db:"mobile_suffix" json:"-"`
	Theme        string `db:"theme" json:"theme"`
	IsAdmin      bool   `db:"is_admin" json:"isAdmin"`
	IsActive     bool   `db:"is_active" json:"isActive"`
	CreatedAt    int64  `db:"created_at" json:"createdAt"`
}

func (Profile) TableName() string {
	return "profile"
}

//func (u *Contacts) APIFormat() gqtypes.Contacts {
//	return gqtypes.Contacts{
//		ID:        u.ID,
//		Username:  u.Username,
//		Email:     u.Email,
//		FirstName: u.FirstName,
//		LastName:  u.LastName,
//		//FullName: fmt.Sprintf("%s %s", u.FirstName, u.LastName),
//		Bio:      u.Bio,
//		Location: u.Location,
//		Avatar:   u.Avatar,
//		IsActive: u.IsActive,
//		IsAdmin:  u.IsAdmin,
//	}
//}
