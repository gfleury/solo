//go:build !windows && !freebsd
// +build !windows,!freebsd

package vpn

import (
	"github.com/gfleury/solo/client/netlink"
	"github.com/mudler/water"
)

func (i *VPNInterface) createInterface() error {
	var err error
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = i.config.InterfaceName

	i.networkInterface, err = water.New(config)
	return err
}

func (i *VPNInterface) prepareInterface() error {

	link, err := netlink.LinkByName(i.config.InterfaceName)
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(i.config.InterfaceAddress)
	if err != nil {
		return err
	}

	err = netlink.LinkSetMTU(link, i.config.InterfaceMTU)
	if err != nil {
		return err
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		return err
	}

	return nil
}
