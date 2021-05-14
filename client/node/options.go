/*
Copyright © 2021-2022 Ettore Di Giacinto <mudler@mocaccino.org>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package node

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/gfleury/solo/client/crypto"
	discovery "github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/utils"
	"github.com/pkg/errors"
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

func (y YAMLConnectionConfig) copy(cfg *Config, d *discovery.DHT) {

	d.OTPKey = y.DiscoveryKey

	cfg.BroadcastKey = y.BroadcastKey
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

func FromBase64(enableDHT bool, bb string, d *discovery.DHT) func(cfg *Config) error {
	if d == nil {
		d = discovery.NewDHT()
	}
	return func(cfg *Config) error {
		if len(cfg.DiscoveryService) == 0 {
			cfg.DiscoveryService = append(cfg.DiscoveryService, d)
		}
		d.DiscoveryPeers = cfg.DiscoveryPeers
		if len(bb) == 0 {
			return nil
		}
		configDec, err := base64.StdEncoding.DecodeString(bb)
		if err != nil {
			return err
		}
		t := YAMLConnectionConfig{}

		if err := yaml.Unmarshal(configDec, &t); err != nil {
			return errors.Wrap(err, "parsing yaml")
		}
		t.copy(cfg, d)
		return nil
	}
}
