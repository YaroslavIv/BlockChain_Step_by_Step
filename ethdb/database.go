package ethdb

import "io"

type KeyValueReader interface {
	Has(key []byte) (bool, error)

	Get(key []byte) ([]byte, error)
}

type KeyValueWriter interface {
	Put(key, value []byte) error

	Delete(key []byte) error
}

type KeyValueStore interface {
	KeyValueReader
	KeyValueWriter
	Batcher

	io.Closer
}

type Reader interface {
	KeyValueReader
}

type Writer interface {
	KeyValueWriter
}

type Database interface {
	Reader
	Writer
	Batcher

	io.Closer
}
