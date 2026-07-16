package models

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
