package leveldb

import (
	"bcsbs/ethdb"
	"fmt"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	minCache   = 16
	minHandles = 16
)

type Database struct {
	fn string
	db *leveldb.DB

	quitLock sync.Mutex
	quitChan chan chan error
}

func New(file string, cache int, handles int, namespace string, readonly bool) (*Database, error) {
	return NewCustom(file, namespace, func(options *opt.Options) {
		if cache < minCache {
			cache = minCache
		}
		if handles < minHandles {
			handles = minHandles
		}
		options.OpenFilesCacheCapacity = handles
		options.BlockCacheCapacity = cache / 2 * opt.MiB
		options.WriteBuffer = cache / 4 * opt.MiB
		if readonly {
			options.ReadOnly = true
		}
	})
}

func NewCustom(file, namespace string, customize func(options *opt.Options)) (*Database, error) {
	options := configureOptions(customize)
	db, err := leveldb.OpenFile(file, options)

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}
	if err != nil {
		return nil, err
	}
	ldb := &Database{
		fn:       file,
		db:       db,
		quitChan: make(chan chan error),
	}

	return ldb, nil
}

func configureOptions(customizeFn func(*opt.Options)) *opt.Options {
	options := &opt.Options{
		Filter:                 filter.NewBloomFilter(10),
		DisableSeeksCompaction: true,
	}

	if customizeFn != nil {
		customizeFn(options)
	}
	return options
}

// io.Closer

func (db *Database) Close() error {
	db.quitLock.Lock()
	defer db.quitLock.Unlock()

	if db.quitChan != nil {
		errc := make(chan error)
		db.quitChan <- errc
		if err := <-errc; err != nil {
			fmt.Println("Metrics collection failed", "err", err)
		}
		db.quitChan = nil
	}
	return db.db.Close()
}

// KeyValueReader

func (db *Database) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

func (db *Database) Get(key []byte) ([]byte, error) {
	dat, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// KeyValueWriter

func (db *Database) Put(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

func (db *Database) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

// Batcher

func (db *Database) NewBatch() ethdb.Batch {
	return &batch{
		db: db.db,
		b:  new(leveldb.Batch),
	}
}

func (db *Database) NewBatchWithSize(size int) ethdb.Batch {
	return &batch{
		db: db.db,
		b:  leveldb.MakeBatch(size),
	}
}

// Batch

type batch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

func (b *batch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(key) + len(value)
	return nil
}

func (b *batch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += len(key)
	return nil
}

func (b *batch) ValueSize() int {
	return b.size
}

func (b *batch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	return b.b.Replay(&replayer{writer: w})
}

// replayer

type replayer struct {
	writer  ethdb.KeyValueWriter
	failure error
}

func (r *replayer) Put(key, value []byte) {
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Put(key, value)
}

func (r *replayer) Delete(key []byte) {
	if r.failure != nil {
		return
	}
	r.failure = r.writer.Delete(key)
}
