package api

import (
	"context"
	"encoding/json"
	"github.com/evgfitil/gophermart.git/internal/database"
	"github.com/evgfitil/gophermart.git/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func RegisterHandler(db database.DBStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var user models.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			http.Error(res, "invalid request body", http.StatusBadRequest)
			return
		}

		isUnique, err := db.IsUserUnique(requestContext, user.Username)
		if err != nil {
			http.Error(res, "database error", http.StatusInternalServerError)
			return
		}
		if !isUnique {
			http.Error(res, "user already exists", http.StatusConflict)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "error hashing password", http.StatusInternalServerError)
			return
		}

		err = db.CreateUser(requestContext, user.Username, string(hashedPassword))
		if err != nil {
			http.Error(res, "error creating user", http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
		res.Write([]byte("User registered successfully"))
	}
}
