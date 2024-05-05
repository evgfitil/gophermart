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

type Storage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) error
	GetUserByUsername(ctx context.Context, username string) (string, error)
	GetUserID(ctx context.Context, username string) (int, error)
	GetOrders(ctx context.Context, userID int) ([]models.Order, error)
	IsUserUnique(ctx context.Context, username string) (bool, error)
	Ping(ctx context.Context) error
	ProcessOrder(ctx context.Context, order models.Order) error
}

func Ping(s Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := s.Ping(req.Context())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
