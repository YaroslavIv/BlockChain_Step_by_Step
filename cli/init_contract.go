package cli

import (
	"bcsbs/core/types"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

func (cli *CLI) initContract(address, key string, amount int) {
	code := initCode(address)

	private_key, err := crypto.HexToECDSA(key)
	if err != nil {
		panic(err)
	}

	tx := types.NewContractCreation(0, big.NewInt(int64(amount)), code)
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
