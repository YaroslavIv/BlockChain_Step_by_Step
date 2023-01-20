package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type LegacyTx struct {
	Sender *common.Address

	Nonce   uint64
	To      *common.Address
	Data    []byte
	V, R, S *big.Int
}

func NewTransaction(nonce uint64, to common.Address, data []byte) *Transaction {
	return NewTX(&LegacyTx{
		Nonce: nonce,
		To:    &to,
		Data:  data,
	})
}

func (tx *LegacyTx) copy() TxData {
	cpy := &LegacyTx{
		Nonce: tx.Nonce,
		To:    copyAddressPtr(tx.To),
		Data:  common.CopyBytes(tx.Data),
		V:     new(big.Int),
		R:     new(big.Int),
		S:     new(big.Int),
	}

	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}

	return cpy
}

func (tx *LegacyTx) txType() byte                                 { return LegacyTxType }
func (tx *LegacyTx) data() []byte                                 { return tx.Data }
func (tx *LegacyTx) nonce() uint64                                { return tx.Nonce }
func (tx *LegacyTx) to() *common.Address                          { return tx.To }
func (tx *LegacyTx) sender() *common.Address                      { return tx.Sender }
func (tx *LegacyTx) rawSignatureValues() (v, r, s *big.Int)       { return tx.V, tx.R, tx.S }
func (tx *LegacyTx) setSignatureValues(chainID, v, r, s *big.Int) { tx.V, tx.R, tx.S = v, r, s }
func (tx *LegacyTx) setSender(sender *common.Address)             { tx.Sender = sender }
