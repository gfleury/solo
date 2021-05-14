package vpn

import (
	"github.com/gfleury/solo/client/logger"
)

type PacketLogger struct {
	*BasicIOPacket
	logger logger.Logger
}

func (w *PacketLogger) InboundChain(packet *VPNPacket) (*VPNPacket, error) {
	w.logger.Infof("> %s -> %s %d %d %d", packet.header.GetDstID(), packet.header.GetSrcID(), packet.header.Version, packet.header.Count, packet.header.Size)

	return w.callNext(packet)
}

func (w *PacketLogger) OutboundChain(packet *VPNPacket) *VPNPacket {
	w.logger.Infof("< %s -> %s %d %d %d", packet.header.GetDstID(), packet.header.GetSrcID(), packet.header.Version, packet.header.Count, packet.header.Size)

	return w.callPrevious(packet)
}
