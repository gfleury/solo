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
package discovery

import (
	"context"
	"crypto/sha256"
	"sync"
	"time"

	"github.com/gfleury/solo/client/crypto"
	"github.com/gfleury/solo/client/utils"

	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	discovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

const (
	DHT_FOUND = "DHT_FOUND"
)

type DHT struct {
	*dht.IpfsDHT

	OTPKeyReceiver    chan crypto.OTPKey
	OTPKey            crypto.OTPKey
	Rendezvous        string
	latestRendezvous  string
	DiscoveryPeers    AddrList
	DiscoveryInterval time.Duration
	dhtOptions        []dht.Option
}

func NewDHT(d ...dht.Option) *DHT {
	return &DHT{dhtOptions: d, OTPKeyReceiver: make(chan crypto.OTPKey)}
}

func (d *DHT) Option(ctx context.Context) func(c *libp2p.Config) error {
	return libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		// make the DHT with the given Host
		return d.startDHT(ctx, h)
	})
}
func (d *DHT) GetNextRendezvous() string {
	totp := d.OTPKey.TOTP(sha256.New)

	rv := crypto.MD5(totp)
	d.latestRendezvous = rv
	return rv
}

func (d *DHT) startDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	if d.IpfsDHT == nil {
		// Start a DHT, for use in peer discovery. We can't just make a new DHT
		// client because we want each peer to maintain its own local copy of the
		// DHT, so that the bootstrapping node of the DHT can go down without
		// inhibiting future peer discovery.

		kad, err := dht.New(ctx, h, d.dhtOptions...)
		if err != nil {
			return nil, err
		}
		d.IpfsDHT = kad
	}

	return d.IpfsDHT, nil
}

func (d *DHT) Run(c log.StandardLogger, ctx context.Context, host host.Host) error {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	_, err := d.startDHT(ctx, host)
	if err != nil {
		return err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	c.Info("Bootstrapping DHT")
	if err = d.IpfsDHT.Bootstrap(ctx); err != nil {
		return err
	}

	connect := func() {
		d.bootstrapPeers(c, ctx, host)
		rv := d.GetNextRendezvous()
		c.Debugf("Announcing with key: %s", rv)
		d.announceAndConnect(c, ctx, host, rv)
	}

	go func() {
		// Bootstrap DHT peers, so we can have connectivity with it
		d.bootstrapPeers(c, ctx, host)

		// Wait to receive the OTPKey from ConfigurationDiscovery
		d.OTPKey = <-d.OTPKeyReceiver

		t := utils.NewBackoffTicker(utils.BackoffMaxInterval(d.DiscoveryInterval))
		defer t.Stop()
		for {
			select {
			case <-t.C:
				connect()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (d *DHT) bootstrapPeers(c log.StandardLogger, ctx context.Context, host host.Host) {
	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range d.DiscoveryPeers {
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			panic(err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if host.Network().Connectedness(peerinfo.ID) != network.Connected {
				if err := host.Connect(ctx, *peerinfo); err != nil {
					c.Debug(err.Error())
				} else {
					c.Debug("Connection established with bootstrap node:", *peerinfo)
				}
			}
		}()
	}
	wg.Wait()
}

func (d *DHT) FindClosePeers(logger log.StandardLogger) func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
	return func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
		peerChan := make(chan peer.AddrInfo, numPeers)
		toStream := []peer.AddrInfo{}

		go func() {
			closestPeers, err := d.GetClosestPeers(ctx, d.PeerID().String())
			if err != nil {
				logger.Error(err)
				close(peerChan)
				return
			}

			for _, p := range closestPeers {
				addrs := d.Host().Peerstore().Addrs(p)
				if len(addrs) == 0 {
					continue
				}
				logger.Debugf("[relay discovery] Found close peer '%s'", p.String())
				toStream = append(toStream, peer.AddrInfo{ID: p, Addrs: addrs})
			}

			if len(toStream) > numPeers {
				toStream = toStream[0 : numPeers-1]
			}

			for _, t := range toStream {
				peerChan <- t
			}

			close(peerChan)
		}()

		return peerChan
	}
}

func (d *DHT) announceAndConnect(l log.StandardLogger, ctx context.Context, host host.Host, rv string) error {
	l.Debugf("Announcing ourselves with addresses: %v", host.Addrs())
	routingDiscovery := discovery.NewRoutingDiscovery(d.IpfsDHT)
	routingDiscovery.Advertise(ctx, rv)

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	l.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, rv)
	if err != nil {
		return err
	}

	for p := range peerChan {
		// Don't dial ourselves or peers without address
		if p.ID == host.ID() || len(p.Addrs) == 0 {
			continue
		}

		TagPeerAsFound(host, p.ID)

		if host.Network().Connectedness(p.ID) != network.Connected {
			l.Debugf("Found peer %s with %d addresses", p.ID, len(p.Addrs))
			timeoutCtx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
			if err := host.Connect(timeoutCtx, p); err != nil {
				l.Debugf("Failed connecting to %s with addresses: %s", p.ID, p.Addrs)
			} else {
				l.Debugf("Connected to %s with %d addresses", p.ID, len(p.Addrs))
			}
			cancelFunc()
		} else {
			l.Debugf("Known peer (already connected): %s with %d addresses", p.ID, len(p.Addrs))
		}
	}

	return nil
}

func TagPeerAsFound(myself host.Host, peerIDFound peer.ID) {
	myself.ConnManager().UpsertTag(peerIDFound, DHT_FOUND, func(i int) int { return 0 })
}
