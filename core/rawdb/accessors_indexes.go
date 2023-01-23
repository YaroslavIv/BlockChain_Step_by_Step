package rawdb

import (
	"bcsbs/core/types"
	"bcsbs/ethdb"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func ReadTxLookupEntry(db ethdb.Reader, hash common.Hash) *uint64 {
	data, _ := db.Get(txLookupKey(hash))
	if len(data) == 0 {
		return nil
	}

	if len(data) < common.HashLength {
		number := new(big.Int).SetBytes(data).Uint64()
		return &number
	}

	return nil
}

func WriteTxLookupEntries(db ethdb.KeyValueWriter, number uint64, hashes []common.Hash) {
	numberBytes := new(big.Int).SetUint64(number).Bytes()
	for _, hash := range hashes {
		writeTxLookupEntry(db, hash, numberBytes)
	}
}

func writeTxLookupEntry(db ethdb.KeyValueWriter, hash common.Hash, numberBytes []byte) {
	if err := db.Put(txLookupKey(hash), numberBytes); err != nil {
		fmt.Println("Failed to store transaction lookup entry", "err", err)
	}
}

func WriteTxLookupEntriesByBlock(db ethdb.KeyValueWriter, block *types.Block) {
	numberBytes := block.Number().Bytes()
	for _, tx := range block.Transactions() {
		writeTxLookupEntry(db, tx.Hash(), numberBytes)
	}
}

func DeleteTxLookupEntry(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(txLookupKey(hash)); err != nil {
		fmt.Println("Failed to delete transaction lookup entry", "err", err)
	}
}

func DeleteTxLookupEntries(db ethdb.KeyValueWriter, hashes []common.Hash) {
	for _, hash := range hashes {
		DeleteTxLookupEntry(db, hash)
	}
}

func ReadTransaction(db ethdb.Reader, hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64) {

	blockNumber := ReadTxLookupEntry(db, hash)
	if blockNumber == nil {
		return nil, common.Hash{}, 0, 0
	}

	blockHash := ReadCanonicalHash(db, *blockNumber)
	if blockHash == (common.Hash{}) {
		return nil, common.Hash{}, 0, 0
	}

	body := ReadBody(db, blockHash, *blockNumber)
	if body == nil {
		fmt.Println("Transaction referenced missing", "number", *blockNumber, "hash", blockHash)
		return nil, common.Hash{}, 0, 0
	}

	for txIndex, tx := range body.Transactions {
		if tx.Hash() == hash {
			return tx, blockHash, *blockNumber, uint64(txIndex)
		}
	}

	fmt.Println("Transaction not found", "number", *blockNumber, "hash", blockHash, "txhash", hash)
	return nil, common.Hash{}, 0, 0
}
