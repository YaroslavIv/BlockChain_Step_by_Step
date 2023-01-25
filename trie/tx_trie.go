package trie

import (
	"bcsbs/core/rawdb"
	"bcsbs/core/types"
	"bcsbs/ethdb"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type TxTrie struct {
	cache map[string][]byte
	db    ethdb.Database
}

func NewTxTrie(db ethdb.Database) (*TxTrie, error) {
	t := &TxTrie{
		cache: make(map[string][]byte),
		db:    db,
	}

	return t, nil
}

func Byte2StateAccount(res []byte) (*types.StateAccount, error) {
	var ret types.StateAccount
	if res != nil {
		err := rlp.DecodeBytes(res, &ret)
		return &ret, err
	}
	return nil, fmt.Errorf("Not correct res")
}

func StateAccount2Byte(acc *types.StateAccount) ([]byte, error) {
	data, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (t *TxTrie) TryGet(key []byte) ([]byte, error) {
	res := t.cache[string(key)]
	if res != nil {
		return res, nil
	}

	if t.db == nil {
		return nil, fmt.Errorf("Not find key and not DB")
	}

	if acc := rawdb.ReadAccountData(t.db, common.BytesToAddress(key)); acc != nil {
		if data, err := rlp.EncodeToBytes(acc); err == nil {
			t.cache[string(key)] = data
			return data, nil
		}
	}

	return nil, fmt.Errorf("Not find key")
}

func (t *TxTrie) TryUpdate(key []byte, value []byte) error {
	if t.db != nil {
		rawdb.WriteAccountData(t.db, common.BytesToAddress(key), value)
	}

	t.cache[string(key)] = value
	return nil
}
