package crypto

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OTPTestSuite struct {
	suite.Suite
}

func TestOTPTestSuite(t *testing.T) {
	suite.Run(t, new(OTPTestSuite))
}

func (s *OTPTestSuite) SetupTest() {
}

func (s *OTPTestSuite) TestOTPGeneration() {
	key0 := &OTPKey{
		Key:       "0key1234",
		KeyLength: 16,
		Interval:  2,
	}

	key0_value := key0.TOTP(sha256.New)

	s.Equal(key0_value, key0.TOTP(sha256.New))
	time.Sleep(2 * time.Second)

	s.NotEqual(key0_value, key0.TOTP(sha256.New))

	key0.Interval = 5
	now := time.Now()
	last := time.Now()
	for key0.TOTP(sha256.New) == key0.TOTP(sha256.New) {
		last = time.Now()
	}

	s.WithinDuration(now, last, 30*time.Second)
}
