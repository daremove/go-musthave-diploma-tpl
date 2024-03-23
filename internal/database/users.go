package database

import (
	"context"
	"errors"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicateUser = errors.New("user is duplicated")
)

const (
	InsertUserQuery = `
		INSERT INTO
			users (login, hash)
		VALUES ($1, $2)
	`
	SelectUserQuery = `
		SELECT
			login,
			hash
		FROM
		    users
		WHERE
		    login = $1
	`
)

func (d *Database) SaveUser(ctx context.Context, user models.UserWithHash) error {
	if _, err := d.db.Exec(ctx, InsertUserQuery, user.Login, user.Hash); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateUser
		}

		return err
	}

	return nil
}

func (d *Database) FindUser(ctx context.Context, login string) (*models.UserWithHash, error) {
	user := &models.UserWithHash{}

	if err := d.db.QueryRow(ctx, SelectUserQuery, login).Scan(&user.Login, &user.Hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	return user, nil
}
