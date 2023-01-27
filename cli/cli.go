package cli

import (
	"flag"
	"fmt"
	"os"
)

type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  startserver -address ADDRESS - Start Server")
	fmt.Println("  initcontract -address ADDRESS -key KEY -amount AMOUNT - Create contract")
	fmt.Println("  sendtx -address ADDRESS -key KEY -amount AMOUNT - Send coin")
	fmt.Println("  move -address ADDRESS -x X -y Y -key KEY - Send position")
	fmt.Println("  createwallet -dir DIR - Generates a new key-pair and saves it into the wallet file")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	startServerCmd := flag.NewFlagSet("startserver", flag.ExitOnError)
	initContractCmd := flag.NewFlagSet("initcontract", flag.ExitOnError)
	sendTxCmd := flag.NewFlagSet("sendtx", flag.ExitOnError)
	moveCmd := flag.NewFlagSet("move", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	startServerAddress := startServerCmd.String("address", "", "The address Coinbase")

	initContractAddress := initContractCmd.String("address", "", "The address player")
	initContractKey := initContractCmd.String("key", "", "the private key")
	initContractAmount := initContractCmd.Int("amount", 0, "the bet")

	sendTxAddress := sendTxCmd.String("address", "", "The To player")
	sendTxKey := sendTxCmd.String("key", "", "the private key")
	sendTxAmount := sendTxCmd.Int("amount", 0, "the amount of coun")

	moveAddress := moveCmd.String("address", "", "the adress contract")
	moveX := moveCmd.Int("x", -1, "the coordinat X")
	moveY := moveCmd.Int("y", -1, "the coordinat Y")
	moveKey := moveCmd.String("key", "", "the private key")

	createwalletDir := createWalletCmd.String("dir", "./", "the dir save file")
	createwalletPassphrase := createWalletCmd.String("passphrase", "", "the crypto phrase")

	switch os.Args[1] {
	case "startserver":
		err := startServerCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "initcontract":
		err := initContractCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "sendtx":
		err := sendTxCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "move":
		err := moveCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}

	}

	if startServerCmd.Parsed() {
		if *startServerAddress == "" {
			startServerCmd.Usage()
			os.Exit(1)
		}
		cli.startServer(*startServerAddress)

	} else if initContractCmd.Parsed() {
		if *initContractAddress == "" || *initContractKey == "" || *initContractAmount < 0 {
			initContractCmd.Usage()
			os.Exit(1)
		}
		cli.initContract(*initContractAddress, *initContractKey, *initContractAmount)

	} else if sendTxCmd.Parsed() {
		if *sendTxAddress == "" || *sendTxKey == "" || *sendTxAmount < 0 {
			sendTxCmd.Usage()
			os.Exit(1)
		}
		cli.sendTx(*sendTxAddress, *sendTxKey, *sendTxAmount)

	} else if moveCmd.Parsed() {
		if *moveAddress == "" || *moveX < 0 || *moveX > 2 ||
			*moveY < 0 || *moveY > 2 || *moveKey == "" {
			moveCmd.Usage()
			os.Exit(1)
		}
		cli.move(*moveAddress, *moveX, *moveY, *moveKey)

	} else if createWalletCmd.Parsed() {
		cli.createWallet(*createwalletDir, *createwalletPassphrase)

	} else {
		cli.printUsage()
		os.Exit(1)
	}

}
