package core

import (
	"bcsbs/core/state"
	"bcsbs/core/types"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	txSlotSize = 32 * 1024
)

var (
	ErrAlreadyKnown       = errors.New("already known")
	ErrInvalidSender      = errors.New("invalid sender")
	ErrReplaceUnderpriced = errors.New("replacement transaction underpriced")
	ErrNegativeValue      = errors.New("negative value")
)

type blockChain interface {
	StateAt() (*state.StateDB, error)
}

type TxPool struct {
	chain  blockChain
	signer types.Signer
	mu     sync.RWMutex

	txsCh chan<- NewTxsEvent

	currentState *state.StateDB

	pending map[common.Address]*txList
	queue   map[common.Address]*txList
	beats   map[common.Address]time.Time
	all     *txLookup
}

func NewTxPool(chain blockChain, signer types.Signer) *TxPool {
	pool := &TxPool{
		chain:  chain,
		signer: signer,

		txsCh: make(chan NewTxsEvent),

		pending: make(map[common.Address]*txList),
		queue:   make(map[common.Address]*txList),
		beats:   make(map[common.Address]time.Time),
		all:     newTxLookup(),
	}

	pool.reset()
	return pool
}

func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) {
	pool.txsCh = ch
}

func (pool *TxPool) reset() {
	statedb, err := pool.chain.StateAt()
	if err != nil {
		fmt.Println(fmt.Errorf("Failed to reset txpool state %s", err))
		return
	}
	pool.currentState = statedb
}

func (pool *TxPool) Stats() (int, int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.stats()
}

func (pool *TxPool) stats() (int, int) {
	pending := 0
	for _, list := range pool.pending {
		pending += list.Len()
	}
	queued := 0
	for _, list := range pool.queue {
		queued += list.Len()
	}
	return pending, queued
}

func (pool *TxPool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		pending[addr] = list.Flatten()
	}
	queued := make(map[common.Address]types.Transactions)
	for addr, list := range pool.queue {
		queued[addr] = list.Flatten()
	}
	return pending, queued
}

func (pool *TxPool) Update() {
	for addr, txl := range pool.queue {
		if pool.pending[addr] == nil {
			pool.pending[addr] = newTxList(false)
		}

		for _, tx := range txl.txs.items {
			pool.queue[addr].Remove(tx)
			pool.pending[addr].Add(tx)
		}
	}
}

func (pool *TxPool) RemoveTx(hash common.Hash) {
	tx := pool.all.Get(hash)
	if tx == nil {
		return
	}

	addr, _ := types.Sender(pool.signer, tx)
	if pool.pending[addr] != nil {
		delete(pool.pending, addr)
	}
}

func (pool *TxPool) Pending() map[common.Address]types.Transactions {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.pending {
		txs := list.Flatten()

		if len(txs) > 0 {
			pending[addr] = txs
		}
	}

	return pending
}

func (pool *TxPool) validateTx(tx *types.Transaction) error {
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}

	from, err := types.Sender(pool.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}

	if pool.currentState.GetNonce(from) > tx.Nonce() {
		return ErrNonceTooLow
	}

	if pool.currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		return ErrInsufficientFunds
	}

	return nil
}

func (pool *TxPool) add(tx *types.Transaction) (replaced bool, err error) {
	hash := tx.Hash()
	if pool.all.Get(hash) != nil {
		fmt.Println(fmt.Errorf("Discarding already known transaction %s", hash))
		return false, ErrAlreadyKnown
	}

	if err := pool.validateTx(tx); err != nil {
		fmt.Println(fmt.Errorf("Discarding invalid transaction hash: %s err: %s", hash, err))
		return false, err
	}

	replaced, err = pool.enqueueTx(hash, tx, true)
	if err != nil {
		return false, err
	}

	return replaced, nil
}

func (pool *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction, addAll bool) (bool, error) {
	// Try to insert the transaction into the future queue
	from, _ := types.Sender(pool.signer, tx) // already validated
	if pool.queue[from] == nil {
		pool.queue[from] = newTxList(false)
	}
	inserted, old := pool.queue[from].Add(tx)
	if !inserted {
		return false, ErrReplaceUnderpriced
	}

	if old != nil {
		pool.all.Remove(old.Hash())
	}

	if pool.all.Get(hash) == nil && !addAll {
		fmt.Println(fmt.Errorf("Missing transaction in lookup set, please report the issue hash : %s", hash))
	}
	if addAll {
		pool.all.Add(tx)
	}
	// If we never record the heartbeat, do it right now.
	if _, exist := pool.beats[from]; !exist {
		pool.beats[from] = time.Now()
	}
	return old != nil, nil
}

func (pool *TxPool) AddLocalAndUpdate(tx *types.Transaction) error {
	errs := pool.AddLocalsAndUpdate([]*types.Transaction{tx})
	return errs[0]
}

func (pool *TxPool) AddLocal(tx *types.Transaction) error {
	errs := pool.AddLocals([]*types.Transaction{tx})
	return errs[0]
}

func (pool *TxPool) AddLocals(txs []*types.Transaction) []error {
	err := pool.addTxs(txs, true)
	pool.txsCh <- NewTxsEvent{txs}

	return err
}

func (pool *TxPool) AddLocalsAndUpdate(txs []*types.Transaction) []error {
	err := pool.addTxs(txs, true)
	pool.Update()
	pool.txsCh <- NewTxsEvent{txs}

	return err
}

func (pool *TxPool) addTxs(txs []*types.Transaction, sync bool) []error {
	var (
		errs = make([]error, len(txs))
		news = make([]*types.Transaction, 0, len(txs))
	)

	for i, tx := range txs {
		if pool.all.Get(tx.Hash()) != nil {
			errs[i] = ErrAlreadyKnown
			continue
		}

		_, err := types.Sender(pool.signer, tx)
		if err != nil {
			errs[i] = ErrInvalidSender
			continue
		}

		news = append(news, tx)
	}
	if len(news) == 0 {
		return errs
	}

	pool.mu.Lock()
	newErrs, _ := pool.addTxsLocked(news)
	pool.mu.Unlock()

	var nilSlot = 0
	for _, err := range newErrs {
		for errs[nilSlot] != nil {
			nilSlot++
		}
		errs[nilSlot] = err
		nilSlot++
	}

	return errs
}

func (pool *TxPool) addTxsLocked(txs []*types.Transaction) ([]error, *accountSet) {
	dirty := newAccountSet(pool.signer)
	errs := make([]error, len(txs))
	for i, tx := range txs {
		replaced, err := pool.add(tx)
		errs[i] = err
		if err == nil && !replaced {
			dirty.addTx(tx)
		}
	}
	return errs, dirty
}

type txLookup struct {
	slots  int
	lock   sync.RWMutex
	locals map[common.Hash]*types.Transaction
}

func newTxLookup() *txLookup {
	return &txLookup{
		locals: make(map[common.Hash]*types.Transaction),
	}
}

func (t *txLookup) Get(hash common.Hash) *types.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	tx := t.locals[hash]
	return tx
}

func (t *txLookup) Add(tx *types.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.slots += numSlots(tx)

	t.locals[tx.Hash()] = tx
}

func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tx, ok := t.locals[hash]

	if !ok {
		fmt.Println(fmt.Errorf("No transaction found to be deleted hash: %s", hash))
		return
	}
	t.slots -= numSlots(tx)

	delete(t.locals, hash)
}

func numSlots(tx *types.Transaction) int {
	return int((tx.Size() + txSlotSize - 1) / txSlotSize)
}

type accountSet struct {
	accounts map[common.Address]struct{}
	signer   types.Signer
	cache    *[]common.Address
}

func newAccountSet(signer types.Signer, addrs ...common.Address) *accountSet {
	as := &accountSet{
		accounts: make(map[common.Address]struct{}),
		signer:   signer,
	}
	for _, addr := range addrs {
		as.add(addr)
	}
	return as
}

func (as *accountSet) addTx(tx *types.Transaction) {
	if addr, err := types.Sender(as.signer, tx); err == nil {
		as.add(addr)
	}
}

func (as *accountSet) add(addr common.Address) {
	as.accounts[addr] = struct{}{}
	as.cache = nil
}
