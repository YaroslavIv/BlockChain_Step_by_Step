package ethash

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/crypto"
)

func hashimotoFull(hash []byte, nonce uint64) []byte {

	return hashimoto(hash, nonce)
}

func hashimoto(hash []byte, nonce uint64) []byte {

	seed := make([]byte, 40)
	copy(seed, hash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)

	return crypto.Keccak256(seed)
}
