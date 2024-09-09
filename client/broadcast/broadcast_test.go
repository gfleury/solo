package broadcast_test

import (
	"context"
	"testing"
	"time"

	"github.com/gfleury/solo/client/broadcast"
	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/broadcast/protocol"
	"github.com/gfleury/solo/client/broadcast/prp"
	"github.com/gfleury/solo/client/crypto"
	"github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/vpn"
	"github.com/gfleury/solo/common/models"
	"github.com/ipfs/go-log"
	"github.com/stretchr/testify/suite"
)

type BroadcastTestSuite struct {
	suite.Suite
	otpInterval int
}

func TestBroadcastTestSuite(t *testing.T) {
	suite.Run(t, new(BroadcastTestSuite))
}

func (s *BroadcastTestSuite) SetupTest() {
	s.otpInterval = 120
}

func (s *BroadcastTestSuite) TestBroadcastPubSubShortIntervalOTP() {
	s.otpInterval = 10
	s.TestBroadcastPubSubOTP()
}

func (s *BroadcastTestSuite) TestBroadcastPubSubOTP() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	h1, _ := vpn.NewTestHost("0")
	h2, _ := vpn.NewTestHost("0")

	err := vpn.TestConnectHosts(ctx, h1, h2)
	s.NoError(err)

	logger := logger.New(log.LevelDebug)
	otpKey := crypto.OTPKey{
		Key:       "otpKey",
		KeyLength: 16,
		Interval:  s.otpInterval,
	}

	b1 := broadcast.NewBroadcaster(logger, &otpKey, 1024)
	b2 := broadcast.NewBroadcaster(logger, &otpKey, 1024)

	go func() {
		b1.Start(ctx, h1, "10.2.3.1")
	}()

	go func() {
		b2.Start(ctx, h2, "10.2.3.2")
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				b1.SendPacket(ctx, metapacket.NewMetaPacket(protocol.Type_PRP, &prp.PRPacket{PRPType: prp.PRPReply, IP: "10.2.3.4", Machine: models.NetworkNode{PeerID: h1.ID().String()}}))
				b2.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.1")))
				b1.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.2")))
				time.Sleep(2 * time.Second)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, _, _ := b2.Lookup("10.2.3.4")
			if m != nil {
				m, _, _ = b2.Lookup("10.2.3.1")
				if m != nil {
					m, _, _ = b1.Lookup("10.2.3.2")
					if m != nil {
						return
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *BroadcastTestSuite) TestBroadcastStreamSeal() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	h1, _ := vpn.NewTestHost("0")
	h2, _ := vpn.NewTestHost("0")

	err := vpn.TestConnectHosts(ctx, h1, h2)
	s.NoError(err)

	logger := logger.New(log.LevelDebug)

	otpKey := crypto.OTPKey{
		Key:       "supersecret",
		KeyLength: 32,
		Interval:  120,
	}

	b1 := broadcast.NewStreamBroadcaster(logger, discovery.AddrList{}, otpKey, false)
	b2 := broadcast.NewStreamBroadcaster(logger, discovery.AddrList{}, otpKey, false)

	go func() {
		b1.Start(ctx, h1, "10.2.3.1")
	}()

	go func() {
		b2.Start(ctx, h2, "10.2.3.2")
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				b1.SendPacket(ctx, metapacket.NewMetaPacket(protocol.Type_PRP, &prp.PRPacket{PRPType: prp.PRPReply, IP: "10.2.3.4", Machine: models.NetworkNode{PeerID: h1.ID().String()}}))
				b2.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.1")))
				b1.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket("10.2.3.2")))
				time.Sleep(2 * time.Second)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, _, _ := b2.Lookup("10.2.3.4")
			if m != nil {
				m, _, _ = b2.Lookup("10.2.3.1")
				if m != nil {
					m, _, _ = b1.Lookup("10.2.3.2")
					if m != nil {
						return
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}
