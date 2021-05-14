//go:build darwin
// +build darwin

package netlink

import (
	"fmt"
	"net"
	"os/exec"
)

func LinkSetMTU(link Link, mtu int) error {
	if err := exec.Command("ifconfig", link.name, "mtu", fmt.Sprint(mtu)).Run(); err != nil {
		return err
	}
	return nil
}

func LinkByName(name string) (Link, error) {
	link := Link{name: name}
	return link, nil
}

func AddrAdd(link Link, addr *Addr) error {
	if err := exec.Command("ifconfig", link.name, "inet", addr.IP.String(), addr.IP.String(), "up").Run(); err != nil {
		return err
	}

	// route add  10.1.0.0/24 -iface utun4
	_, net, err := net.ParseCIDR(addr.String())
	if err != nil {
		return err
	}
	if err := exec.Command("route", "add", net.String(), "-iface", link.name).Run(); err != nil {
		return err
	}

	return nil
}
