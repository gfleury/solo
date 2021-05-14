package vpn

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	HEADER_SIZE                = 88
	PEER_ID_SIZE               = 38
	VPN_DATA     VPNPacketType = iota
	VPN_NOISEHANDSHAKE
)

type VPNPacketType uint8

type Header struct {
	Version  uint8              // 1 byte
	Size     uint32             // 4 bytes
	Count    uint32             // 4 bytes
	Type     uint8              // 1 byte
	Reserved [2]byte            // 3 bytes
	DstID    [PEER_ID_SIZE]byte // 38 bytes
	SrcID    [PEER_ID_SIZE]byte // 38 bytes

}

type VPNPacket struct {
	header        Header
	networkPacket Packet
}

func DeslicePeerID(id []byte) [PEER_ID_SIZE]byte {
	ret := [PEER_ID_SIZE]byte{}
	for i := 0; len(id) == PEER_ID_SIZE && i < PEER_ID_SIZE; i++ {
		ret[i] = id[i]
	}
	return ret
}

func NewVPNPacket(t VPNPacketType, packet Packet, dst []byte, src []byte) *VPNPacket {
	vpnPacket := &VPNPacket{Header{
		Version:  1,
		Count:    0,
		Type:     t.Uint8(),
		Reserved: [2]byte{0, 0},
		Size:     uint32(len(packet)),
		DstID:    DeslicePeerID(dst),
		SrcID:    DeslicePeerID(src),
	}, packet,
	}
	return vpnPacket
}

func (t VPNPacketType) Uint8() uint8 {
	return uint8(t)
}

// Unmarshal VPNPacket TO byte slice b
func (p VPNPacket) Read(b []byte) (int, error) {
	bBuffer := bytes.NewBuffer([]byte{})

	err := binary.Write(bBuffer, binary.BigEndian, &p.header)
	if err != nil {
		return 0, err
	}

	for idx := 0; len(p.networkPacket) == int(p.header.Size) && idx < int(p.header.Size); idx++ {
		err := binary.Write(bBuffer, binary.BigEndian, &p.networkPacket[idx])
		if err != nil {
			return 0, err
		}
	}

	return copy(b, bBuffer.Bytes()), io.EOF
}

// Marshal VPNPacket FROM byte slice b
func (p *VPNPacket) Write(b []byte) (int, error) {
	bReader := bytes.NewReader(b)
	n := 0

	err := binary.Read(bReader, binary.BigEndian, &p.header)
	if err != nil {
		return n, err
	}

	// check if there is enough data to read a full packet
	if len(b) < HEADER_SIZE+int(p.header.Size) {
		return n, io.EOF
	}

	n += HEADER_SIZE

	networkPacket := make(Packet, p.header.Size)

	for idx := 0; idx < int(p.header.Size); idx++ {
		err := binary.Read(bReader, binary.BigEndian, &networkPacket[idx])
		if err != nil {
			return 0, err
		}
	}
	n += int(p.header.Size)

	p.networkPacket = networkPacket
	return n, nil
}

func (p VPNPacket) Equal(b VPNPacket) bool {
	return p.header.Size == b.header.Size && bytes.Equal(p.networkPacket, b.networkPacket)
}

func (h *Header) GetDstID() peer.ID {
	dstID, _ := peer.IDFromBytes(h.DstID[:])
	return dstID
}

func (h *Header) GetSrcID() peer.ID {
	srcID, _ := peer.IDFromBytes(h.SrcID[:])
	return srcID
}
