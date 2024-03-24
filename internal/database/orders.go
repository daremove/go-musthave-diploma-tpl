package database

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
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
			status,
			uploaded_at
		FROM
		    orders
		WHERE
		    id = $1
	`
	SelectOrdersWithAccrualQuery = `
		SELECT
			o.id,
			user_id,
			status,
			uploaded_at,
			SUM(coalesce(amount, 0))
		FROM
		    orders o
			LEFT JOIN accrual_flow af ON o.id = af.order_id
		WHERE
		    user_id = $1
		GROUP BY 
		    o.id
	`
)

type OrderDB struct {
	ID         string
	UserId     string
	Status     OrderStatusDB
	UploadedAt time.Time
}

type OrderWithAccrualDB struct {
	OrderDB
	Accrual float64
}

type OrderStatusDB struct {
	models.OrderStatus
}

func (s *OrderStatusDB) Scan(value interface{}) error {
	strVal, ok := value.(string)

	if !ok {
		return fmt.Errorf("OrderStatus must be a string, got %T instead", value)
	}

	*s = OrderStatusDB{models.OrderStatus(strVal)}

	return nil
}

func (s *OrderStatusDB) Value() (driver.Value, error) {
	return string(s.OrderStatus), nil
}

func (d *Database) CreateOrder(ctx context.Context, orderId, userId string) error {
	if _, err := d.db.Exec(ctx, InsertOrderQuery, orderId, userId); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateOrder
		}

		return err
	}

	return nil
}

func (d *Database) FindOrder(ctx context.Context, orderId string) (*OrderDB, error) {
	order := &OrderDB{}

	if err := d.db.QueryRow(ctx, SelectOrderQuery, orderId).Scan(&order.ID, &order.UserId, &order.Status, &order.UploadedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return order, nil
}

func (d *Database) FindOrdersWithAccrual(ctx context.Context, userId string) (*[]OrderWithAccrualDB, error) {
	var result []OrderWithAccrualDB

	rows, err := d.db.Query(ctx, SelectOrdersWithAccrualQuery, userId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item OrderWithAccrualDB

		if err := rows.Scan(&item.ID, &item.UserId, &item.Status, &item.UploadedAt, &item.Accrual); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}
