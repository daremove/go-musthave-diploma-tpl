package services

import (
	"context"
	"errors"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/database"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/utils"
	"sort"
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

	FindOrder(ctx context.Context, orderId string) (*database.OrderDB, error)

	FindOrdersWithAccrual(ctx context.Context, userId string) (*[]database.OrderWithAccrualDB, error)
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

func (o *OrderService) GetOrders(ctx context.Context, userId string) ([]models.Order, error) {
	orders, err := o.storage.FindOrdersWithAccrual(ctx, userId)

	if err != nil {
		return []models.Order{}, err
	}

	if orders == nil {
		return []models.Order{}, nil
	}

	var result []models.Order

	for _, order := range *orders {
		accrual := order.Accrual
		result = append(result, models.Order{
			ID:         order.ID,
			Status:     order.Status.OrderStatus,
			UploadedAt: utils.RFC3339Date{Time: order.UploadedAt},
			Accrual:    &accrual,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].UploadedAt.Time.Before(result[j].UploadedAt.Time)
	})

	return result, nil
}
