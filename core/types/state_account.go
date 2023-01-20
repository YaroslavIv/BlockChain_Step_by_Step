package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type StateAccount struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash
	CodeHash []byte
}
