package l1

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"ipvn7/pkg/l0"
)

// Estados de la Máquina de Estados Finitos (FSM) de salud de pares
const (
	HealthStateHealthy     = 0 // Flujo óptimo y pacing respetado
	HealthStateDegraded    = 1 // Incremento de RTT o ligera varianza de Jitter
	HealthStateUnstable    = 2 // Pérdida moderada de paquetes; activa Make-Before-Break
	HealthStateUnreachable = 3 // Ausencia total de ACKs o pérdida crítica
	HealthStateQuarantined = 4 // Anomalía maliciosa o aislamiento de vecindario
)

// PeerLocator representa la dirección física efímera de un par
type PeerLocator struct {
	PhysicalAddr *net.UDPAddr
	LastSeen     time.Time
	LatencyMs    float64
}

// PeerNode representa un par en la red en malla
type PeerNode struct {
	DID             string
	PublicKey       ed25519.PublicKey
	Locator         PeerLocator
	RingIndex       int
	HealthState     int
	JitterMs        float64
	LossRate        float64
	HealthScore     float64
	QuarantineUntil time.Time
}

// KleinbergRouter implementa el enrutamiento geométrico elástico de Mundo Pequeño
type KleinbergRouter struct {
	mu          sync.RWMutex
	LocalDID    string
	LocalPub    ed25519.PublicKey
	config      RouterConfig
	rings       [][]*PeerNode
	peerIndex   map[string]*PeerNode
	TotalRoutes int
}

// NewKleinbergRouterWithConfig inicializa el enrutador con configuración elástica explícita
func NewKleinbergRouterWithConfig(localID *l0.Identity, cfg RouterConfig) *KleinbergRouter {
	if cfg.NumRings <= 0 {
		cfg.NumRings = DefaultNumRings
	}
	if cfg.PeersPerRing <= 0 {
		cfg.PeersPerRing = DefaultPeersPerRing
	}
	if cfg.MaxTotalPeers <= 0 {
		cfg.MaxTotalPeers = cfg.NumRings * cfg.PeersPerRing
	}
	if cfg.AlphaLatencyWeight < 0.0 || cfg.AlphaLatencyWeight > 1.0 {
		cfg.AlphaLatencyWeight = 0.35
	}

	rings := make([][]*PeerNode, cfg.NumRings)
	for i := 0; i < cfg.NumRings; i++ {
		rings[i] = make([]*PeerNode, 0, cfg.PeersPerRing)
	}

	return &KleinbergRouter{
		LocalDID:  localID.DID(),
		LocalPub:  localID.PublicKey,
		config:    cfg,
		rings:     rings,
		peerIndex: make(map[string]*PeerNode),
	}
}

// NewKleinbergRouter inicializa el enrutador con la configuración estándar (16 anillos, 256 pares)
func NewKleinbergRouter(localID *l0.Identity) *KleinbergRouter {
	return NewKleinbergRouterWithConfig(localID, DefaultRouterConfig())
}

// GetConfig retorna la configuración activa del enrutador
func (r *KleinbergRouter) GetConfig() RouterConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// AddOrUpdatePeer inserta o actualiza un par con política de desalojo inteligente (Smart Eviction)
func (r *KleinbergRouter) AddOrUpdatePeer(did string, addr *net.UDPAddr, latencyMs float64) error {
	if did == r.LocalDID || did == "" {
		return nil // Jamás registrarse a sí mismo como par externo
	}

	pub, err := l0.PublicKeyFromDID(did)
	if err != nil {
		return fmt.Errorf("did inválido: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()


	// 1. Si el par ya existe, actualizamos su localizador (IP Roaming sin caída)
	if existing, found := r.peerIndex[did]; found {
		existing.Locator.PhysicalAddr = addr
		existing.Locator.LastSeen = time.Now()
		existing.Locator.LatencyMs = latencyMs
		return nil
	}

	ring := XORKeyDistanceN(r.LocalPub, pub, r.config.NumRings)

	// 2. Si el anillo específico está lleno, aplicar política de desalojo local
	if len(r.rings[ring]) >= r.config.PeersPerRing {
		evictIdx := -1

		// Prioridad 1: Desalojar nodos en Cuarentena o Inaccesibles
		for i, p := range r.rings[ring] {
			if p.HealthState == HealthStateQuarantined || p.HealthState == HealthStateUnreachable {
				evictIdx = i
				break
			}
		}

		// Prioridad 2: Desalojar nodos inestables con degradación severa
		if evictIdx == -1 {
			for i, p := range r.rings[ring] {
				if p.HealthState == HealthStateUnstable {
					evictIdx = i
					break
				}
			}
		}

		// Prioridad 3: Reemplazo competitivo si el nuevo par ofrece mejor latencia que el peor del anillo
		if evictIdx == -1 {
			worstIdx := 0
			worstLatency := r.rings[ring][0].Locator.LatencyMs
			for i, p := range r.rings[ring] {
				if p.Locator.LatencyMs > worstLatency {
					worstLatency = p.Locator.LatencyMs
					worstIdx = i
				}
			}
			if latencyMs < worstLatency {
				evictIdx = worstIdx
			}
		}

		if evictIdx >= 0 {
			oldPeer := r.rings[ring][evictIdx]
			delete(r.peerIndex, oldPeer.DID)

			newPeer := &PeerNode{
				DID:       did,
				PublicKey: pub,
				Locator: PeerLocator{
					PhysicalAddr: addr,
					LastSeen:     time.Now(),
					LatencyMs:    latencyMs,
				},
				RingIndex:   ring,
				HealthState: HealthStateHealthy,
				HealthScore: 100.0,
			}
			r.rings[ring][evictIdx] = newPeer
			r.peerIndex[did] = newPeer
		}
		return nil
	}

	// 3. Si la capacidad global está al límite, desalojar el par globalmente más deteriorado
	if len(r.peerIndex) >= r.config.MaxTotalPeers {
		globalWorstDID := ""
		globalWorstRing := -1
		globalWorstIdx := -1
		worstScore := 999.0

		for rIdx, ringList := range r.rings {
			if len(ringList) <= 1 {
				continue // No vaciar anillos únicos para no fragmentar el árbol de enrutamiento
			}
			for pIdx, p := range ringList {
				score := p.HealthScore - (p.Locator.LatencyMs / 10.0)
				if score < worstScore {
					worstScore = score
					globalWorstDID = p.DID
					globalWorstRing = rIdx
					globalWorstIdx = pIdx
				}
			}
		}

		if globalWorstDID != "" && globalWorstRing >= 0 && globalWorstIdx >= 0 {
			delete(r.peerIndex, globalWorstDID)
			// Remover del slice
			r.rings[globalWorstRing] = append(r.rings[globalWorstRing][:globalWorstIdx], r.rings[globalWorstRing][globalWorstIdx+1:]...)
		}
	}

	newPeer := &PeerNode{
		DID:       did,
		PublicKey: pub,
		Locator: PeerLocator{
			PhysicalAddr: addr,
			LastSeen:     time.Now(),
			LatencyMs:    latencyMs,
		},
		RingIndex:   ring,
		HealthState: HealthStateHealthy,
		HealthScore: 100.0,
	}

	r.rings[ring] = append(r.rings[ring], newPeer)
	r.peerIndex[did] = newPeer
	r.TotalRoutes++
	return nil
}


// UpdatePeerHealth actualiza las métricas empíricas y transiciona la FSM del par
func (r *KleinbergRouter) UpdatePeerHealth(did string, pacingCompliance, jitterMs, lossRate float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer, found := r.peerIndex[did]
	if !found {
		return
	}

	score, state := CalculateHealthScore(pacingCompliance, jitterMs, lossRate)
	peer.JitterMs = jitterMs
	peer.LossRate = lossRate
	peer.HealthScore = score
	peer.HealthState = state
}

// FindNextHop implementa el reenvío voraz hacia la clave destino con métrica híbrida 2D (XOR + EWMA Latency)
// y conmutación proactiva Make-Before-Break
func (r *KleinbergRouter) FindNextHop(destDID string) (*PeerNode, error) {
	pubDest, err := l0.PublicKeyFromDID(destDID)
	if err != nil {
		return nil, fmt.Errorf("did destino inválido: %w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Si está en índice directo y saludable, entrega directa inmediata
	if direct, found := r.peerIndex[destDID]; found {
		if direct.HealthState == HealthStateHealthy {
			return direct, nil
		}
	}

	// 2. Búsqueda voraz con función métrica híbrida 2D (distancia XOR, latencia y salud FSM)
	return FindBestPeerCandidate(r.peerIndex, pubDest, r.config.NumRings, r.config.AlphaLatencyWeight)
}

// GetAllPeers retorna una instantánea de todos los pares registrados
func (r *KleinbergRouter) GetAllPeers() []*PeerNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]*PeerNode, 0, len(r.peerIndex))
	for _, p := range r.peerIndex {
		pCopy := *p
		res = append(res, &pCopy)
	}
	return res
}

// RemovePeer elimina un par por su DID
func (r *KleinbergRouter) RemovePeer(did string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, found := r.peerIndex[did]
	if !found {
		return
	}
	delete(r.peerIndex, did)
	ring := p.RingIndex
	if ring >= 0 && ring < len(r.rings) {
		newRing := make([]*PeerNode, 0, len(r.rings[ring]))
		for _, item := range r.rings[ring] {
			if item.DID != did {
				newRing = append(newRing, item)
			}
		}
		r.rings[ring] = newRing
	}
}

// HandleRoamingUpdate procesa un paquete de roaming y actualiza el localizador
func (r *KleinbergRouter) HandleRoamingUpdate(pkt *l0.Packet, newAddr *net.UDPAddr) error {
	valid, err := pkt.VerifyPacketSignature()
	if err != nil || !valid {
		return errors.New("firma de paquete de roaming inválida")
	}

	return r.AddOrUpdatePeer(pkt.SourceDID, newAddr, 1.1)
}



