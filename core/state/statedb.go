package state

import (
	"bcsbs/core/types"
	"bcsbs/core/vm"
	"bcsbs/trie"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

type StateDB struct {
	tx_trie      Trie
	storage_trie Trie

	evm *vm.EVM

	stateObjects map[common.Address]*stateObject
}

func New(tx_trie, storage_trie Trie, blockCtx *vm.BlockContext) (*StateDB, error) {
	sdb := &StateDB{
		tx_trie:      tx_trie,
		storage_trie: storage_trie,

		stateObjects: make(map[common.Address]*stateObject),
	}
	sdb.evm = vm.NewEVM(sdb, blockCtx)
	return sdb, nil
}

func (s *StateDB) Copy() *StateDB {
	st := &StateDB{
		tx_trie:      s.tx_trie,
		storage_trie: s.storage_trie,
		stateObjects: make(map[common.Address]*stateObject),
	}

	st.stateObjects[common.Address{}] = nil

	for addr, state := range s.stateObjects {
		if state != nil {
			st.stateObjects[addr] = state.deepCopy(s)
		}
	}

	return st
}

// STATE

func (s *StateDB) getStateObject(addr common.Address) *stateObject {
	if obj := s.getDeletedStateObject(addr); obj != nil && !obj.deleted {
		return obj
	}
	return nil
}

func (s *StateDB) getDeletedStateObject(addr common.Address) *stateObject {
	if obj := s.stateObjects[addr]; obj != nil {
		return obj
	}

	var data *types.StateAccount

	if data == nil {
		data_byte, err := s.tx_trie.TryGet(addr.Bytes())
		data, _ = trie.Byte2StateAccount(data_byte)
		if err != nil {
			fmt.Printf("getDeleteStateObject (%x) error: %s\n", addr.Bytes(), err)
			return nil
		}
		if data == nil {
			return nil
		}
	}

	obj := newObject(s, addr, *data)
	s.setStateObject(obj)
	return obj
}

func (s *StateDB) setStateObject(object *stateObject) {
	s.stateObjects[object.Address()] = object
}

func (s *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		stateObject, _ = s.createObject(addr)
	}
	return stateObject
}

func (s *StateDB) CreateAccount(addr common.Address) {
	newObj, prev := s.createObject(addr)
	if prev != nil {
		newObj.setBalance(prev.data.Balance)
	}
}

func (s *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = s.getDeletedStateObject(addr)

	newobj = newObject(s, addr, types.StateAccount{})

	s.setStateObject(newobj)
	if prev != nil && !prev.deleted {
		return newobj, prev
	}
	return newobj, nil
}

// UPDATE

func (s *StateDB) ApplyTx(tx *types.Transaction) bool {
	stateObject := s.GetOrNewStateObject(*tx.Sender())
	if stateObject != nil && stateObject.Balance().Cmp(tx.Value()) >= 0 {
		if *tx.To() != (common.Address{}) {
			s.evm.Call(vm.AccountRef(*tx.Sender()), *tx.To(), tx.Data(), tx.Value())
		} else {
			s.evm.Create(vm.AccountRef(*tx.Sender()), tx.Data(), tx.Value())
		}

		return true
	}
	return false
}

func (s *StateDB) UpdateStateObject(obj *stateObject) {
	addr := obj.Address()

	if data, err := trie.StateAccount2Byte(&obj.data); err == nil {
		if err := s.tx_trie.TryUpdate(addr[:], data); err != nil {
			panic(fmt.Errorf("updateStateObject (%x) error: %v", addr[:], err))
		}
	} else {
		panic(fmt.Errorf("StateAccount2Byte error: %v", err))
	}

}

// GET

func (s *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}

func (s *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

func (s *StateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(s.storage_trie, hash)
	}
	return common.Hash{}
}

func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (s *StateDB) GetCode(addr common.Address) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code()
	}
	return nil
}

// SET

func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
		s.UpdateStateObject(stateObject)
	}
}

func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
		s.UpdateStateObject(stateObject)
	}
}

func (s *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
		s.UpdateStateObject(stateObject)
	}
}

func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
		s.UpdateStateObject(stateObject)
	}
}

func (s *StateDB) SetState(addr common.Address, key, value common.Hash) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(s.storage_trie, key, value)
	}
}

func (s *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
		s.UpdateStateObject(stateObject)
	}
}
