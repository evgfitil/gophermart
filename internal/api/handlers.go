package api

import (
	"context"
	"github.com/evgfitil/gophermart.git/internal/models"
	"net/http"
	"time"
)

const (
	requestTimeout = 1 * time.Second
)

type UserStorage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) error
	GetUserByUsername(ctx context.Context, username string) (string, error)
	GetUserID(ctx context.Context, username string) (int, error)
	IsUserUnique(ctx context.Context, username string) (bool, error)

	/*
		Новые методы:
		1. GetUserBalance(ctx context.Context, userID int) (float64, error) - метод для получения баланса пользователем
		2. WithdrawBalance(ctx context.Context, userID int, amount float64) error - метод для списания баллов с проверкой их баланса
	*/

}

type OrderStorage interface {
	GetOrders(ctx context.Context, userID int) ([]models.Order, error)
	ProcessOrder(ctx context.Context, order models.Order) error

	/*
	   Новые методы:
	   1. UpdateOrderAccrual(ctx context.Context, orderID int, accrual float64) error - обновление начислений баллов в заказах
	*/
}

type TransactionStorage interface {
	// AddTransaction добавляет новую транзакцию в бд
	AddTransaction(ctx context.Context, transaction models.Transaction) (models.Transaction, error)

	// GetTransactions возвращает список транзакций пользователя
	GetTransactions(ctx context.Context, userID int) ([]models.Transaction, error)
}

// GetBalanceHandler возвращает баланс пользователя
func GetBalanceHandler(us UserStorage, ts TransactionStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

// WithdrawBalanceHandler запрос на списание средств
func WithdrawBalanceHandler(us UserStorage, ts TransactionStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

// GetWithdrawalsHandler возвращает список транзакций пользователя
func GetWithdrawalsHandler(ts TransactionStorage, us UserStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
