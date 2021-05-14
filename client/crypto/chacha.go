package crypto

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

func Seal(plaintext []byte, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func Unseal(ciphertext []byte, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aead.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce := ciphertext[:aead.NonceSize()]
	ciphertestWithoutNonce := ciphertext[aead.NonceSize():]

	return aead.Open(nil, nonce, ciphertestWithoutNonce, nil)
}
