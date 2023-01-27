package cli

import (
	"bcsbs/core/vm"
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/rpc/json"
)

func initCode(addr string) []byte {
	var code string
	for _, i := range []string{
		"INIT",

		"CALLER", "PUSH1", "00", "SSTORE",
		"PUSH20", addr[2:], "PUSH1", "01", "SSTORE",
		"PUSH1", "01", "PUSH1", "02", "SSTORE",
	} {
		if len(i) > 2 && len(i) < 20 {
			code += strconv.FormatInt(int64(vm.StringToOp(i)), 16)
		} else {
			code += i
		}
	}

	code_hex, err := hex.DecodeString(code)
	if err != nil {
		panic(fmt.Sprintf("code: %x err: %s", code, err))
	}

	return code_hex
}

func moveCode(position int) []byte {
	var code string
	for _, i := range []string{
		"MOVE",

		"PUSH1", fmt.Sprintf("0%x", position), "SLOAD", "ISZERO",
		"PUSH1", "09", "JUMPI", "00",

		"PUSH1", "02", "SLOAD",

		"PUSH1", "00", "SLOAD",
		"CALLER",
		"03",

		"ISZERO", "AND", "PUSH1", "30", "JUMPI",

		"PUSH1", "02", "SLOAD", "ISZERO",

		"PUSH1", "01", "SLOAD",
		"CALLER",
		"03",
		"ISZERO", "AND", "PUSH1", "25", "JUMPI", "00",

		"PUSH1", "02", "PUSH1", fmt.Sprintf("0%x", position), "SSTORE", "PUSH1", "01", "PUSH1", "02", "SSTORE", "00",
		"PUSH1", "01", "PUSH1", fmt.Sprintf("0%x", position), "SSTORE", "PUSH1", "00", "PUSH1", "02", "SSTORE", "00",
	} {
		if len(i) > 2 && len(i) < 20 {
			code += strconv.FormatInt(int64(vm.StringToOp(i)), 16)
		} else {
			code += i
		}
	}

	code_hex, err := hex.DecodeString(code)
	if err != nil {
		panic(err)
	}

	return code_hex
}

func send(url string, sign []byte) {

	txArgs := &TxArgs{
		Sign: common.Bytes2Hex(sign),
	}

	message, err := json.EncodeClientRequest("server.SendRawTransaction", txArgs)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var result Response
	err = json.DecodeClientResponse(resp.Body, &result)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
