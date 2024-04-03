package database

import (
	"context"
	"time"
)

const (
	InsertWithdrawalQuery = `
		INSERT INTO
			withdrawal_flow (order_id, user_id, amount)
		VALUES ($1, $2, $3)
	`
	SelectWithdrawalFlowQuery = `
		SELECT
		    order_id,
			amount,
		    processed_at
		FROM
		    withdrawal_flow
		WHERE
		    user_id = $1
	`
)

type WithdrawalFlowItemDB struct {
	OrderID     string
	Amount      float64
	ProcessedAt time.Time
}

func (d *Database) CreateWithdrawal(ctx context.Context, orderID, userID string, amount float64) error {
	if _, err := d.db.Exec(ctx, InsertWithdrawalQuery, orderID, userID, amount); err != nil {
		return err
	}

	return nil
}

func (d *Database) FindWithdrawalFlow(ctx context.Context, userID string) (*[]WithdrawalFlowItemDB, error) {
	var result []WithdrawalFlowItemDB

	rows, err := d.db.Query(ctx, SelectWithdrawalFlowQuery, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item WithdrawalFlowItemDB

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
