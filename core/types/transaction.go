package types

import (
	"container/heap"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	LegacyTxType = iota
)

var (
	ErrInvalidSig         = errors.New("invalid transaction v, r, s values")
	ErrTxTypeNotSupported = errors.New("transaction type not supported")
)

type TxData interface {
	txType() byte
	copy() TxData

	sender() *common.Address
	data() []byte
	value() *big.Int
	nonce() uint64
	to() *common.Address

	rawSignatureValues() (v, r, s *big.Int)
	setSignatureValues(chainID, v, r, s *big.Int)
	setSender(sender *common.Address)
}

type Transaction struct {
	inner TxData
	time  time.Time

	size atomic.Value
	hash atomic.Value
	from atomic.Value
}

func NewTX(inner TxData) *Transaction {
	tx := new(Transaction)
	tx.setDecoded(inner.copy(), 0)
	return tx
}

func (tx *Transaction) setDecoded(inner TxData, size int) {
	tx.inner = inner
	tx.time = time.Now()
	if size > 0 {
		tx.size.Store(common.StorageSize(size))
	}
}

func (tx *Transaction) Type() uint8             { return tx.inner.txType() }
func (tx *Transaction) Data() []byte            { return tx.inner.data() }
func (tx *Transaction) Value() *big.Int         { return new(big.Int).Set(tx.inner.value()) }
func (tx *Transaction) Nonce() uint64           { return tx.inner.nonce() }
func (tx *Transaction) To() *common.Address     { return copyAddressPtr(tx.inner.to()) }
func (tx *Transaction) Sender() *common.Address { return copyAddressPtr(tx.inner.sender()) }

func (tx *Transaction) Cost() *big.Int {
	return tx.Value()
}

func (tx *Transaction) RawSignatureValues() (v, r, s *big.Int) {
	return tx.inner.rawSignatureValues()
}

func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	var h common.Hash
	if tx.Type() == LegacyTxType {
		h = rlpHash(tx.inner)
	} else {
		h = prefixedRlpHash(tx.Type(), tx.inner)
	}
	tx.hash.Store(h)
	return h
}

func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.inner)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := tx.inner.copy()
	cpy.setSignatureValues(signer.ChainID(), v, r, s)
	tx_sign := &Transaction{inner: cpy, time: tx.time}

	sender, err := signer.Sender(tx_sign)
	cpy.setSender(copyAddressPtr(&sender))

	return tx_sign, err
}

func (tx *Transaction) String() string {
	return fmt.Sprintf("\tSender: %s\n", tx.Sender()) +
		fmt.Sprintf("\tTo: %s\n", tx.To()) +
		fmt.Sprintf("\tValue: %d\n", tx.Value()) +
		fmt.Sprintf("\tNonce: %d\n", tx.Nonce()) +
		fmt.Sprintf("\tData: %s\n", tx.Data())
}

type Transactions []*Transaction

func (s Transactions) Len() int { return len(s) }

func copyAddressPtr(a *common.Address) *common.Address {
	if a == nil {
		return nil
	}
	cpy := *a
	return &cpy
}

type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce() < s[j].Nonce() }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type TxWithMinerFee struct {
	tx       *Transaction
	minerFee *big.Int
}

type TxByPriceAndTime []*TxWithMinerFee

func (s TxByPriceAndTime) Len() int { return len(s) }
func (s TxByPriceAndTime) Less(i, j int) bool {
	cmp := s[i].minerFee.Cmp(s[j].minerFee)
	if cmp == 0 {
		return s[i].tx.time.Before(s[j].tx.time)
	}
	return cmp > 0
}
func (s TxByPriceAndTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TxByPriceAndTime) Push(x interface{}) {
	*s = append(*s, x.(*TxWithMinerFee))
}

func (s *TxByPriceAndTime) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TransactionsByPriceAndNonce struct {
	txs    map[common.Address]Transactions // Per account nonce-sorted list of transactions
	heads  TxByPriceAndTime                // Next transaction for each unique account (price heap)
	signer Signer                          // Signer for the set of transactions
}

func NewTransactionsByPriceAndNonce(signer Signer, txs map[common.Address]Transactions) *TransactionsByPriceAndNonce {
	heads := make(TxByPriceAndTime, 0, len(txs))
	for from, accTxs := range txs {
		acc, _ := Sender(signer, accTxs[0])
		wrapped, err := NewTxWithMinerFee(accTxs[0])
		if acc != from || err != nil {
			delete(txs, from)
			continue
		}
		heads = append(heads, wrapped)
		txs[from] = accTxs[1:]
	}
	heap.Init(&heads)

	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0].tx
}

func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}

func (t *TransactionsByPriceAndNonce) Shift() {
	acc, _ := Sender(t.signer, t.heads[0].tx)
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		if wrapped, err := NewTxWithMinerFee(txs[0]); err == nil {
			t.heads[0], t.txs[acc] = wrapped, txs[1:]
			heap.Fix(&t.heads, 0)
			return
		}
	}
	heap.Pop(&t.heads)
}

func NewTxWithMinerFee(tx *Transaction) (*TxWithMinerFee, error) {
	return &TxWithMinerFee{
		tx:       tx,
		minerFee: common.Big0,
	}, nil
}
