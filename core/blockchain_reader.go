package core

import (
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/core/types"

	"github.com/ethereum/go-ethereum/common"
)

func (bc *BlockChain) CurrentHeader() *types.Header {
	return rawdb.ReadHeadHeader(bc.db)
}

func (bc *BlockChain) CurrentBlock() *types.Block {
	return rawdb.ReadHeadBlock(bc.db)
}

func (bc *BlockChain) StateAt() (*state.StateDB, error) {
	return bc.statedb, nil
}

func (bc *BlockChain) GetBlockByHash(hash common.Hash) *types.Block {
	if number := rawdb.ReadHeaderNumber(bc.db, hash); number != nil {
		return rawdb.ReadBlock(bc.db, hash, *number)
	}
	return nil
}

func (bc *BlockChain) HasBlock(hash common.Hash, number uint64) bool {
	return rawdb.HasHeader(bc.db, hash, number) &&
		rawdb.HasBody(bc.db, hash, number)
}
