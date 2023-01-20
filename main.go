package main

import (
	"bcsbs/core"
	"bcsbs/core/types"
	"fmt"
	"math/big"
	"time"
)

func main() {
	genesis := core.DefaultGenesisBlock()

	bc := &core.BlockChain{}

	b := genesis.ToBlock()
	bc.InsertChain(types.Blocks{b})

	var i int64
	for i = 1; i < 10; i++ {
		h := &types.Header{
			ParentHash: b.Hash(),
			Time:       uint64(time.Now().Unix()),
			Number:     big.NewInt(i),
		}
		b = types.NewBlock(h, fmt.Sprintf("Data: %d", i))
		bc.InsertChain(types.Blocks{b})
	}

	fmt.Println(bc)

}
