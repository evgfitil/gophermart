package api

import (
	"context"
	"net/http"
	"time"
)

const (
	requestTimeout = 1 * time.Second
)

type Storage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) error
	GetUserByUsername(ctx context.Context, username string) (string, error)
	IsUserUnique(ctx context.Context, username string) (bool, error)
	Ping(ctx context.Context) error
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
