//go:build darwin

package l1

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"ipvn7/pkg/l0"
)

// NativeTunAdapter en macOS utun (requiere privilegios root / sandbox network entitlement)
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
	// En macOS, la apertura de /dev/net/tun o utun requiere privilegios root de kernel.
	// Si no se cuenta con ellos, se retorna error para activar el UserspaceVirtualAdapter de cero fricción.
	return nil, errors.New("macOS utun nativo requiere privilegios de superusuario o NetworkExtension entitlement: conmutando a modo espacio de usuario")
}

func (a *NativeTunAdapter) ReadPacket() ([]byte, error) {
	if a.closed.Load() {
		return nil, errors.New("adaptador TUN cerrado")
	}
	return nil, errors.New("no implementado en macOS sin driver kernel")
}

func (a *NativeTunAdapter) WritePacket(data []byte) error {
	if a.closed.Load() {
		return errors.New("adaptador TUN cerrado")
	}
	return nil
}

func (a *NativeTunAdapter) Close() error {
	a.closeOnce.Do(func() {
		a.closed.Store(true)
		if a.file != nil {
			_ = a.file.Close()
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
