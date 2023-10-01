package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/gfleury/solo/client/crypto"
	"github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/logger"
	rendezvous "github.com/gfleury/solo/rendezvous/node"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/suite"
)

type DHTTestSuite struct {
	suite.Suite
}

func TestDHTTestSuite(t *testing.T) {
	suite.Run(t, new(DHTTestSuite))
}

func (s *DHTTestSuite) SetupTest() {
}

func (s *DHTTestSuite) TestDHT() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rendezvous, err := rendezvous.NewRendezvousHost(ctx, logger.New(log.LevelDebug), "")
	if err != nil {
		s.FailNow(err.Error())
	}
	rendezvous.Start()

	bootstrappers, err := rendezvous.GetAddrs()
	if err != nil {
		s.FailNow(err.Error())
	}

	key := crypto.OTPKey{
		Key:       "sharedKey",
		KeyLength: 16,
		Interval:  20,
	}

	dht := discovery.NewDHT()
	dht.DiscoveryPeers = bootstrappers
	dht.OTPKey = key
	dht.DiscoveryInterval = 10 * time.Second

	dht2 := discovery.NewDHT()
	dht2.DiscoveryPeers = bootstrappers
	dht2.OTPKey = key
	dht2.DiscoveryInterval = 10 * time.Second

	h, _ := libp2p.New()
	h2, _ := libp2p.New()

	log := logger.New(log.LevelDebug)
	dht.Run(log, ctx, h)
	dht2.Run(log, ctx, h2)

	startTime := time.Now()

	for len(h.Network().Peers()) < 2 && time.Since(startTime) < 10*time.Second {
	}

	for h.Network().Connectedness(h2.ID()) != network.Connected && time.Since(startTime) < 20*time.Second {
	}

	s.Equal(h.Network().Connectedness(h2.ID()), network.Connected)
}

func (s *DHTTestSuite) TestDHT600Seconds() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rendezvous, err := rendezvous.NewRendezvousHost(ctx, logger.New(log.LevelDebug), "")
	if err != nil {
		s.FailNow(err.Error())
	}
	rendezvous.Start()

	bootstrappers, err := rendezvous.GetAddrs()
	if err != nil {
		s.FailNow(err.Error())
	}

	key := crypto.OTPKey{
		Key:       "sharedKey",
		KeyLength: 16,
		Interval:  20,
	}

	dht := discovery.NewDHT()
	dht.DiscoveryPeers = bootstrappers
	dht.OTPKey = key
	dht.DiscoveryInterval = 600 * time.Second

	dht2 := discovery.NewDHT()
	dht2.DiscoveryPeers = bootstrappers
	dht2.OTPKey = key
	dht2.DiscoveryInterval = 600 * time.Second

	h, _ := libp2p.New()
	h2, _ := libp2p.New()

	log := logger.New(log.LevelDebug)
	dht.Run(log, ctx, h)
	dht2.Run(log, ctx, h2)

	startTime := time.Now()

	for len(h.Network().Peers()) < 2 && time.Since(startTime) < 10*time.Second {
	}

	for h.Network().Connectedness(h2.ID()) != network.Connected && time.Since(startTime) < 20*time.Second {
	}

	s.Equal(h.Network().Connectedness(h2.ID()), network.Connected)
}
