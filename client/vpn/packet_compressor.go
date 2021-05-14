package vpn

import "github.com/gfleury/solo/client/utils"

type PacketCompressor struct {
	BasicIOPacket
}

func (w *PacketCompressor) InboundChain(packet *VPNPacket) (*VPNPacket, error) {
	b, err := utils.Decompress(packet.networkPacket)
	if err != nil {
		packet = &VPNPacket{}
		return packet, err
	}
	packet.networkPacket = b
	return w.callNext(packet)
}

func (w *PacketCompressor) OutboundChain(packet *VPNPacket) *VPNPacket {
	b := utils.Compress(packet.networkPacket)
	packet.networkPacket = b
	return w.callPrevious(packet)
}
