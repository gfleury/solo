package node_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/broadcast/protocol"
	"github.com/gfleury/solo/client/broadcast/prp"
	"github.com/gfleury/solo/client/config"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/node"
	"github.com/gfleury/solo/cmd"
	"github.com/gfleury/solo/common/models"
	rendezvous "github.com/gfleury/solo/server/core-api/rendezvous"
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
	s.token = models.GenerateNewConnectionData(120).Base64()

	s.l = logger.New(log.LevelDebug)

}

func (s *NodeTestSuite) TestNodeDiscoveryBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetAllLoggers(log.LevelDebug)
	rendezvous, err := rendezvous.NewRendezvousHost(ctx, logger.New(log.LevelDebug), "")
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
		RandomPort:        true,
		StandaloneMode:    true,
		InterfaceAddress:  "10.2.3.1/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10,
	})
	e2, _ := node.NewWithConfig(config.Config{
		Token:             s.token,
		RandomIdentity:    true,
		RandomPort:        true,
		StandaloneMode:    true,
		InterfaceAddress:  "10.2.3.2/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10,
	})

	go e.Start(ctx)
	go e2.Start(ctx)
	time.Sleep(3 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				e.Broadcaster.SendPacket(ctx, metapacket.NewMetaPacket(protocol.Type_PRP, &prp.PRPacket{PRPType: prp.PRPReply, IP: "10.2.3.4", Machine: models.NetworkNode{OwnPeerIdentification: e.Host().ID().String()}}))
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
	token := models.GenerateNewConnectionData(120).Base64()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rendezvous, err := rendezvous.NewRendezvousHost(ctx, logger.New(log.LevelDebug), "")
	if err != nil {
		s.FailNow(err.Error())
	}
	rendezvous.Start()

	bootstrapAddrs, err := rendezvous.GetAddrs()
	if err != nil {
		s.FailNow(err.Error())
	}

	e, _ := node.NewWithConfig(config.Config{
		Token:             token,
		RandomIdentity:    true,
		RandomPort:        true,
		StandaloneMode:    true,
		InterfaceAddress:  "10.2.3.1/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10})
	e2, _ := node.NewWithConfig(config.Config{
		Token:             token,
		RandomIdentity:    true,
		RandomPort:        true,
		StandaloneMode:    true,
		InterfaceAddress:  "10.2.3.2/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10})

	go e.Start(ctx)
	go e2.Start(ctx)
	time.Sleep(3 * time.Second)

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

	e3, _ := node.NewWithConfig(config.Config{
		Token:             token,
		RandomIdentity:    true,
		RandomPort:        true,
		StandaloneMode:    true,
		InterfaceAddress:  "10.2.3.3/24",
		DiscoveryPeers:    bootstrapAddrs.StringSlice(),
		LogLevel:          "debug",
		DiscoveryInterval: 10})

	go e3.Start(ctx)
	time.Sleep(3 * time.Second)

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

func XXXTestTryConnectingRemote(t *testing.T) {
	config := config.Config{
		Token:                "dnBucHJlc2hhcmVka2V5OiA0cDFua1QwdTczcUhmaTRGa3FVQWM5bkxwUE9HWGJaQgpicm9hZGNhc3RrZXk6CiAga2V5OiBIZ0Z3ak5tMjJHYm4yc05zczJQdTd5Q21TbHp3MjhKYwogIGtleWxlbmd0aDogMzIKICBpbnRlcnZhbDogOTAwMApkaXNjb3ZlcnlrZXk6CiAga2V5OiA3c01NQkt3UElYNjB1Zlg2cTQ0b29TakNWVlFFdVYwWAogIGtleWxlbmd0aDogMzIKICBpbnRlcnZhbDogOTAwMAo=",
		InterfaceAddress:     "10.1.0.1/24",
		InterfaceName:        "",
		CreateInterface:      false,
		Libp2pLogLevel:       "info",
		LogLevel:             "debug",
		DiscoveryPeers:       cmd.DEFAULT_DISCOVERY_PEERS,
		PublicDiscoveryPeers: false,
		DiscoveryInterval:    60,
		InterfaceMTU:         1500,
		MaxConnections:       1500,
		HolePunch:            true,
		RandomIdentity:       false,
		RandomPort:           false,
	}

	e, err := node.NewWithConfig(config)
	if err != nil {
		fmt.Printf("failed to create new node: %s\n", err)
		return
	}

	ctx := context.Background()

	go func() {
		for {
			time.Sleep(20 * time.Second)

			fmt.Println(e.Host().Addrs())

			for _, peerID := range e.Host().Network().Peers() {
				addrInfo := e.Host().Network().Peerstore().PeerInfo(peerID)
				e.Host().Connect(ctx, addrInfo)
			}
		}
	}()

	_ = e.Start(ctx)
	if err != nil {
		fmt.Printf("failed to start node: %s\n", err)
		return
	}

}
