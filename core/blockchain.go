package core

import (
	"bcsbs/core/types"
	"fmt"
)

type BlockChain struct {
	blocks types.Blocks
}

func (bc *BlockChain) InsertChain(chain types.Blocks) (int, error) {
	if len(chain) == 0 {
		return 0, nil
	}

	if len(bc.blocks) > 0 {
		head := bc.blocks[len(bc.blocks)-1]
		if chain[0].NumberU64() != head.NumberU64()+1 || chain[0].ParentHash() != head.Hash() {
			return 0, fmt.Errorf("non contiguous insert: item 0 is #%d [%x..], item 1 is #%d [%x..] (parent [%x..])", head.NumberU64(),
				head.Hash().Bytes()[:4], chain[0].NumberU64(), chain[0].Hash().Bytes()[:4], chain[0].ParentHash().Bytes()[:4])
		}
	}

	for i := 1; i < len(chain); i++ {
		block, prev := chain[i], chain[i-1]
		if block.NumberU64() != prev.NumberU64()+1 || block.ParentHash() != prev.Hash() {

			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x..], item %d is #%d [%x..] (parent [%x..])", i-1, prev.NumberU64(),
				prev.Hash().Bytes()[:4], i, block.NumberU64(), block.Hash().Bytes()[:4], block.ParentHash().Bytes()[:4])
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
