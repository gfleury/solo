//go:build !windows && !freebsd
// +build !windows,!freebsd

package vpn

import (
	"bytes"
	"time"

	"github.com/gfleury/solo/client/netlink"
	"github.com/mudler/water"
)

type memoryBuffer struct {
	read  *bytes.Buffer
	write *bytes.Buffer
}

func (m *memoryBuffer) Read(b []byte) (int, error) {
	for m.read.Len() < 1 {
		// Block all empty reads
		time.Sleep(time.Second)
	}
	return m.read.Read(b)
}

func (m *memoryBuffer) Write(b []byte) (int, error) {
	return m.write.Write(b)
}

func (memoryBuffer) Close() error { return nil }

func newMemoryBuffer() *memoryBuffer {
	return &memoryBuffer{bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})}
}

func (i *VPNInterface) createInterface(createInterface bool) error {
	var err error
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = i.config.InterfaceName

	if createInterface {
		i.networkInterface, err = water.New(config)
	} else {
		i.networkInterface = &water.Interface{ReadWriteCloser: newMemoryBuffer()}
	}
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
