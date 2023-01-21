package core

import "errors"

var (
	ErrNonceTooLow       = errors.New("nonce too low")
	ErrInsufficientFunds = errors.New("insufficient funds for gas * price + value")
)
