package water

import (
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/tun"
)

type wintun struct {
	dev tun.Device
}

func (w *wintun) Close() error {
	return w.dev.Close()
}

func (w *wintun) Write(b []byte) (int, error) {
	return w.dev.Write(b, 0)
}

func (w *wintun) Read(b []byte) (int, error) {
	return w.dev.Read(b, 0)
}

func openDev(config Config) (ifce *Interface, err error) {
	if config.DeviceType == TAP {
		return nil, err
	}
	id := &windows.GUID{
		0x0000000,
		0xFFFF,
		0xFFFF,
		[8]byte{0xFF, 0xe9, 0x76, 0xe5, 0x8c, 0x74, 0x06, 0x3e},
	}
	tun.WintunTunnelType = "DtlsLink"
	dev, err := tun.CreateTUNWithRequestedGUID(config.PlatformSpecificParams.Name, id, 0)
	if err != nil {
		return nil, err
	}
	wintun := &wintun{dev: dev}
	ifce = &Interface{isTAP: false, ReadWriteCloser: wintun, name: config.PlatformSpecificParams.Name}
	return ifce, nil
}
