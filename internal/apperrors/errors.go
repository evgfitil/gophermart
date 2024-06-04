package apperrors

import "errors"

var (
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderNumberTaken   = errors.New("order number already taken by another user")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)
