package models

type User struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

// todo mode to db layer
type UserDB struct {
	ID    string
	Login string
	Hash  string
}
