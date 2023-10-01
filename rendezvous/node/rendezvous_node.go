package rendezvous

import (
	"context"
	"time"

	"github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/node"
	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/security/noise"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

const (
	DEFAULT_RENDEZVOUS_BASE_PORT = 5533
)

var logger = log.Logger("rendezvous")

type RendezvousHost struct {
	ctx   context.Context
	host  host.Host
	dht   *dht.IpfsDHT
	relay *relay.Relay
}

func NewRendezvousHost(ctx context.Context, name string, opts ...libp2p.Option) (*RendezvousHost, error) {
	rendezvous := &RendezvousHost{
		ctx: ctx,
	}

	if name != "" {
		identity := node.NewIdentityWithName(logger, name)

		privateKey, err := identity.LoadOrGeneratePrivateKey(0)
		if err != nil {
			return nil, err
		}

		// Use the keypair we generated
		opts = append(opts, libp2p.Identity(privateKey))
	}

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	identify.ActivationThresh = 1

	finalOpts := append(opts,
		[]libp2p.Option{
			// Multiple listen addresses
			libp2p.ListenAddrs([]multiaddr.Multiaddr(nil)...),
			// support noise connections
			libp2p.Security(noise.ID, noise.New),
			// Let's prevent our peer from having too many
			// connections by attaching a connection manager.
			libp2p.ConnectionManager(connmgr),
			// If you want to help other peers to figure out if they are behind
			// NATs, you can launch the server-side of AutoNAT too (AutoRelay
			// already runs the client)
			//
			// This service is highly rate-limited and should not cause any
			// performance issues.
			libp2p.EnableNATService(),
			// Let this host use the DHT to find other hosts
			libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
				dstore := dsync.MutexWrap(ds.NewMapDatastore())
				var err error
				rendezvous.dht, err = dht.New(ctx, host, dht.Datastore(dstore), dht.Mode(dht.ModeServer), dht.DisableAutoRefresh(), dht.MaxRecordAge(120*time.Second))
				return rendezvous.dht, err
			}),
		}...)

	log.SetAllLoggers(log.LevelInfo)

	rendezvous.host, err = libp2p.New(finalOpts...)
	if err != nil {
		return nil, err
	}

	limit := relay.DefaultLimit()
	resources := relay.DefaultResources()

	rendezvous.relay, err = relay.New(rendezvous.host, relay.WithLimit(limit), relay.WithResources(resources), relay.WithACL(&ACLFilter{}))
	if err != nil {
		return nil, err
	}

	return rendezvous, nil
}

func (r *RendezvousHost) Start() error {
	logger.Info("Host created. We are:", r.host.ID())
	logger.Info(r.host.Addrs())

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	logger.Info("Bootstrapping the DHT")
	if err := r.dht.Bootstrap(r.ctx); err != nil {
		return err
	}

	return nil
}

func (r *RendezvousHost) GetAddrs() (discovery.AddrList, error) {
	// print the node's PeerInfo in multiaddr format
	peerInfo := peer.AddrInfo{
		ID:    r.host.ID(),
		Addrs: r.host.Addrs(),
	}

	return peer.AddrInfoToP2pAddrs(&peerInfo)
}
