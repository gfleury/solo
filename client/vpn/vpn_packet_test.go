package vpn

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/netip"
	"testing"

	"github.com/mudler/water"
	"github.com/stretchr/testify/suite"
	"golang.zx2c4.com/wireguard/tun/tuntest"
)

type VPNPacketTestSuite struct {
	suite.Suite
}

func TestVPNPacketTestSuite(t *testing.T) {
	suite.Run(t, new(VPNPacketTestSuite))
}

func (s *VPNPacketTestSuite) SetupTest() {
}

func (s *VPNPacketTestSuite) TestHostIDSizeMatchesOurHeader() {
	h1, _ := NewTestHost("0")
	s.Equal(PEER_ID_SIZE, len([]byte(h1.ID())))
}

func (s *VPNPacketTestSuite) TestEOFWhileWrite() {
	testInterface := NewTestPacketBuffer()
	v := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
	}

	testInterfacePacketBuffer(s, testInterface, v)
}

func (s *VPNPacketTestSuite) TestEOFWhileWriteWithRawCopy() {
	testInterface := NewTestPacketBuffer()
	v := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
		chain:            NewIOChainPacket(&RawCopy{}),
	}

	testInterfacePacketBuffer(s, testInterface, v)
}

func testInterfacePacketBuffer(s *VPNPacketTestSuite, testInterface *TestPacketBuffer, v *VPNInterface) {

	// Full packet
	b := bytes.NewBuffer([]byte{})

	// version       uint8   // 1 byte
	// count         uint64  // 8 bytes
	// size          uint32  // 4 bytes
	// reserved      [2]byte // 2 bytes
	// dstPeerID     peer.ID
	// srcPeerID     peer.ID
	// networkPacket Packet

	header := Header{Version: uint8(1),
		Count:    uint32(0),
		Size:     uint32(4),
		Type:     VPN_DATA.Uint8(),
		Reserved: [2]byte{},
		DstID:    [PEER_ID_SIZE]byte{},
		SrcID:    [PEER_ID_SIZE]byte{},
	}
	binary.Write(b, binary.BigEndian, &header)

	binary.Write(b, binary.BigEndian, []byte{1, 2, 3, 4})

	n, err := io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(HEADER_SIZE+4), n)

	// First header than network packet
	err = binary.Write(b, binary.BigEndian, &header)
	s.NoError(err)

	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(HEADER_SIZE), n)

	binary.Write(b, binary.BigEndian, []byte{1, 2, 3, 4})
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(4), n)

	s.ElementsMatch(testInterface.myPackets, []byte{1, 2, 3, 4, 1, 2, 3, 4})

	// Clean myPackets
	testInterface.myPackets = []byte{}

	// Header + half network packet
	binary.Write(b, binary.BigEndian, &header)
	binary.Write(b, binary.BigEndian, []byte{1, 2})
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(90), n)

	binary.Write(b, binary.BigEndian, []byte{3, 4})
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(2), n)

	s.ElementsMatch(testInterface.myPackets, []byte{1, 2, 3, 4})

	// Clean myPackets
	testInterface.myPackets = []byte{}

	// Empty buffer
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(0), n)
	s.Equal(v.buffer.Len(), 0)
	s.Equal(testInterface.MyPacketsLen(), 0)

	// Half header
	binary.Write(b, binary.BigEndian, &header.Version)
	binary.Write(b, binary.BigEndian, &header.Size)
	binary.Write(b, binary.BigEndian, &header.Count)
	binary.Write(b, binary.BigEndian, &header.Type)
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(10), n)

	binary.Write(b, binary.BigEndian, &header.Reserved)
	binary.Write(b, binary.BigEndian, &header.DstID)
	binary.Write(b, binary.BigEndian, &header.SrcID)
	binary.Write(b, binary.BigEndian, []byte{1, 2, 3, 4})
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(82), n)

	s.ElementsMatch(testInterface.myPackets, []byte{1, 2, 3, 4})

}

func (s *VPNPacketTestSuite) TestBugDoMilenio() {
	testInterface := NewTestPacketBuffer()
	v := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
	}

	pingPacket := tuntest.Ping(netip.MustParseAddr("10.1.1.1"), netip.MustParseAddr("10.1.1.1"))

	// n, err := io.Copy(v, NewVPNPacket(pingPacket))
	// s.NoError(err)
	// s.Equal(n, int64(36))

	// s.ElementsMatch(testInterface.myPackets, pingPacket)

	// testInterface.myPackets = []byte{}

	header := Header{Type: VPN_DATA.Uint8(), Size: uint32(len(pingPacket))}

	n, err := io.Copy(v, VPNPacket{header, nil})
	if err != io.ErrUnexpectedEOF {
		s.NoError(err)
	}
	s.Equal(n, int64(HEADER_SIZE))

	s.Equal(testInterface.MyPacketsLen(), 0)

	b := bytes.NewBuffer([]byte{})
	secondHalf := pingPacket[16:]
	firstHalf := pingPacket[:16]
	nn, err := b.Write(firstHalf)
	s.NoError(err)
	s.Equal(nn, 16)
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(n, int64(16))

	s.Equal(testInterface.MyPacketsLen(), 0)

	nn, err = b.Write(secondHalf)
	s.NoError(err)
	s.Equal(nn, 16)
	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(n, int64(16))

	s.ElementsMatch(testInterface.myPackets, pingPacket)

}

func (s *VPNPacketTestSuite) TestVPNPacketTest() {
	b := []byte{}
	bBuffer := bytes.NewBuffer(b)
	pingPacket := tuntest.Ping(netip.MustParseAddr("10.1.1.1"), netip.MustParseAddr("10.1.1.1"))

	vpnPacket := NewVPNPacket(VPN_DATA, pingPacket, []byte{}, []byte{})

	emptyVPNPacket := VPNPacket{}

	n1, _ := io.Copy(bBuffer, vpnPacket)
	s.Equal(len(pingPacket)+HEADER_SIZE, int(n1))

	n2, err := io.Copy(&emptyVPNPacket, bBuffer)
	s.NoError(err)
	s.Equal(len(pingPacket)+HEADER_SIZE, int(n2))

	s.True(vpnPacket.Equal(emptyVPNPacket))
}
