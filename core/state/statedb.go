package state

import (
	"bcsbs/core/types"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var (
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

type StateDB struct {
	trie Trie

	stateObjects map[common.Address]*stateObject
}

func New(trie Trie) (*StateDB, error) {
	sdb := &StateDB{
		trie: trie,

		stateObjects: make(map[common.Address]*stateObject),
	}

	return sdb, nil
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
		var err error
		data, err = s.trie.TryGetAccount(addr.Bytes())
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
	stateObject := s.getStateObject(*tx.Sender())
	if stateObject != nil && stateObject.Balance().Cmp(tx.Value()) >= 0 {
		stateObject.setNonce(stateObject.Nonce() + 1)
		stateObject.AddBalance(tx.Value().Neg(tx.Value()))
		s.AddBalance(*tx.To(), tx.Value())
		return true
	}
	return false
}

func (s *StateDB) UpdateStateObject(obj *stateObject) {
	addr := obj.Address()
	if err := s.trie.TryUpdateAccount(addr[:], &obj.data); err != nil {
		panic(fmt.Errorf("updateStateObject (%x) error: %v", addr[:], err))
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

// SET

func (s *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

func (s *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

func (s *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := s.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}
