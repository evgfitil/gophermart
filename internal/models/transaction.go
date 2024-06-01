package models

import "time"

type Transaction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	OrderNumber string    `json:"order_number"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	TransactionTypeAccrual    = "accrual"
	TransactionTypeWithdrawal = "withdrawal"
)
