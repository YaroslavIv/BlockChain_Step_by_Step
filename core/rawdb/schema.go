package rawdb

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
)

var (
	headHeaderKey = []byte("LastHeader")
	headBlockKey  = []byte("LastBlock")

	headerPrefix       = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
	headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)
	headerHashSuffix   = []byte("n") // headerPrefix + num (uint64 big endian) + headerHashSuffix -> hash

	blockBodyPrefix = []byte("b") // blockBodyPrefix + num (uint64 big endian) + hash -> block body

	txLookupPrefix = []byte("l") // txLookupPrefix + hash -> transaction/receipt lookup metadata

	accountDataPrefix = []byte("a") // accountDataPrefix + addr -> StateAccount

	storagePrefix = []byte("s") // storagePrefix + root + key -> Storage metadata
)

func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func headerNumberKey(hash common.Hash) []byte {
	return append(headerNumberPrefix, hash.Bytes()...)
}

func headerHashKey(number uint64) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)
}

func headerKey(number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

func blockBodyKey(number uint64, hash common.Hash) []byte {
	return append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

func txLookupKey(hash common.Hash) []byte {
	return append(txLookupPrefix, hash.Bytes()...)
}

func accountData(addr common.Address) []byte {
	return append(accountDataPrefix, addr.Bytes()...)
}

func storage(key []byte) []byte {
	return append(storagePrefix, key...)
}
