// Package devicebridge implementa el aislamiento estricto de radio de impacto (Blast-Radius Quarantine)
// garantizando que anomalías en periféricos IoT (Shadow DIDs) jamás aíslen el nodo raíz residencial.
package devicebridge

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrRootImmuneToIoTQuarantine = errors.New("blast_radius: el nodo raíz es inmune al aislamiento por anomalías IoT")
	ErrInvalidDID                = errors.New("blast_radius: DID inválido")
)

// QuarantineStatus representa el estado de contención de un periférico
type QuarantineStatus struct {
	DID            string    `json:"did"`
	Reason         string    `json:"reason"`
	QuarantinedAt  time.Time `json:"quarantined_at"`
	IsShadowLeaf   bool      `json:"is_shadow_leaf"`
	RootProtected  bool      `json:"root_protected"`
}

// BlastRadiusGuard gestiona el cortafuegos de contención para periféricos IoT
type BlastRadiusGuard struct {
	mu           sync.RWMutex
	quarantined  map[string]*QuarantineStatus
	rootDID      string
}

// NewBlastRadiusGuard inicializa el centinela de radio de impacto
func NewBlastRadiusGuard(rootDID string) *BlastRadiusGuard {
	return &BlastRadiusGuard{
		quarantined: make(map[string]*QuarantineStatus),
		rootDID:     rootDID,
	}
}

// IsShadowDevice verifica si el identificador pertenece a un periférico hoja IoT
func IsShadowDevice(did string) bool {
	return strings.HasPrefix(did, "did:ipvn7:shadow:") || strings.Contains(did, ":iot:")
}

// QuarantineDevice aísla exclusivamente un periférico IoT sin afectar al nodo residencial
func (g *BlastRadiusGuard) QuarantineDevice(targetDID, reason string) (*QuarantineStatus, error) {
	if targetDID == "" {
		return nil, ErrInvalidDID
	}

	// Invariante de seguridad: El nodo raíz jamás puede ser aislado por fallas de periféricos
	if targetDID == g.rootDID || !IsShadowDevice(targetDID) {
		return nil, ErrRootImmuneToIoTQuarantine
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	status := &QuarantineStatus{
		DID:           targetDID,
		Reason:        reason,
		QuarantinedAt: time.Now(),
		IsShadowLeaf:  true,
		RootProtected: true,
	}
	g.quarantined[targetDID] = status
	return status, nil
}

// IsQuarantined retorna si un dispositivo específico está bajo aislamiento
func (g *BlastRadiusGuard) IsQuarantined(did string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.quarantined[did]
	return exists
}

// ReleaseDevice levanta la cuarentena de un periférico
func (g *BlastRadiusGuard) ReleaseDevice(did string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.quarantined, did)
}

// EvaluateTelemetryAnomaly procesa telemetría corrupta o ruidosa de un dispositivo
func (g *BlastRadiusGuard) EvaluateTelemetryAnomaly(srcDID string, anomalyScore float64) (quarantined bool, rootUnaffected bool) {
	if !IsShadowDevice(srcDID) {
		// Los nodos de computación y usuarios no son penalizados por el evaluador IoT
		return false, true
	}

	if anomalyScore >= 0.85 {
		_, _ = g.QuarantineDevice(srcDID, "Telemetría corrupta detectada por centinela IoT")
		return true, true
	}

	return false, true
}

// TotalQuarantined retorna el conteo de dispositivos bajo aislamiento
func (g *BlastRadiusGuard) TotalQuarantined() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.quarantined)
}
