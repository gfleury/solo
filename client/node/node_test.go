package node_test

import (
	"context"
	"testing"
	"time"

	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/broadcast/protocol"
	"github.com/gfleury/solo/client/broadcast/prp"
	"github.com/gfleury/solo/client/config"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/node"
	"github.com/gfleury/solo/client/types"
	"github.com/gfleury/solo/rendezvous"
	"github.com/ipfs/go-log"
	"github.com/stretchr/testify/suite"
)

type NodeTestSuite struct {
	suite.Suite
	token string
	l     *logger.Logger
}

func TestNodeTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}

func (s *NodeTestSuite) SetupTest() {
	// Trigger key rotation on a low frequency to test everything works in between
	s.token = node.GenerateNewConnectionData(120).Base64()

	s.l = logger.New(log.LevelDebug)

}

func (s *NodeTestSuite) TestNodeDiscoveryBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetAllLoggers(log.LevelDebug)
	rendezvous, err := rendezvous.NewRendezvousHost(ctx, "")
	if err != nil {
		s.FailNow(err.Error())
	}
	rendezvous.Start()

	bootstrapAddrs, err := rendezvous.GetAddrs()
	if err != nil {
		s.FailNow(err.Error())
	}

	e, _ := node.NewWithConfig(config.Config{
		Token:             s.token,
		RandomIdentity:    true,
		InterfaceAddress:  "10.2.3.1/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10,
	})
	e2, _ := node.NewWithConfig(config.Config{
		Token:             s.token,
		RandomIdentity:    true,
		InterfaceAddress:  "10.2.3.2/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10,
	})

	e.Start(ctx)
	e2.Start(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				e.Broadcaster.SendPacket(ctx, metapacket.NewMetaPacket(protocol.Type_PRP, &prp.PRPacket{PRPType: prp.PRPReply, IP: "10.2.3.4", Machine: types.Machine{PeerID: e.Host().ID().String()}}))
				e2.Broadcaster.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.1")))
				e.Broadcaster.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.2")))
				time.Sleep(2 * time.Second)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, _, _ := e2.Broadcaster.Lookup("10.2.3.4")
			if m != nil {
				m, _, _ = e2.Broadcaster.Lookup("10.2.3.1")
				if m != nil {
					m, _, _ = e.Broadcaster.Lookup("10.2.3.2")
					if m != nil {
						return
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *NodeTestSuite) TestShortIntervalOTP() {
	token := node.GenerateNewConnectionData(120).Base64()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rendezvous, err := rendezvous.NewRendezvousHost(ctx, "")
	if err != nil {
		s.FailNow(err.Error())
	}
	rendezvous.Start()

	bootstrapAddrs, err := rendezvous.GetAddrs()
	if err != nil {
		s.FailNow(err.Error())
	}

	e, _ := node.NewWithConfig(config.Config{Token: token, RandomIdentity: true, InterfaceAddress: "10.2.3.1/24", DiscoveryPeers: bootstrapAddrs.StringSlice()})
	e2, _ := node.NewWithConfig(config.Config{Token: token, RandomIdentity: true, InterfaceAddress: "10.2.3.2/24", DiscoveryPeers: bootstrapAddrs.StringSlice()})

	e.Start(ctx)
	e2.Start(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				e2.Broadcaster.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.1")))
				e.Broadcaster.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.2")))
				e.Broadcaster.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.3")))
				time.Sleep(2 * time.Second)
			}
		}
	}()

out:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, _, _ := e2.Broadcaster.Lookup("10.2.3.1")
			if m != nil {
				m, _, _ = e.Broadcaster.Lookup("10.2.3.2")
				if m != nil {
					break out
				}
			}

			time.Sleep(2 * time.Second)
		}
	}

	e3, _ := node.NewWithConfig(config.Config{Token: token, RandomIdentity: true, InterfaceAddress: "10.2.3.3/24", DiscoveryPeers: bootstrapAddrs.StringSlice()})

	e3.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, _, _ := e3.Broadcaster.Lookup("10.2.3.1")
			if m != nil {
				m, _, _ = e3.Broadcaster.Lookup("10.2.3.2")
				if m != nil {
					m, _, _ = e.Broadcaster.Lookup("10.2.3.3")
					if m != nil {
						return
					}
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}
