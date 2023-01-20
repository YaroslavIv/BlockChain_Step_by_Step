package core

import (
	"bcsbs/core/types"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Genesis struct {
	Timestamp  uint64
	Number     uint64
	ParentHash common.Hash
}

func DefaultGenesisBlock() *Genesis {
	return &Genesis{}
}

func (g *Genesis) ToBlock() *types.Block {
	head := &types.Header{
		ParentHash: g.ParentHash,
		Number:     new(big.Int).SetUint64(g.Number),
		Time:       g.Timestamp,
	}
	block := types.NewBlock(head, nil)

	if block.Number().Sign() != 0 {
		err := fmt.Errorf("can't commit genesis block with number > 0")
		panic(err)
	}

	return block
}
