package ethdb

type Batch interface {
	KeyValueWriter

	ValueSize() int

	Write() error

	Reset()

	Replay(w KeyValueWriter) error
}

type Batcher interface {
	NewBatch() Batch

	NewBatchWithSize(size int) Batch
}
