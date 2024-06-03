package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/evgfitil/gophermart.git/internal/apperrors"
	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func (db *DBStorage) isOrderExists(ctx context.Context, tx *sql.Tx, orderID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE order_number = $1)`
	err := tx.QueryRowContext(ctx, query, orderID).Scan(&exists)
	return exists, err
}

func (db *DBStorage) returnOrderOwnerID(ctx context.Context, tx *sql.Tx, orderID string) (int, error) {
	var ownerID int
	query := `SELECT user_id FROM orders WHERE order_number = $1`
	err := tx.QueryRowContext(ctx, query, orderID).Scan(&ownerID)
	return ownerID, err
}

func (db *DBStorage) GetOrders(ctx context.Context, userID int) ([]models.Order, error) {
	var orders []models.Order

	query := `SELECT order_number, status, accrual, uploaded_at FROM orders WHERE user_id = $1`
	rows, err := db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Sugar.Errorf("error retrieving orders: %v", err)
		}
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.OrderNumber, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			logger.Sugar.Errorf("error retrieving order: %v", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("error after row iteration: %v", err)
		return nil, err
	}

	return orders, nil
}

func (db *DBStorage) GetNewOrders(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order

	query := `SELECT id, order_number, user_id, status, uploaded_at FROM orders WHERE status NOT IN ('PROCESSED', 'INVALID')`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Sugar.Errorf("error retrieving orders: %v", err)
		}
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &order.Status, &order.UploadedAt)
		if err != nil {
			logger.Sugar.Errorf("error retrieving order: %v", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("error after row iteration: %v", err)
		return nil, err
	}
	return orders, nil
}

func (db *DBStorage) ProcessOrder(ctx context.Context, order models.Order) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if exists, _ := db.isOrderExists(ctx, tx, order.OrderNumber); exists {
		var ownerID int
		ownerID, err = db.returnOrderOwnerID(ctx, tx, order.OrderNumber)
		if err != nil {
			return err
		}
		if ownerID != order.UserID {
			return apperrors.ErrOrderNumberTaken
		} else {
			return apperrors.ErrOrderAlreadyExists
		}
	}

	query := `INSERT INTO orders (user_id, order_number, status, uploaded_at) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, query, order.UserID, order.OrderNumber, order.Status, order.UploadedAt)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *DBStorage) UpdateOrderAccrual(ctx context.Context, orderNumber string, accrual float64) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Sugar.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	updateOrderQuery := `UPDATE orders SET accrual = $1, status = 'PROCESSED' WHERE order_number = $2`
	_, err = tx.ExecContext(ctx, updateOrderQuery, accrual, orderNumber)
	if err != nil {
		logger.Sugar.Errorf("error updating order: %v", err)
		return err
	}

	var userID int
	var orderID string
	getUserAndOrderIDQuery := `SELECT user_id, id FROM orders WHERE order_number = $1`
	row := tx.QueryRowContext(ctx, getUserAndOrderIDQuery, orderNumber)
	if err = row.Scan(&userID, &orderID); err != nil {
		logger.Sugar.Errorf("error getting user_id for order: %v", err)
		return err
	}

	addTransactionQuery := `INSERT INTO transactions (user_id, type, amount, order_number) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, addTransactionQuery, userID, "accrual", accrual, orderID)
	if err != nil {
		logger.Sugar.Errorf("error adding transaction: %v", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Sugar.Errorf("error committing transaction: %v", err)
		return err
	}
	return nil
}

func (db *DBStorage) UpdateOrderStatus(ctx context.Context, orderNumber string, status string) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		logger.Sugar.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	updateOrderStatusQuery := `UPDATE orders SET status = $1 WHERE order_number = $2`
	_, err = tx.ExecContext(ctx, updateOrderStatusQuery, status, orderNumber)
	if err != nil {
		logger.Sugar.Errorf("error updating order: %v", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Sugar.Errorf("error committing transaction: %v", err)
		return err
	}

	return nil
}
