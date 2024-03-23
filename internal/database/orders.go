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
	ErrDuplicateOrder = errors.New("order is duplicated")
)

const (
	InsertOrderQuery = `
		INSERT INTO
			orders (id, user_id)
		VALUES ($1, $2)
	`
	SelectOrderQuery = `
		SELECT
		    id,
			user_id,
			uploaded_at
		FROM
		    orders
		WHERE
		    id = $1
	`
)

func (d *Database) CreateOrder(ctx context.Context, orderId, userId string) error {
	tx, err := d.db.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, InsertOrderQuery, orderId, userId); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateOrder
		}

		return err
	}

	if err := createAccrual(ctx, tx, orderId); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (d *Database) FindOrder(ctx context.Context, orderId string) (*models.OrderDB, error) {
	order := &models.OrderDB{}

	if err := d.db.QueryRow(ctx, SelectOrderQuery, orderId).Scan(&order.ID, &order.UserId, &order.UploadedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return order, nil
}
