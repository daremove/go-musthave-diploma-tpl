package models

type User struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

type UserDB struct {
	ID    string
	Login string
	Hash  string
}
