package vm

import (
	"errors"
	"fmt"
)

var (
	ErrWriteProtection          = errors.New("write protection")
	ErrContractAddressCollision = errors.New("contract address collision")
	ErrInsufficientBalance      = errors.New("insufficient balance for transfer")
	ErrInvalidJump              = errors.New("invalid jump destination")
	ErrNonceUintOverflow        = errors.New("nonce uint64 overflow")

	errStopToken = errors.New("stop token")
)

type ErrStackUnderflow struct {
	stackLen int
	required int
}

func (e *ErrStackUnderflow) Error() string {
	return fmt.Sprintf("stack underflow (%d <=> %d)", e.stackLen, e.required)
}

type ErrStackOverflow struct {
	stackLen int
	limit    int
}

func (e *ErrStackOverflow) Error() string {
	return fmt.Sprintf("stack limit reached %d (%d)", e.stackLen, e.limit)
}
