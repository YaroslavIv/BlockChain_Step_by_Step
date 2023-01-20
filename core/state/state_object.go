package state

import (
	"bcsbs/core/types"
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var emptyCodeHash = crypto.Keccak256(nil)

type stateObject struct {
	address  common.Address
	addrHash common.Hash
	data     types.StateAccount
	db       *StateDB

	deleted bool
}

func newObject(db *StateDB, address common.Address, data types.StateAccount) *stateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	if data.Root == (common.Hash{}) {
		data.Root = emptyRoot
	}

	return &stateObject{
		db:       db,
		address:  address,
		addrHash: crypto.Keccak256Hash(address[:]),
		data:     data,
	}
}

// GET

func (s *stateObject) empty() bool {
	return s.data.Nonce == 0 && s.data.Balance.Sign() == 0 && bytes.Equal(s.data.CodeHash, emptyCodeHash)
}

func (s *stateObject) Address() common.Address {
	return s.address
}

func (s *stateObject) Balance() *big.Int {
	return s.data.Balance
}

func (s *stateObject) Nonce() uint64 {
	return s.data.Nonce
}

// SET
// -PUB

func (s *stateObject) AddBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		if s.empty() {
			s.touch()
		}
		return
	}
	s.SetBalance(new(big.Int).Add(s.Balance(), amount))
}

func (s *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Sub(s.Balance(), amount))
}

func (s *stateObject) SetBalance(amount *big.Int) {
	s.setBalance(amount)
}

func (s *stateObject) SetNonce(nonce uint64) {
	s.setNonce(nonce)
}

// -PRI

func (s *stateObject) setBalance(amount *big.Int) {
	s.data.Balance = amount
}

func (s *stateObject) setNonce(nonce uint64) {
	s.data.Nonce = nonce
}

func (s *stateObject) touch() {}
