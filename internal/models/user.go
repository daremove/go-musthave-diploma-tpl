package models

type User struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

type UserWithHash struct {
	Login string
	Hash  string
}
