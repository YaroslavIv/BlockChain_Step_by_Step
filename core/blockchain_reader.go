package core

import "bcsbs/core/types"

func (bc *BlockChain) CurrentHeader() *types.Header {
	return bc.CurrentBlock().Header()
}

func (bc *BlockChain) CurrentBlock() *types.Block {
	return bc.blocks[len(bc.blocks)-1]
}
