package database

import (
	"context"
	"database/sql"
	"errors"

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
