package rawdb

import (
	"bcsbs/core/types"
	"bcsbs/ethdb"
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// Header Hash

func ReadCanonicalHash(db ethdb.Reader, number uint64) common.Hash {
	data, _ := db.Get(headerHashKey(number))
	return common.BytesToHash(data)
}

func WriteCanonicalHash(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	if err := db.Put(headerHashKey(number), hash.Bytes()); err != nil {
		fmt.Println("Failed to store number to hash mapping", "err", err)
	}
}

func DeleteCanonicalHash(db ethdb.KeyValueWriter, number uint64) {
	if err := db.Delete(headerHashKey(number)); err != nil {
		fmt.Println("Failed to delete number to hash mapping", "err", err)
	}
}

// Header Number

func ReadHeaderNumber(db ethdb.KeyValueReader, hash common.Hash) *uint64 {
	data, _ := db.Get(headerNumberKey(hash))
	if len(data) != 8 {
		return nil
	}
	number := binary.BigEndian.Uint64(data)
	return &number
}

func WriteHeaderNumber(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	key := headerNumberKey(hash)
	enc := encodeBlockNumber(number)

	if err := db.Put(key, enc); err != nil {
		fmt.Println("Failed to store hash to number mapping", "err", err)
	}
}

func DeleteHeaderNumber(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(headerNumberKey(hash)); err != nil {
		fmt.Println("Failed to delete hash to number mapping", "err", err)
	}
}

// Current

func ReadHeadHeaderHash(db ethdb.KeyValueReader) common.Hash {
	data, _ := db.Get(headHeaderKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

func WriteHeadHeaderHash(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Put(headHeaderKey, hash.Bytes()); err != nil {
		fmt.Println("Failed to store last header's hash", "err", err)
	}
}

func ReadHeadBlockHash(db ethdb.KeyValueReader) common.Hash {
	data, _ := db.Get(headBlockKey)
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

func WriteHeadBlockHash(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Put(headBlockKey, hash.Bytes()); err != nil {
		fmt.Println("Failed to store last block's hash", "err", err)
	}
}

func ReadHeadHeader(db ethdb.Reader) *types.Header {
	headHeaderHash := ReadHeadHeaderHash(db)
	if headHeaderHash == (common.Hash{}) {
		return nil
	}
	headHeaderNumber := ReadHeaderNumber(db, headHeaderHash)
	if headHeaderNumber == nil {
		return nil
	}
	return ReadHeader(db, headHeaderHash, *headHeaderNumber)
}

func ReadHeadBlock(db ethdb.Reader) *types.Block {
	headBlockHash := ReadHeadBlockHash(db)
	if headBlockHash == (common.Hash{}) {
		return nil
	}
	headBlockNumber := ReadHeaderNumber(db, headBlockHash)
	if headBlockNumber == nil {
		return nil
	}
	return ReadBlock(db, headBlockHash, *headBlockNumber)
}

// Header

func HasHeader(db ethdb.Reader, hash common.Hash, number uint64) bool {
	if has, err := db.Has(headerKey(number, hash)); !has || err != nil {
		return false
	}
	return true
}

func ReadHeaderRLP(db ethdb.Reader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(headerKey(number, hash))
	return data
}

func ReadHeader(db ethdb.Reader, hash common.Hash, number uint64) *types.Header {
	data := ReadHeaderRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	header := new(types.Header)
	if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
		fmt.Println("Invalid block header RLP", "hash", hash, "err", err)
		return nil
	}
	return header
}

func WriteHeader(db ethdb.KeyValueWriter, header *types.Header) {
	var (
		hash   = header.Hash()
		number = header.Number.Uint64()
	)
	WriteHeaderNumber(db, hash, number)

	data, err := rlp.EncodeToBytes(header)
	if err != nil {
		fmt.Println("Failed to RLP encode header", "err", err)
	}
	key := headerKey(number, hash)
	if err := db.Put(key, data); err != nil {
		fmt.Println("Failed to store header", "err", err)
	}
}

func DeleteHeader(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	deleteHeaderWithoutNumber(db, hash, number)
	if err := db.Delete(headerNumberKey(hash)); err != nil {
		fmt.Println("Failed to delete hash to number mapping", "err", err)
	}
}

func deleteHeaderWithoutNumber(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	if err := db.Delete(headerKey(number, hash)); err != nil {
		fmt.Println("Failed to delete header", "err", err)
	}
}

// Body

func HasBody(db ethdb.Reader, hash common.Hash, number uint64) bool {
	if has, err := db.Has(blockBodyKey(number, hash)); !has || err != nil {
		return false
	}
	return true
}

func ReadBodyRLP(db ethdb.Reader, hash common.Hash, number uint64) rlp.RawValue {
	data, _ := db.Get(blockBodyKey(number, hash))
	return data
}

func ReadBody(db ethdb.Reader, hash common.Hash, number uint64) *types.Body {
	data := ReadBodyRLP(db, hash, number)
	if len(data) == 0 {
		return nil
	}
	body := new(types.Body)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		fmt.Println("Invalid block body RLP", "hash", hash, "err", err)
		return nil
	}
	return body
}

func WriteBodyRLP(db ethdb.KeyValueWriter, hash common.Hash, number uint64, rlp rlp.RawValue) {
	if err := db.Put(blockBodyKey(number, hash), rlp); err != nil {
		fmt.Println("Failed to store block body", "err", err)
	}
}

func WriteBody(db ethdb.KeyValueWriter, hash common.Hash, number uint64, body *types.Body) {
	data, err := rlp.EncodeToBytes(body)
	if err != nil {
		fmt.Println("Failed to RLP encode body", "err", err)
	}
	WriteBodyRLP(db, hash, number, data)
}

func DeleteBody(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	if err := db.Delete(blockBodyKey(number, hash)); err != nil {
		fmt.Println("Failed to delete block body", "err", err)
	}
}

// Block

func ReadBlock(db ethdb.Reader, hash common.Hash, number uint64) *types.Block {
	header := ReadHeader(db, hash, number)
	if header == nil {
		return nil
	}
	body := ReadBody(db, hash, number)
	if body == nil {
		return nil
	}
	return types.NewBlockWithHeader(header).WithBody(body.Transactions)
}

func WriteBlock(db ethdb.KeyValueWriter, block *types.Block) {
	WriteBody(db, block.Hash(), block.NumberU64(), block.Body())
	WriteHeader(db, block.Header())
}

func DeleteBlock(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	DeleteHeader(db, hash, number)
	DeleteBody(db, hash, number)
}

func DeleteBlockWithoutNumber(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
	deleteHeaderWithoutNumber(db, hash, number)
	DeleteBody(db, hash, number)
}
