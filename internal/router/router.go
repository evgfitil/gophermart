package router

import (
	"github.com/evgfitil/gophermart.git/internal/api"
	"github.com/evgfitil/gophermart.git/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func ApiRouter(db database.DBStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Get("/ping", api.Ping(db))
	r.Route("/api/user", func(r chi.Router) {
		//r.Post("/register", api.Register)
	})
	return r
}
