package database

import (
	"context"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
)

func (db *Database) SaveUser(ctx context.Context, login, hash string) error {
	return nil
}

func (db *Database) FindUser(ctx context.Context, login string) (*models.User, error) {
	l := "alex"
	p := "empty"
	return &models.User{Login: &l, Password: &p}, nil
}
