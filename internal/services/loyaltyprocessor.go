package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/evgfitil/gophermart.git/internal/models"
)

const (
	defaultRetrySeconds = 1
)

type OrderStorage interface {
	GetNewOrders(ctx context.Context) ([]models.Order, error)
	GetOrders(ctx context.Context, userID int) ([]models.Order, error)
	ProcessOrder(ctx context.Context, order models.Order) error
	UpdateOrderAccrual(ctx context.Context, orderNumber string, accrual float64) error
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
}

type TransactionStorage interface {
	//// AddTransaction добавляет новую транзакцию в бд
	//AddTransaction(ctx context.Context, transaction models.Transaction) (models.Transaction, error)
	//
	//// GetTransactions Возвращает список транзакций пользователя
	//GetTransactions(ctx context.Context, userID int) ([]models.Transaction, error)
}

// LoyaltyProcessorService опрашивает систему расчета начислений балоов, обновляет заказы и добавляет операцию зачисления баллов в таблицу транзакций

type LoyaltyProcessorService struct {
	AccrualURL         string
	OrderStorage       OrderStorage
	TransactionStorage TransactionStorage
	client             *resty.Client
}

func NewLoyaltyProcessorService(URL string, os OrderStorage, ts TransactionStorage) *LoyaltyProcessorService {
	client := resty.New()
	return &LoyaltyProcessorService{
		AccrualURL:         URL,
		OrderStorage:       os,
		TransactionStorage: ts,
		client:             client,
	}
}

func (lps *LoyaltyProcessorService) updateOrder(ctx context.Context, order models.Order) error {
	switch order.Status {
	case "PROCESSED":
		fmt.Println("PROCESSED")
		if err := lps.OrderStorage.UpdateOrderAccrual(ctx, order.ID, order.Accrual); err != nil {
			logger.Sugar.Errorln("update order accrual failed", err)
			return err
		}
	default:
		if err := lps.OrderStorage.UpdateOrderStatus(ctx, order.ID, order.Status); err != nil {
			logger.Sugar.Errorln("update order status failed", err)
			return err
		}
	}
	return nil
}

func (lps *LoyaltyProcessorService) CheckAccrual(ctx context.Context, orders []models.Order) {
	for _, order := range orders {
		retryCount := 0
		resp, err := lps.client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			Get(lps.AccrualURL + "/api/orders/" + order.OrderNumber)
		if err != nil {
			logger.Sugar.Errorln("Error making request to accrual service: ", err)
			continue
		}

		if resp.StatusCode() == http.StatusNoContent {
			logger.Sugar.Infof("Order %s is not registered in accrual service", order.OrderNumber)
			continue
		}

		if resp.StatusCode() == http.StatusTooManyRequests {
			retryAfter := resp.Header().Get("Retry-After")
			var retrySeconds int
			retrySeconds, err = strconv.Atoi(retryAfter)
			if err != nil {
				logger.Sugar.Errorln("Error converting Retry-After header to int: ", err)
				retrySeconds = defaultRetrySeconds
			}
			time.Sleep(time.Duration(retrySeconds) * time.Second)
		}

		if resp.StatusCode() == http.StatusInternalServerError {
			delay := time.Duration(math.Pow(2, float64(retryCount))) * time.Second
			logger.Sugar.Errorf("internal server error: error making request to accrual service, retrying in %v", delay)
			time.Sleep(delay)
			retryCount++
		}

		var result struct {
			Order   string  `json:"order"`
			Status  string  `json:"status"`
			Accrual float64 `json:"accrual"`
		}

		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			logger.Sugar.Errorln("Error unmarshalling response from accrual service: ", err)
		}
		logger.Sugar.Infoln("Processed order ", order.OrderNumber, " with status ", result.Status)

		if err = lps.updateOrder(ctx, order); err != nil {
			logger.Sugar.Errorf("error updating order %s with status %s", order.OrderNumber, result.Status)
		}
	}
}

func (lps *LoyaltyProcessorService) Start(ctx context.Context, interval time.Duration) {
	logger.Sugar.Infoln("Starting loyaltyProcessorService")
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				orders, err := lps.OrderStorage.GetNewOrders(ctx)
				if err != nil {
					logger.Sugar.Errorln("Error fetching new orders: ", err)
					continue
				}
				if len(orders) == 0 {
					continue
				}
				lps.CheckAccrual(ctx, orders)
			}
		}
	}()
}
