package l1

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"ipvn7/pkg/l0"
)

const (
	DefaultMTU = l0.MaxPacketSize // 1280 bytes determinista
)

// AdapterMode define el modo de ejecución del adaptador de red virtual
type AdapterMode string

const (
	ModeKernelVirtual    AdapterMode = "KERNEL_NATIVE"
	ModeUserspaceVirtual AdapterMode = "USERSPACE_FASTPATH"
)

// TunAdapter define la interfaz abstracta del adaptador de red virtual
type TunAdapter interface {
	ReadPacket() ([]byte, error)
	WritePacket(data []byte) error
	Close() error
	MTU() int
	IPv4() net.IP
	IPv6() net.IP
	Mode() AdapterMode
	IsUserspace() bool
}

// UserspaceVirtualAdapter proporciona una interfaz virtual en memoria de ultra-alta velocidad
// y cero privilegios con búferes sincronizados mediante canales Go.
type UserspaceVirtualAdapter struct {
	ipv4      net.IP
	ipv6      net.IP
	mtu       int
	inbound   chan []byte
	outbound  chan []byte
	closed    atomic.Bool
	closeOnce sync.Once
}

// NewUserspaceVirtualAdapter inicializa el adaptador virtual en memoria
func NewUserspaceVirtualAdapter(id *l0.Identity) *UserspaceVirtualAdapter {
	return &UserspaceVirtualAdapter{
		ipv4:     id.IPv4(),
		ipv6:     id.IPv6(),
		mtu:      DefaultMTU,
		inbound:  make(chan []byte, 1024),
		outbound: make(chan []byte, 1024),
	}
}


func (a *UserspaceVirtualAdapter) ReadPacket() ([]byte, error) {
	if a.closed.Load() {
		return nil, errors.New("adaptador TUN cerrado")
	}
	pkt, ok := <-a.outbound
	if !ok {
		return nil, errors.New("adaptador TUN cerrado")
	}
	return pkt, nil
}

func (a *UserspaceVirtualAdapter) WritePacket(data []byte) error {
	if a.closed.Load() {
		return errors.New("adaptador TUN cerrado")
	}
	if len(data) > a.mtu {
		return fmt.Errorf("paquete excede MTU de %d bytes (tamaño: %d)", a.mtu, len(data))
	}
	select {
	case a.inbound <- data:
		return nil
	default:
		// Fast-path drop si la cola está llena (Sustained flow)
		return errors.New("cola TUN saturada: paquete descartado")
	}
}

// InjectPacket inyecta un datagrama directamente en el flujo de recepción del adaptador
func (a *UserspaceVirtualAdapter) InjectPacket(data []byte) error {
	if a.closed.Load() {
		return errors.New("adaptador TUN cerrado")
	}
	select {
	case a.outbound <- data:
		return nil
	default:
		return errors.New("búfer de salida saturado")
	}
}

func (a *UserspaceVirtualAdapter) Close() error {
	a.closeOnce.Do(func() {
		a.closed.Store(true)
		close(a.inbound)
		close(a.outbound)
	})
	return nil
}

func (a *UserspaceVirtualAdapter) MTU() int {
	return a.mtu
}

func (a *UserspaceVirtualAdapter) IPv4() net.IP {
	return a.ipv4
}

func (a *UserspaceVirtualAdapter) IPv6() net.IP {
	return a.ipv6
}

func (a *UserspaceVirtualAdapter) Mode() AdapterMode {
	return ModeUserspaceVirtual
}

func (a *UserspaceVirtualAdapter) IsUserspace() bool {
	return true
}

// TunConfig parametriza la configuración del adaptador virtual
type TunConfig struct {
	DeviceName   string
	IPv4CIDR     string
	IPv6CIDR     string
	MTU          int
	PreferNative bool
}

// DefaultTunConfig genera la configuración canónica de ipvn7
func DefaultTunConfig(id *l0.Identity) TunConfig {
	return TunConfig{
		DeviceName:   "ipvn7-tun0",
		IPv4CIDR:     fmt.Sprintf("%s/16", id.IPv4().String()),
		IPv6CIDR:     fmt.Sprintf("%s/64", id.IPv6().String()),
		MTU:          DefaultMTU,
		PreferNative: true,
	}
}

// CreateTunAdapter instancia el adaptador apropiado con fallback automático seguro.
// Si preferNative es verdadero y el host cuenta con permisos (CAP_NET_ADMIN o root en Linux),
// se intentará abrir la interfaz de kernel. De lo contrario, se activa el adaptador virtual de userspace.
func CreateTunAdapter(id *l0.Identity, preferNative bool) (TunAdapter, error) {
	if preferNative {
		native, err := TryCreateNativeTun(id, "ipvn7-tun0")
		if err == nil && native != nil {
			return native, nil
		}
	}
	return NewUserspaceVirtualAdapter(id), nil
}

// TryCreateNativeTun intenta inicializar un TUN nativo del sistema operativo
func TryCreateNativeTun(id *l0.Identity, devName string) (TunAdapter, error) {
	return createNativeTun(id, devName)
}

// NewUserspaceTunAdapter alias canónico de alta velocidad para entornos no privilegiados
func NewUserspaceTunAdapter(id *l0.Identity) *UserspaceVirtualAdapter {
	return NewUserspaceVirtualAdapter(id)
}


