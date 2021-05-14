package vpn

type RawCopy struct {
	*BasicIOPacket
}

func (w *RawCopy) InboundChain(packet *VPNPacket) (*VPNPacket, error) {
	return w.callNext(packet)
}

func (w *RawCopy) OutboundChain(packet *VPNPacket) *VPNPacket {
	return w.callPrevious(packet)
}
