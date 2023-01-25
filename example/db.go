package example

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/trie"
	"fmt"
	"math"
	"math/big"
)

func DB() {

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	db, _ := rawdb.NewLevelDBDatabase("./my_geth", 0, 0, "", false)
	state_trie, _ := trie.NewTxTrie(db)
	statedb, _ := state.New(state_trie, nil, nil)

	bc := core.NewBlockChain(db, engine, nil, statedb)

	fmt.Println(bc)
	addr := rawdb.ReadHeadBlock(db).Coinbase()
	fmt.Printf("Balance(%s) : %d\n", addr, statedb.GetBalance(addr))

}
