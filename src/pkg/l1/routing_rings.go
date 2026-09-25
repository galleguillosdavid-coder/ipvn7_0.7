package l1

import (
	"encoding/binary"
	"math/bits"
)

// RouterProfile define el perfil topológico del nodo según su hardware y función de red
type RouterProfile string

const (
	ProfileMicro    RouterProfile = "micro"    // IoT / Móvil / Edge (8 anillos, 64 pares)
	ProfileStandard RouterProfile = "standard" // Desktop / Laptop / Servidor estándar (16 anillos, 256 pares)
	ProfileBackbone RouterProfile = "backbone" // Gateway / Anchor / Servidor dedicado (32 anillos, 1024 pares)
	ProfileCustom   RouterProfile = "custom"   // Parámetros N-anillos configurables ad-hoc
)

// Constantes canónicas y valores por defecto (compatibilidad hacia atrás)
const (
	DefaultNumRings     = 16
	DefaultPeersPerRing = 16
	DefaultMaxPeers     = DefaultNumRings * DefaultPeersPerRing // 256 pares por defecto
	NumRings            = DefaultNumRings                       // Retrocompatibilidad con código existente
	PeersPerRing        = DefaultPeersPerRing
	MaxPeers            = DefaultMaxPeers
)

// RouterConfig define el dimensionamiento elástico y las heurísticas de enrutamiento
type RouterConfig struct {
	Profile            RouterProfile `json:"profile"`
	NumRings           int           `json:"num_rings"`             // N anillos logarítmicos (4 a 64)
	PeersPerRing       int           `json:"peers_per_ring"`        // Capacidad acotada por anillo
	MaxTotalPeers      int           `json:"max_total_peers"`       // Capacidad total de memoria
	AlphaLatencyWeight float64       `json:"alpha_latency_weight"`  // Ponderación de latencia RTT (0.0=XOR puro, 0.35=híbrido 2D)
	AutoRebalance      bool          `json:"auto_rebalance"`        // Prospección periódica de anillos vacíos
}

// DefaultRouterConfig retorna la configuración estándar optimizada (16 anillos, 256 pares)
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		Profile:            ProfileStandard,
		NumRings:           DefaultNumRings,
		PeersPerRing:       DefaultPeersPerRing,
		MaxTotalPeers:      DefaultMaxPeers,
		AlphaLatencyWeight: 0.35,
		AutoRebalance:      true,
	}
}

// MicroRouterConfig retorna la configuración liviana para IoT o dispositivos móviles (8 anillos, 64 pares)
func MicroRouterConfig() RouterConfig {
	return RouterConfig{
		Profile:            ProfileMicro,
		NumRings:           8,
		PeersPerRing:       8,
		MaxTotalPeers:      64,
		AlphaLatencyWeight: 0.25,
		AutoRebalance:      false,
	}
}

// BackboneRouterConfig retorna la configuración de alto rendimiento para Gateways y Anclas (32 anillos, 1024 pares)
func BackboneRouterConfig() RouterConfig {
	return RouterConfig{
		Profile:            ProfileBackbone,
		NumRings:           32,
		PeersPerRing:       32,
		MaxTotalPeers:      1024,
		AlphaLatencyWeight: 0.40,
		AutoRebalance:      true,
	}
}

// XORKeyDistanceN calcula la distancia XOR de 256 bits y determina el anillo logarítmico (0 a numRings-1)
func XORKeyDistanceN(keyA, keyB []byte, numRings int) int {
	if numRings <= 0 {
		numRings = DefaultNumRings
	}
	if len(keyA) != 32 || len(keyB) != 32 {
		return numRings - 1
	}

	// Contar ceros iniciales en XOR (Leading Zeros)
	lz := 0
	for i := 0; i < 4; i++ {
		chunkA := binary.BigEndian.Uint64(keyA[i*8 : (i+1)*8])
		chunkB := binary.BigEndian.Uint64(keyB[i*8 : (i+1)*8])
		xor := chunkA ^ chunkB
		if xor == 0 {
			lz += 64
		} else {
			lz += bits.LeadingZeros64(xor)
			break
		}
	}

	// Mapear los 256 bits a numRings anillos logarítmicos
	// Mayor cantidad de ceros iniciales = nodos más cercanos = Anillo 0
	bitsPerRing := 256 / numRings
	if bitsPerRing <= 0 {
		bitsPerRing = 1
	}
	ring := (256 - lz) / bitsPerRing
	if ring >= numRings {
		ring = numRings - 1
	}
	if ring < 0 {
		ring = 0
	}
	return ring
}

// XORKeyDistance calcula la distancia logarítmica usando la cantidad de anillos por defecto
func XORKeyDistance(keyA, keyB []byte) int {
	return XORKeyDistanceN(keyA, keyB, DefaultNumRings)
}

// GetRingDistribution entrega la cantidad de pares alojados en cada uno de los N anillos
func (r *KleinbergRouter) GetRingDistribution() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dist := make([]int, len(r.rings))
	for i, ringPeers := range r.rings {
		dist[i] = len(ringPeers)
	}
	return dist
}

// GetRingPeers entrega los pares pertenecientes a un anillo específico
func (r *KleinbergRouter) GetRingPeers(ringIndex int) []*PeerNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ringIndex < 0 || ringIndex >= len(r.rings) {
		return nil
	}

	res := make([]*PeerNode, len(r.rings[ringIndex]))
	copy(res, r.rings[ringIndex])
	return res
}

// RebalanceRings detecta anillos despoblados (huecos topológicos) para prospección proactiva
func (r *KleinbergRouter) RebalanceRings() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	starvedRings := make([]int, 0)
	for i, ringPeers := range r.rings {
		if len(ringPeers) == 0 {
			starvedRings = append(starvedRings, i)
		}
	}
	return starvedRings
}
