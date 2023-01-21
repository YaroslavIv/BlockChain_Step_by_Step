package types

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
)

type BlockNonce [8]byte

func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

type Header struct {
	ParentHash common.Hash
	TxHash     common.Hash
	Number     *big.Int
	Time       uint64
	Nonce      BlockNonce
}

func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

type Body struct {
	Transactions Transactions
}

type Block struct {
	header       *Header
	transactions Transactions

	hash atomic.Value
}

func NewBlock(header *Header, txs []*Transaction) *Block {
	b := &Block{
		header: CopyHeader(header),
	}

	if len(txs) != 0 {
		b.header.TxHash = rlpHash(txs)
		b.transactions = make(Transactions, len(txs))
		copy(b.transactions, txs)
	}

	return b
}

func NewBlockWithHeader(header *Header) *Block {
	return &Block{header: CopyHeader(header)}
}

func CopyHeader(h *Header) *Header {
	cpy := *h
	return &cpy
}

func (b *Block) WithSeal(header *Header) *Block {
	cpy := *header

	return &Block{
		header: &cpy,
	}
}

func (b *Block) WithSealAndTx(header *Header, txs *Transactions) *Block {
	cpy := *header
	transactions := *txs

	return &Block{
		header:       &cpy,
		transactions: transactions,
	}
}

func (b *Block) Hash() common.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := b.header.Hash()
	b.hash.Store(v)
	return v
}

func (b *Block) String() string {
	tx_info := fmt.Sprintf("Tx Len: %d\n", len(b.transactions))
	for _, tx := range b.transactions {
		tx_info += tx.String() + "\n"
	}
	return fmt.Sprintf("Block: %d\n", b.NumberU64()) +
		fmt.Sprintf("ParentHash: %s\n", b.header.ParentHash) +
		fmt.Sprintf("Time: %d\n", b.header.Time) +
		fmt.Sprintf("Hash: %s\n", b.Hash()) +
		fmt.Sprintf("TxHash: %s\n", b.TxHash()) +
		fmt.Sprintf("Nonce: %d\n", b.Nonce()) + tx_info

}

func (b *Block) Body() *Body             { return &Body{b.transactions} }
func (b *Block) Header() *Header         { return CopyHeader(b.header) }
func (b *Block) NumberU64() uint64       { return b.header.Number.Uint64() }
func (b *Block) Number() *big.Int        { return new(big.Int).Set(b.header.Number) }
func (b *Block) ParentHash() common.Hash { return b.header.ParentHash }
func (b *Block) Nonce() uint64           { return binary.BigEndian.Uint64(b.header.Nonce[:]) }
func (b *Block) TxHash() common.Hash     { return b.header.TxHash }

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}
