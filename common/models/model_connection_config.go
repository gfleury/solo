package models

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/gfleury/solo/client/crypto"
	"github.com/gfleury/solo/client/utils"
	"gopkg.in/yaml.v2"
)

type YAMLConnectionConfig struct {
	VPNPreSharedKey string
	BroadcastKey    crypto.OTPKey
	DiscoveryKey    crypto.OTPKey
}

// Read from Base64 Token
func YAMLConnectionConfigFromToken(s string) (*YAMLConnectionConfig, error) {
	y := &YAMLConnectionConfig{}
	bytesData, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytesData, y)
	return y, err
}

// Base64 returns the base64 string representation of the connection
func (y YAMLConnectionConfig) Base64() string {
	bytesData, _ := yaml.Marshal(y)
	return base64.StdEncoding.EncodeToString(bytesData)
}

// YAML returns the connection config as yaml string
func (y YAMLConnectionConfig) YAML() string {
	bytesData, _ := yaml.Marshal(y)
	return string(bytesData)
}

func GenerateNewConnectionData(i ...int) *YAMLConnectionConfig {
	defaultInterval := 600
	keyLength := sha256.Size

	if len(i) >= 3 {
		keyLength = i[2]
		defaultInterval = i[0]
	} else if len(i) >= 2 {
		defaultInterval = i[0]
	} else if len(i) == 1 {
		defaultInterval = i[0]
	}

	return &YAMLConnectionConfig{
		VPNPreSharedKey: utils.RandStringRunes(sha256.Size),
		BroadcastKey: crypto.OTPKey{
			Key:       utils.RandStringRunes(keyLength),
			Interval:  defaultInterval,
			KeyLength: keyLength,
		},
		DiscoveryKey: crypto.OTPKey{
			Key:       utils.RandStringRunes(keyLength),
			Interval:  defaultInterval,
			KeyLength: keyLength,
		},
	}
}
