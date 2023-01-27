package core

import "bcsbs/core/types"

type NewTxsEvent struct{ Txs []*types.Transaction }
