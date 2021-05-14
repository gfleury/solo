package crypto

type Sealer interface {
	Seal([]byte, []byte) ([]byte, error)
	Unseal([]byte, []byte) ([]byte, error)
}

// Uses chacha20poly1305
type DefaultSealer struct{}

func (*DefaultSealer) Seal(message []byte, key []byte) ([]byte, error) {
	return Seal(message, key)
}

func (*DefaultSealer) Unseal(message []byte, key []byte) ([]byte, error) {
	return Unseal(message, key)
}
