package l0

import (
	"errors"
	"fmt"
)

var (
	ErrXDPUnsupportedPlatform = errors.New("xdp: plataforma no soporta eBPF/XDP de forma nativa (requiere Linux kernel >= 4.18)")
	ErrXDPInterfaceNotFound   = errors.New("xdp: interfaz de red física no encontrada")
	ErrXDPAttachFailed        = errors.New("xdp: fallo al adjuntar programa XDP en la interfaz")
)

// XDPAttachMode define el modo de enlace en el subsistema de red del kernel
type XDPAttachMode int

const (
	XDPModeNative  XDPAttachMode = 1 // En la tarjeta física de red (NIC driver)
	XDPModeGeneric XDPAttachMode = 2 // Genérico en el stack SKB del kernel
	XDPModeOffload XDPAttachMode = 3 // Directo en el hardware SmartNIC / FPGA
)

// XDPConfig encapsula las opciones para adjuntar el filtro de kernel
type XDPConfig struct {
	InterfaceName string
	ObjectPath    string
	Mode          XDPAttachMode
}

// XDPManager define la interfaz operativa para el ciclo de vida del filtro de conmutación
type XDPManager interface {
	Attach(cfg XDPConfig) error
	Detach(iface string) error
	GetStats() (*XDPFilterStats, error)
	IsAttached(iface string) bool
}

// FormatAttachMode retorna la representación en cadena del modo XDP
func (m XDPAttachMode) String() string {
	switch m {
	case XDPModeNative:
		return "xdp-native (NIC driver)"
	case XDPModeGeneric:
		return "xdp-generic (SKB fallback)"
	case XDPModeOffload:
		return "xdp-offload (Hardware SmartNIC)"
	default:
		return fmt.Sprintf("xdp-unknown (%d)", m)
	}
}
