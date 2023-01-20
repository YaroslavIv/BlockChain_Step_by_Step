package state

import "bcsbs/core/types"

type Trie interface {
	TryGetAccount(key []byte) (*types.StateAccount, error)

	TryUpdateAccount(key []byte, account *types.StateAccount) error
}
