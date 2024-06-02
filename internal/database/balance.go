package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/evgfitil/gophermart.git/internal/apperrors"
	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func (db *DBStorage) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	var userBalance models.Balance

	query := `
        SELECT
            COALESCE(SUM(CASE when type = 'accrual' THEN amount ELSE 0 END), 0) -
            COALESCE(SUM(CASE when type = 'withdrawal' THEN amount ELSE 0 END), 0) AS current,
            COALESCE(SUM(CASE when type = 'withdrawal' THEN amount ELSE 0 END), 0) AS withdrawn
        FROM transactions
        WHERE user_id = $1
    `

	err := db.conn.QueryRowContext(ctx, query, userID).Scan(&userBalance.Current, &userBalance.Withdrawn)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return &models.Balance{Current: 0, Withdrawn: 0}, nil
		}
		logger.Sugar.Errorf("error retrieving user balance: %v", err)
		return &models.Balance{}, err
	}

	return &userBalance, nil
}

func (db *DBStorage) GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	query := `SELECT (order_number, amount, created_at) FROM transactions WHERE user_id=$1 AND type = 'withdrawal' ORDER BY created_at DESC`
	rows, err := db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Sugar.Errorf("error retrieving withdrawals: %v", err)
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal models.Withdrawal
		err = rows.Scan(&withdrawal.OrderNumber, &withdrawal.Amount, &withdrawal.CreatedAt)
		if err != nil {
			logger.Sugar.Errorf("error retrieving withdrawals: %v", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("error after row iteration: %v", err)
		return nil, err
	}

	return withdrawals, nil
}

func (db *DBStorage) WithdrawUserBalance(ctx context.Context, transaction *models.Transaction) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Sugar.Errorf("error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	var currentBalance float64
	balanceQuery := `
	SELECT
	    COALESCE(SUM(CASE WHEN type = 'accrual' THEN amount ELSE 0 END), 0) -
	    COALESCE(SUM(CASE WHEN type = 'withdrawal' THEN amount ELSE 0 END), 0) AS current
	FROM transactions
	WHERE user_id = $1;
    `
	err = tx.QueryRowContext(ctx, balanceQuery, transaction.UserID).Scan(&currentBalance)
	if err != nil {
		logger.Sugar.Errorf("error retrieving user balance: %v", err)
		return err
	}

	if currentBalance < transaction.Amount {
		return apperrors.ErrInsufficientFunds
	}

	var orderExists bool
	orderQuery := `SELECT EXISTS(SELECT 1 FROM orders WHERE order_number = $1)`
	err = tx.QueryRowContext(ctx, orderQuery, transaction.OrderNumber).Scan(&orderExists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Sugar.Errorf("error querying order for user balance: %v", err)
		return err
	}

	if orderExists {
		return apperrors.ErrOrderAlreadyExists
	}

	createOrderQuery := `INSERT INTO orders (user_id, order_number, status, uploaded_at) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, createOrderQuery, transaction.UserID, transaction.OrderNumber, "PROCESSED", time.Now())
	if err != nil {
		logger.Sugar.Errorf("error inserting order for user withdraw: %v", err)
		return err
	}

	createTransactionQuery := `INSERT INTO transactions (user_id, type, amount, order_number, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, createTransactionQuery, transaction.UserID, transaction.Type, transaction.Amount, transaction.OrderNumber, time.Now())
	if err != nil {
		logger.Sugar.Errorf("error inserting transaction for withdraw: %v", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Sugar.Errorf("error committing transaction: %v", err)
		return err
	}

	return nil
}
