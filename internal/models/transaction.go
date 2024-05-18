package models

import "time"

type Transaction struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Amount    int       `json:"amount"`
	OrderID   int       `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	TransactionTypeAccrual    = "accrual"
	TransactionTypeWithdrawal = "withdrawal"
)
