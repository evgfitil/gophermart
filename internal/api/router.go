package api

import (
	"github.com/evgfitil/gophermart.git/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
)

func Router(db database.DBStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(jwtauth.Verifier(tokenAuth))
	r.With(jwtauth.Authenticator).Get("/ping", Ping(db))
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", RegisterHandler(db))
		r.Post("/login", AuthHandler(db))
	})
	return r
}
