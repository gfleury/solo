package vpn

import "golang.org/x/exp/slices"

type IOChainPacket interface {
	InboundChain(*VPNPacket) (*VPNPacket, error)
	OutboundChain(*VPNPacket) *VPNPacket
	SetNext(IOChainPacket)
	SetPrevious(IOChainPacket)
}

func NewIOChainPacket(ioChain ...IOChainPacket) IOChainPacket {
	b := &BasicIOPacket{}
	var root IOChainPacket

	root = b
	for _, chain := range ioChain {
		root.SetPrevious(chain)
		root = chain
	}

	slices.Reverse(ioChain)

	root = b
	for _, chain := range ioChain {
		root.SetNext(chain)
		root = chain
	}

	return b
}

type BasicIOPacket struct {
	next     IOChainPacket
	previous IOChainPacket
}

func (b *BasicIOPacket) SetNext(c IOChainPacket) {
	b.next = c
}

func (b *BasicIOPacket) SetPrevious(c IOChainPacket) {
	b.previous = c
}

func (b *BasicIOPacket) callNext(packet *VPNPacket) (*VPNPacket, error) {
	if b != nil && b.next != nil {
		return b.next.InboundChain(packet)
	}
	return packet, nil
}

func (b *BasicIOPacket) callPrevious(packet *VPNPacket) *VPNPacket {
	if b != nil && b.previous != nil {
		return b.previous.OutboundChain(packet)
	}
	return packet
}

func (b *BasicIOPacket) InboundChain(packet *VPNPacket) (*VPNPacket, error) {
	return b.next.InboundChain(packet)
}

func (b *BasicIOPacket) OutboundChain(packet *VPNPacket) *VPNPacket {
	return b.previous.OutboundChain(packet)
}
