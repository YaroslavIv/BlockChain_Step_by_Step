package example

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"bcsbs/miner"
	"bcsbs/trie"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Backend struct {
	bc   *core.BlockChain
	pool *core.TxPool
}

func (b *Backend) BlockChain() *core.BlockChain {
	return b.bc
}
func (b *Backend) TxPool() *core.TxPool {
	return b.pool
}

func Miner() {
	genesis := core.DefaultGenesisBlock()

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	db, _ := rawdb.NewLevelDBDatabase("./my_geth", 0, 0, "", false)
	state_trie, _ := trie.NewTxTrie(db)
	statedb, _ := state.New(state_trie, nil, nil)

	bc := core.NewBlockChain(db, engine, genesis, statedb)

	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	signer := types.HomesteadSigner{}

	pool := core.NewTxPool(bc, signer)
	backend := &Backend{
		bc:   bc,
		pool: pool,
	}
	miner := miner.New(backend, engine)

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
	miner.Start(addr)
	time.Sleep(time.Second * 3)

	pool.AddLocals(txs)
	pool.Update()

	for i := 0; i < 3; i++ {
		miner.Start(addr)
		time.Sleep(time.Second * 3)
	}

	fmt.Println(bc)
	fmt.Printf("Balance(%s) : %d", addr, statedb.GetBalance(addr))
}
