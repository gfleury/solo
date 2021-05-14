package vpn

import (
	"context"
	"fmt"
	"io"
	"time"

	gonoise "github.com/flynn/noise"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/gfleury/solo/client/broadcast"
	"github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/protocol"

	libp2p_protocol "github.com/libp2p/go-libp2p/core/protocol"

	"github.com/pkg/errors"
)

type VPNService struct {
	logger log.StandardLogger

	// VPN Interface
	vpnInterface *VPNInterface
	Config       InterfaceConfig
	broadcast    broadcast.Broadcaster

	// Frame processing timeout
	timeout time.Duration
}

type VPNHost interface {
	NewStream(ctx context.Context, p peer.ID, pids ...libp2p_protocol.ID) (network.Stream, error)
	ID() peer.ID
	PrivateKey() *gonoise.DHKey
	PeerPublicKey(peer.ID) []byte
}

func VPNNetworkService(config InterfaceConfig) *VPNService {
	vpnService := &VPNService{
		timeout: 5 * time.Second,
		logger:  logger.New(log.LevelDebug),
		Config:  config,
	}

	return vpnService
}

type WrapperHost struct {
	host host.Host
}

func (w *WrapperHost) NewStream(ctx context.Context, p peer.ID, pids ...libp2p_protocol.ID) (network.Stream, error) {
	return w.host.NewStream(ctx, p, pids...)
}

func (w *WrapperHost) ID() peer.ID {
	return w.host.ID()
}

func (w *WrapperHost) PrivateKey() *gonoise.DHKey {
	privKey := w.host.Peerstore().PrivKey(w.host.ID())
	pubKey := w.host.Peerstore().PubKey(w.host.ID())
	if privKey != nil && pubKey != nil {
		privBytes, err := privKey.Raw()
		pubBytes, err2 := pubKey.Raw()
		if err != nil || err2 != nil {
			panic(fmt.Sprint(err, err2))
		}
		return &gonoise.DHKey{
			Private: privBytes,
			Public:  pubBytes,
		}
	}
	panic("host do not have private key")
}

func (w *WrapperHost) PeerPublicKey(p peer.ID) []byte {
	pubKey := w.host.Peerstore().PubKey(p)
	if pubKey != nil {
		pubBytes, err := pubKey.Raw()
		if err != nil {
			return nil
		}
		return pubBytes
	}
	return nil
}

func NewWrapperHost(h host.Host) VPNHost {
	return &WrapperHost{host: h}
}

func (v *VPNService) Run(ctx context.Context, logger log.StandardLogger, host host.Host, broadcast broadcast.Broadcaster) error {
	var err error

	v.logger = logger
	v.broadcast = broadcast

	// Create and configure Network Interface used on the VPN Service
	if v.vpnInterface == nil {
		v.vpnInterface, err = newInterface(&v.Config, NewWrapperHost(host))
		if err != nil {
			return err
		}
	}

	// Set the VPN P2P stream handler (for incoming VPNPacket streams)
	host.SetStreamHandler(protocol.ALLEIN.ID(), v.dataStreamHandler())

	if v.Config.CreateInterface {
		if err := v.vpnInterface.prepareInterface(); err != nil {
			return err
		}
	}

	// Announce ourselves on the network
	v.broadcast.AnnounceMyself(ctx)

	// read packets from the network interface
	go v.readPackets(ctx)

	return nil
}

func (v *VPNService) dataStreamHandler() func(stream network.Stream) {
	return func(stream network.Stream) {
		// TODO: Verify Inbound Frames
		dstID := stream.Conn().RemotePeer()
		streamKey := v.vpnInterface.getInboundStreamKey(dstID)

		v.logger.Debugf("New data stream inbound from: %s", streamKey)
		v.vpnInterface.streamMap.New(streamKey, stream)
		n, err := io.Copy(v.vpnInterface, stream)
		if err != nil {
			v.logger.Errorf("Failed to copy all data (copied only: %d) into network interface: %s", n, err)
			stream.Reset()
		}
		v.logger.Debugf("Finish and remove noiseStream handler: %s", dstID)

		// Stream ist tot
		v.vpnInterface.streamMap.Delete(streamKey)
	}
}

// Handles OUTGOING packets on the VPNService
// Packets are written to a libp2p stream
func (v *VPNService) handlePacket(packet Packet) error {
	ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
	defer cancel()

	if len(packet) < 1 {
		return fmt.Errorf("packet size is less than 1")
	}

	dstIp, err := packet.DstIp()
	if err != nil {
		return err
	}

	dst := dstIp.String()

	notFoundErr := fmt.Errorf("'%s' not found in the routing table", dst)

	// Query the routing table
	dstNode, found, wasLookupNotLongAgo := v.broadcast.Lookup(dst)
	if !found {
		if !wasLookupNotLongAgo {
			// Send a PRPRequest to all nodes
			err = v.broadcast.PRPRequest(ctx, dst)
			if err != nil {
				return err
			}
		}
		return notFoundErr
	}

	dstID, err := peer.Decode(dstNode.PeerID)
	if err != nil {
		return errors.Wrap(err, "could not decode peer")
	}

	return v.vpnInterface.handlePacket(ctx, dstID, packet)
}

func (v *VPNService) readPackets(ctx context.Context) {
	defer func() {
		// VPNService clean-up go-routine
		v.vpnInterface.networkInterface.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			packet, n, err := v.vpnInterface.ReadPacket()
			if err != nil {
				continue
			}

			if err := v.handlePacket(packet[:n]); err != nil {
				v.logger.Errorf("Handle packet error: %s", err)
				continue
			}

		}
	}
}

func NewTestHost(port string, opts ...libp2p.Option) (host.Host, error) {
	m, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/" + port)
	if err != nil {
		return nil, err
	}
	return libp2p.New(append(opts, libp2p.ListenAddrs([]multiaddr.Multiaddr{m}...))...)
}

func TestConnectHosts(ctx context.Context, h1, h2 host.Host) error {

	h1PeerInfo := peer.AddrInfo{
		ID:    h1.ID(),
		Addrs: h1.Addrs(),
	}

	h2PeerInfo := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}

	err := h1.Connect(ctx, h2PeerInfo)
	if err != nil {
		return err
	}
	discovery.TagPeerAsFound(h1, h2.ID())
	err = h2.Connect(ctx, h1PeerInfo)
	if err != nil {
		return err
	}
	discovery.TagPeerAsFound(h2, h1.ID())

	return nil
}
