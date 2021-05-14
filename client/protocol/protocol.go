package protocol

import (
	p2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ALLEIN         Protocol = "/allein/0.1"
	BROADCAST      Protocol = "/broadcast/0.1"
	NOISEHANDSHAKE Protocol = "/noisehandshake/0.1"
)

type Protocol string

func (p Protocol) ID() p2pprotocol.ID {
	return p2pprotocol.ID(string(p))
}
