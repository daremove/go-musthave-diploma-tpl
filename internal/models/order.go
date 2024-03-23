package models

import "time"

type OrderDB struct {
	ID         string
	UserId     string
	UploadedAt time.Time
}
