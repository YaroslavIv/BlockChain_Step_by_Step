package vm

type (
	executionFunc func(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error)
)

type operation struct {
	execute executionFunc

	minStack int
	maxStack int
}

var (
	frontierInstructionSet = newFrontierInstructionSet()
)

type JumpTable [256]*operation

func newFrontierInstructionSet() JumpTable {
	tbl := JumpTable{
		// 0x0
		STOP: {
			execute:  opStop,
			minStack: minStack(0, 0),
			maxStack: maxStack(0, 0),
		},
		SUB: {
			execute:  opSub,
			minStack: minStack(2, 1),
			maxStack: maxStack(2, 1),
		},

		// 0x10
		ISZERO: {
			execute:  opIszero,
			minStack: minStack(1, 1),
			maxStack: maxStack(1, 1),
		},
		AND: {
			execute:  opAnd,
			minStack: minStack(2, 1),
			maxStack: maxStack(2, 1),
		},
		NOT: {
			execute:  opNot,
			minStack: minStack(1, 1),
			maxStack: maxStack(1, 1),
		},

		// 0x30
		CALLER: {
			execute:  opCaller,
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},

		// 0x50
		POP: {
			execute:  opPop,
			minStack: minStack(1, 0),
			maxStack: maxStack(1, 0),
		},
		SLOAD: {
			execute:  opSload,
			minStack: minStack(1, 1),
			maxStack: maxStack(1, 1),
		},
		SSTORE: {
			execute:  opSstore,
			minStack: minStack(2, 0),
			maxStack: maxStack(2, 0),
		},
		JUMPI: {
			execute:  opJumpi,
			minStack: minStack(2, 0),
			maxStack: maxStack(2, 0),
		},

		// 0x60
		PUSH1: {
			execute:  opPush1,
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH2: {
			execute:  makePush(2, 2),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH3: {
			execute:  makePush(3, 3),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH4: {
			execute:  makePush(4, 4),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH5: {
			execute:  makePush(5, 5),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH6: {
			execute:  makePush(6, 6),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH7: {
			execute:  makePush(7, 7),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH8: {
			execute:  makePush(8, 8),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH9: {
			execute:  makePush(9, 9),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH10: {
			execute:  makePush(10, 10),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH11: {
			execute:  makePush(11, 11),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH12: {
			execute:  makePush(12, 12),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH13: {
			execute:  makePush(13, 13),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH14: {
			execute:  makePush(14, 14),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH15: {
			execute:  makePush(15, 15),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH16: {
			execute:  makePush(16, 16),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH17: {
			execute:  makePush(17, 17),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH18: {
			execute:  makePush(18, 18),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH19: {
			execute:  makePush(19, 19),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH20: {
			execute:  makePush(20, 20),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH21: {
			execute:  makePush(21, 21),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH22: {
			execute:  makePush(22, 22),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH23: {
			execute:  makePush(23, 23),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH24: {
			execute:  makePush(24, 24),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH25: {
			execute:  makePush(25, 25),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH26: {
			execute:  makePush(26, 26),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH27: {
			execute:  makePush(27, 27),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH28: {
			execute:  makePush(28, 28),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH29: {
			execute:  makePush(29, 29),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH30: {
			execute:  makePush(30, 30),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH31: {
			execute:  makePush(31, 31),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},
		PUSH32: {
			execute:  makePush(32, 32),
			minStack: minStack(0, 1),
			maxStack: maxStack(0, 1),
		},

		// 0x90
		SWAP1: {
			execute:  makeSwap(1),
			minStack: minSwapStack(2),
			maxStack: maxSwapStack(2),
		},

		// 0xb0
		INIT: {
			execute:  opInit,
			minStack: minStack(0, 0),
			maxStack: maxStack(0, 0),
		},
		MOVE: {
			execute:  opMove,
			minStack: minStack(0, 0),
			maxStack: maxStack(0, 0),
		},
	}

	return tbl
}
