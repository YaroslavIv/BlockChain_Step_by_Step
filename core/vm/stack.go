package vm

import (
	"sync"

	"github.com/holiman/uint256"
)

var stackPool = sync.Pool{
	New: func() interface{} {
		return &Stack{data: make([]uint256.Int, 0, 16)}
	},
}

type Stack struct {
	data []uint256.Int
}

func newstack() *Stack {
	return stackPool.Get().(*Stack)
}

func returnStack(s *Stack) {
	s.data = s.data[:0]
	stackPool.Put(s)
}

func (st *Stack) len() int {
	return len(st.data)
}

func (st *Stack) pop() (ret uint256.Int) {
	ret = st.data[st.len()-1]
	st.data = st.data[:st.len()-1]
	return
}

func (st *Stack) peek() *uint256.Int {
	return &st.data[st.len()-1]
}

func (st *Stack) push(d *uint256.Int) {
	st.data = append(st.data, *d)
}

func (st *Stack) swap(n int) {
	st.data[st.len()-n], st.data[st.len()-1] = st.data[st.len()-1], st.data[st.len()-n]
}
