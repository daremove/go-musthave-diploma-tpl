package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/database"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/logger"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type AccrualService struct {
	storage          accrualStorage
	jobQueueService  accrualJobQueueService
	externalEndpoint string
}

type accrualStorage interface {
	UpdateOrderStatus(ctx context.Context, orderId string, status database.OrderStatusDB) error

	CreateAccrual(ctx context.Context, orderId string, amount float64) error
}

type accrualJobQueueService interface {
	Enqueue(job Job)

	ScheduleJob(job Job, delay time.Duration)

	PauseAndResume(delay time.Duration)
}

func NewAccrualService(storage accrualStorage, jobQueueService accrualJobQueueService, externalEndpoint string) *AccrualService {
	return &AccrualService{
		storage:          storage,
		jobQueueService:  jobQueueService,
		externalEndpoint: externalEndpoint,
	}
}

func (as *AccrualService) CalculateAccrual(orderId string) {
	as.jobQueueService.Enqueue(func(ctx context.Context) {
		data, retryAfter, err := as.fetchAccrualData(orderId)

		if err != nil {
			if errors.Is(err, errNoOrder) {
				logger.Log.Info("order isn't registered", zap.String("orderId", orderId))
				// todo maybe need to enqueue
				return
			}

			logger.Log.Error("failed to fetch accrual data", zap.Error(err))
			// todo add timeout enqueue
			return
		}

		if retryAfter > 0 {
			logger.Log.Info("got retryAfter", zap.Int("retryAfter", retryAfter), zap.String("orderId", orderId))
			as.jobQueueService.PauseAndResume(time.Second * time.Duration(retryAfter))
			as.jobQueueService.Enqueue(func(ctx context.Context) {
				as.CalculateAccrual(orderId)
			})
			logger.Log.Info("enqueued new job after pausing", zap.Int("retryAfter", retryAfter), zap.String("orderId", orderId))
			return
		}

		logger.Log.Info("got accrual data",
			zap.String("orderId", orderId),
			zap.String("status", string(data.Status)),
		)

		if data.Status == AccrualStatusRegistered {
			as.jobQueueService.ScheduleJob(func(ctx context.Context) {
				as.CalculateAccrual(orderId)
			}, time.Minute)
			logger.Log.Info("enqueued new schedule job", zap.String("orderId", orderId))

			return
		}

		if data.Status == AccrualStatusProcessed ||
			data.Status == AccrualStatusProcessing ||
			data.Status == AccrualStatusInvalid {
			if err := as.storage.UpdateOrderStatus(ctx, orderId, database.OrderStatusDB{OrderStatus: models.OrderStatus(data.Status)}); err != nil {
				logger.Log.Error("failed to update status", zap.Error(err))
				return
			}

			logger.Log.Info("updated order status",
				zap.String("orderId", orderId),
				zap.String("status", string(data.Status)),
			)

			if data.Accrual == nil {
				return
			}

			if err := as.storage.CreateAccrual(ctx, orderId, *data.Accrual); err != nil {
				logger.Log.Error("failed to create accrual", zap.Error(err))
				return
			}

			logger.Log.Info("saved accrual value",
				zap.String("orderId", orderId),
				zap.String("status", string(data.Status)),
				zap.Float64("accrual", *data.Accrual),
			)

			return
		}

		logger.Log.Error("status isn't defined", zap.String("status", string(data.Status)))
	})
}

type accrualOrderStatus string

const (
	AccrualStatusRegistered accrualOrderStatus = "REGISTERED"
	AccrualStatusInvalid    accrualOrderStatus = "INVALID"
	AccrualStatusProcessing accrualOrderStatus = "PROCESSING"
	AccrualStatusProcessed  accrualOrderStatus = "PROCESSED"
)

type accrualDataResponse struct {
	ID      string             `json:"order"`
	Status  accrualOrderStatus `json:"status"`
	Accrual *float64           `json:"accrual,omitempty"`
}

var (
	errNoOrder = errors.New("order isn't registered")
	errServer  = errors.New("internal server error")
)

const defaultRetryAfterDuration = 60

func (as *AccrualService) fetchAccrualData(orderId string) (data *accrualDataResponse, retryAfter int, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/orders/%s", as.externalEndpoint, orderId), nil)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := client.Do(req)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to send data by using GET method: %w", err)
	}

	if res.StatusCode == http.StatusNoContent {
		return nil, 0, errNoOrder
	}

	if res.StatusCode == http.StatusTooManyRequests {
		retryAfter, err := strconv.Atoi(res.Header.Get("Retry-After"))

		if err != nil {
			retryAfter = defaultRetryAfterDuration
		}

		return nil, retryAfter, nil
	}

	if res.StatusCode == http.StatusInternalServerError {
		return nil, 0, errServer
	}

	var parsedData accrualDataResponse
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(res.Body); err != nil {
		return nil, 0, fmt.Errorf("failed to read from response body: %w", err)
	}

	if err := json.Unmarshal(buf.Bytes(), &parsedData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &parsedData, 0, nil
}
