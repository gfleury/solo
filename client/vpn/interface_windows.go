//go:build windows
// +build windows

/*
Copyright © 2021-2022 Ettore Di Giacinto <mudler@mocaccino.org>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package vpn

import (
	"bytes"
	"net/netip"
	"time"

	"golang.org/x/sys/windows"

	"github.com/mudler/water"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
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
	// find interface created by water
	guid, err := windows.GUIDFromString("{00000000-FFFF-FFFF-FFE9-76E58C74063E}")
	if err != nil {
		return err
	}
	luid, err := winipcfg.LUIDFromGUID(&guid)
	if err != nil {
		return err
	}

	prefix, err := netip.ParsePrefix(i.config.InterfaceAddress)
	if err != nil {
		return err
	}
	addresses := append([]netip.Prefix{}, prefix)
	if err := luid.SetIPAddresses(addresses); err != nil {
		return err
	}

	iface, err := luid.IPInterface(windows.AF_INET)
	if err != nil {
		return err
	}
	iface.NLMTU = uint32(i.config.InterfaceMTU)
	if err := iface.Set(); err != nil {
		return err
	}
	return nil
}
