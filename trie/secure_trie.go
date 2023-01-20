package trie

import (
	"bcsbs/core/types"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

type StateTrie struct {
	db map[string][]byte
}

func NewStateTrie() (*StateTrie, error) {
	t := &StateTrie{
		db: make(map[string][]byte),
	}

	return t, nil
}

func (t *StateTrie) TryGetAccount(key []byte) (*types.StateAccount, error) {
	var ret types.StateAccount
	res := t.db[string(key)]
	if res == nil {
		return nil, fmt.Errorf("Not find key")
	}
	err := rlp.DecodeBytes(res, &ret)
	return &ret, err
}

func (t *StateTrie) TryUpdateAccount(key []byte, acc *types.StateAccount) error {
	data, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return err
	}
	t.db[string(key)] = data
	return nil
}
