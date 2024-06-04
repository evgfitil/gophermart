package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
	"time"
)

const (
	requestTimeout = 1 * time.Second
)

func Router(os OrderStorage, us UserStorage, bs BalanceStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", HandleUserRegistration(us))
		r.Post("/login", HandleUserLogin(us))
	})
	r.With(jwtauth.Authenticator).Route("/api/user/balance", func(r chi.Router) {
		r.Get("/", HandleGetUserBalance(bs))
		r.Post("/withdraw", HandleWithdrawBalance(bs))
	})
	r.With(jwtauth.Authenticator).Route("/api/user/orders", func(r chi.Router) {
		r.Post("/", HandleUploadOrder(os, us))
		r.Get("/", HandleGetUserOrders(os, us))
	})
	r.With(jwtauth.Authenticator).Route("/api/user/withdrawals", func(r chi.Router) {
		r.Get("/", HandleGetWithdrawals(bs))
	})
	return r
}
