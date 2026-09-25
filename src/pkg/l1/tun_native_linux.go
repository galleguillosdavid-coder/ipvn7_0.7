//go:build linux

package l1

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
	"ipvn7/pkg/l0"
)

type NativeTunAdapter struct {
	file      *os.File
	name      string
	ipv4      net.IP
	ipv6      net.IP
	mtu       int
	closed    atomic.Bool
	closeOnce sync.Once
}

func createNativeTun(id *l0.Identity, devName string) (TunAdapter, error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("error abriendo /dev/net/tun (requiere CAP_NET_ADMIN): %w", err)
	}

	var ifr struct {
		name  [16]byte
		flags uint16
		_     [22]byte
	}
	copy(ifr.name[:], []byte(devName))
	ifr.flags = unix.IFF_TUN | unix.IFF_NO_PI

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		file.Fd(),
		uintptr(unix.TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr)),
	)
	if errno != 0 {
		file.Close()
		return nil, fmt.Errorf("ioctl TUNSETIFF falló: %v", errno)
	}

	return &NativeTunAdapter{
		file: file,
		name: devName,
		ipv4: id.IPv4(),
		ipv6: id.IPv6(),
		mtu:  DefaultMTU,
	}, nil
}

func (a *NativeTunAdapter) ReadPacket() ([]byte, error) {
	if a.closed.Load() {
		return nil, errors.New("adaptador TUN cerrado")
	}
	buf := make([]byte, a.mtu)
	n, err := a.file.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (a *NativeTunAdapter) WritePacket(data []byte) error {
	if a.closed.Load() {
		return errors.New("adaptador TUN cerrado")
	}
	_, err := a.file.Write(data)
	return err
}

func (a *NativeTunAdapter) Close() error {
	a.closeOnce.Do(func() {
		a.closed.Store(true)
		if a.file != nil {
			a.file.Close()
		}
	})
	return nil
}

func (a *NativeTunAdapter) MTU() int {
	return a.mtu
}

func (a *NativeTunAdapter) IPv4() net.IP {
	return a.ipv4
}

func (a *NativeTunAdapter) IPv6() net.IP {
	return a.ipv6
}

func (a *NativeTunAdapter) Mode() AdapterMode {
	return ModeKernelVirtual
}

func (a *NativeTunAdapter) IsUserspace() bool {
	return false
}
