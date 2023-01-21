package example

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/types"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func Transactions() {
	genesis := core.DefaultGenesisBlock()

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	bc := core.NewBlockChain(engine, genesis, nil)

	key, _ := crypto.GenerateKey()
	signer := types.HomesteadSigner{}

	var i, j int64
	var k uint64
	block := make(chan *types.Block)
	for i = 1; i < 5; i++ {
		time.Sleep(time.Second)
		cb := bc.CurrentBlock()
		h := &types.Header{
			ParentHash: cb.Hash(),
			Time:       uint64(time.Now().Unix()),
			Number:     big.NewInt(i),
		}
		var txs []*types.Transaction
		for j = 0; j < i; j++ {
			tx := types.NewTransaction(k, common.BytesToAddress([]byte("Rustam")), []byte(fmt.Sprintf("%d + %d = %d", j, j, j*2)))
			if tx_sign, err := types.SignTx(tx, signer, key); err != nil {
				panic(err)
			} else {
				txs = append(txs, tx_sign)
				k += 1
			}
		}
		engine.Seal(types.NewBlock(h, txs), block, nil)
		bc.AddBlock(<-block)
	}

	fmt.Println(bc)
}
