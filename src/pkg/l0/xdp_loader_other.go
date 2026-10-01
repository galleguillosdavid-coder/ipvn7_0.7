//go:build !linux

package l0

import (
	"sync"
)

// FallbackXDPManager provee el gestor en sistemas operativos sin soporte de kernel Linux eBPF/XDP
type FallbackXDPManager struct {
	mu       sync.Mutex
	attached map[string]XDPConfig
	stats    XDPFilterStats
}

// NewFallbackXDPManager inicializa el gestor de compatibilidad para plataformas no-Linux
func NewFallbackXDPManager() *FallbackXDPManager {
	return &FallbackXDPManager{
		attached: make(map[string]XDPConfig),
	}
}

// DefaultXDPManager retorna el gestor eBPF de fallback para plataformas no-Linux
func DefaultXDPManager() XDPManager {
	return NewFallbackXDPManager()
}

// Attach devuelve ErrXDPUnsupportedPlatform en plataformas no-Linux
func (m *FallbackXDPManager) Attach(cfg XDPConfig) error {
	return ErrXDPUnsupportedPlatform
}

// Detach devuelve nil o limpia entradas locales
func (m *FallbackXDPManager) Detach(iface string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.attached, iface)
	return nil
}

// IsAttached retorna falso en plataformas sin soporte XDP de kernel
func (m *FallbackXDPManager) IsAttached(iface string) bool {
	return false
}

// GetStats retorna las estadísticas de filtrado
func (m *FallbackXDPManager) GetStats() (*XDPFilterStats, error) {
	return &m.stats, nil
}
