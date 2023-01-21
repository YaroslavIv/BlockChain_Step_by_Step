package example

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"bcsbs/trie"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func PoolTx() {
	genesis := core.DefaultGenesisBlock()

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	state_trie, _ := trie.NewStateTrie()
	statedb, _ := state.New(state_trie)

	bc := core.NewBlockChain(engine, genesis, statedb)

	key, _ := crypto.GenerateKey()
	signer := types.HomesteadSigner{}

	pool := core.NewTxPool(bc, signer)

	var k uint64
	var i, j int64
	var txs []*types.Transaction
	for i = 1; i < 5; i++ {
		for j = 0; j < i; j++ {
			tx := types.NewTransaction(k, common.BytesToAddress([]byte("Rustam")), big.NewInt(i*j), []byte(fmt.Sprintf("%d + %d = %d", i, j, i+j)))
			if tx_sign, err := types.SignTx(tx, signer, key); err != nil {
				panic(err)
			} else {
				txs = append(txs, tx_sign)
				k += 1
			}
		}
	}

	pool.AddLocals(txs)

	pending, queued := pool.Content()

	out := "Pending\n"
	for addr, txs := range pending {
		out += fmt.Sprintf("\tAddr: %s\n", addr)
		for i, tx := range txs {
			out += fmt.Sprintf("\t%d: \n%s\n", i, tx)
		}
	}

	out += "\nQueued\n"
	for addr, txs := range queued {
		out += fmt.Sprintf("\tAddr: %s\n\n", addr)
		for i, tx := range txs {
			out += fmt.Sprintf("\t\n%d: \n%s\n", i, tx)
		}
	}

	fmt.Println(out)
}
