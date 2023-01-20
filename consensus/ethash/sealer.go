package ethash

import (
	"bcsbs/core/types"
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"
)

func (ethash *Ethash) Seal(block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {

	abort := make(chan struct{})

	ethash.lock.Lock()
	threads := ethash.threads
	if ethash.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			ethash.lock.Unlock()
			return err
		}
		ethash.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	ethash.lock.Unlock()

	if threads == 0 {
		threads = runtime.NumCPU()
	}
	if threads < 0 {
		threads = 0
	}

	var (
		pend   sync.WaitGroup
		locals = make(chan *types.Block)
	)
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			ethash.mine(block, id, nonce, abort, locals)
		}(i, uint64(ethash.rand.Int63()))
	}

	go func() {
		var result *types.Block
		select {
		case result = <-locals:
			select {
			case results <- result:
			default:
				fmt.Println("Sealing result is not read by miner")
			}
		}
		pend.Wait()
	}()
	return nil
}

func (ethash *Ethash) mine(block *types.Block, id int, seed uint64, abort chan struct{}, found chan *types.Block) {
	var (
		header = block.Header()
		body   = block.Body()
		hash   = ethash.SealHash(header).Bytes()
		target = new(big.Int).Div(two256, ethash.Target)
	)

	var (
		attempts  = int64(0)
		nonce     = seed
		powBuffer = new(big.Int)
	)

search:
	for {
		select {
		case <-abort:
			fmt.Println("Ethash nonce search aborted")
			break search
		default:
			attempts++
			if (attempts % (1 << 15)) == 0 {
				attempts = 0
			}

			result := hashimotoFull(hash, nonce)
			if powBuffer.SetBytes(result).Cmp(target) <= 0 {
				header = types.CopyHeader(header)
				header.Nonce = types.EncodeNonce(nonce)

				select {
				case found <- block.WithSealAndBody(header, body):
					fmt.Println("Ethash nonce found and reported nonce:", nonce)
				case <-abort:
					fmt.Println("Ethash nonce found but discarded")
				}
				break search
			}
			nonce++
		}
	}
}
