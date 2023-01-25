package trie

import (
	"bcsbs/core/rawdb"
	"bcsbs/ethdb"
	"fmt"
)

type StorageTrie struct {
	cache map[string][]byte
	db    ethdb.Database
}

func NewStorageTrie(db ethdb.Database) (*StorageTrie, error) {
	t := &StorageTrie{
		cache: make(map[string][]byte),
		db:    db,
	}

	return t, nil
}

func get_key(root, key []byte) string {
	return string(append(root, key...))
}

func (t *StorageTrie) TryGet(key []byte) ([]byte, error) {
	res := t.cache[string(key)]
	if res != nil {
		return res, nil
	}

	if t.db == nil {
		return nil, fmt.Errorf("Not find key and not DB")
	}

	if val := rawdb.ReadStorage(t.db, key); val != nil {
		t.cache[string(key)] = val
		return val, nil
	}

	return nil, fmt.Errorf("Not find key")
}

func (t *StorageTrie) TryUpdate(key, value []byte) error {
	if t.db != nil {
		rawdb.WriteStorage(t.db, key, value)
	}

	t.cache[string(key)] = value
	return nil
}
