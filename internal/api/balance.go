package api

import (
	"context"
	"encoding/json"
	"github.com/go-chi/jwtauth"
	"net/http"
)

func GetUserBalance(us UserStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		_, claims, err := jwtauth.FromContext(requestContext)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims == nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		username, ok := claims["user_id"].(string)
		if !ok {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := us.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}
		userBalance, err := us.GetUserBalance(requestContext, userID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(userBalance)
	}
}
