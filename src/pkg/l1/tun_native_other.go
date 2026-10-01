//go:build !windows && !linux && !darwin

package l1

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"ipvn7/pkg/l0"
)

type NativeTunAdapter struct {
	name      string
	ipv4      net.IP
	ipv6      net.IP
	mtu       int
	closed    atomic.Bool
	closeOnce sync.Once
}

func createNativeTun(id *l0.Identity, devName string) (TunAdapter, error) {
	return nil, errors.New("adaptador TUN de kernel no soportado en esta plataforma: fallback automático a modo espacio de usuario")
}

func (a *NativeTunAdapter) ReadPacket() ([]byte, error) {
	return nil, errors.New("plataforma no soportada")
}

func (a *NativeTunAdapter) WritePacket(data []byte) error {
	return errors.New("plataforma no soportada")
}

func (a *NativeTunAdapter) Close() error {
	a.closeOnce.Do(func() {
		a.closed.Store(true)
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
