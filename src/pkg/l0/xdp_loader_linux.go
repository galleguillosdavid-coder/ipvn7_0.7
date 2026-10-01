//go:build linux

package l0

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
)

// LinuxXDPManager gestiona la vinculación de programas XDP con el kernel de Linux
type LinuxXDPManager struct {
	mu       sync.RWMutex
	attached map[string]XDPConfig
	stats    XDPFilterStats
}

// NewLinuxXDPManager inicializa el gestor nativo de eBPF para Linux
func NewLinuxXDPManager() *LinuxXDPManager {
	return &LinuxXDPManager{
		attached: make(map[string]XDPConfig),
	}
}

// DefaultXDPManager retorna el gestor eBPF de producción para la plataforma actual
func DefaultXDPManager() XDPManager {
	return NewLinuxXDPManager()
}

// Attach vincula el objeto ELF eBPF en la interfaz especificada
func (m *LinuxXDPManager) Attach(cfg XDPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Verificar existencia de interfaz física
	iface, err := net.InterfaceByName(cfg.InterfaceName)
	if err != nil || iface == nil {
		return fmt.Errorf("%w: %s", ErrXDPInterfaceNotFound, cfg.InterfaceName)
	}

	modeFlag := "xdp"
	switch cfg.Mode {
	case XDPModeGeneric:
		modeFlag = "xdpgeneric"
	case XDPModeOffload:
		modeFlag = "xdpoffload"
	default:
		modeFlag = "xdp"
	}

	// 2. Ejecutar vinculación física mediante netlink / iproute2
	cmd := exec.Command("ip", "link", "set", "dev", cfg.InterfaceName, modeFlag, "obj", cfg.ObjectPath, "sec", "xdp")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w en %s: %s (err: %v)", ErrXDPAttachFailed, cfg.InterfaceName, strings.TrimSpace(stderr.String()), err)
	}

	m.attached[cfg.InterfaceName] = cfg
	return nil
}

// Detach desvincula el programa XDP de la interfaz de red
func (m *LinuxXDPManager) Detach(ifaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := exec.Command("ip", "link", "set", "dev", ifaceName, "xdp", "off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fallo al desvincular XDP de %s: %s", ifaceName, strings.TrimSpace(stderr.String()))
	}

	delete(m.attached, ifaceName)
	return nil
}

// IsAttached consulta el estado real de la interfaz consultando los atributos del enlace
func (m *LinuxXDPManager) IsAttached(ifaceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out, err := exec.Command("ip", "link", "show", "dev", ifaceName).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "xdp")
}

// GetStats retorna las métricas agregadas de filtrado de tramas
func (m *LinuxXDPManager) GetStats() (*XDPFilterStats, error) {
	return &m.stats, nil
}
