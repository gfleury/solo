package vpn

import (
	"time"
)

// Define the maximum storage size as 1MB
const MaxStorageSize = 5 * 1024 * 1024 // 5MB

// PacketRingWrapper wraps a packet with additional metadata
type packetRingWrapper struct {
	retryCount int
	packet     Packet
	timestamp  time.Time
}

// PacketRing manages packets with a maximum storage size
type PacketRing struct {
	wrappedPackets []packetRingWrapper
	totalSize      int
}

// AddPacket adds a packet to the storage, replacing the oldest one if necessary
func (s *PacketRing) AddPacket(packet Packet) {
	packetSize := len(packet)
	now := time.Now()

	// Remove packets older than 20 seconds
	s.removeOldPackets(now)

	// If adding the packet would exceed the limit, remove oldest packets
	for s.totalSize+packetSize > MaxStorageSize {
		oldest := s.wrappedPackets[0]
		oldestSize := len(oldest.packet)
		s.wrappedPackets = s.wrappedPackets[1:]
		s.totalSize -= oldestSize
	}

	// Add the new packet
	s.wrappedPackets = append(s.wrappedPackets, packetRingWrapper{packet: packet, retryCount: 0, timestamp: now})
	s.totalSize += packetSize
}

// ProcessPackets processes each packet in the slice, removing it after a successful operation
func (s *PacketRing) ProcessPackets(processFunc func(Packet) error) {
	now := time.Now()
	i := 0
	for i < len(s.wrappedPackets) {
		// Skip packets older than 20 seconds
		if now.Sub(s.wrappedPackets[i].timestamp) > 20*time.Second {
			s.wrappedPackets = append(s.wrappedPackets[:i], s.wrappedPackets[i+1:]...)
			continue
		}

		// Perform the operation
		err := processFunc(s.wrappedPackets[i].packet)
		if err == nil {
			s.wrappedPackets = append(s.wrappedPackets[:i], s.wrappedPackets[i+1:]...)
		} else {
			i++
		}
	}
}

// removeOldPackets removes packets older than 20 seconds from the storage
func (s *PacketRing) removeOldPackets(now time.Time) {
	i := 0
	for i < len(s.wrappedPackets) {
		if now.Sub(s.wrappedPackets[i].timestamp) > 20*time.Second {
			oldestSize := len(s.wrappedPackets[i].packet)
			s.wrappedPackets = append(s.wrappedPackets[:i], s.wrappedPackets[i+1:]...)
			s.totalSize -= oldestSize
		} else {
			i++
		}
	}
}
