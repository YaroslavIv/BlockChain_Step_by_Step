package main

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/types"
	"fmt"
	"math"
	"math/big"
	"time"
)

func main() {
	genesis := core.DefaultGenesisBlock()

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	bc := core.NewBlockChain(engine, genesis)

	var i int64
	block := make(chan *types.Block)
	for i = 1; i < 10; i++ {
		time.Sleep(time.Second)
		cb := bc.CurrentBlock()
		h := &types.Header{
			ParentHash: cb.Hash(),
			Time:       uint64(time.Now().Unix()),
			Number:     big.NewInt(i),
		}
		engine.Seal(types.NewBlock(h, fmt.Sprintf("%d + %d = %d", i, i, i*2)), block, nil)
		bc.AddBlock(<-block)
	}

	fmt.Println(bc)

}
