package rendezvous

import (
	"context"
	"fmt"
	"time"

	"github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/node"
	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
)

const (
	DEFAULT_RENDEZVOUS_BASE_PORT = 5533
)

var logger = log.Logger("rendezvous")

type RendezvousHost struct {
	ctx  context.Context
	host host.Host
	dht  *dht.IpfsDHT
}

var ListenAddrs = func(cfg *libp2p.Config) error {
	addrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", DEFAULT_RENDEZVOUS_BASE_PORT),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", DEFAULT_RENDEZVOUS_BASE_PORT),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", DEFAULT_RENDEZVOUS_BASE_PORT+1),
		fmt.Sprintf("/ip6/::/tcp/%d", DEFAULT_RENDEZVOUS_BASE_PORT),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1", DEFAULT_RENDEZVOUS_BASE_PORT),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1/webtransport", DEFAULT_RENDEZVOUS_BASE_PORT+1),
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

	finalOpts := append(opts,
		[]libp2p.Option{
			// Multiple listen addresses
			libp2p.ListenAddrs([]multiaddr.Multiaddr(nil)...),
			// support noise connections
			libp2p.Security(noise.ID, noise.New),
			// support any other default transports (TCP)
			ListenAddrs,
			// Let's prevent our peer from having too many
			// connections by attaching a connection manager.
			libp2p.ConnectionManager(connmgr),
			// Attempt to open ports using uPNP for NATed hosts.
			libp2p.NATPortMap(),
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
				rendezvous.dht, err = dht.New(ctx, host, dht.Datastore(dstore), dht.Mode(dht.ModeAutoServer))
				return rendezvous.dht, err
			}),
			libp2p.EnableHolePunching(),
			libp2p.EnableRelay(),
			libp2p.EnableRelayService(),
		}...)

	rendezvous.host, err = libp2p.New(finalOpts...)
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
	peerInfo := peerstore.AddrInfo{
		ID:    r.host.ID(),
		Addrs: r.host.Addrs(),
	}

	return peerstore.AddrInfoToP2pAddrs(&peerInfo)
}
