package core

import (
	"bcsbs/core/state"
	"bcsbs/core/types"

	"github.com/ethereum/go-ethereum/common"
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

func (bc *BlockChain) GetBlockByHash(hash common.Hash) *types.Block {
	for _, block := range bc.blocks {
		if block.Hash() == hash {
			return block
		}
	}
	return nil
}

func (bc *BlockChain) HasBlock(hash common.Hash, number uint64) bool {
	if len(bc.blocks) < int(number)+1 {
		return false
	} else if bc.blocks[number].Hash() != hash {
		return false
	}
	return true
}
