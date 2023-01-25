package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type ContractRef interface {
	Address() common.Address
}

type AccountRef common.Address

func (ar AccountRef) Address() common.Address { return (common.Address)(ar) }

type Contract struct {
	CallerAddress common.Address

	caller ContractRef
	self   ContractRef

	Code     []byte
	CodeHash common.Hash
	CodeAddr *common.Address

	value *big.Int
}

func NewContract(caller, object ContractRef, value *big.Int) *Contract {
	c := &Contract{CallerAddress: caller.Address(), caller: caller, self: object}

	c.value = value

	return c
}

func (c *Contract) validJumpdest(dest *uint256.Int) bool {
	udest, overflow := dest.Uint64WithOverflow()

	if overflow || udest >= uint64(len(c.Code)) {
		return false
	}

	return true
}

func (c *Contract) GetOp(n uint64) OpCode {
	if n < uint64(len(c.Code)) {
		return OpCode(c.Code[n])
	}

	return STOP
}

func (c *Contract) Caller() common.Address {
	return c.CallerAddress
}

func (c *Contract) Address() common.Address {
	return c.self.Address()
}

func (c *Contract) Value() *big.Int {
	return c.value
}

func (c *Contract) SetCallCode(addr *common.Address, hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
	c.CodeAddr = addr
}

func (c *Contract) SetCodeOptionalHash(addr *common.Address, codeAndHash *codeAndHash) {
	c.Code = codeAndHash.code
	c.CodeHash = codeAndHash.hash
	c.CodeAddr = addr
}
