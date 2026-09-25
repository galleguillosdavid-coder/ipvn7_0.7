package l1

import (
	"errors"
	"sync"
	"time"
)

const (
	DefaultMaxConsecutiveLosses = 3
	DefaultDegradedLatencyMs    = 500.0
	DefaultRecoverySuccesses    = 2 // Histeresis anti-flapping: requiere 2 éxitos consecutivos para restaurar
	LinkHealingCooldown         = 5 * time.Second
)

var (
	ErrNoAlternativePeer = errors.New("link_healing: no hay pares alternativos disponibles en los anillos")
	ErrPeerNotFound      = errors.New("link_healing: par objetivo no encontrado")
)

// PeerHealthStats registra métricas en tiempo real de salud de un enlace P2P
type PeerHealthStats struct {
	DID                  string
	ConsecutiveLosses    int
	ConsecutiveSuccesses int
	LastLatencyMs        float64
	LastEvaluation       time.Time
	Degraded             bool
	FailoverCount        uint32
	LastFailoverTime     time.Time
	HoldDownUntil        time.Time
	FlapPenaltyCount     int
}

// LinkHealingEngine supervisa la degradación de enlaces y ejecuta failover O(1)
type LinkHealingEngine struct {
	mu     sync.RWMutex
	router *KleinbergRouter
	stats  map[string]*PeerHealthStats
}

// NewLinkHealingEngine instancia el centinela de auto-reparación de enlaces
func NewLinkHealingEngine(router *KleinbergRouter) *LinkHealingEngine {
	return &LinkHealingEngine{
		router: router,
		stats:  make(map[string]*PeerHealthStats),
	}
}

// RecordProbeResult procesa el resultado de un sondeo o datagrama y determina si requiere failover
func (h *LinkHealingEngine) RecordProbeResult(did string, success bool, latencyMs float64) (bool, string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	stat, exists := h.stats[did]
	now := time.Now()
	if !exists {
		stat = &PeerHealthStats{
			DID:            did,
			LastEvaluation: now,
		}
		h.stats[did] = stat
	}

	stat.LastEvaluation = now
	stat.LastLatencyMs = latencyMs

	if !success {
		stat.ConsecutiveLosses++
		stat.ConsecutiveSuccesses = 0
		if stat.ConsecutiveLosses >= DefaultMaxConsecutiveLosses {
			stat.Degraded = true
		}
	} else {
		stat.ConsecutiveLosses = 0
		if latencyMs > 1500.0 {
			stat.Degraded = true
			stat.ConsecutiveSuccesses = 0
		} else if latencyMs > DefaultDegradedLatencyMs {
			stat.Degraded = true
			stat.ConsecutiveSuccesses = 0
		} else {
			stat.ConsecutiveSuccesses++
			// Histeresis anti-flapping: exige estabilidad consecutiva y que el periodo Hold-Down haya expirado
			if stat.ConsecutiveSuccesses >= DefaultRecoverySuccesses {
				if !now.Before(stat.HoldDownUntil) {
					stat.Degraded = false
					stat.FlapPenaltyCount = 0 // Enlace estabilizado de forma continua
				}
			}
		}
	}

	if (!success || latencyMs > DefaultDegradedLatencyMs) && stat.Degraded {
		// RFC 2439 Flap Damping: si ocurren failovers repetidos en < 10s, activar penalización exponencial Hold-Down
		if !stat.LastFailoverTime.IsZero() && now.Sub(stat.LastFailoverTime) < 10*time.Second {
			stat.FlapPenaltyCount++
			penaltyDuration := time.Duration(stat.FlapPenaltyCount) * 2 * time.Second
			if penaltyDuration > 30*time.Second {
				penaltyDuration = 30 * time.Second
			}
			stat.HoldDownUntil = now.Add(penaltyDuration)
		}
		stat.LastFailoverTime = now

		// Seleccionar par alternativo en anillo concéntrico
		altDID, err := h.selectAlternativePeerLocked(did)
		if err == nil {
			stat.FailoverCount++
			return true, altDID
		}
	}

	return false, ""
}

// selectAlternativePeerLocked busca el par óptimo disponible excluyendo al degradado
func (h *LinkHealingEngine) selectAlternativePeerLocked(degradedDID string) (string, error) {
	if h.router == nil {
		return "", ErrNoAlternativePeer
	}

	allPeers := h.router.GetAllPeers()
	var bestCandidate *PeerNode
	var lowestLatency float64 = 999999.0

	for _, p := range allPeers {
		if p.DID == degradedDID {
			continue
		}
		// Evaluar si el par alternativo está saludable
		if p.HealthState == 0 { // 0 = Healthy
			rtt := 1.0
			if p.Locator.LatencyMs > 0 {
				rtt = p.Locator.LatencyMs
			}
			if bestCandidate == nil || rtt < lowestLatency {
				bestCandidate = p
				lowestLatency = rtt
			}
		}
	}

	if bestCandidate == nil {
		return "", ErrNoAlternativePeer
	}

	return bestCandidate.DID, nil
}

// GetPeerStats retorna una copia de las estadísticas de salud del enlace
func (h *LinkHealingEngine) GetPeerStats(did string) (*PeerHealthStats, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	stat, ok := h.stats[did]
	if !ok {
		return nil, false
	}
	statCopy := *stat
	return &statCopy, true
}
