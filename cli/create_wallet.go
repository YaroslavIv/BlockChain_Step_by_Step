package cli

import (
	"bcsbs/account/keystore"
	"fmt"
)

func (cli *CLI) createWallet(dir, passphrase string) {
	ks := keystore.NewPlaintextKeyStore(dir)

	key, err := ks.NewAccount(passphrase)
	if err != nil {
		panic(err)
	}
	fmt.Println(key)
}
