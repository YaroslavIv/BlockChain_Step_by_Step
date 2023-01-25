package vm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var emptyCodeHash = crypto.Keccak256Hash(nil)

type (
	CanTransferFunc func(StateDB, common.Address, *big.Int) bool
	TransferFunc    func(StateDB, common.Address, common.Address, *big.Int)
)

type BlockContext struct {
	Transfer    TransferFunc
	CanTransfer CanTransferFunc
}

type EVM struct {
	Context BlockContext

	StateDB StateDB

	interpreter *EVMInterpreter
}

func NewEVM(statedb StateDB, blockCtx *BlockContext) *EVM {
	evm := &EVM{
		StateDB: statedb,
		Context: *blockCtx,
	}
	evm.interpreter = NewEVMInterpreter(evm)
	return evm
}

func (evm *EVM) Call(caller ContractRef, addr common.Address, input []byte, value *big.Int) (ret []byte, err error) {
	if value.Sign() != 0 && !evm.Context.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, ErrInsufficientBalance
	}

	evm.Context.Transfer(evm.StateDB, caller.Address(), addr, value)

	nonce := evm.StateDB.GetNonce(caller.Address())
	if nonce+1 < nonce {
		return nil, ErrNonceUintOverflow
	}
	evm.StateDB.SetNonce(caller.Address(), nonce+1)

	code := evm.StateDB.GetCode(addr)
	if len(code) == 0 {
		ret, err = nil, nil
	} else {
		addrCopy := addr
		contract := NewContract(caller, AccountRef(addrCopy), value)
		contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), input)
		ret, err = evm.interpreter.Run(contract, nil, false)
	}

	if err != nil {
		fmt.Println(err)
	}

	return ret, err
}

type codeAndHash struct {
	code []byte
	hash common.Hash
}

func (evm *EVM) create(caller ContractRef, codeAndHash *codeAndHash, value *big.Int, address common.Address, typ OpCode) ([]byte, common.Address, error) {

	if !evm.Context.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, common.Address{}, ErrInsufficientBalance
	}

	nonce := evm.StateDB.GetNonce(caller.Address())
	if nonce+1 < nonce {
		return nil, common.Address{}, ErrNonceUintOverflow
	}

	evm.StateDB.SetNonce(caller.Address(), nonce+1)

	contractHash := evm.StateDB.GetCodeHash(address)
	if evm.StateDB.GetNonce(address) != 0 || (contractHash != (common.Hash{}) && contractHash != emptyCodeHash) {
		return nil, common.Address{}, ErrContractAddressCollision
	}

	evm.StateDB.CreateAccount(address)
	evm.Context.Transfer(evm.StateDB, caller.Address(), address, value)

	contract := NewContract(caller, AccountRef(address), value)
	evm.StateDB.SetCode(address, codeAndHash.code)
	contract.SetCodeOptionalHash(&address, codeAndHash)

	ret, err := evm.interpreter.Run(contract, nil, false)
	if err != nil {
		fmt.Printf("res: %x err: %v addr: %s\n", ret, err, address)
	} else {
		evm.StateDB.SetCode(address, contract.Code)
	}

	return ret, address, err
}

func (evm *EVM) Create(caller ContractRef, code []byte, value *big.Int) (ret []byte, contractAddr common.Address, err error) {
	contractAddr = crypto.CreateAddress(caller.Address(), evm.StateDB.GetNonce(caller.Address()))
	return evm.create(caller, &codeAndHash{code: code}, value, contractAddr, CREATE)
}
