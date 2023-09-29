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
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	conngater "github.com/libp2p/go-libp2p/p2p/net/conngater"
	multiaddr "github.com/multiformats/go-multiaddr"
)

const (
	DEFAULT_BASE_PORT = 5544
)

func ListenAddrs(randomPort bool, ports ...int) func(cfg *libp2p.Config) error {
	var port, webTransportPort int
	if len(ports) == 0 {
		port = DEFAULT_BASE_PORT
		webTransportPort = DEFAULT_BASE_PORT + 1
	} else {
		port = ports[0]
		if len(ports) > 1 {
			webTransportPort = ports[1]
		} else {
			webTransportPort = port + 1
		}
	}
	if randomPort {
		port = 0
		webTransportPort = 0
	}
	return func(cfg *libp2p.Config) error {
		addrs := []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", webTransportPort),
			fmt.Sprintf("/ip6/::/tcp/%d", port),
			fmt.Sprintf("/ip6/::/udp/%d/quic-v1", port),
			fmt.Sprintf("/ip6/::/udp/%d/quic-v1/webtransport", webTransportPort),
		}
		listenAddrs := make([]multiaddr.Multiaddr, 0, len(addrs))
		for _, s := range addrs {
			addr, err := multiaddr.NewMultiaddr(s)
			if err != nil {
				return err
			}
			listenAddrs = append(listenAddrs, addr)
		}
		return cfg.Apply(libp2p.ListenAddrs(listenAddrs...))
	}
}

// Host returns the libp2p peer host
func (e *Node) Host() host.Host {
	return e.host
}

// ConnectionGater returns the underlying libp2p conngater
func (e *Node) ConnectionGater() *conngater.BasicConnectionGater {
	return e.cg
}

// BlockSubnet blocks the CIDR subnet from connections
func (e *Node) BlockSubnet(cidr string) error {
	// Avoid to loopback traffic by trying to connect to nodes in via VPN
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	return e.ConnectionGater().BlockSubnet(n)
}

func (e *Node) genHost(ctx context.Context) (host.Host, error) {
	opts := e.config.Options

	cg, err := conngater.NewBasicConnectionGater(nil)
	if err != nil {
		return nil, err
	}

	e.cg = cg

	if e.config.InterfaceAddress != "" {
		e.BlockSubnet(e.config.InterfaceAddress)
	}

	e.BlockSubnet("127.0.0.0/8")

	if !e.config.RandomIdentity {
		// generate Identity privkey if its not already persisted
		identity := NewIdentity(e.config.Logger)

		privateKey, err := identity.LoadOrGeneratePrivateKey(0)
		if err != nil {
			return nil, err
		}

		opts = append(opts, libp2p.ConnectionGater(cg), libp2p.Identity(privateKey))
	}

	if len(e.config.ListenAddresses) > 0 {
		addrs := []multiaddr.Multiaddr{}
		for _, l := range e.config.ListenAddresses {
			addrs = append(addrs, []multiaddr.Multiaddr(l)...)
		}
		opts = append(opts, libp2p.ListenAddrs(addrs...))
	} else {
		opts = append(opts, ListenAddrs(e.config.RandomPort))
	}

	for _, d := range e.config.DiscoveryService {
		opts = append(opts, d.Option(ctx))
	}

	opts = append(opts, e.config.AdditionalOptions...)

	opts = append(opts, libp2p.FallbackDefaults)

	return libp2p.New(opts...)
}
