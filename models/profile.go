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
}

func (Contacts) TableName() string {
	return "contacts"
}

type Profile struct {
	ID           int64  `db:"id,pk" json:"id"`
	TelegramID   int64  `db:",uq" json:"telegramId"`
	Username     string `db:",uq" json:"username"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Timezone     string `db:"timezone" json:"timezone"`
	MobileNumber string `db:"mobile_number" json:"mobileNumber"`
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
