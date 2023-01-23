package trie

import (
	"bcsbs/core/rawdb"
	"bcsbs/core/types"
	"bcsbs/ethdb"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type StateTrie struct {
	cache map[string][]byte
	db    ethdb.Database
}

func NewStateTrie(db ethdb.Database) (*StateTrie, error) {
	t := &StateTrie{
		cache: make(map[string][]byte),
		db:    db,
	}

	return t, nil
}

func (t *StateTrie) TryGetAccount(key []byte) (*types.StateAccount, error) {
	var ret types.StateAccount
	res := t.cache[string(key)]
	if res != nil {
		err := rlp.DecodeBytes(res, &ret)
		return &ret, err
	}

	if t.db == nil {
		return nil, fmt.Errorf("Not find key and not DB")
	}

	if acc := rawdb.ReadAccountData(t.db, common.BytesToAddress(key)); acc != nil {
		if data, err := rlp.EncodeToBytes(acc); err == nil {
			t.cache[string(key)] = data
		}

		return acc, nil
	}

	return nil, fmt.Errorf("Not find key")
}

func (t *StateTrie) TryUpdateAccount(key []byte, acc *types.StateAccount) error {
	data, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return err
	}
	if t.db != nil {
		rawdb.WriteAccountData(t.db, common.BytesToAddress(key), acc)
	}

	t.cache[string(key)] = data
	return nil
}
