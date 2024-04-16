package api

import (
	"github.com/evgfitil/gophermart.git/internal/database"
	"net/http"
)

func Ping(db database.DBStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := db.Ping(req.Context())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
