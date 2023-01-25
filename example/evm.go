package example

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"bcsbs/core/vm"
	"bcsbs/miner"
	"bcsbs/trie"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

func initContract(nonce uint64, signer types.Signer, private_key *ecdsa.PrivateKey, addr common.Address, amount *big.Int) *types.Transaction {
	var code string
	for _, i := range []string{
		"INIT",

		"CALLER", "PUSH1", "00", "SSTORE",
		"PUSH20", addr.String()[2:], "PUSH1", "01", "SSTORE",
		"PUSH1", "01", "PUSH1", "02", "SSTORE",
	} {
		if len(i) > 2 && len(i) < 20 {
			code += strconv.FormatInt(int64(vm.StringToOp(i)), 16)
		} else {
			code += i
		}
	}

	code_hex, _ := hex.DecodeString(code)

	tx := types.NewContractCreation(nonce, amount, code_hex)
	if tx_sign, err := types.SignTx(tx, signer, private_key); err != nil {
		panic(err)
	} else {
		return tx_sign
	}
}

func move(nonce uint64, signer types.Signer, private_key *ecdsa.PrivateKey, addr_contract common.Address, x, y int) *types.Transaction {
	z := x*3 + y + 3
	if z < 3 || z > 11 {
		return nil
	}

	var code string
	for _, i := range []string{
		"MOVE",

		"PUSH1", fmt.Sprintf("0%x", z), "SLOAD", "ISZERO",
		"PUSH1", "09", "JUMPI", "00",

		"PUSH1", "02", "SLOAD",

		"PUSH1", "00", "SLOAD",
		"CALLER",
		"03",

		"ISZERO", "AND", "PUSH1", "30", "JUMPI",

		"PUSH1", "02", "SLOAD", "ISZERO",

		"PUSH1", "01", "SLOAD",
		"CALLER",
		"03",
		"ISZERO", "AND", "PUSH1", "25", "JUMPI", "00",

		"PUSH1", "02", "PUSH1", fmt.Sprintf("0%x", z), "SSTORE", "PUSH1", "01", "PUSH1", "02", "SSTORE", "00",
		"PUSH1", "01", "PUSH1", fmt.Sprintf("0%x", z), "SSTORE", "PUSH1", "00", "PUSH1", "02", "SSTORE", "00",
	} {
		if len(i) > 2 && len(i) < 20 {
			code += strconv.FormatInt(int64(vm.StringToOp(i)), 16)
		} else {
			code += i
		}
	}

	code_hex, _ := hex.DecodeString(code)

	tx := types.NewTransaction(nonce, addr_contract, big.NewInt(0), code_hex)
	if tx_sign, err := types.SignTx(tx, signer, private_key); err != nil {
		panic(err)
	} else {
		return tx_sign
	}
}

func show(addr_contract common.Address, statedb *state.StateDB) {
	in1 := statedb.GetState(addr_contract, crypto.Keccak256Hash(append(addr_contract.Bytes(), new(uint256.Int).SetUint64(0).Bytes()...)))
	in2 := statedb.GetState(addr_contract, crypto.Keccak256Hash(append(addr_contract.Bytes(), new(uint256.Int).SetUint64(1).Bytes()...)))
	in3 := statedb.GetState(addr_contract, crypto.Keccak256Hash(append(addr_contract.Bytes(), new(uint256.Int).SetUint64(2).Bytes()...)))

	addr1 := common.BytesToAddress(in1.Bytes()[12:])
	addr2 := common.BytesToAddress(in2.Bytes()[12:])
	value := statedb.GetBalance(addr_contract)
	is_addr1 := new(big.Int).SetBytes(in3[:]).Int64()

	var field_out string
	var i uint64
	for i = 3; i < 12; i++ {
		in := statedb.GetState(addr_contract, crypto.Keccak256Hash(append(addr_contract.Bytes(), new(uint256.Int).SetUint64(i).Bytes()...)))

		z := new(big.Int).SetBytes(in[:]).Int64()
		if z == 1 {
			field_out += "x"
		} else if z == 2 {
			field_out += "o"
		} else {
			field_out += " "
		}

		if i != 11 {
			if i%3 == 2 {
				field_out += "\n- - -\n"
			} else {
				field_out += "|"
			}
		} else {
			field_out += "\n\n"
		}
	}

	out := fmt.Sprintf("Contract: %s\n", addr_contract) +
		fmt.Sprintf("\tAddr1: %x\n", addr1) +
		fmt.Sprintf("\tAddr2: %x\n", addr2) +
		fmt.Sprintf("\tIs_addr1: %d\n", is_addr1) +
		fmt.Sprintf("\tValue: %d\n", value) +
		field_out

	fmt.Println(out)
}

func EVM() {
	genesis := core.DefaultGenesisBlock()

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	db, _ := rawdb.NewLevelDBDatabase("./my_geth", 0, 0, "", false)
	tx_trie, _ := trie.NewTxTrie(db)
	storage_trie, _ := trie.NewStorageTrie(db)
	blockCtx := core.NewEVMBlockContext()

	statedb, _ := state.New(tx_trie, storage_trie, &blockCtx)

	bc := core.NewBlockChain(db, engine, genesis, statedb)

	key1, _ := crypto.GenerateKey()
	key2, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(key1.PublicKey)
	addr2 := crypto.PubkeyToAddress(key2.PublicKey)
	signer := types.HomesteadSigner{}

	pool := core.NewTxPool(bc, signer)
	backend := &Backend{
		bc:   bc,
		pool: pool,
	}
	miner := miner.New(backend, engine)

	miner.Start(addr1)
	time.Sleep(time.Second * 3)

	addr_contract := crypto.CreateAddress(bc.CurrentHeader().Coinbase, 0)
	pool.AddLocal(initContract(statedb.GetNonce(addr1), signer, key1, addr2, big.NewInt(10)))
	pool.Update()
	miner.Start(addr1)
	time.Sleep(time.Second * 3)

	pool.AddLocal(move(statedb.GetNonce(addr1), signer, key1, addr_contract, 0, 0))
	pool.Update()
	miner.Start(addr1)
	time.Sleep(time.Second * 3)

	pool.AddLocal(move(statedb.GetNonce(addr2), signer, key2, addr_contract, 1, 1))
	pool.Update()
	miner.Start(addr1)
	time.Sleep(time.Second * 3)

	pool.AddLocal(move(statedb.GetNonce(addr1), signer, key1, addr_contract, 1, 0))
	pool.Update()
	miner.Start(addr1)
	time.Sleep(time.Second * 3)

	fmt.Println(bc)

	show(addr_contract, statedb)
}
