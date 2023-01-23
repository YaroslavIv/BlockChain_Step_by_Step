package example

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"bcsbs/trie"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func Transactions_2() {
	genesis := core.DefaultGenesisBlock()

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	db, _ := rawdb.NewLevelDBDatabase("./my_geth", 0, 0, "", false)
	state_trie, _ := trie.NewStateTrie(db)
	statedb, _ := state.New(state_trie)

	bc := core.NewBlockChain(db, engine, genesis, statedb)

	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	signer := types.HomesteadSigner{}

	statedb.AddBalance(addr, big.NewInt(1_000_000))

	pool := core.NewTxPool(bc, signer)

	var k uint64
	var i, j int64
	var txs []*types.Transaction
	to := common.BytesToAddress([]byte("Rustam"))
	for i = 1; i < 5; i++ {
		for j = 0; j < i; j++ {
			tx := types.NewTransaction(k, to, big.NewInt(i*j), []byte(fmt.Sprintf("%d + %d = %d", i, j, i+j)))
			if tx_sign, err := types.SignTx(tx, signer, key); err != nil {
				panic(err)
			} else {
				txs = append(txs, tx_sign)
				k += 1
			}
		}
	}

	pool.AddLocals(txs)

	_, queued := pool.Content()

	cb := bc.CurrentBlock()
	h := &types.Header{
		ParentHash: cb.Hash(),
		Time:       uint64(time.Now().Unix()),
		Number:     big.NewInt(1),
	}
	block := make(chan *types.Block)
	for _, txs := range queued {
		engine.Seal(types.NewBlock(h, txs), block, nil)
		bc.AddBlock(<-block)
	}

	out := fmt.Sprintf("Addr: %s\n", addr) +
		fmt.Sprintf("\tBalance: %d\n", statedb.GetBalance(addr)) +
		fmt.Sprintf("\tNonce: %d\n", statedb.GetNonce(addr)) +
		"\n" +
		fmt.Sprintf("Addr: %s\n", to) +
		fmt.Sprintf("\tBalance: %d\n", statedb.GetBalance(to)) +
		fmt.Sprintf("\tNonce: %d\n", statedb.GetNonce(to))

	fmt.Println(bc)
	fmt.Println(out)
}
