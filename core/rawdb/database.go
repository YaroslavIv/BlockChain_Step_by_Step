package rawdb

import (
	"bcsbs/ethdb"
	"bcsbs/ethdb/leveldb"
)

func NewLevelDBDatabase(file string, cache int, handles int, namespace string, readonly bool) (ethdb.Database, error) {
	db, err := leveldb.New(file, cache, handles, namespace, readonly)
	if err != nil {
		return nil, err
	}
	return NewDatabase(db), nil
}

type nofreezedb struct {
	ethdb.KeyValueStore
}

func NewDatabase(db ethdb.KeyValueStore) ethdb.Database {
	return &nofreezedb{KeyValueStore: db}
}
