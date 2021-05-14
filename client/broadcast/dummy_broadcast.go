package broadcast

import (
	"context"

	"github.com/gfleury/solo/client/broadcast/metapacket"
	"github.com/gfleury/solo/client/types"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DummyBroadcast struct {
	table map[string]peer.ID
}

func NewDummyBroadcast() *DummyBroadcast {
	return &DummyBroadcast{
		table: make(map[string]peer.ID),
	}
}

func (b *DummyBroadcast) Lookup(dstIP string) (*types.Machine, bool, bool) {
	if peer, found := b.table[dstIP]; found {
		return &types.Machine{PeerID: peer.String(), IP: dstIP, OS: "linux"}, found, false
	}
	return nil, false, false
}

func (b *DummyBroadcast) Start(ctx context.Context, host host.Host, s string) error {
	return nil
}

func (b *DummyBroadcast) SendPacket(ctx context.Context, packet *metapacket.MetaPacket) error {
	return nil
}

func (b *DummyBroadcast) PRPRequest(ctx context.Context, unknownDstIP string) error {
	return nil
}

func (b DummyBroadcast) AddFakePeer(ip string, peer peer.ID) {
	b.table[ip] = peer
}

func (b DummyBroadcast) AnnounceMyself(ctx context.Context) error {
	return nil
}
