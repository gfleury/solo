package utils

import (
	"strings"

	"github.com/multiformats/go-multiaddr"
)

func DefaultAddrsFactory(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
	goodAddrs := []multiaddr.Multiaddr{}

	for i := range addrs {
		if strings.Contains(addrs[i].String(), "127.0.0.1") ||
			strings.Contains(addrs[i].String(), "/::1/") {
			continue
		}
		goodAddrs = append(goodAddrs, addrs[i])
	}
	return goodAddrs
}
