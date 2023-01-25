package vm

type ScopeContext struct {
	Stack    *Stack
	Contract *Contract
}

type EVMInterpreter struct {
	evm *EVM

	JumpTable *JumpTable

	readOnly bool
}

func NewEVMInterpreter(evm *EVM) *EVMInterpreter {
	in := &EVMInterpreter{
		evm:       evm,
		JumpTable: &frontierInstructionSet,
	}

	return in
}

func (in *EVMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {

	if readOnly && !in.readOnly {
		in.readOnly = true
		defer func() { in.readOnly = false }()
	}

	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		op          OpCode
		stack       = newstack()
		callContext = &ScopeContext{
			Stack:    stack,
			Contract: contract,
		}

		pc = uint64(0)

		res []byte
	)

	defer func() {
		returnStack(stack)
	}()

	for {
		op = contract.GetOp(pc)
		operation := in.JumpTable[op]

		if sLen := stack.len(); sLen < operation.minStack {
			return nil, &ErrStackUnderflow{stackLen: sLen, required: operation.minStack}
		} else if sLen > operation.maxStack {
			return nil, &ErrStackOverflow{stackLen: sLen, limit: operation.maxStack}
		}

		res, err = operation.execute(&pc, in, callContext)
		if err != nil {
			break
		}
		pc++
	}

	if err == errStopToken {
		err = nil
	}

	return res, err
}
