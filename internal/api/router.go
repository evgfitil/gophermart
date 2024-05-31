package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
)

func Router(os OrderStorage, us UserStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", RegisterHandler(us))
		r.Post("/login", AuthHandler(us))
	})
	r.With(jwtauth.Authenticator).Route("/api/user/orders", func(r chi.Router) {
		r.Post("/", UploadOrderHandler(os, us))
		r.Get("/", GetOrders(os, us))
	})
	return r
}
