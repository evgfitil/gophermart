package api

import (
	"context"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/evgfitil/gophermart.git/internal/apperrors"
	"github.com/evgfitil/gophermart.git/internal/models"
	"github.com/go-chi/jwtauth"
	"io"
	"net/http"
	"time"
)

func UploadOrderHandler(s Storage) http.HandlerFunc {
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
			http.Error(res, err.Error(), http.StatusUnauthorized)
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
		userID, err = s.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}

		order := models.Order{
			UserID:      userID,
			OrderNumber: orderNumber,
			Status:      "NEW",
			UploadedAt:  time.Now(),
		}

		if err = s.ProcessOrder(requestContext, order); err != nil {
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
