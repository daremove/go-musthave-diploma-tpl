package database

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicateAccrual = errors.New("accrual is duplicated")
)

const (
	InsertAccrualQuery = `
		INSERT INTO
			accruals (id)
		VALUES ($1)
	`
)

func createAccrual(ctx context.Context, db DBExecutor, orderId string) error {
	if _, err := db.Exec(ctx, InsertAccrualQuery, orderId); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateAccrual
		}

		return err
	}

	return nil
}
