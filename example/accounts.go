package example

import (
	"bcsbs/core/state"
	"bcsbs/trie"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

func Accounts() {
	state_trie, _ := trie.NewTxTrie(nil)
	statedb, _ := state.New(state_trie, nil, nil)

	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)

	statedb.CreateAccount(addr)
	statedb.AddBalance(addr, big.NewInt(1_000_000))
	statedb.SetNonce(addr, 2)
	statedb.UpdateStateObject(statedb.GetOrNewStateObject(addr))

	out := fmt.Sprintf("Addr: %s\n", addr) +
		fmt.Sprintf("\tBalance: %d\n", statedb.GetBalance(addr)) +
		fmt.Sprintf("\tNonce: %d\n", statedb.GetNonce(addr))

	fmt.Println(out)

}
