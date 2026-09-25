// Package l1 implementa el descubrimiento de MTU de ruta (Path MTU Discovery) adaptativo
// y ajuste dinámico de tamaño de datagramas para la malla ipvn7 v0.7.
package l1

import (
	"sync"
	"time"
)

// Constantes canónicas de escalones de MTU en ipvn7
const (
	MTUTierLoRa     = 256  // Módulos de radio de baja tasa (SX1276 / Ebyte)
	MTUTierConstrain = 512  // Enlaces celulares 2G/3G degradados o BLE
	MTUTierCanonical = 1280 // Estándar inmutable ipvn7 y mínimo IPv6
	MTUTierLAN       = 1420 // Enlaces Ethernet / Wi-Fi directos de alta capacidad
)

var canonicalMTUTiers = []int{MTUTierLAN, MTUTierCanonical, MTUTierConstrain, MTUTierLoRa}

// PeerMTUState almacena el estado de MTU adaptativo para un par específico
type PeerMTUState struct {
	EffectiveMTU   int
	TierIndex      int
	ConsecutiveLoss int
	LastProbe      time.Time
	LastSuccess    time.Time
}

// AdaptiveMTUManager orquesta la adaptación dinámica del tamaño de paquete por par
type AdaptiveMTUManager struct {
	mu    sync.RWMutex
	peers map[string]*PeerMTUState
}

// NewAdaptiveMTUManager inicializa el gestor de MTU adaptativo
func NewAdaptiveMTUManager() *AdaptiveMTUManager {
	return &AdaptiveMTUManager{
		peers: make(map[string]*PeerMTUState),
	}
}

// GetEffectiveMTU retorna el MTU seguro actual para un par dado (por defecto 1280B)
func (m *AdaptiveMTUManager) GetEffectiveMTU(peerDID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.peers[peerDID]
	if !exists || state == nil {
		return MTUTierCanonical
	}
	return state.EffectiveMTU
}

// OnPacketLoss reporta pérdida de datagramas consecutiva, degradando el MTU al siguiente escalón
func (m *AdaptiveMTUManager) OnPacketLoss(peerDID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.getOrCreatePeer(peerDID)
	state.ConsecutiveLoss++

	// Si se observan 3 o más pérdidas consecutivas, degradar al siguiente escalón
	if state.ConsecutiveLoss >= 3 && state.TierIndex < len(canonicalMTUTiers)-1 {
		state.TierIndex++
		state.EffectiveMTU = canonicalMTUTiers[state.TierIndex]
		state.ConsecutiveLoss = 0
	}

	return state.EffectiveMTU
}

// OnSuccessProbe confirma la entrega exitosa de una sonda o datagrama de tamaño dado
func (m *AdaptiveMTUManager) OnSuccessProbe(peerDID string, packetSize int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.getOrCreatePeer(peerDID)
	state.ConsecutiveLoss = 0
	state.LastSuccess = time.Now()

	// Si el tamaño exitoso es mayor que el efectivo actual, actualizarlo
	if packetSize > state.EffectiveMTU {
		for i, tier := range canonicalMTUTiers {
			if packetSize >= tier {
				state.TierIndex = i
				state.EffectiveMTU = tier
				break
			}
		}
	}
}

// Reset restaura el MTU de un par al valor canónico estándar (1280B)
func (m *AdaptiveMTUManager) Reset(peerDID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if state, exists := m.peers[peerDID]; exists {
		state.TierIndex = 1 // MTUTierCanonical
		state.EffectiveMTU = MTUTierCanonical
		state.ConsecutiveLoss = 0
	}
}

func (m *AdaptiveMTUManager) getOrCreatePeer(peerDID string) *PeerMTUState {
	state, exists := m.peers[peerDID]
	if !exists {
		state = &PeerMTUState{
			EffectiveMTU:   MTUTierCanonical,
			TierIndex:      1, // Índice de MTUTierCanonical en canonicalMTUTiers
			LastProbe:      time.Now(),
			LastSuccess:    time.Now(),
		}
		m.peers[peerDID] = state
	}
	return state
}
