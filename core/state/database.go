package state

type Trie interface {
	TryGet(key []byte) ([]byte, error)

	TryUpdate(key, val []byte) error
}
