package water

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func ioctl(fd uintptr, request uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

func setupFd(config Config, fd uintptr) (name string, err error) {
	var flags uint16 = syscall.IFF_NO_PI
	if config.DeviceType == TUN {
		flags |= syscall.IFF_TUN
	} else {
		flags |= syscall.IFF_TAP
	}
	if config.PlatformSpecificParams.MultiQueue {
		flags |= 0x0100
	}

	if name, err = createInterface(fd, config.Name, flags); err != nil {
		return "", err
	}

	if err = setDeviceOptions(fd, config); err != nil {
		return "", err
	}

	return name, nil
}

func createInterface(fd uintptr, ifName string, flags uint16) (createdIFName string, err error) {
	var req ifReq
	req.Flags = flags
	copy(req.Name[:], ifName)

	err = ioctl(fd, syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req)))
	if err != nil {
		return
	}

	createdIFName = strings.Trim(string(req.Name[:]), "\x00")
	return
}

func setDeviceOptions(fd uintptr, config Config) (err error) {
	if config.Permissions != nil {
		if err = ioctl(fd, syscall.TUNSETOWNER, uintptr(config.Permissions.Owner)); err != nil {
			return
		}
		if err = ioctl(fd, syscall.TUNSETGROUP, uintptr(config.Permissions.Group)); err != nil {
			return
		}
	}

	// Set additional options to make IO faster
	if err = ioctl(fd, syscall.TUNSETNOCSUM, 1); err != nil {
		return err
	}

	// set clear the persist flag
	value := 0
	if config.Persist {
		value = 1
	}
	return ioctl(fd, syscall.TUNSETPERSIST, uintptr(value))
}
