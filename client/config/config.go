package config

import (
	"github.com/gfleury/solo/client/discovery"
)

type Config struct {
	Token                string
	InterfaceAddress     string
	InterfaceName        string
	CreateInterface      bool
	Libp2pLogLevel       string
	LogLevel             string
	DiscoveryPeers       []string
	PublicDiscoveryPeers bool
	DiscoveryInterval    int
	InterfaceMTU         int
	MaxConnections       int
	HolePunch            bool
	RandomIdentity       bool
	RandomPort           bool
}

func Peers2List(peers []string) discovery.AddrList {
	addrsList := discovery.AddrList{}
	for _, p := range peers {
		err := addrsList.Set(p)
		if err != nil {
			panic(err)
		}
	}
	return addrsList
}
