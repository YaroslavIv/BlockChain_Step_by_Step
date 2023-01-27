package cli

import (
	"bcsbs/consensus/ethash"
	"bcsbs/core"
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"bcsbs/core/vm"
	"bcsbs/miner"
	"bcsbs/trie"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/holiman/uint256"
)

type TxArgs struct {
	Sign string
}

type Response struct {
	Result string
}

type Server struct {
	pool    *core.TxPool
	statedb *state.StateDB
}

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

func show(addr_contract common.Address, statedb *state.StateDB) string {
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

	return out
}

func NewServer(addr common.Address) *Server {

	engine := &ethash.Ethash{
		Target: big.NewInt(int64(math.Pow(16, 3))),
	}

	db, _ := rawdb.NewLevelDBDatabase("./my_geth", 0, 0, "", false)
	tx_trie, _ := trie.NewTxTrie(db)
	storage_trie, _ := trie.NewStorageTrie(db)
	blockCtx := core.NewEVMBlockContext()

	statedb, _ := state.New(tx_trie, storage_trie, &blockCtx)

	var genesis *core.Genesis
	if rawdb.ReadHeadBlockHash(db) == (common.Hash{}) {
		genesis = core.DefaultGenesisBlock()
	}

	bc := core.NewBlockChain(db, engine, genesis, statedb)

	signer := types.HomesteadSigner{}
	pool := core.NewTxPool(bc, signer)

	backend := &Backend{
		bc:   bc,
		pool: pool,
	}
	miner := miner.New(backend, engine)

	miner.Start(addr)

	server := &Server{
		pool:    pool,
		statedb: statedb,
	}

	return server
}

func (s *Server) SendRawTransaction(r *http.Request, args *TxArgs, result *Response) error {
	sign := common.Hex2Bytes(args.Sign)
	tx := new(types.Transaction)
	tx.UnmarshalBinary(sign)
	tx.SetNonce(s.statedb.GetNonce(*tx.Sender()))

	s.pool.AddLocalAndUpdate(tx)

	time.Sleep(time.Second * 3)
	if len(tx.Data()) < 1 {
		*result = Response{Result: tx.Text()}
	} else if tx.Data()[0] == byte(vm.INIT) {
		AddrContract := crypto.CreateAddress(*tx.Sender(), s.statedb.GetNonce(*tx.Sender())-1)
		*result = Response{Result: fmt.Sprintf("AddrContract: %s", AddrContract)}
	} else if tx.Data()[0] == byte(vm.MOVE) {
		*result = Response{Result: show(*tx.To(), s.statedb)}
	}
	return nil
}

func (cli *CLI) startServer(address string) {
	addr := common.HexToAddress(address)

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	server := NewServer(addr)

	rpcServer.RegisterService(server, "server")

	router := mux.NewRouter()
	router.Handle("/delivery", rpcServer)
	http.ListenAndServe(":1337", router)

}
