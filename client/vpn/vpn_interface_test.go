package vpn

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gfleury/solo/client/crypto/noise"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/utils"
	"github.com/gfleury/solo/client/vpn/stream_map"

	noisepkg "github.com/flynn/noise"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2p_protocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mudler/water"
	"github.com/stretchr/testify/suite"
	"golang.zx2c4.com/wireguard/tun/tuntest"
)

type VPNInterfaceTestSuite struct {
	suite.Suite
}

func TestVPNInterfaceTestSuite(t *testing.T) {
	suite.Run(t, new(VPNInterfaceTestSuite))
}

func (s *VPNInterfaceTestSuite) SetupTest() {
}

type FakeHost struct {
	stream *TestPacketBuffer
}

func (h *FakeHost) NewStream(ctx context.Context, p peer.ID, pids ...libp2p_protocol.ID) (io.ReadWriter, error) {
	return h.stream, nil
}

type PacketReader struct {
	sync.Mutex
	count  int
	packet Packet
}

func (p *PacketReader) Read(r []byte) (int, error) {
	p.Lock()
	defer p.Unlock()
	if p.count < 1 {
		return 0, io.EOF
	}
	p.count--
	n := copy(r, p.packet)
	return n, nil
}

type TestPacketBufferFeed struct {
	t *TestPacketBuffer
}

func (t *TestPacketBufferFeed) Write(b []byte) (int, error) {
	t.t.Lock()
	defer t.t.Unlock()
	packet := make(Packet, len(b))
	n := copy(packet, b)
	t.t.packets = append(t.t.packets, packet) // Add
	return n, nil
}

type TestPacketBuffer struct {
	sync.Mutex
	packets   []Packet
	myPackets []byte
	ip        netip.Addr
}

func (t *TestPacketBuffer) Read(b []byte) (int, error) {
	t.Lock()
	defer t.Unlock()
	for len(t.packets) < 1 {
		return 0, io.EOF
	}

	packet := t.packets[0]
	n := copy(b, packet)
	t.packets = t.packets[1:] // Remove

	return n, nil
}

func (t *TestPacketBuffer) Write(b []byte) (int, error) {
	t.Lock()
	defer t.Unlock()
	p := Packet(b)
	dstIP, err := p.DstIp()
	if err == nil && strings.Contains(dstIP.String(), t.ip.String()) {
		srcIP, err := p.SrcIp()
		if err != nil {
			return 0, err
		}

		fmt.Println("Received Packet fuer mich")
		reply := tuntest.Ping(netip.AddrFrom4([4]byte(srcIP)), netip.AddrFrom4([4]byte(dstIP)))
		t.packets = append(t.packets, Packet(reply))
	}
	t.myPackets = append(t.myPackets, b...) // Add
	return len(b), nil
}

func (t *TestPacketBuffer) PacketsLen() int {
	t.Lock()
	defer t.Unlock()
	return len(t.packets)
}

func (t *TestPacketBuffer) MyPacketsLen() int {
	t.Lock()
	defer t.Unlock()
	return len(t.myPackets)
}

func (t *TestPacketBuffer) Close() error {
	return nil
}

func NewTestPacketBuffer(ips ...netip.Addr) *TestPacketBuffer {
	i := &TestPacketBuffer{
		packets:   []Packet{},
		myPackets: []byte{},
	}
	if len(ips) > 0 {
		i.ip = ips[0]
	}
	return i
}

func (s *VPNInterfaceTestSuite) TestVPNInterface5() {
	s.runVPNInterfaceTest(5)
}

func (s *VPNInterfaceTestSuite) TestVPNInterfaceLots() {
	s.runVPNInterfaceTest(10000000)
}

func (s *VPNInterfaceTestSuite) runVPNInterfaceTest(packetCount int) {
	pingPacket := tuntest.Ping(netip.MustParseAddr("10.1.1.1"), netip.MustParseAddr("10.1.1.1"))

	testInterface := NewTestPacketBuffer()
	hostStream := NewTestPacketBuffer()
	p := &PacketReader{
		packet: pingPacket,
		count:  packetCount,
	}
	v := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
	}

	// Write packets on network interface
	io.Copy(&TestPacketBufferFeed{t: testInterface}, p)

	// Read from VPNInterface
	io.Copy(hostStream, v)

	// Wait packet to arrive on hostStream
	now := time.Now()
	for hostStream.MyPacketsLen() < packetCount && time.Since(now) < 5*time.Second {
	}

	s.Equal(len(pingPacket)*packetCount, hostStream.MyPacketsLen())

	s.True(Diff(bytes.NewReader(hostStream.myPackets), &PacketReader{packet: pingPacket, count: packetCount}))
}

func (s *VPNInterfaceTestSuite) TestEOFWhileWriteWithPacketCompressor() {
	testInterface := NewTestPacketBuffer()
	v := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
		chain:            NewIOChainPacket(&PacketCompressor{}),
	}

	pingPacket := tuntest.Ping(netip.MustParseAddr("10.1.1.1"), netip.MustParseAddr("10.1.1.1"))

	b := bytes.NewBuffer([]byte{})
	n, err := v.writeStream(b, NewVPNPacket(VPN_DATA, pingPacket, nil, nil))
	s.NoError(err)
	s.Equal(int64(52), n)

	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(140), n)

	packetReader := bytes.NewReader(testInterface.myPackets)
	bb, _ := io.ReadAll(packetReader)
	s.True(bytes.Equal(pingPacket, bb))

}

func (s *VPNInterfaceTestSuite) TestPacketLogger() {
	testInterface := NewTestPacketBuffer()
	logger := logger.New(log.LevelDebug)
	v := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
		chain: NewIOChainPacket(
			&PacketLogger{
				logger: *logger,
			}),
	}

	pingPacket := tuntest.Ping(netip.MustParseAddr("10.1.1.1"), netip.MustParseAddr("10.1.1.1"))

	prvKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 4096)
	s.NoError(err)

	id, err := peer.IDFromPrivateKey(prvKey)
	s.NoError(err)

	logger.Infof("Logging packet: %s", id)
	b := bytes.NewBuffer([]byte{})
	n, err := v.writeStream(b, NewVPNPacket(VPN_DATA, pingPacket, []byte(id), []byte(id)))
	s.NoError(err)
	s.Equal(int64(32), n)

	n, err = io.Copy(v, b)
	s.NoError(err)
	s.Equal(int64(120), n)

	packetReader := bytes.NewReader(testInterface.myPackets)
	bb, _ := io.ReadAll(packetReader)
	s.True(bytes.Equal(pingPacket, bb))

}

func (s *VPNInterfaceTestSuite) TestPacketNoisySimple() {

	kI, err := noisepkg.DH25519.GenerateKeypair(rand.Reader)
	s.Require().NoError(err)

	kR, err := noisepkg.DH25519.GenerateKeypair(rand.Reader)
	s.Require().NoError(err)

	i, err := noise.NewNoiseStreamInitiator(&kI, kR.Public, []byte("supersecretsupersecretsupersecre"))
	s.Require().NoError(err)
	r, err := noise.NewNoiseStreamReceiver(&kR, kI.Public, []byte("supersecretsupersecretsupersecre"))
	s.Require().NoError(err)

	replyMsg, err := i.DoHandshake(nil)
	s.Require().NoError(err)

	replyMsg, err = r.DoHandshake(replyMsg)
	s.Require().NoError(err)

	_, err = i.DoHandshake(replyMsg)
	s.Require().NoError(err)

	streamMap := stream_map.NewNoiseStreamMap()

	prvKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 4096)
	s.NoError(err)

	id, err := peer.IDFromPrivateKey(prvKey)
	s.NoError(err)

	stream := utils.NewFakeStream(id.Pretty())
	streamMap.NewWithNoise(getStreamKey(id, id), stream, i)

	testInterface := NewTestPacketBuffer()
	v1 := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
		chain: NewIOChainPacket(
			&PacketNoisy{
				streamMap: streamMap,
			},
		),
	}

	pingPacket := tuntest.Ping(netip.MustParseAddr("10.1.1.1"), netip.MustParseAddr("10.1.1.1"))
	vpnPacket := NewVPNPacket(VPN_DATA, pingPacket, []byte(id), []byte(id))

	time.Sleep(1 * time.Second)
	b := bytes.NewBuffer([]byte{})
	n, err := v1.writeStream(b, vpnPacket)
	s.NoError(err)
	s.Equal(int64(48), n)

	streamMap.NewWithNoise(getStreamKey(id, id), stream, r)

	v2 := &VPNInterface{
		networkInterface: &water.Interface{ReadWriteCloser: testInterface},
		config:           &InterfaceConfig{InterfaceMTU: 1420},
		buffer:           bytes.NewBuffer(make([]byte, 0)),
		chain: NewIOChainPacket(
			&PacketNoisy{
				streamMap: streamMap,
			},
		),
	}

	n, err = io.Copy(v2, b)
	s.NoError(err)
	s.Equal(int64(136), n)

	s.ElementsMatch(pingPacket, testInterface.myPackets)
}
