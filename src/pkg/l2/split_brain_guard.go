package l2

import (
	"fmt"
	"sync"
	"time"
)

// PartitionState representa el estado de aislamiento de red del nodo
type PartitionState string

const (
	StateHealthy     PartitionState = "HEALTHY"
	StatePartitioned PartitionState = "PARTITIONED" // En riesgo de split-brain
	StateRecovering  PartitionState = "RECOVERING"
)

// PeerLiveness registro de actividad física de un par
type PeerLiveness struct {
	PeerDID    string    `json:"peer_did"`
	LastSeen   time.Time `json:"last_seen"`
	IsActive   bool      `json:"is_active"`
}

// SplitBrainGuard protege al nodo de divergir durante particiones de red L2
type SplitBrainGuard struct {
	mu                  sync.RWMutex
	peers               map[string]*PeerLiveness
	state               PartitionState
	heartbeatTimeout    time.Duration
	partitionThreshold float64 // p.ej. 0.50 (50% de pares mínimos activos)
	stateChangedAt      time.Time
	recoveryWindow      time.Duration
}

// NewSplitBrainGuard inicializa el guardián de particiones
func NewSplitBrainGuard(timeout time.Duration, threshold float64) *SplitBrainGuard {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if threshold <= 0 {
		threshold = 0.50
	}
	return &SplitBrainGuard{
		peers:               make(map[string]*PeerLiveness),
		state:               StateHealthy,
		heartbeatTimeout:    timeout,
		partitionThreshold: threshold,
		stateChangedAt:      time.Now(),
		recoveryWindow:      5 * time.Second,
	}
}

// RecordHeartbeat registra un latido vivo recibido de un par remoto
func (g *SplitBrainGuard) RecordHeartbeat(peerDID string) {
	g.RecordHeartbeatAt(peerDID, time.Now())
}

// RecordHeartbeatAt registra un latido vivo con marca temporal explícita
func (g *SplitBrainGuard) RecordHeartbeatAt(peerDID string, at time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()

	p, exists := g.peers[peerDID]
	if !exists {
		p = &PeerLiveness{PeerDID: peerDID}
		g.peers[peerDID] = p
	}
	p.LastSeen = at
	p.IsActive = true
}

// RegisterPeer añade formalmente un par al censo de la malla
func (g *SplitBrainGuard) RegisterPeer(peerDID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.peers[peerDID]; !exists {
		g.peers[peerDID] = &PeerLiveness{
			PeerDID:  peerDID,
			LastSeen: time.Time{},
			IsActive: false,
		}
	}
}

// AuditLiveness evalúa la salud del enjambre y previene split-brain
func (g *SplitBrainGuard) AuditLiveness(now time.Time) (PartitionState, string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	total := len(g.peers)
	if total == 0 {
		// Modo nodo solitario/bootstrap
		return g.state, "Standalone mode (0 registered peers)"
	}

	activeCount := 0
	for _, p := range g.peers {
		if now.Sub(p.LastSeen) <= g.heartbeatTimeout {
			p.IsActive = true
			activeCount++
		} else {
			p.IsActive = false
		}
	}

	ratio := float64(activeCount) / float64(total)

	if ratio < g.partitionThreshold {
		if g.state != StatePartitioned {
			g.state = StatePartitioned
			g.stateChangedAt = now
		}
		return g.state, fmt.Sprintf("Partition detected: %d/%d peers active (%.1f%% < %.1f%%)",
			activeCount, total, ratio*100, g.partitionThreshold*100)
	}

	// Recuperación o estado saludable
	if g.state == StatePartitioned {
		g.state = StateRecovering
		g.stateChangedAt = now
		return g.state, fmt.Sprintf("Recovering partition: quorum restored (%d/%d active)", activeCount, total)
	}

	if g.state == StateRecovering {
		if now.Sub(g.stateChangedAt) >= g.recoveryWindow {
			g.state = StateHealthy
			g.stateChangedAt = now
			return g.state, "Recovered to HEALTHY state after stability window"
		}
		return g.state, fmt.Sprintf("Stabilizing recovery (remaining %v)", g.recoveryWindow-now.Sub(g.stateChangedAt))
	}

	return g.state, fmt.Sprintf("Mesh healthy: %d/%d peers active (%.1f%%)", activeCount, total, ratio*100)
}

// IsSafeToWrite indica si el nodo tiene quórum suficiente para aplicar mutaciones de red
func (g *SplitBrainGuard) IsSafeToWrite() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state == StateHealthy
}

// GetStatus retorna el snapshot del estado del guardián
func (g *SplitBrainGuard) GetStatus() (PartitionState, int, int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	active := 0
	for _, p := range g.peers {
		if p.IsActive {
			active++
		}
	}
	return g.state, active, len(g.peers)
}
