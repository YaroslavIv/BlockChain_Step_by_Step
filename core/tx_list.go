package core

import (
	"bcsbs/core/types"
	"container/heap"
	"sort"
)

type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type txList struct {
	txs *txSortedMap
}

func newTxList(strict bool) *txList {
	return &txList{
		txs: newTxSortedMap(),
	}
}

func (l *txList) Len() int {
	return l.txs.Len()
}

func (l *txList) Add(tx *types.Transaction) (bool, *types.Transaction) {
	old := l.txs.Get(tx.Nonce())

	l.txs.Put(tx)

	return true, old
}

func (l *txList) Flatten() types.Transactions {
	return l.txs.Flatten()
}

type txSortedMap struct {
	items map[uint64]*types.Transaction
	index *nonceHeap
	cache types.Transactions
}

func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]*types.Transaction),
		index: new(nonceHeap),
	}
}

func (m *txSortedMap) Len() int {
	return len(m.items)
}

func (m *txSortedMap) Get(nonce uint64) *types.Transaction {
	return m.items[nonce]
}

func (m *txSortedMap) Put(tx *types.Transaction) {
	nonce := tx.Nonce()
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce], m.cache = tx, nil
}

func (m *txSortedMap) Flatten() types.Transactions {
	cache := m.flatten()
	txs := make(types.Transactions, len(cache))
	copy(txs, cache)
	return txs
}

func (m *txSortedMap) flatten() types.Transactions {
	if m.cache == nil {
		m.cache = make(types.Transactions, 0, len(m.items))
		for _, tx := range m.items {
			m.cache = append(m.cache, tx)
		}
		sort.Sort(types.TxByNonce(m.cache))
	}
	return m.cache
}
