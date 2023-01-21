package core

import (
	"bcsbs/consensus"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"fmt"
	"sync"
)

type BlockChain struct {
	blocks types.Blocks

	genesisBlock *types.Block

	engine consensus.Engine

	statedb *state.StateDB

	mu sync.Mutex
}

func NewBlockChain(engine consensus.Engine, genesis *Genesis, statedb *state.StateDB) *BlockChain {
	bc := &BlockChain{
		engine:       engine,
		genesisBlock: genesis.ToBlock(),
		statedb:      statedb,
	}

	bc.blocks = append(bc.blocks, bc.genesisBlock)

	return bc
}

func (bc *BlockChain) AddBlock(block *types.Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if err := bc.engine.VerifyHeader(bc.CurrentHeader(), block.Header(), true); err != nil {
		panic(err)
	}

	for _, tx := range block.Body().Transactions {
		bc.statedb.ApplyTx(tx)
	}

	bc.blocks = append(bc.blocks, block)
}

func (bc *BlockChain) InsertChain(chain types.Blocks) (int, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(chain) == 0 {
		return 0, nil
	}

	head := bc.CurrentBlock()
	if chain[0].NumberU64() != head.NumberU64()+1 || chain[0].ParentHash() != head.Hash() {
		return 0, fmt.Errorf("non contiguous insert: item 0 is #%d [%x..], item 1 is #%d [%x..] (parent [%x..])", head.NumberU64(),
			head.Hash().Bytes()[:4], chain[0].NumberU64(), chain[0].Hash().Bytes()[:4], chain[0].ParentHash().Bytes()[:4])
	}
	if err := bc.engine.VerifyHeader(head.Header(), chain[0].Header(), true); err != nil {
		return 0, err
	}

	for i := 1; i < len(chain); i++ {
		block, prev := chain[i], chain[i-1]
		if block.NumberU64() != prev.NumberU64()+1 || block.ParentHash() != prev.Hash() {

			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x..], item %d is #%d [%x..] (parent [%x..])", i-1, prev.NumberU64(),
				prev.Hash().Bytes()[:4], i, block.NumberU64(), block.Hash().Bytes()[:4], block.ParentHash().Bytes()[:4])
		}
		if err := bc.engine.VerifyHeader(prev.Header(), block.Header(), true); err != nil {
			return 0, err
		}
	}

	return bc.insertChain(chain, true, true)
}

func (bc *BlockChain) insertChain(chain types.Blocks, verifySeals, setHead bool) (int, error) {
	for _, block := range chain {
		bc.blocks = append(bc.blocks, block)
	}
	return len(chain), nil
}

func (bc *BlockChain) String() string {
	var out string

	for _, block := range bc.blocks {
		out += block.String() + "\n"
	}

	return out
}
