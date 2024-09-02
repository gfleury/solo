package vpn

// Define the maximum storage size as 1MB
const MaxStorageSize = 5 * 1024 * 1024 // 1MB

type packetRingWrapper struct {
	retryCount int
	packet     Packet
}

// Storage struct to manage Packets
type PacketRing struct {
	wrappedPackets []packetRingWrapper
	totalSize      int
}

// AddPacket adds an Packet to the storage, replacing the oldest one if necessary
func (s *PacketRing) AddPacket(packet Packet) {
	packetSize := len(packet)

	// If adding the Packet would exceed the limit, remove oldest Packets
	for s.totalSize+packetSize > MaxStorageSize {
		oldest := s.wrappedPackets[0]
		oldestSize := len(oldest.packet)
		s.wrappedPackets = s.wrappedPackets[1:]
		s.totalSize -= oldestSize
	}

	// Add the new Packet
	s.wrappedPackets = append(s.wrappedPackets, packetRingWrapper{packet: packet, retryCount: 0})
	s.totalSize += packetSize
}

// ProcessObjects processes each object in the slice, removing it after a successful operation
func (s *PacketRing) ProcessPackets(processFunc func(Packet) error) {
	i := 0
	for i < len(s.wrappedPackets) {
		// Perform the operation
		err := processFunc(s.wrappedPackets[i].packet)
		if err == nil {
			s.wrappedPackets = append(s.wrappedPackets[:i], s.wrappedPackets[i+1:]...)
		} else {
			i++
		}
	}

}
