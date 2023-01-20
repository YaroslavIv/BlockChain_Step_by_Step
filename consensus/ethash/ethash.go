package ethash

import (
	"math/big"
	"math/rand"
	"sync"
)

var (
	two256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
)

type Ethash struct {
	Target *big.Int

	rand    *rand.Rand
	threads int

	lock sync.Mutex
}
