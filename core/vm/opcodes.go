package vm

import "fmt"

type OpCode byte

func (op OpCode) IsPush() bool {
	switch op {
	case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
		return true
	}
	return false
}

// 0x0 range - arithmetic ops.
const (
	STOP OpCode = 0x0
	SUB  OpCode = 0x3
)

// 0x10
const (
	ISZERO OpCode = 0x15
	AND    OpCode = 0x16
	NOT    OpCode = 0x19
)

// 0x30
const (
	CALLER OpCode = 0x33
)

// 0x50 range - 'storage' and execution.
const (
	POP    OpCode = 0x50
	SLOAD  OpCode = 0x54
	SSTORE OpCode = 0x55
	JUMPI  OpCode = 0x57
)

// 0x60 range - pushes.
const (
	PUSH1 OpCode = 0x60 + iota
	PUSH2
	PUSH3
	PUSH4
	PUSH5
	PUSH6
	PUSH7
	PUSH8
	PUSH9
	PUSH10
	PUSH11
	PUSH12
	PUSH13
	PUSH14
	PUSH15
	PUSH16
	PUSH17
	PUSH18
	PUSH19
	PUSH20
	PUSH21
	PUSH22
	PUSH23
	PUSH24
	PUSH25
	PUSH26
	PUSH27
	PUSH28
	PUSH29
	PUSH30
	PUSH31
	PUSH32
)

// 0x90
const (
	SWAP1 = 0x90 + iota
)

// 0xb0 - Game
const (
	INIT OpCode = 0xb0
	MOVE OpCode = 0xb1
)

// 0xf0
const (
	CREATE OpCode = 0xf0
)

var opCodeToString = map[OpCode]string{
	// 0x0
	STOP: "STOP",
	SUB:  "SUB",

	// 0x10
	ISZERO: "ISZERO",
	AND:    "AND",
	NOT:    "NOT",

	// 0x30
	CALLER: "CALLER",

	// 0x50
	POP:    "POP",
	SLOAD:  "SLOAD",
	SSTORE: "SSTORE",
	JUMPI:  "JUMPI",

	// 0x60
	PUSH1:  "PUSH1",
	PUSH2:  "PUSH2",
	PUSH3:  "PUSH3",
	PUSH4:  "PUSH4",
	PUSH5:  "PUSH5",
	PUSH6:  "PUSH6",
	PUSH7:  "PUSH7",
	PUSH8:  "PUSH8",
	PUSH9:  "PUSH9",
	PUSH10: "PUSH10",
	PUSH11: "PUSH11",
	PUSH12: "PUSH12",
	PUSH13: "PUSH13",
	PUSH14: "PUSH14",
	PUSH15: "PUSH15",
	PUSH16: "PUSH16",
	PUSH17: "PUSH17",
	PUSH18: "PUSH18",
	PUSH19: "PUSH19",
	PUSH20: "PUSH20",
	PUSH21: "PUSH21",
	PUSH22: "PUSH22",
	PUSH23: "PUSH23",
	PUSH24: "PUSH24",
	PUSH25: "PUSH25",
	PUSH26: "PUSH26",
	PUSH27: "PUSH27",
	PUSH28: "PUSH28",
	PUSH29: "PUSH29",
	PUSH30: "PUSH30",
	PUSH31: "PUSH31",
	PUSH32: "PUSH32",

	// 0x90
	SWAP1: "SWAP1",

	// 0xb0
	INIT: "INIT",
	MOVE: "MOVE",

	// 0xf0
	CREATE: "CREATE",
}

func (op OpCode) String() string {
	str := opCodeToString[op]
	if len(str) == 0 {
		return fmt.Sprintf("opcode %#x not defined", int(op))
	}

	return str
}

var stringToOp = map[string]OpCode{
	// 0x0
	"STOP": STOP,
	"SUB":  SUB,

	// 0x10
	"ISZERO": ISZERO,
	"AND":    AND,
	"NOT":    NOT,

	// 0x30
	"CALLER": CALLER,

	// 0x50
	"POP":    POP,
	"SLOAD":  SLOAD,
	"SSTORE": SSTORE,
	"JUMPI":  JUMPI,

	// 0x60
	"PUSH1":  PUSH1,
	"PUSH2":  PUSH2,
	"PUSH3":  PUSH3,
	"PUSH4":  PUSH4,
	"PUSH5":  PUSH5,
	"PUSH6":  PUSH6,
	"PUSH7":  PUSH7,
	"PUSH8":  PUSH8,
	"PUSH9":  PUSH9,
	"PUSH10": PUSH10,
	"PUSH11": PUSH11,
	"PUSH12": PUSH12,
	"PUSH13": PUSH13,
	"PUSH14": PUSH14,
	"PUSH15": PUSH15,
	"PUSH16": PUSH16,
	"PUSH17": PUSH17,
	"PUSH18": PUSH18,
	"PUSH19": PUSH19,
	"PUSH20": PUSH20,
	"PUSH21": PUSH21,
	"PUSH22": PUSH22,
	"PUSH23": PUSH23,
	"PUSH24": PUSH24,
	"PUSH25": PUSH25,
	"PUSH26": PUSH26,
	"PUSH27": PUSH27,
	"PUSH28": PUSH28,
	"PUSH29": PUSH29,
	"PUSH30": PUSH30,
	"PUSH31": PUSH31,
	"PUSH32": PUSH32,

	// 0x90
	"SWAP1": SWAP1,

	// 0xb0
	"INIT": INIT,
	"MOVE": MOVE,

	// 0xf0
	"CREATE": CREATE,
}

func StringToOp(str string) OpCode {
	return stringToOp[str]
}
