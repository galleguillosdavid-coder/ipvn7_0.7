// Package l1 implementa el selector dinámico del gateway de salida óptimo con histéresis.
package l1

import (
	"sync"
	"time"
)

// EgressSelector evalúa la malla y selecciona la mejor pasarela predeterminada
type EgressSelector struct {
	mu                 sync.RWMutex
	mode               int
	specificDID        string
	activeDID          string
	activeScore        float64
	lastSwitch         time.Time
	registry           *EgressRegistry
	hysteresisMargin   float64
	minObservationTime time.Duration
	onGatewayChanged   func(oldDID, newDID string)
}

// NewEgressSelector inicializa el selector con parámetros canónicos
func NewEgressSelector(reg *EgressRegistry) *EgressSelector {
	if reg == nil {
		reg = NewEgressRegistry()
	}
	return &EgressSelector{
		mode:               EgressModeAuto,
		registry:           reg,
		hysteresisMargin:   EgressHysteresisMargin,
		minObservationTime: EgressMinObservationTime,
	}
}

// SetOnGatewayChanged registra un callback invocado ante una conmutación de gateway
func (s *EgressSelector) SetOnGatewayChanged(fn func(oldDID, newDID string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onGatewayChanged = fn
}

// SetMode configura el modo de operación y opcionalmente el DID objetivo
func (s *EgressSelector) SetMode(mode int, specificDID string) {
	s.mu.Lock()
	oldDID := s.activeDID
	s.mode = mode
	s.specificDID = specificDID

	if mode == EgressModeDirect {
		s.activeDID = ""
		s.activeScore = 0
		s.lastSwitch = time.Now()
	}
	newDID := s.activeDID
	cb := s.onGatewayChanged
	s.mu.Unlock()

	if cb != nil && oldDID != newDID {
		cb(oldDID, newDID)
	}
}

// GetStatus retorna el estado actual del selector
func (s *EgressSelector) GetStatus() (mode int, activeDID string, activeScore float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode, s.activeDID, s.activeScore
}

// EvaluatePeers analiza una lista de pares físicos y conmuta al gateway óptimo
func (s *EgressSelector) EvaluatePeers(peers []*PeerNode, now time.Time) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mode == EgressModeDirect {
		s.activeDID = ""
		s.activeScore = 0
		return ""
	}

	if s.mode == EgressModeSpecific {
		return s.evaluateSpecific(peers, now)
	}

	return s.evaluateAuto(peers, now)
}

func (s *EgressSelector) evaluateSpecific(peers []*PeerNode, now time.Time) string {
	targetDID := s.specificDID
	if targetDID == "" {
		s.activeDID = ""
		return ""
	}

	for _, p := range peers {
		if p.DID == targetDID && p.HealthState != HealthStateUnreachable && p.HealthState != HealthStateQuarantined {
			if s.activeDID != targetDID {
				old := s.activeDID
				s.activeDID = targetDID
				s.lastSwitch = now
				if s.onGatewayChanged != nil {
					go s.onGatewayChanged(old, targetDID)
				}
			}
			return targetDID
		}
	}

	// Si el nodo específico está inalcanzable, fallback automático a salida directa local
	if s.activeDID != "" {
		old := s.activeDID
		s.activeDID = ""
		s.lastSwitch = now
		if s.onGatewayChanged != nil {
			go s.onGatewayChanged(old, "")
		}
	}
	return ""
}

func (s *EgressSelector) evaluateAuto(peers []*PeerNode, now time.Time) string {
	var bestDID string
	bestScore := 999999.0

	// Buscar el candidato óptimo entre los pares que anuncian salida a Internet
	for _, p := range peers {
		if p.HealthState == HealthStateUnreachable || p.HealthState == HealthStateQuarantined {
			continue
		}
		cap, ok := s.registry.Get(p.DID)
		if !ok || !cap.CanExit {
			continue
		}

		score := CalculateGatewayScore(p.Locator.LatencyMs, p.LossRate, cap.ActiveSessions)
		if score < bestScore {
			bestScore = score
			bestDID = p.DID
		}
	}

	// Si no hay candidatos viables con salida a Internet, fallback a salida local directa
	if bestDID == "" {
		if s.activeDID != "" {
			old := s.activeDID
			s.activeDID = ""
			s.activeScore = 0
			s.lastSwitch = now
			if s.onGatewayChanged != nil {
				go s.onGatewayChanged(old, "")
			}
		}
		return ""
	}

	// Si ya tenemos un gateway activo, aplicar regla de histéresis del 25% para evitar flapping
	if s.activeDID != "" {
		if s.activeDID == bestDID {
			s.activeScore = bestScore
			return s.activeDID
		}

		// Solo conmutar si el nuevo gateway supera al actual en al menos un 25% sostenido
		threshold := s.activeScore * (1.0 - s.hysteresisMargin)
		if bestScore >= threshold && now.Sub(s.lastSwitch) < s.minObservationTime {
			return s.activeDID // Mantener el actual por estabilidad
		}
	}

	// Conmutación exitosa al nuevo gateway óptimo
	old := s.activeDID
	s.activeDID = bestDID
	s.activeScore = bestScore
	s.lastSwitch = now

	if s.onGatewayChanged != nil && old != bestDID {
		go s.onGatewayChanged(old, bestDID)
	}

	return s.activeDID
}
