package ethash

import (
	"bcsbs/consensus"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

var (
	FrontierBlockReward = big.NewInt(5e+18)
)

var (
	errOlderBlockTime = errors.New("timestamp older than parent")
	errInvalidPoW     = errors.New("invalid proof-of-work")
	errParentHash     = errors.New("invalid parentHash")
)

func (ethash *Ethash) SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	enc := []interface{}{
		header.ParentHash,
		header.TxHash,
		header.Number,
		header.Time,
	}

	rlp.Encode(hasher, enc)
	hasher.Sum(hash[:0])
	return hash
}

func (ethash *Ethash) VerifyHeader(parent, header *types.Header, seal bool) error {
	return ethash.verifyHeader(header, parent, seal)
}

func (ethash *Ethash) verifyHeader(header, parent *types.Header, seal bool) error {
	if header.Time <= parent.Time {
		return errOlderBlockTime
	}

	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}

	if parent.Hash() != header.ParentHash {
		return errParentHash
	}

	if seal {
		if err := ethash.verifySeal(header); err != nil {
			return err
		}
	}

	return nil
}

func (ethash *Ethash) verifySeal(header *types.Header) error {

	var result []byte

	result = hashimotoFull(ethash.SealHash(header).Bytes(), header.Nonce.Uint64())
	target := new(big.Int).Div(two256, ethash.Target)

	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}

	return nil
}

func (ethash *Ethash) Prepare(parent, header *types.Header) error {
	return nil
}

func (ethash *Ethash) Finalize(header *types.Header, state *state.StateDB, txs []*types.Transaction) {
	accumulateRewards(state, header)
}

func (ethash *Ethash) FinalizeAndAssemble(header *types.Header, state *state.StateDB, txs []*types.Transaction) (*types.Block, error) {
	ethash.Finalize(header, state, txs)

	return types.NewBlock(header, txs), nil
}

func accumulateRewards(state *state.StateDB, header *types.Header) {
	blockReward := FrontierBlockReward

	state.AddBalance(header.Coinbase, blockReward)
}
