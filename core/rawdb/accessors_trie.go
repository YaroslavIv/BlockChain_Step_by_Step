package rawdb

import (
	"bcsbs/core/types"
	"bcsbs/ethdb"
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func HasAccountData(db ethdb.Reader, addr common.Address) bool {
	if has, err := db.Has(accountData(addr)); !has || err != nil {
		return false
	}
	return true
}

func ReadAccountDataRLP(db ethdb.Reader, addr common.Address) rlp.RawValue {
	data, _ := db.Get(accountData(addr))
	return data
}

func ReadAccountData(db ethdb.Reader, addr common.Address) *types.StateAccount {
	data := ReadAccountDataRLP(db, addr)
	if len(data) == 0 {
		return nil
	}

	acc := new(types.StateAccount)
	if err := rlp.Decode(bytes.NewReader(data), acc); err != nil {
		fmt.Println("Invalid block Account RLP", "addr", addr, "err", err)
		return nil
	}
	return acc
}

func WriteAccountData(db ethdb.Writer, addr common.Address, acc *types.StateAccount) {
	data, err := rlp.EncodeToBytes(acc)
	if err != nil {
		fmt.Println("Failed to RLP encode account", "err", err)
	}
	key := accountData(addr)
	if err := db.Put(key, data); err != nil {
		fmt.Println("Failed to store header", "err", err)
	}
}
