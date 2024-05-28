package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/go-chi/jwtauth"

	"github.com/evgfitil/gophermart.git/internal/apperrors"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func GetOrders(os OrderStorage, us UserStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		_, claims, err := jwtauth.FromContext(requestContext)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims == nil {
			http.Error(res, "no claims available", http.StatusUnauthorized)
			return
		}

		username, ok := claims["user_id"].(string)
		if !ok {
			http.Error(res, "No required claim available", http.StatusUnauthorized)
			return
		}

		userID, err := us.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			return
		}

		userOrders, err := os.GetOrders(requestContext, userID)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(userOrders)
	}
}

func UploadOrderHandler(os OrderStorage, us UserStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		_, claims, err := jwtauth.FromContext(requestContext)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims == nil {
			http.Error(res, "no claims available", http.StatusUnauthorized)
			return
		}

		username, ok := claims["user_id"].(string)
		if !ok {
			http.Error(res, "No required claim available", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(res, "failed to read the request body", http.StatusInternalServerError)
			return
		}

		orderNumber := string(body)
		if orderNumber == "" {
			http.Error(res, "empty orderNumber", http.StatusUnprocessableEntity)
			return
		}
		if err = goluhn.Validate(orderNumber); err != nil {
			http.Error(res, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		var userID int
		userID, err = us.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}

		order := models.Order{
			UserID:      userID,
			OrderNumber: orderNumber,
			Status:      "NEW",
			UploadedAt:  time.Now(),
		}

		if err = os.ProcessOrder(requestContext, order); err != nil {
			switch err {
			case apperrors.ErrOrderAlreadyExists:
				http.Error(res, "order already exists", http.StatusOK)
				return
			case apperrors.ErrOrderNumberTaken:
				http.Error(res, err.Error(), http.StatusConflict)
				return
			default:
				http.Error(res, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusAccepted)
		res.Write([]byte("upload in successfully"))
	}
}
