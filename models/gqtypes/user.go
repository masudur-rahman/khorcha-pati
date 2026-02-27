package gqtypes

type Profile struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Avatar    string `json:"avatar"`
	IsActive  bool   `json:"isActive"`
	IsAdmin   bool   `json:"isAdmin"`
}

type UserParams struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Avatar    string `json:"avatar"`
}
