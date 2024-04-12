package vpn

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"

	"github.com/gfleury/solo/client/crypto/noise"
	"github.com/gfleury/solo/client/protocol"
	"github.com/gfleury/solo/client/vpn/stream_map"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mudler/water"
)

const (
	TUN_INFO_HEADER_SIZE = 4
)

type InterfaceConfig struct {
	// VPNService Encryption key
	PreSharedKey string

	// MTU on interface level
	InterfaceMTU     int
	CreateInterface  bool
	InterfaceName    string
	InterfaceAddress string
}

type VPNInterface struct {
	host             VPNHost
	networkInterface *water.Interface
	hasInfoHeader    bool
	config           *InterfaceConfig

	buffer    *bytes.Buffer
	streamMap *stream_map.AlleinStreamMap
	chain     IOChainPacket
}

func newInterface(config *InterfaceConfig, host VPNHost) (*VPNInterface, error) {
	streamMap := stream_map.NewNoiseStreamMap()
	var err error
	i := &VPNInterface{
		config: config,
		buffer: bytes.NewBuffer(make([]byte, 0)),
		chain: NewIOChainPacket(
			&PacketCompressor{},
			&PacketNoisy{streamMap: streamMap},
		),
		streamMap: streamMap,
		host:      host,
	}
	switch runtime.GOOS {
	case "darwin":
		i.hasInfoHeader = true
	default:
		i.hasInfoHeader = false
	}

	err = i.createInterface(config.CreateInterface)

	return i, err
}

// Just a wrapper for ReadPacket so it can be used as io.Reader
func (v *VPNInterface) Read(b []byte) (int, error) {
	packet, n, err := v.ReadPacket()
	return copy(b, packet[:n]), err
}

// Reads a packet from the TUN interface
// This is called when there is traffic on the TUN interface
// Tip: OUTGOING TRAFFIC (from the client perspective)
func (v *VPNInterface) ReadPacket() (Packet, int, error) {
	packet := make(Packet, v.config.InterfaceMTU)

	n, err := v.networkInterface.Read([]byte(packet))
	if err != nil {
		return packet, n, err
	}

	if v.hasInfoHeader && n > 4 {
		return packet[TUN_INFO_HEADER_SIZE:n], n - TUN_INFO_HEADER_SIZE, err
	}
	return packet[:n], n, err
}

// Tip: OUTGOING TRAFFIC (p2pnetwork)
func (v *VPNInterface) writeStream(stream io.ReadWriter, packet *VPNPacket) (int64, error) {

	if packet.header.Type == VPN_NOISEHANDSHAKE.Uint8() {
		dstID := packet.header.GetDstID()
		streamKey := v.getOutboundStreamKey(dstID)

		// NOISE HANDSHAKE
		// Setup noise handshake stream as initiator
		noiseStream, err := noise.NewNoiseStreamInitiator(v.host.PrivateKey(), v.host.PeerPublicKey(dstID), []byte(v.config.PreSharedKey))
		if err != nil {
			return 0, fmt.Errorf("could not open stream noise to %s: %w", dstID, err)
		}
		reply, err := noiseStream.DoHandshake(nil)
		if err != nil {
			return 0, fmt.Errorf("failed to DoHandshake: %s", err)
		}
		_, err = io.Copy(stream, NewVPNPacket(VPN_NOISEHANDSHAKE, reply, []byte(dstID), []byte(v.host.ID())))
		if err != nil {
			return 0, fmt.Errorf("failed to write handshake msg into stream: %s", err)
		}

		reply = make([]byte, v.config.InterfaceMTU+HEADER_SIZE)
		replySize, err := stream.Read(reply)
		if err != nil {
			return 0, fmt.Errorf("failed to read handshake msg into stream: %s", err)
		}
		vpnPacket := &VPNPacket{}

		_, err = vpnPacket.Write(reply[:replySize])
		if err != nil {
			return 0, fmt.Errorf("failed to marshall VPNPacket: %s", err)
		}

		_, err = noiseStream.DoHandshake(vpnPacket.networkPacket)
		if err != nil {
			return 0, fmt.Errorf("failed to final handshake phase: %s", err)
		}

		v.streamMap.NewWithNoise(streamKey, stream, noiseStream)

		packet.header.Type = VPN_DATA.Uint8()
	}
	n, err := io.Copy(stream, v.OutboundChain(packet))

	// Remove the 4 bytes from the VPNPacket size header

	return n - HEADER_SIZE, err
}

func (v *VPNInterface) handlePacket(ctx context.Context, dstID peer.ID, packet Packet) error {
	streamKey := v.getOutboundStreamKey(dstID)
	// Open a  Data stream if necessary
	if soloStream, ok := v.streamMap.Get(streamKey); ok {
		// TODO: Return read bytes here and aggregate somewhere
		_, err := v.writeStream(soloStream.Stream, NewVPNPacket(VPN_DATA, packet, []byte(dstID), []byte(v.host.ID())))
		if err == nil {
			return nil
		}

		// Stream ist tot
		// v.logger.Debugf("Finish and remove noiseStream and data stream: %s %s", dstID, err)
		soloStream.Stream.(network.Stream).Reset()
		v.streamMap.Delete(streamKey)

		return err
	} else {
		// v.logger.Debugf("Create new data stream for %s", streamKey)
		stream, err := v.host.NewStream(ctx, dstID, protocol.ALLEIN.ID())
		if err != nil {
			return fmt.Errorf("could not open stream to %s: %w", dstID, err)
		}

		// Set first type as VPN_NOISEHANDSHAKE to force handshake insive the VPNInterface
		_, err = v.writeStream(stream, NewVPNPacket(VPN_NOISEHANDSHAKE, packet, []byte(dstID), []byte(v.host.ID())))
		// v.logger.Debugf("Stream created sucessfuly: %s", streamKey)

		return err
	}
}

// Writes packet on the TUN Interface
// This is called when there is data on the incoming stream
// Tip: INCOMING TRAFFIC (from the client perspective)
func (v *VPNInterface) Write(b []byte) (int, error) {
	n, err := v.buffer.Write(b)
	if err != nil {
		return n, err
	}

	// Wait until the size is complete (4 bytes)
	for v.buffer.Len() > HEADER_SIZE {
		p := &VPNPacket{}
		err := binary.Read(v.buffer, binary.BigEndian, &p.header)
		if err != nil {
			if err == io.EOF {
				_, err := io.Copy(v.buffer, p)
				return n, err
			}
			return n, err
		}

		networkPacket := make(Packet, 0)

		for idx := 0; idx < int(p.header.Size); idx++ {
			var nb byte
			err := binary.Read(v.buffer, binary.BigEndian, &nb)
			if err != nil {
				if err == io.EOF {
					_, err := io.Copy(v.buffer, p)
					v.buffer.Write(networkPacket)
					return n, err
				}
				return n, err
			}
			networkPacket = append(networkPacket, nb)
		}
		p.networkPacket = networkPacket

		switch p.header.Type {
		case VPN_NOISEHANDSHAKE.Uint8():
			// Noise handshake
			dstID := p.header.GetSrcID()

			streamKey := v.getInboundStreamKey(dstID)
			soloStream, found := v.streamMap.Get(streamKey)
			if found && soloStream.NoiseStream == nil {
				noiseStream, err := noise.NewNoiseStreamReceiver(v.host.PrivateKey(), v.host.PeerPublicKey(dstID), []byte(v.config.PreSharedKey))
				if err != nil {
					return 0, fmt.Errorf("failed to create noise stream on receiver side: %s", err)
				}
				reply, err := noiseStream.DoHandshake(p.networkPacket)
				if err != nil {
					return 0, fmt.Errorf("failed to noise handshake: %s", err)
				}
				if reply != nil {
					_, err = io.Copy(soloStream.Stream, NewVPNPacket(VPN_NOISEHANDSHAKE, reply, p.header.SrcID[:], []byte(v.host.ID())))
					if err != nil {
						return 0, fmt.Errorf("failed to write msg into incomingStream: %s", err)
					}
				}

				soloStream.NoiseStream = noiseStream
				v.streamMap.Put(streamKey, soloStream)
			}
		case VPN_DATA.Uint8():
			ioProcessedPacket, err := v.InboundChain(p)
			if err != nil {
				return n, fmt.Errorf("packet has been dropped by InboundChain: %s", err)
			}
			_, err = v.writeToNetworkInterface(ioProcessedPacket.networkPacket)
			if err != nil {
				_, err2 := io.Copy(v.buffer, p)
				return n, fmt.Errorf("network write error: %s, packet buffer %s", err, err2)
			}
		}
	}
	return n, nil
}

func (v *VPNInterface) getOutboundStreamKey(dstID peer.ID) string {
	return v.host.ID().String() + dstID.String()
}

func (v *VPNInterface) getInboundStreamKey(dstID peer.ID) string {
	srcID := v.host.ID()
	return dstID.String() + srcID.String()
}

// Call IOChain from the higher encapsulation level
// Unseal -> Uncompress
func (v *VPNInterface) InboundChain(packet *VPNPacket) (*VPNPacket, error) {
	if v.chain == nil {
		return packet, nil
	}
	// Invoke IOChain for writing and update size afterwards
	p, err := v.chain.InboundChain(packet)
	p.header.Size = uint32(len(p.networkPacket))
	return p, err
}

// Call IOChain from the lower encapsulation level
// Compress -> Seal
func (v *VPNInterface) OutboundChain(packet *VPNPacket) *VPNPacket {
	if v.chain == nil {
		return packet
	}
	// Invoke IOChain for reading and update size afterwards
	p := v.chain.OutboundChain(packet)
	p.header.Size = uint32(len(p.networkPacket))
	return p
}

func (v *VPNInterface) writeToNetworkInterface(packet Packet) (int, error) {
	if v.hasInfoHeader {
		tunInfoHeader, err := packet.GetTunInfoHeader()
		if err != nil {
			return 0, err
		}
		packet = append(tunInfoHeader, packet...)
		n, err := v.networkInterface.Write(packet)
		return n - TUN_INFO_HEADER_SIZE, err
	}
	return v.networkInterface.Write(packet)
}
