package miner

import (
	"bcsbs/consensus"
	"bcsbs/core"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	resultQueueSize = 10
	staleThreshold  = 7
)

var (
	errBlockInterruptedByNewHead = errors.New("new head arrived while building block")
)

type task struct {
	state     *state.StateDB
	block     *types.Block
	createdAt time.Time
}

const (
	commitInterruptNone int32 = iota
	commitInterruptNewHead
	commitInterruptResubmit
)

type newWorkReq struct {
	interrupt *int32
	noempty   bool
	timestamp int64
}

type worker struct {
	engine consensus.Engine
	eth    Backend
	chain  *core.BlockChain

	newWorkCh chan *newWorkReq
	taskCh    chan *task
	resultCh  chan *types.Block
	startCh   chan struct{}
	exitCh    chan struct{}

	wg sync.WaitGroup

	current *environment

	mu       sync.RWMutex
	coinbase common.Address

	pendingMu    sync.RWMutex
	pendingTasks map[common.Hash]*task

	running int32
	newTxs  int32

	noempty uint32

	fullTaskHook func()
}

func newWorker(engine consensus.Engine, eth Backend) *worker {
	worker := &worker{
		engine: engine,
		eth:    eth,
		chain:  eth.BlockChain(),

		newWorkCh: make(chan *newWorkReq),
		taskCh:    make(chan *task),
		resultCh:  make(chan *types.Block, resultQueueSize),
		exitCh:    make(chan struct{}),
		startCh:   make(chan struct{}, 1),

		pendingTasks: make(map[common.Hash]*task),
	}

	worker.wg.Add(4)
	go worker.mainLoop()
	go worker.newWorkLoop()
	go worker.resultLoop()
	go worker.taskLoop()

	// worker.startCh <- struct{}{}

	return worker
}

func (w *worker) setEtherbase(addr common.Address) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.coinbase = addr
}

func (w *worker) start() {
	atomic.StoreInt32(&w.running, 1)
	w.startCh <- struct{}{}
}

func (w *worker) stop() {
	atomic.StoreInt32(&w.running, 0)
}

func (w *worker) isRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

func (w *worker) close() {
	atomic.StoreInt32(&w.running, 0)
	close(w.exitCh)
	w.wg.Wait()
}

func (w *worker) commitWork(interrupt *int32, noempty bool, timestamp int64) {
	start := time.Now()

	var coinbase common.Address
	if w.isRunning() {
		if w.coinbase == (common.Address{}) {
			fmt.Println("Refusing to mine without etherbase")
			return
		}
		coinbase = w.coinbase
	}

	work, err := w.prepareWork(&generateParams{
		timestamp: uint64(timestamp),
		coinbase:  coinbase,
	})
	if err != nil {
		return
	}

	if !noempty && atomic.LoadUint32(&w.noempty) == 0 {
		w.commit(work.copy(), nil, false, start)
	}

	err = w.fillTransactions(interrupt, work)
	if errors.Is(err, errBlockInterruptedByNewHead) {
		fmt.Println(errBlockInterruptedByNewHead)
		return
	}
	w.commit(work.copy(), w.fullTaskHook, true, start)

	w.current = work
}

func (w *worker) commit(env *environment, interval func(), update bool, start time.Time) error {
	if w.isRunning() {
		if interval != nil {
			interval()
		}

		env := env.copy()
		block, err := w.engine.FinalizeAndAssemble(env.header, env.state, env.txs)
		if err != nil {
			return err
		}

		select {
		case w.taskCh <- &task{state: env.state, block: block, createdAt: time.Now()}:
		case <-w.exitCh:
			fmt.Println("Worker has exited")
		}
	}

	return nil
}

type generateParams struct {
	timestamp  uint64
	forceTime  bool
	parentHash common.Hash
	coinbase   common.Address
	noTxs      bool
}

type environment struct {
	signer types.Signer

	state    *state.StateDB
	tcount   int
	coinbase common.Address

	header *types.Header
	txs    []*types.Transaction
}

func (env *environment) copy() *environment {
	cpy := &environment{
		signer:   env.signer,
		state:    env.state,
		tcount:   env.tcount,
		coinbase: env.coinbase,
		header:   types.CopyHeader(env.header),
	}

	cpy.txs = make([]*types.Transaction, len(env.txs))
	copy(cpy.txs, env.txs)
	return cpy
}

func (w *worker) prepareWork(genParams *generateParams) (*environment, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	parent := w.chain.CurrentBlock()
	if genParams.parentHash != (common.Hash{}) {
		parent = w.chain.GetBlockByHash(genParams.parentHash)
	}
	if parent == nil {
		return nil, fmt.Errorf("missing parent")
	}

	timestamp := genParams.timestamp
	if parent.Time() >= timestamp {
		if genParams.forceTime {
			return nil, fmt.Errorf("invalid timestamp, parent %d given %d", parent.Time(), timestamp)
		}
		timestamp = parent.Time() + 1
	}

	num := parent.Number()
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		Time:       timestamp,
		Coinbase:   genParams.coinbase,
	}

	if err := w.engine.Prepare(parent.Header(), header); err != nil {
		fmt.Println("Failed to prepare header for sealing err:", err)
		return nil, err
	}

	env, err := w.makeEnv(parent, header, genParams.coinbase)
	if err != nil {
		fmt.Println("Failed to create sealing context err:", err)
		return nil, err
	}

	return env, nil
}

func (w *worker) fillTransactions(interrupt *int32, env *environment) error {
	pending := w.eth.TxPool().Pending()

	if len(pending) > 0 {
		txs := types.NewTransactionsByPriceAndNonce(env.signer, pending)
		if err := w.commitTransactions(env, txs, interrupt); err != nil {
			return err
		}
	}

	return nil
}

func (w *worker) commitTransactions(env *environment, txs *types.TransactionsByPriceAndNonce, interrupt *int32) error {
	for {
		if interrupt != nil && atomic.LoadInt32(interrupt) != commitInterruptNone {
			return errBlockInterruptedByNewHead
		}

		tx := txs.Peek()
		if tx == nil {
			break
		}

		err := w.commitTransaction(env, tx)
		if err != nil {
			txs.Pop()
		} else {
			env.tcount++
			txs.Shift()
			w.eth.TxPool().RemoveTx(tx.Hash())
		}
	}

	return nil
}

func (w *worker) commitTransaction(env *environment, tx *types.Transaction) error {
	if !env.state.ApplyTx(tx) {
		return fmt.Errorf("Error TX hash: %s", tx.Hash())
	}
	env.txs = append(env.txs, tx)
	return nil
}

func (w *worker) makeEnv(parent *types.Block, header *types.Header, coinbase common.Address) (*environment, error) {
	state, err := w.chain.StateAt()
	if err != nil {
		return nil, err
	}

	env := &environment{
		signer:   types.HomesteadSigner{},
		state:    state,
		coinbase: coinbase,
		header:   header,
	}

	env.tcount = 0
	return env, nil
}

// GO

func (w *worker) mainLoop() {
	defer w.wg.Done()

	for {
		select {
		case req := <-w.newWorkCh:
			w.commitWork(req.interrupt, req.noempty, req.timestamp)
		case <-w.exitCh:
			return
		}
	}
}

func (w *worker) newWorkLoop() {
	defer w.wg.Done()

	var (
		interrupt *int32
		timestamp int64
	)

	commit := func(noempty bool, s int32) {
		if interrupt != nil {
			atomic.StoreInt32(interrupt, s)
		}
		interrupt = new(int32)
		select {
		case w.newWorkCh <- &newWorkReq{interrupt: interrupt, noempty: noempty, timestamp: timestamp}:
		case <-w.exitCh:
			return
		}
		atomic.StoreInt32(&w.newTxs, 0)
	}

	clearPending := func(number uint64) {
		w.pendingMu.Lock()
		for h, t := range w.pendingTasks {
			if t.block.NumberU64()+staleThreshold <= number {
				delete(w.pendingTasks, h)
			}
		}
		w.pendingMu.Unlock()
	}

	for {
		select {
		case <-w.startCh:
			clearPending(w.chain.CurrentBlock().NumberU64())
			timestamp = time.Now().Unix()
			commit(true, commitInterruptNewHead)
		case <-w.exitCh:
			return
		}
	}
}

func (w *worker) resultLoop() {
	defer w.wg.Done()

	for {
		select {
		case block := <-w.resultCh:

			if block == nil {
				continue
			}

			if w.chain.HasBlock(block.Hash(), block.NumberU64()) {
				continue
			}

			var (
				sealhash = w.engine.SealHash(block.Header())
				hash     = block.Hash()
			)

			w.pendingMu.RLock()
			_, exist := w.pendingTasks[sealhash]
			w.pendingMu.RUnlock()

			if !exist {
				fmt.Println("Block found but no relative pending task", "number", block.Number(), "sealhash", sealhash, "hash", hash)
				continue
			}

			err := w.chain.WriteBlockAndSetHead(block)
			if err != nil {
				fmt.Println("Failed writing block to chain", "err", err)
				continue
			}

			fmt.Println("Successfully sealed new block", "number", block.Number(), "sealhash", sealhash, "hash", hash)

		case <-w.exitCh:
			return
		}
	}
}

func (w *worker) taskLoop() {
	defer w.wg.Done()

	var (
		stopCh chan struct{}
		prev   common.Hash
	)

	interrupt := func() {
		if stopCh != nil {
			close(stopCh)
			stopCh = nil
		}
	}
	for {
		select {
		case task := <-w.taskCh:
			sealHash := w.engine.SealHash(task.block.Header())
			if sealHash == prev {
				continue
			}

			interrupt()
			stopCh, prev = make(chan struct{}), sealHash

			w.pendingMu.Lock()
			w.pendingTasks[sealHash] = task
			w.pendingMu.Unlock()

			if err := w.engine.Seal(task.block, w.resultCh, stopCh); err != nil {
				fmt.Println("Block sealing failed", "err", err)
				w.pendingMu.Lock()
				delete(w.pendingTasks, sealHash)
				w.pendingMu.Unlock()
			}
		case <-w.exitCh:
			interrupt()
			return
		}
	}
}
