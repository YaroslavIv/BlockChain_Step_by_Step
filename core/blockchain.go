package core

import (
	"bcsbs/consensus"
	"bcsbs/core/rawdb"
	"bcsbs/core/state"
	"bcsbs/core/types"
	"bcsbs/ethdb"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type BlockChain struct {
	blocks types.Blocks

	db ethdb.Database

	genesisBlock *types.Block

	engine consensus.Engine

	statedb *state.StateDB

	mu sync.Mutex
}

func NewBlockChain(db ethdb.Database, engine consensus.Engine, genesis *Genesis, statedb *state.StateDB) *BlockChain {
	bc := &BlockChain{
		db:      db,
		engine:  engine,
		statedb: statedb,
	}

	if genesis != nil {
		bc.AddGenesis(genesis)
	}

	bc.blocks = append(bc.blocks, bc.genesisBlock)

	return bc
}

func (bc *BlockChain) AddGenesis(genesis *Genesis) {
	bc.genesisBlock = genesis.ToBlock()

	rawdb.WriteHeadHeaderHash(bc.db, bc.genesisBlock.Hash())
	rawdb.WriteHeadBlockHash(bc.db, bc.genesisBlock.Hash())
	rawdb.WriteHeaderNumber(bc.db, bc.genesisBlock.Hash(), bc.genesisBlock.NumberU64())
	rawdb.WriteBlock(bc.db, bc.genesisBlock)
}

func (bc *BlockChain) WriteBlockAndSetHead(block *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return bc.writeBlockAndSetHead(block)
}

func (bc *BlockChain) writeBlockAndSetHead(block *types.Block) error {
	currentBlock := bc.CurrentBlock()
	if block.ParentHash() != currentBlock.Hash() {
		return fmt.Errorf("block.ParentHash != parent.Hash %s != %s", block.ParentHash(), currentBlock.Hash())
	}

	rawdb.WriteHeadHeaderHash(bc.db, block.Hash())
	rawdb.WriteHeadBlockHash(bc.db, block.Hash())
	rawdb.WriteHeaderNumber(bc.db, block.Hash(), block.NumberU64())
	rawdb.WriteTxLookupEntriesByBlock(bc.db, block)
	rawdb.WriteBlock(bc.db, block)

	bc.blocks = append(bc.blocks, block)
	return nil
}

func (bc *BlockChain) AddBlock(block *types.Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if err := bc.engine.VerifyHeader(bc.CurrentHeader(), block.Header(), true); err != nil {
		panic(err)
	}

	for _, tx := range block.Body().Transactions {
		bc.statedb.ApplyTx(tx)
	}

	bc.blocks = append(bc.blocks, block)
}

func (bc *BlockChain) InsertChain(chain types.Blocks) (int, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(chain) == 0 {
		return 0, nil
	}

	head := bc.CurrentBlock()
	if chain[0].NumberU64() != head.NumberU64()+1 || chain[0].ParentHash() != head.Hash() {
		return 0, fmt.Errorf("non contiguous insert: item 0 is #%d [%x..], item 1 is #%d [%x..] (parent [%x..])", head.NumberU64(),
			head.Hash().Bytes()[:4], chain[0].NumberU64(), chain[0].Hash().Bytes()[:4], chain[0].ParentHash().Bytes()[:4])
	}
	if err := bc.engine.VerifyHeader(head.Header(), chain[0].Header(), true); err != nil {
		return 0, err
	}

	for i := 1; i < len(chain); i++ {
		block, prev := chain[i], chain[i-1]
		if block.NumberU64() != prev.NumberU64()+1 || block.ParentHash() != prev.Hash() {

			return 0, fmt.Errorf("non contiguous insert: item %d is #%d [%x..], item %d is #%d [%x..] (parent [%x..])", i-1, prev.NumberU64(),
				prev.Hash().Bytes()[:4], i, block.NumberU64(), block.Hash().Bytes()[:4], block.ParentHash().Bytes()[:4])
		}
		if err := bc.engine.VerifyHeader(prev.Header(), block.Header(), true); err != nil {
			return 0, err
		}
	}

	return bc.insertChain(chain, true, true)
}

func (bc *BlockChain) insertChain(chain types.Blocks, verifySeals, setHead bool) (int, error) {
	for _, block := range chain {
		bc.blocks = append(bc.blocks, block)
	}
	return len(chain), nil
}

func (bc *BlockChain) String() string {
	var out string
	block := bc.CurrentBlock()

	for {
		out += block.String() + "\n"
		if block.ParentHash() == (common.Hash{}) && block.NumberU64() < 1 {
			break
		}

		block = rawdb.ReadBlock(bc.db, block.ParentHash(), block.NumberU64()-1)
	}

	return out
}
