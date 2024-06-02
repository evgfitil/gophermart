package models

import "time"

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Amount      float64   `json:"sum"`
	CreatedAt   time.Time `json:"processed_at"`
}
