package miner

import (
	"bcsbs/consensus"
	"bcsbs/core"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Backend interface {
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
}

type Miner struct {
	worker   *worker
	coinbase common.Address
	eth      Backend
	engine   consensus.Engine

	exitCh  chan struct{}
	startCh chan common.Address
	stopCh  chan struct{}

	wg sync.WaitGroup
}

func New(eth Backend, engine consensus.Engine) *Miner {
	miner := &Miner{
		eth:    eth,
		engine: engine,
		worker: newWorker(engine, eth),

		exitCh:  make(chan struct{}),
		startCh: make(chan common.Address),
		stopCh:  make(chan struct{}),
	}

	miner.wg.Add(1)
	go miner.update()

	return miner
}

func (miner *Miner) update() {
	defer miner.wg.Done()

	for {
		select {
		case addr := <-miner.startCh:
			miner.SetEtherbase(addr)
			miner.worker.start()
		case <-miner.stopCh:
			miner.worker.stop()
		case <-miner.exitCh:
			miner.worker.close()
			return
		}
	}
}

func (miner *Miner) Start(coinbase common.Address) {
	miner.startCh <- coinbase
}

func (miner *Miner) Stop() {
	miner.stopCh <- struct{}{}
}

func (miner *Miner) Close() {
	close(miner.exitCh)
	miner.wg.Wait()
}

func (miner *Miner) Mining() bool {
	return miner.worker.isRunning()
}

func (miner *Miner) SetEtherbase(addr common.Address) {
	miner.coinbase = addr
	miner.worker.setEtherbase(addr)
}
