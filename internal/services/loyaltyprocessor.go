package services

import (
	"context"
	"github.com/evgfitil/gophermart.git/internal/models"
)

type OrderStorage interface {
	GetOrders(ctx context.Context, userID int) ([]models.Order, error)
	ProcessOrder(ctx context.Context, order models.Order) error

	/*
	   Новые методы:
	   1. GetNewOrders(ctx context.Context) ([]models.Order, error) - получение заказов со статусом NEW для обработки сервисом loyaltyprocessor
	   2. UpdateOrderAccrual(ctx context.Context, orderID int, accrual float64) error - обновление начислений баллов в заказах
	*/
}

type TransactionStorage interface {
	// AddTransaction добавляет новую транзакцию в бд
	AddTransaction(ctx context.Context, transaction models.Transaction) (models.Transaction, error)

	// GetTransactions Возвращает список транзакций пользователя
	GetTransactions(ctx context.Context, userID int) ([]models.Transaction, error)
}

// LoyaltyProcessorService опрашивает систему расчета начислений балоов, обновляет заказы и добавляет операцию зачисления баллов в таблицу транзакций

type LoyaltyProcessorService struct {
	AccrualUrl         string
	OrderStorage       OrderStorage
	TransactionStorage TransactionStorage
}

func NewLoyaltyProcessorService(URL string, os OrderStorage, ts TransactionStorage) *LoyaltyProcessorService {
	return &LoyaltyProcessorService{}
}

func (lps *LoyaltyProcessorService) CheckAccrual(ctx context.Context, orders []models.Order) {
}
