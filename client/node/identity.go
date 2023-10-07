package node

import (
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mitchellh/go-homedir"
)

var (
	IDENTITY_STORE_DIR = ""
	ALLEIN_STORE_DIR   = "/.solo"
)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		IDENTITY_STORE_DIR = ALLEIN_STORE_DIR
	}
	// Ignore error here, the store dir will be on the current folder
	IDENTITY_STORE_DIR = fmt.Sprintf("%s/%s", home, ALLEIN_STORE_DIR)
}

type Identity struct {
	Name       string
	PrivateKey crypto.PrivKey
}

func NewIdentity() *Identity {
	return NewIdentityWithName("identity")
}

func NewIdentityWithName(name string) *Identity {
	i := &Identity{Name: name}
	return i
}

func (i *Identity) LoadOrGeneratePrivateKey(seed int64) (crypto.PrivKey, error) {
	// Check if we have any privkey identity cached already
	keyFile := filepath.Join(IDENTITY_STORE_DIR, i.Name)
	dat, err := os.ReadFile(keyFile)
	if err == nil && len(dat) > 0 {
		i.PrivateKey, err = crypto.UnmarshalPrivateKey(dat)
		if err != nil {
			return nil, err
		}
	} else {
		i.PrivateKey, err = genPrivKey(0)
		if err != nil {
			return nil, err
		}

		r, err := crypto.MarshalPrivateKey(i.PrivateKey)
		if err != nil {
			return nil, err
		}

		err = os.MkdirAll(IDENTITY_STORE_DIR, 0700)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(keyFile, r, 0600)
		if err != nil {
			return nil, err
		}

	}
	return i.PrivateKey, nil
}

func genPrivKey(seed int64) (crypto.PrivKey, error) {
	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 4096, r)
	return prvKey, err
}
