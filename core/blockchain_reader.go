package core

import (
	"bcsbs/core/state"
	"bcsbs/core/types"
)

func (bc *BlockChain) CurrentHeader() *types.Header {
	return bc.CurrentBlock().Header()
}

func (bc *BlockChain) CurrentBlock() *types.Block {
	return bc.blocks[len(bc.blocks)-1]
}

func (bc *BlockChain) StateAt() (*state.StateDB, error) {
	return bc.statedb, nil
}
