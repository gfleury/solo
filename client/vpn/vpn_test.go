package vpn

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"net/netip"
	"testing"
	"time"

	"github.com/gfleury/solo/client/broadcast"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/vpn/stream_map"
	"github.com/ipfs/go-log/v2"
	"github.com/mudler/water"
	"github.com/stretchr/testify/suite"
	"golang.zx2c4.com/wireguard/tun/tuntest"
)

type VPNTestSuite struct {
	suite.Suite
}

type VPNPacketTestReader struct {
	count  int
	packet VPNPacket
}

func (p *VPNPacketTestReader) Read(r []byte) (int, error) {
	if p.count < 1 {
		return 0, io.EOF
	}
	m := bytes.NewBuffer([]byte{})
	nn, err := io.Copy(m, p.packet)
	if err != nil {
		return int(nn), err
	}
	p.count--
	n := copy(r, m.Bytes())
	return int(n), nil
}

func TestVPNTestSuite(t *testing.T) {
	suite.Run(t, new(VPNTestSuite))
}

func (s *VPNTestSuite) SetupTest() {
}

func (s *VPNTestSuite) TestVPNTestSimple() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()

	// log.SetAllLoggers(log.LevelDebug)

	ip1 := netip.MustParseAddr("10.0.0.1")
	ip2 := netip.MustParseAddr("10.0.0.2")

	dummyBroadcast := broadcast.NewDummyBroadcast()

	testInterface1 := NewTestPacketBuffer()

	testInterface2 := NewTestPacketBuffer()

	logger := logger.New(log.LevelDebug)

	h1, _ := NewTestHost("0")
	h2, _ := NewTestHost("0")

	s.NotEqual(h1.ID(), h2.ID())

	nsm1 := stream_map.NewNoiseStreamMap()
	nsm2 := stream_map.NewNoiseStreamMap()

	vpn1 := &VPNService{
		vpnInterface: &VPNInterface{
			networkInterface: &water.Interface{ReadWriteCloser: testInterface1},
			config:           &InterfaceConfig{InterfaceMTU: 1420},
			buffer:           bytes.NewBuffer(make([]byte, 0)),
			streamMap:        nsm1,
			chain: &PacketNoisy{
				streamMap: nsm1,
			},
			host: NewWrapperHost(h1),
		},
		broadcast: dummyBroadcast,
		timeout:   2 * time.Second,
	}

	vpn2 := &VPNService{
		vpnInterface: &VPNInterface{
			networkInterface: &water.Interface{ReadWriteCloser: testInterface2},
			config:           &InterfaceConfig{InterfaceMTU: 1420},
			buffer:           bytes.NewBuffer(make([]byte, 0)),
			streamMap:        nsm2,
			chain: &PacketNoisy{
				streamMap: nsm2,
			},
			host: NewWrapperHost(h2),
		},
		broadcast: dummyBroadcast,
		timeout:   2 * time.Second,
	}

	dummyBroadcast.AddFakePeer(ip1.String(), h1.ID())
	dummyBroadcast.AddFakePeer(ip2.String(), h2.ID())

	err := TestConnectHosts(ctx, h1, h2)
	if !s.NoError(err) {
		return
	}

	err = vpn1.Run(ctx, logger, h1, dummyBroadcast)
	s.NoError(err)
	err = vpn2.Run(ctx, logger, h2, dummyBroadcast)
	s.NoError(err)

	pingPacket := tuntest.Ping(ip2, ip1)
	packetCount := 1024 * 100

	n, err := io.Copy(&TestPacketBufferFeed{t: testInterface1}, &PacketReader{packet: pingPacket, count: packetCount})
	s.NoError(err)
	expectedLen := len(pingPacket) * packetCount
	s.Equal(int(n), expectedLen)

	for testInterface2.MyPacketsLen() < expectedLen && ctx.Err() == nil {
	}

	lenInt2 := testInterface2.MyPacketsLen()

	testInterface2.Lock()
	defer testInterface2.Unlock()
	if !s.True(Diff(bytes.NewReader(testInterface2.myPackets), &PacketReader{packet: pingPacket, count: packetCount})) {
		s.Equal(0, len(testInterface1.packets))
		s.Equal(expectedLen, lenInt2)
	}
}

func (s *VPNTestSuite) TestVPNTestWithPacketNoisy() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()

	// log.SetAllLoggers(log.LevelDebug)

	ip1 := netip.MustParseAddr("10.0.0.1")
	ip2 := netip.MustParseAddr("10.0.0.2")

	dummyBroadcast := broadcast.NewDummyBroadcast()

	testInterface1 := NewTestPacketBuffer()

	testInterface2 := NewTestPacketBuffer()

	logger := logger.New(log.LevelDebug)

	h1, _ := NewTestHost("0")
	h2, _ := NewTestHost("0")

	s.NotEqual(h1.ID(), h2.ID())

	nsm1 := stream_map.NewNoiseStreamMap()
	nsm2 := stream_map.NewNoiseStreamMap()

	vpn1 := &VPNService{
		vpnInterface: &VPNInterface{
			networkInterface: &water.Interface{ReadWriteCloser: testInterface1},
			config:           &InterfaceConfig{InterfaceMTU: 1420},
			buffer:           bytes.NewBuffer(make([]byte, 0)),
			streamMap:        nsm1,
			chain: &PacketNoisy{
				streamMap: nsm1,
			},
			host: NewWrapperHost(h1),
		},
		broadcast: dummyBroadcast,
		timeout:   2 * time.Second,
	}

	vpn2 := &VPNService{
		vpnInterface: &VPNInterface{
			networkInterface: &water.Interface{ReadWriteCloser: testInterface2},
			config:           &InterfaceConfig{InterfaceMTU: 1420},
			buffer:           bytes.NewBuffer(make([]byte, 0)),
			streamMap:        nsm2,
			chain: &PacketNoisy{
				streamMap: nsm2,
			},
			host: NewWrapperHost(h2),
		},
		broadcast: dummyBroadcast,
		timeout:   2 * time.Second,
	}

	dummyBroadcast.AddFakePeer(ip1.String(), h1.ID())
	dummyBroadcast.AddFakePeer(ip2.String(), h2.ID())

	err := TestConnectHosts(ctx, h1, h2)
	if !s.NoError(err) {
		return
	}

	err = vpn1.Run(ctx, logger, h1, dummyBroadcast)
	s.NoError(err)
	err = vpn2.Run(ctx, logger, h2, dummyBroadcast)
	s.NoError(err)

	pingPacket := tuntest.Ping(ip2, ip1)
	packetCount := 1024 * 100

	n, err := io.Copy(&TestPacketBufferFeed{t: testInterface1}, &PacketReader{packet: pingPacket, count: packetCount})
	s.NoError(err)
	expectedLen := len(pingPacket) * packetCount
	s.Equal(int(n), expectedLen)

	for testInterface2.MyPacketsLen() < expectedLen && ctx.Err() == nil {
	}

	lenInt2 := testInterface2.MyPacketsLen()

	testInterface2.Lock()
	defer testInterface2.Unlock()
	if !s.True(Diff(bytes.NewReader(testInterface2.myPackets), &PacketReader{packet: pingPacket, count: packetCount})) {
		s.Equal(0, len(testInterface1.packets))
		s.Equal(expectedLen, lenInt2)
	}
}

func Diff(r1, r2 io.Reader) (identical bool, err error) {
	h1, h2 := sha256.New(), sha256.New()

	if _, err = io.Copy(h1, r1); err != nil {
		return
	}
	if _, err = io.Copy(h2, r2); err != nil {
		return
	}
	return bytes.Equal(h1.Sum(nil), h2.Sum(nil)), nil
}

func (s *VPNTestSuite) TestVPNTestWithOneRoundtripPacketNoisy() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()

	// log.SetAllLoggers(log.LevelDebug)

	ip1 := netip.MustParseAddr("10.0.0.1")
	ip2 := netip.MustParseAddr("10.0.0.2")

	dummyBroadcast := broadcast.NewDummyBroadcast()

	testInterface1 := NewTestPacketBuffer(ip1)

	testInterface2 := NewTestPacketBuffer(ip2)

	logger := logger.New(log.LevelDebug)

	h1, _ := NewTestHost("0")
	h2, _ := NewTestHost("0")

	s.NotEqual(h1.ID(), h2.ID())

	nsm1 := stream_map.NewNoiseStreamMap()
	nsm2 := stream_map.NewNoiseStreamMap()

	vpn1 := &VPNService{
		vpnInterface: &VPNInterface{
			networkInterface: &water.Interface{ReadWriteCloser: testInterface1},
			config:           &InterfaceConfig{InterfaceMTU: 1420},
			buffer:           bytes.NewBuffer(make([]byte, 0)),
			streamMap:        nsm1,
			chain: &PacketNoisy{
				streamMap: nsm1,
			},
			host: NewWrapperHost(h1),
		},
		broadcast: dummyBroadcast,
		timeout:   2 * time.Second,
	}

	vpn2 := &VPNService{
		vpnInterface: &VPNInterface{
			networkInterface: &water.Interface{ReadWriteCloser: testInterface2},
			config:           &InterfaceConfig{InterfaceMTU: 1420},
			buffer:           bytes.NewBuffer(make([]byte, 0)),
			streamMap:        nsm2,
			chain: &PacketNoisy{
				streamMap: nsm2,
			},
			host: NewWrapperHost(h2),
		},
		broadcast: dummyBroadcast,
		timeout:   2 * time.Second,
	}

	dummyBroadcast.AddFakePeer(ip1.String(), h1.ID())
	dummyBroadcast.AddFakePeer(ip2.String(), h2.ID())

	err := TestConnectHosts(ctx, h1, h2)
	if !s.NoError(err) {
		return
	}

	err = TestConnectHosts(ctx, h2, h1)
	if !s.NoError(err) {
		return
	}

	err = vpn1.Run(ctx, logger, h1, dummyBroadcast)
	s.NoError(err)
	err = vpn2.Run(ctx, logger, h2, dummyBroadcast)
	s.NoError(err)

	pingPacket := tuntest.Ping(ip2, ip1)
	packetCount := 1

	n, err := io.Copy(&TestPacketBufferFeed{t: testInterface1}, &PacketReader{packet: pingPacket, count: packetCount})
	s.NoError(err)
	expectedLen := len(pingPacket) * packetCount
	s.Equal(int(n), expectedLen)

	for testInterface2.MyPacketsLen() < expectedLen && ctx.Err() == nil {
	}

	lenInt2 := testInterface2.MyPacketsLen()

	testInterface2.Lock()
	if !s.True(Diff(bytes.NewReader(testInterface2.myPackets), &PacketReader{packet: pingPacket, count: packetCount})) {
		s.Equal(0, len(testInterface1.packets))
		s.Equal(expectedLen, lenInt2)
	}
	testInterface2.Unlock()

	for testInterface1.MyPacketsLen() < expectedLen && ctx.Err() == nil {
	}

	testInterface1.Lock()
	pingPacket = tuntest.Ping(ip1, ip2)
	if !s.True(Diff(bytes.NewReader(testInterface1.myPackets), &PacketReader{packet: pingPacket, count: packetCount})) {
		s.Equal(0, len(testInterface2.packets))
		s.Equal(expectedLen, lenInt2)
	}
	testInterface1.Unlock()

}
