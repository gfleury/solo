package utils

import (
	"net"
	"strings"
)

func FetchLocalRoutes(localIP string) ([]string, error) {
	networks := []string{}

	ifaces, err := net.Interfaces()
	if err != nil {
		return networks, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if v.IP.String() == localIP || strings.HasPrefix(v.String(), "127.") || strings.HasPrefix(v.String(), "::1/128") || v.IP.To4() == nil {
					continue
				}
				networks = append(networks, v.String())
			}

		}
	}

	return networks, nil
}
