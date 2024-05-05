package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
)

func Router(s Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(tokenAuth))
	r.With(jwtauth.Authenticator).Get("/ping", Ping(s))
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", RegisterHandler(s))
		r.Post("/login", AuthHandler(s))
	})
	r.With(jwtauth.Authenticator).Route("/api/user/orders", func(r chi.Router) {
		r.Post("/", UploadOrderHandler(s))
		r.Get("/", GetOrders(s))
	})
	return r
}
