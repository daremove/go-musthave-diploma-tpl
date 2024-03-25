package database

import (
	"context"
	"time"
)

const (
	InsertAccrualQuery = `
		INSERT INTO
			accrual_flow (order_id, amount)
		VALUES ($1, $2)
	`
	SelectAccrualFlowQuery = `
		SELECT
		    order_id,
			amount,
		    processed_at
		FROM
		    accrual_flow af
			LEFT JOIN orders o ON af.order_id = o.id
		WHERE
		    user_id = $1
	`
)

type AccrualFlowItemDB struct {
	OrderID     string
	Amount      float64
	ProcessedAt time.Time
}

func (d *Database) CreateAccrual(ctx context.Context, orderID string, amount float64) error {
	if _, err := d.db.Exec(ctx, InsertAccrualQuery, orderID, amount); err != nil {
		return err
	}

	return nil
}

func (d *Database) FindAccrualFlow(ctx context.Context, userID string) (*[]AccrualFlowItemDB, error) {
	var result []AccrualFlowItemDB

	rows, err := d.db.Query(ctx, SelectAccrualFlowQuery, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item AccrualFlowItemDB

		if err := rows.Scan(&item.OrderID, &item.Amount, &item.ProcessedAt); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}
