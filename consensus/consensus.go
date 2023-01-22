package consensus

import (
	"bcsbs/core/state"
	"bcsbs/core/types"

	"github.com/ethereum/go-ethereum/common"
)

type Engine interface {
	VerifyHeader(parent, header *types.Header, seal bool) error

	Seal(block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error

	SealHash(header *types.Header) common.Hash

	Prepare(parent, header *types.Header) error

	Finalize(header *types.Header, state *state.StateDB, txs []*types.Transaction)

	FinalizeAndAssemble(header *types.Header, state *state.StateDB, txs []*types.Transaction) (*types.Block, error)
}
