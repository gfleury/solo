package config

import (
	"github.com/gfleury/solo/client/discovery"
)

type Config struct {
	Token             string
	InterfaceAddress  string
	InterfaceName     string
	CreateInterface   bool
	Libp2pLogLevel    string
	LogLevel          string
	DiscoveryPeers    []string
	DiscoveryInterval int
	InterfaceMTU      int
	MaxConnections    int
	HolePunch         bool
	NatMap            bool
	NatService        bool
	RandomIdentity    bool
}

func Peers2List(peers []string) discovery.AddrList {
	addrsList := discovery.AddrList{}
	for _, p := range peers {
		addrsList.Set(p)
	}
	return addrsList
}
