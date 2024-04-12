package rendezvous

import (
	"github.com/gfleury/solo/client/logger"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// ACLFilter is an Access Control mechanism for relayed connect.
type ACLFilter struct {
	logger *logger.Logger
}

// AllowReserve returns true if a reservation from a peer with the given peer ID and multiaddr
// is allowed.
func (a *ACLFilter) AllowReserve(p peer.ID, src ma.Multiaddr) bool {
	a.logger.Debugf("AllowReserve from Peer ID: %s addr: %s", p, src)
	return true
}

// AllowConnect returns true if a source peer, with a given multiaddr is allowed to connect
// to a destination peer.
func (a *ACLFilter) AllowConnect(src peer.ID, srcAddr ma.Multiaddr, dst peer.ID) bool {
	a.logger.Debugf("AllowConnect from Peer ID %s to ID %s addr: %s", src, dst, srcAddr)
	return true
}
