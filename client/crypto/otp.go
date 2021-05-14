package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"hash"

	"github.com/creachadair/otp"
)

type OTPKey struct {
	Key                 string
	KeyLength, Interval int
}

func (o *OTPKey) TOTP(f func() hash.Hash) string {
	cfg := otp.Config{
		Hash:     f,           // default is sha1.New
		Digits:   o.KeyLength, // default is 6
		TimeStep: otp.TimeWindow(o.Interval),
		Key:      o.Key,
		Format: func(hash []byte, nb int) string {
			return base64.StdEncoding.EncodeToString(hash)[:nb]
		},
	}
	return cfg.TOTP()
}

func (o *OTPKey) TOTPSHA256(f func() hash.Hash) []byte {
	k := sha256.Sum256([]byte(o.TOTP(f)))
	return k[:]
}
