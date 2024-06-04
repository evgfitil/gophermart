package api

import (
	"context"
	"encoding/json"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/go-chi/jwtauth"
	"net/http"

	"github.com/evgfitil/gophermart.git/internal/apperrors"
	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/evgfitil/gophermart.git/internal/models"
)

type transactionRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type BalanceStorage interface {
	GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID int) (*models.Balance, error)
	GetUserID(ctx context.Context, username string) (int, error)
	WithdrawUserBalance(ctx context.Context, transaction *models.Transaction) error
}

func HandleGetUserBalance(bs BalanceStorage) http.HandlerFunc {
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
			http.Error(res, "no required claims available", http.StatusUnauthorized)
			return
		}

		userID, err := bs.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}
		userBalance, err := bs.GetUserBalance(requestContext, userID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(userBalance)
	}
}

func HandleGetWithdrawals(bs BalanceStorage) http.HandlerFunc {
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
			http.Error(res, "no required claims available", http.StatusUnauthorized)
			return
		}

		userID, err := bs.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		var withdrawals []models.Withdrawal
		withdrawals, err = bs.GetWithdrawals(requestContext, userID)
		if err != nil {
			logger.Sugar.Errorf("error retrieving withdrawals: %v", err)
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(withdrawals)
	}
}

func HandleWithdrawBalance(bs BalanceStorage) http.HandlerFunc {
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
			http.Error(res, "no required claims available", http.StatusUnauthorized)
			return
		}

		userID, err := bs.GetUserID(requestContext, username)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		var currentRequest transactionRequest
		if err = json.NewDecoder(req.Body).Decode(&currentRequest); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err = goluhn.Validate(currentRequest.Order); err != nil {
			http.Error(res, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		var currentTransaction models.Transaction
		currentTransaction.UserID = userID
		currentTransaction.Amount = currentRequest.Sum
		currentTransaction.OrderNumber = currentRequest.Order
		currentTransaction.Type = models.TransactionTypeWithdrawal

		err = bs.WithdrawUserBalance(requestContext, &currentTransaction)
		if err != nil {
			switch err {
			case apperrors.ErrInsufficientFunds:
				http.Error(res, err.Error(), http.StatusPaymentRequired)
			case apperrors.ErrOrderAlreadyExists:
				http.Error(res, err.Error(), http.StatusUnprocessableEntity)
			default:
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}
