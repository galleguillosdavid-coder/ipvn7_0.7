// Package l1 implementa estructuras canónicas y serialización para pasarelas de salida soberanas.
package l1

import (
	"errors"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// Modos de selección de puerta de enlace predeterminada
const (
	EgressModeAuto     = 0 // Selección automática del gateway de menor latencia y pérdida
	EgressModeSpecific = 1 // Salida fijada a un DID específico de la malla
	EgressModeDirect   = 2 // Salida directa exclusiva por el ISP local (sin túnel)
)

// Constantes de estabilidad y dimensionamiento (Axioma III y DEC-130)
const (
	DefaultEgressMSS         = 1220 // MSS Clamping para tramas fijas de 1280B (1280 - 60)
	EgressHysteresisMargin   = 0.25 // 25% de mejora requerida para conmutar pasarela
	EgressMinObservationTime = 3 * time.Second
)

// EgressCapability describe la capacidad de un nodo para actuar como pasarela hacia el Internet público
type EgressCapability struct {
	CanExit        bool      `cbor:"1,keyasint" json:"can_exit"`
	BandwidthMbps  uint32    `cbor:"2,keyasint" json:"bandwidth_mbps"`
	ActiveSessions uint16    `cbor:"3,keyasint" json:"active_sessions"`
	CountryCode    string    `cbor:"4,keyasint,omitempty" json:"country_code,omitempty"`
	UpdatedAt      time.Time `cbor:"5,keyasint" json:"updated_at"`
}

// GatewayScore consolida la evaluación física de un gateway candidato
type GatewayScore struct {
	DID         string    `json:"did"`
	Score       float64   `json:"score"` // Menor es mejor
	RTTMs       float64   `json:"rtt_ms"`
	LossRate    float64   `json:"loss_rate"`
	LoadRatio   float64   `json:"load_ratio"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// CalculateGatewayScore calcula la métrica ponderada de calidad de pasarela
// Métrica: RTT (40%) + Pérdida (40%) + Carga (20%). Menor puntuación = Mejor pasarela.
func CalculateGatewayScore(rttMs, lossRate float64, activeSessions uint16) float64 {
	if rttMs < 0.1 {
		rttMs = 0.1
	}
	// Factor de saturación: 0 a 100 basado en sesiones activas (escala de 50 sesiones max)
	loadFactor := float64(activeSessions) * 2.0
	if loadFactor > 100.0 {
		loadFactor = 100.0
	}
	// Pérdida en escala porcentual 0-100
	lossFactor := lossRate * 100.0
	if lossFactor > 100.0 {
		lossFactor = 100.0
	}

	return (rttMs * 0.40) + (lossFactor * 0.40) + (loadFactor * 0.20)
}

// EncodeEgressCapability serializa la capacidad en CBOR determinista RFC 8949
func EncodeEgressCapability(cap *EgressCapability) ([]byte, error) {
	if cap == nil {
		return nil, errors.New("capacidad nula")
	}
	opts := cbor.CanonicalEncOptions()
	em, err := opts.EncMode()
	if err != nil {
		return nil, err
	}
	return em.Marshal(cap)
}

// DecodeEgressCapability deserializa la capacidad desde CBOR determinista
func DecodeEgressCapability(data []byte) (*EgressCapability, error) {
	if len(data) == 0 {
		return nil, errors.New("datos vacíos")
	}
	var cap EgressCapability
	if err := cbor.Unmarshal(data, &cap); err != nil {
		return nil, err
	}
	return &cap, nil
}

// EgressRegistry almacena en memoria atómica las capacidades anunciadas por los pares
type EgressRegistry struct {
	mu           sync.RWMutex
	capabilities map[string]*EgressCapability
}

// NewEgressRegistry inicializa el registro en memoria
func NewEgressRegistry() *EgressRegistry {
	return &EgressRegistry{
		capabilities: make(map[string]*EgressCapability),
	}
}

// Set registra o actualiza la capacidad de un par
func (r *EgressRegistry) Set(did string, cap *EgressCapability) {
	if did == "" || cap == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.capabilities[did] = cap
}

// Get obtiene la capacidad de un par si existe
func (r *EgressRegistry) Get(did string) (*EgressCapability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cap, ok := r.capabilities[did]
	return cap, ok
}

// Remove elimina un par del registro
func (r *EgressRegistry) Remove(did string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.capabilities, did)
}

// ListGateways retorna todos los DIDs que tienen CanExit == true
func (r *EgressRegistry) ListGateways() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gateways := make([]string, 0, len(r.capabilities))
	for did, cap := range r.capabilities {
		if cap.CanExit {
			gateways = append(gateways, did)
		}
	}
	return gateways
}
