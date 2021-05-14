package broadcast

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/broadcast/prp"
	"github.com/gfleury/solo/client/crypto"
	"github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/protocol"
	"github.com/gfleury/solo/client/types"
	"github.com/ipfs/go-log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type StreamBroadcaster struct {
	sync.Mutex

	discoveryPeersIDs []peer.ID
	selfHost          host.Host
	ready             bool
	otpKey            crypto.OTPKey
	sealer            crypto.Sealer
	logger            log.StandardLogger
	PRPTable          *prp.PRPTableType
}

func NewStreamBroadcaster(
	logger log.StandardLogger,
	discoveryPeers discovery.AddrList,
	otpKey crypto.OTPKey,
) Broadcaster {
	discoveryPeersIDs := make([]peer.ID, len(discoveryPeers))
	for _, discoveryPeer := range discoveryPeers {
		i, _ := peer.AddrInfoFromP2pAddr(discoveryPeer)
		discoveryPeersIDs = append(discoveryPeersIDs, i.ID)
	}

	return &StreamBroadcaster{
		otpKey:            otpKey,
		sealer:            &crypto.DefaultSealer{},
		logger:            logger,
		PRPTable:          prp.NewPRPTable(),
		discoveryPeersIDs: discoveryPeersIDs,
	}
}
func (m *StreamBroadcaster) Lookup(dstIP string) (*types.Machine, bool, bool) {
	return m.PRPTable.Lookup(dstIP)
}

func (m *StreamBroadcaster) StreamHandler() func(stream network.Stream) {
	var mutex sync.Mutex
	msg := make([]byte, 1500)

	return func(stream network.Stream) {
		mutex.Lock()
		defer mutex.Unlock()

		m.logger.Debugf("New broadcast stream for %s", stream.Conn().RemotePeer())
		n, err := stream.Read(msg)
		if err != nil {
			m.logger.Warnf("Fail to receive message: %s", err.Error())
			return
		}

		unsealedPacket, err := m.sealer.Unseal(msg[:n], m.otpKey.TOTPSHA256(sha256.New))
		if err != nil {
			m.logger.Warnf("Fail to unseal receiving message: %s", err.Error())
			return
		}

		cm := &metapacket.MetaPacket{}
		err = json.Unmarshal(unsealedPacket, cm)
		if err != nil {
			m.logger.Errorf("Unable to unmarshal received MetaPacket: %s", err)
			return
		}

		cm.SenderID = stream.Conn().RemotePeer().String()

		if payload := cm.GetPayload(); payload != nil {
			replyPayload, err := payload.Process(m.logger, m.PRPTable)
			if err != nil {
				m.logger.Errorf("Unable to process received MetaPacket Payload: %s", err)
				return
			}
			if replyPayload != nil {
				ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelFunc()

				m.SendPacket(ctx, metapacket.NewFromPayload(replyPayload))
			}
		}
		stream.Reset()
	}
}

func (m *StreamBroadcaster) Ready() bool {
	m.Lock()
	defer m.Unlock()
	return m.ready
}

func (m *StreamBroadcaster) SendPacket(ctx context.Context, packet *metapacket.MetaPacket) error {

	ctxTimeout, cancelFunc := context.WithTimeout(ctx, 2*time.Second)
	defer cancelFunc()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if !m.Ready() {
		err := fmt.Errorf("Broadcaster still not ready")
		m.logger.Error(err)
		return err
	}

	peersIDs := m.selfHost.Network().Peers()
	regularPeersIDs := []peer.ID{}

	for _, peerID := range peersIDs {
		// Filter peers that were not found by DHT
		if !IsPeerFoundByDiscovery(m.selfHost, peerID) {
			continue
		}
		regularPeersIDs = append(regularPeersIDs, peerID)
	}

	m.logger.Debugf("Broadcasting to peers: %s", regularPeersIDs)

	for _, peerID := range regularPeersIDs {

		bytesPacket, err := json.Marshal(packet)
		if err != nil {
			m.logger.Errorf("Broadcast to peer %s failed with: %s", peerID, err)
			return err
		}
		sealedPacket, err := m.sealer.Seal(bytesPacket, m.otpKey.TOTPSHA256(sha256.New))
		if err != nil {
			m.logger.Errorf("Broadcast to peer %s failed with: %s", peerID, err)
			return err
		}

		stream, err := m.selfHost.NewStream(ctxTimeout, peerID, protocol.BROADCAST.ID())
		if err != nil {
			m.logger.Errorf("Broadcast to peer %s failed with: %s", peerID, err)
			return err
		}

		n, err := stream.Write(sealedPacket)
		if err != nil {
			m.logger.Errorf("Broadcast to peer %s failed with: %s", peerID, err)
			return err
		} else if n != len(sealedPacket) {
			m.logger.Errorf("Wrote wrong amount of bytes into broadcast stream, expected %s wrote %d", len(sealedPacket), n)
			return err
		}
	}

	return nil
}

func (m *StreamBroadcaster) Start(ctx context.Context, host host.Host, myIP string) error {
	m.Lock()
	defer m.Unlock()

	// Insert myself on the PRPTable
	myselfMachine := newMachine(host, myIP)
	m.PRPTable.InsertMyselfEntry(&myselfMachine)
	m.selfHost = host

	// Set the VPN P2P stream handler (for incoming VPNPacket streams)
	host.SetStreamHandler(protocol.BROADCAST.ID(), m.StreamHandler())

	m.ready = true
	return nil
}

func (m *StreamBroadcaster) PRPRequest(ctx context.Context, unknownDstIP string) error {
	return m.SendPacket(ctx, metapacket.NewFromPayload(prp.NewPRPRequestPacket(unknownDstIP)))
}

func (m *StreamBroadcaster) AnnounceMyself(ctx context.Context) error {
	return m.SendPacket(ctx, metapacket.NewFromPayload(m.PRPTable.PRPReplyMyself(true)))
}

func IsPeerFoundByDiscovery(host host.Host, peerID peer.ID) bool {
	tags := host.ConnManager().GetTagInfo(peerID)
	if tags != nil {
		_, ok := tags.Tags[discovery.DHT_FOUND]
		return ok
	}
	return false
}
