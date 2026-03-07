package models

type Contacts struct {
	ID               int64  `db:"id,pk"`
	UserID           int64  `db:",uqs"`
	NickName         string `db:",uqs"`
	FullName         string
	Email            string `db:"email,uqs"`
	ContactInfo      string
	NetBalance       float64
	LastTxnTimestamp int64
}

func (Contacts) TableName() string {
	return "contacts"
}

type Profile struct {
	ID         int64  `db:"id,pk"`
	TelegramID int64  `db:",uq"`
	Username   string `db:",uq"`
	FirstName  string
	LastName   string
	Timezone   string `db:"timezone"`
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
