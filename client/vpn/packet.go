package vpn

import (
	"fmt"
	"net"
	"syscall"
)

type Packet []byte

func (packet Packet) IpVersion() uint8 {
	return uint8(packet[0]) >> 4
}

func (packet Packet) GetTunInfoHeader() ([]byte, error) {
	var err error
	tunInfoHeader := make(Packet, 4)

	switch packet.IpVersion() {
	case 4:
		tunInfoHeader[3] = syscall.AF_INET
	case 6:
		tunInfoHeader[3] = syscall.AF_INET6
	default:
		err = fmt.Errorf("cannot identify IP Header version: %d", packet.IpVersion())
	}

	return tunInfoHeader, err
}

func (packet Packet) DstIp() (net.IP, error) {
	var dstIP net.IP
	var err error
	switch packet.IpVersion() {
	case 4:
		dstIP = net.IP(packet[16:20])
	case 6:
		dstIP = net.IP(packet[24:40])
	default:
		err = fmt.Errorf("cannot identify IP Header version: %d", packet.IpVersion())
	}

	return dstIP, err
}

func (packet Packet) SrcIp() (net.IP, error) {
	var srcIP net.IP
	var err error
	switch packet.IpVersion() {
	case 4:
		srcIP = net.IP(packet[12:16])
	case 6:
		srcIP = net.IP(packet[8:24])
	default:
		err = fmt.Errorf("cannot identify IP Header version: %d", packet.IpVersion())
	}

	return srcIP, err
}
