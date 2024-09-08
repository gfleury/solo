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
	"context"
	"encoding/base64"

	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/gfleury/solo/client/broadcast"
	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/crypto"
	discovery "github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/common/models"
)

// Config is the node configuration
type Config struct {
	BroadcastKey crypto.OTPKey

	// ListenAddresses is the discovery peer initial bootstrap addresses
	ListenAddresses []discovery.AddrList

	// Insecure disables secure p2p e2e encrypted communication
	RandomIdentity bool
	RandomPort     bool

	DiscoveryService []DiscoveryService
	NetworkServices  []NetworkService
	Logger           log.StandardLogger

	InterfaceAddress   string
	InterfaceMTU       int
	PublishLocalRoutes bool

	AdditionalOptions, Options []libp2p.Option

	DiscoveryPeers       discovery.AddrList
	PublicDiscoveryPeers bool
	StandaloneMode       bool

	ConnectionConfigToken string
	Sealer                crypto.Sealer
}

type StreamHandler func(*Node) func(stream network.Stream)

type Handler func(*metapacket.MetaPacket, chan *metapacket.MetaPacket) error

type DiscoveryService interface {
	Run(log.StandardLogger, context.Context, host.Host) error
	Option(context.Context) func(c *libp2p.Config) error
}

type NetworkService interface {
	Run(context.Context, log.StandardLogger, host.Host, broadcast.Broadcaster) error
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
		t := models.YAMLConnectionConfig{}

		if err := yaml.Unmarshal(configDec, &t); err != nil {
			return errors.Wrap(err, "parsing yaml")
		}

		// Old copy()
		d.OTPKey = t.DiscoveryKey
		cfg.BroadcastKey = t.BroadcastKey

		return nil
	}
}
