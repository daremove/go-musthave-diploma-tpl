package services

import (
	"context"
	"errors"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/database"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"strconv"
)

var (
	ErrDuplicateOrder               = errors.New("order is duplicated")
	ErrDuplicateOrderByOriginalUser = errors.New("order is duplicated by the same user")
)

type OrderService struct {
	storage orderStorage
}

type orderStorage interface {
	CreateOrder(ctx context.Context, orderId string, userId string) error

	FindOrder(ctx context.Context, orderId string) (*models.OrderDB, error)
}

func NewOrderService(storage orderStorage) *OrderService {
	return &OrderService{storage}
}

func (o *OrderService) VerifyOrderId(orderId string) bool {
	var sum int
	var alternate bool

	for i := len(orderId) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(orderId[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func (o *OrderService) CreateOrder(ctx context.Context, orderId, userId string) error {
	if err := o.storage.CreateOrder(ctx, orderId, userId); err != nil {
		if !errors.Is(err, database.ErrDuplicateOrder) {
			return err
		}

		order, errOrder := o.storage.FindOrder(ctx, orderId)

		if errOrder != nil {
			return errOrder
		}

		if order.UserId == userId {
			return ErrDuplicateOrderByOriginalUser
		}

		return ErrDuplicateOrder
	}

	return nil
}
