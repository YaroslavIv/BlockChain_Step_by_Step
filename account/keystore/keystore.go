package keystore

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"path/filepath"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
)

const KeyStoreScheme = "keystore"

type KeyStore struct {
	storage keyStore

	mu sync.RWMutex
}

func NewPlaintextKeyStore(keydir string) *KeyStore {
	keydir, _ = filepath.Abs(keydir)
	ks := &KeyStore{storage: &keyStorePlain{keydir}}
	ks.init(keydir)
	return ks
}

func (ks *KeyStore) init(keydir string) {}

func (ks *KeyStore) NewAccount(passphrase string) (*Key, error) {
	key, _, err := storeNewKey(ks.storage, crand.Reader, passphrase)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (ks *KeyStore) Find(addr common.Address) (accounts.Account, error) {
	a := accounts.Account{
		Address: addr,
		URL:     accounts.URL{Scheme: KeyStoreScheme, Path: ks.storage.JoinPath(keyFileName(addr))},
	}

	return a, nil
}

func (ks *KeyStore) Load(addr common.Address, passphrase string) (accounts.Account, *Key, error) {
	a, err := ks.Find(addr)
	if err != nil {
		panic(err)
	}

	key, err := ks.storage.GetKey(a.Address, a.URL.Path, passphrase)
	return a, key, err
}

func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}
