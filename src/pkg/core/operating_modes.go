// Package core implementa el controlador formal de Modos Operativos Tri-Estado
// (MODE_OPEN, MODE_SAFE, MODE_DEGRADED) especificado en ip7uin_MVP_Spec_v1.1.docx (Sección 4.8)
// y en la Arquitectura Natural de IPv8.
package core

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// NodeOperatingMode define los estados operacionales soberanos del nodo
type NodeOperatingMode string

const (
	// ModeOpen: Operación nominal completa. Descubrimiento adaptativo activo y admisión estándar.
	ModeOpen NodeOperatingMode = "MODE_OPEN"

	// ModeSafe: Aislamiento estricto (Dark Node). Balizas apagadas, solo admite pares anclados en disco.
	ModeSafe NodeOperatingMode = "MODE_SAFE"

	// ModeDegraded: Contingencia defensiva ante congestión (>30% loss). Cuotas rígidas y prioridad exclusiva a control.
	ModeDegraded NodeOperatingMode = "MODE_DEGRADED"
)

// ModeTransition registra la trazabilidad histórica de los cambios de estado operativo
type ModeTransition struct {
	FromMode  NodeOperatingMode `json:"from_mode"`
	ToMode    NodeOperatingMode `json:"to_mode"`
	Reason    string            `json:"reason"`
	Timestamp time.Time         `json:"timestamp"`
}

// OperatingModeController gobierna los modos del nodo con exclusión mutua
type OperatingModeController struct {
	mu          sync.RWMutex
	currentMode NodeOperatingMode
	history     []ModeTransition
}

// NewOperatingModeController inicializa el controlador en modo abierto nominal
func NewOperatingModeController(initialMode NodeOperatingMode) *OperatingModeController {
	if initialMode == "" {
		initialMode = ModeOpen
	}
	return &OperatingModeController{
		currentMode: initialMode,
		history:     make([]ModeTransition, 0, 32),
	}
}

// SetMode cambia el modo operativo registrando la causa y la marca temporal
func (c *OperatingModeController) SetMode(newMode NodeOperatingMode, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch newMode {
	case ModeOpen, ModeSafe, ModeDegraded:
	default:
		return fmt.Errorf("modo operativo invalido: %s", newMode)
	}

	if c.currentMode == newMode {
		return nil // No-op idempotente
	}

	transition := ModeTransition{
		FromMode:  c.currentMode,
		ToMode:    newMode,
		Reason:    reason,
		Timestamp: time.Now().UTC(),
	}

	// Mantener histórico acotado a 32 transiciones (O(1) memoria)
	if len(c.history) >= 32 {
		c.history = c.history[1:]
	}
	c.history = append(c.history, transition)
	c.currentMode = newMode
	return nil
}

// CurrentMode devuelve el modo operativo activo
func (c *OperatingModeController) CurrentMode() NodeOperatingMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMode
}

// AllowsDiscovery evalúa si el modo permite emitir o procesar balizas de descubrimiento
func (c *OperatingModeController) AllowsDiscovery() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMode == ModeOpen
}

// AllowsUnpinnedPeers indica si se permite aceptar conexiones de pares no registrados en disco
func (c *OperatingModeController) AllowsUnpinnedPeers() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMode != ModeSafe
}

// IsDegraded indica si el nodo está operando bajo restricciones de contingencia
func (c *OperatingModeController) IsDegraded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMode == ModeDegraded
}

// History devuelve una copia defensiva del histórico de transiciones
func (c *OperatingModeController) History() []ModeTransition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]ModeTransition, len(c.history))
	copy(res, c.history)
	return res
}

// ValidateModeString valida y convierte una cadena al tipo formal
func ValidateModeString(m string) (NodeOperatingMode, error) {
	switch m {
	case string(ModeOpen), "open", "OPEN":
		return ModeOpen, nil
	case string(ModeSafe), "safe", "SAFE":
		return ModeSafe, nil
	case string(ModeDegraded), "degraded", "DEGRADED":
		return ModeDegraded, nil
	default:
		return "", errors.New("modo desconocido (opciones validas: MODE_OPEN, MODE_SAFE, MODE_DEGRADED)")
	}
}
