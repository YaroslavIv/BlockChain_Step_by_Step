package cli

import (
	"bcsbs/core/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (cli *CLI) move(address string, x, y int, key string) {
	addr_contract := common.HexToAddress(address)
	position := x*3 + y + 3
	code := moveCode(position)

	tx := types.NewTransaction(0, addr_contract, big.NewInt(0), code)

	private_key, err := crypto.HexToECDSA(key)
	if err != nil {
		panic(err)
	}

	tx_sign, err := types.SignTx(tx, types.HomesteadSigner{}, private_key)
	if err != nil {
		panic(err)
	}

	sign, err := tx_sign.MarshalBinary()
	if err != nil {
		panic(err)
	}

	send("http://93.157.234.29:1337/delivery", sign)
}
