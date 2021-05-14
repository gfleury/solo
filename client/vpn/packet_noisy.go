package vpn

import (
	"fmt"

	"github.com/gfleury/solo/client/vpn/stream_map"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PacketNoisy struct {
	BasicIOPacket
	streamMap *stream_map.AlleinStreamMap
}

func getStreamKey(dstID, srcID peer.ID) string {
	return srcID.Pretty() + dstID.Pretty()
}

// Unseal Packet TO byte slice b
func (p *PacketNoisy) InboundChain(packet *VPNPacket) (*VPNPacket, error) {
	streamKey := getStreamKey(packet.header.GetDstID(), packet.header.GetSrcID())
	if stream, found := p.streamMap.Get(streamKey); !found {
		// Could not find the noiseStream to encrypt the packet, set it to empty
		packet.networkPacket = []byte{}
		err := fmt.Errorf("did not found NoiseStream for id: %s", streamKey)
		return packet, err
	} else {
		decryptedNetworkPacket, err := stream.NoiseStream.Decrypt(packet.networkPacket)
		if err != nil {
			// Fail to Decrypt
			packet.networkPacket = []byte{}
			return packet, err
		} else {
			packet.networkPacket = decryptedNetworkPacket
		}
	}
	return p.callNext(packet)
}

// Seal Packet FROM byte slice b
func (p *PacketNoisy) OutboundChain(packet *VPNPacket) *VPNPacket {
	if stream, found := p.streamMap.Get(getStreamKey(packet.header.GetDstID(), packet.header.GetSrcID())); !found {
		// Could not find the noiseStream to encrypt the packet, set it to empty
		packet.networkPacket = []byte{}
	} else {
		encryptedNetworkPacket, err := stream.NoiseStream.Encrypt(packet.networkPacket)
		if err != nil {
			packet.networkPacket = []byte{}
		} else {
			packet.networkPacket = encryptedNetworkPacket
		}
	}
	return p.callPrevious(packet)
}
