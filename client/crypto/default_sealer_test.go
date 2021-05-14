package crypto_test

import (
	"crypto/sha256"

	. "github.com/gfleury/solo/client/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gfleury/solo/client/crypto"
)

var _ = Describe("Crypto utilities", func() {
	Context("AESSealer", func() {
		It("Encode/decode", func() {
			key := sha256.Sum256([]byte(RandStringRunes(32)))
			message := []byte("foo")

			s := &DefaultSealer{}

			encoded, err := s.Seal(message, key[:])
			Expect(err).ToNot(HaveOccurred())
			Expect(encoded).ToNot(Equal(key))
			Expect(len(encoded)).To(Equal(43))

			// Encode again
			encoded2, err := s.Seal(message, key[:])
			Expect(err).ToNot(HaveOccurred())

			// should differ
			Expect(encoded2).ToNot(Equal(encoded))

			// Decrypt and check
			decoded, err := s.Unseal(encoded, key[:])
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(message))

			decoded, err = s.Unseal(encoded2, key[:])
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(message))
		})
	})
})
