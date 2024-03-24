package database

import (
	"context"
)

const (
	InsertAccrualQuery = `
		INSERT INTO
			accrual_flow (order_id, amount)
		VALUES ($1, $2)
	`
)

func (d *Database) CreateAccrual(ctx context.Context, orderId string, amount float64) error {
	if _, err := d.db.Exec(ctx, InsertAccrualQuery, orderId, amount); err != nil {
		return err
	}

	return nil
}
