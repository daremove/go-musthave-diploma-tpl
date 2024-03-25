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
	UpdateOrderStatusQuery = `
		UPDATE
			orders
		SET
			status = $2
		WHERE
		    id = $1
	`
	SelectAllUnprocessedOrdersQuery = `
		SELECT
			id,
			user_id,
			status,
			uploaded_at
		FROM
		    orders
		WHERE
		    status NOT IN ('INVALID', 'PROCESSED')
	`
)

type OrderDB struct {
	ID         string
	UserID     string
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

func (s OrderStatusDB) Value() (driver.Value, error) {
	return string(s.OrderStatus), nil
}

func (d *Database) CreateOrder(ctx context.Context, orderID, userID string) error {
	if _, err := d.db.Exec(ctx, InsertOrderQuery, orderID, userID); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateOrder
		}

		return err
	}

	return nil
}

func (d *Database) FindOrder(ctx context.Context, orderID string) (*OrderDB, error) {
	order := &OrderDB{}

	if err := d.db.QueryRow(ctx, SelectOrderQuery, orderID).Scan(&order.ID, &order.UserID, &order.Status, &order.UploadedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return order, nil
}

func (d *Database) FindOrdersWithAccrual(ctx context.Context, userID string) (*[]OrderWithAccrualDB, error) {
	var result []OrderWithAccrualDB

	rows, err := d.db.Query(ctx, SelectOrdersWithAccrualQuery, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item OrderWithAccrualDB

		if err := rows.Scan(&item.ID, &item.UserID, &item.Status, &item.UploadedAt, &item.Accrual); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}

func (d *Database) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatusDB) error {
	if _, err := d.db.Exec(ctx, UpdateOrderStatusQuery, orderID, status); err != nil {
		return err
	}

	return nil
}

func (d *Database) FindAllUnprocessedOrders(ctx context.Context) (*[]OrderDB, error) {
	var result []OrderDB

	rows, err := d.db.Query(ctx, SelectAllUnprocessedOrdersQuery)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item OrderDB

		if err := rows.Scan(&item.ID, &item.UserID, &item.Status, &item.UploadedAt); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}
