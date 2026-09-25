// Package l1 implementa el centinela anti-envenenamiento de tablas de enrutamiento (Adversarial Route Poisoning Guard)
// impidiendo que tráfico sintético, alucinaciones de modelos de IA o tormentas de señalización desvíen tráfico.
package l1

import (
	"crypto/sha256"
	"errors"
	"math"
	"sync"
	"time"
)

var (
	ErrAdversarialLoopDetected    = errors.New("route_guard: bucle de enrutamiento o auto-peering detectado")
	ErrMetricDivergence           = errors.New("route_guard: la ruta propuesta no reduce la distancia XOR de Kleinberg")
	ErrPhysicalDiscrepancyTooHigh = errors.New("route_guard: discrepancia crítica entre latencia de IA y medición física UDP")
	ErrSignalingStormDampened     = errors.New("route_guard: oscilación de ruta amortiguada para prevenir avalanchas de señalización")
)

// RouteValidationResult contiene el dictamen de admisión de una ruta
type RouteValidationResult struct {
	Admitted        bool
	RejectionReason string
	SourceDID       string
	NextHopDID      string
	TargetDID       string
	Timestamp       time.Time
}

// AdversarialRouteGuard audita cada propuesta de ruta antes de permitir su conmutación
type AdversarialRouteGuard struct {
	mu           sync.RWMutex
	flapCounters map[string]int
	lastFlapTime map[string]time.Time
	maxFlaps     int
}

// NewAdversarialRouteGuard inicializa el centinela anti-envenenamiento
func NewAdversarialRouteGuard() *AdversarialRouteGuard {
	return &AdversarialRouteGuard{
		flapCounters: make(map[string]int),
		lastFlapTime: make(map[string]time.Time),
		maxFlaps:     3,
	}
}

// ValidateRouteProposal verifica matemáticamente que una ruta sea segura, real y no un ataque sintético
func (g *AdversarialRouteGuard) ValidateRouteProposal(
	srcDID, targetDID, proposedNextHop string,
	aiPredictedLatencyMs float64,
	physicalMeasuredRTT float64,
) (bool, error) {
	// 1. Verificación geométrica básica: prohibido auto-enrutamiento o bucle directo
	if proposedNextHop == srcDID || proposedNextHop == "" {
		return false, ErrAdversarialLoopDetected
	}

	// 2. Verificación de discrepancia física: la IA no puede engañar la física del socket UDP
	if physicalMeasuredRTT > 0 {
		discrepancy := math.Abs(aiPredictedLatencyMs - physicalMeasuredRTT) / physicalMeasuredRTT
		if discrepancy > 0.40 && aiPredictedLatencyMs < physicalMeasuredRTT {
			// La IA afirma que el enlace es milagrosamente rápido cuando el socket físico mide alta latencia/pérdida
			return false, ErrPhysicalDiscrepancyTooHigh
		}
	}

	// 3. Verificación de distancia métrica XOR hacia el objetivo
	srcHash := sha256.Sum256([]byte(srcDID))
	nextHash := sha256.Sum256([]byte(proposedNextHop))
	tgtHash := sha256.Sum256([]byte(targetDID))

	currDist := computeXORDistance(srcHash[:], tgtHash[:])
	proposedDist := computeXORDistance(nextHash[:], tgtHash[:])

	// Si no es el destino final, el salto debe acortar la distancia o mantenerse en el mismo anillo
	if proposedNextHop != targetDID && proposedDist > currDist {
		return false, ErrMetricDivergence
	}

	// 4. Amortiguador de tormentas de señalización (Anti-Flap Dampener)
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	routeKey := srcDID + "->" + targetDID
	if lastTime, exists := g.lastFlapTime[routeKey]; exists {
		if now.Sub(lastTime) < 5*time.Second {
			g.flapCounters[routeKey]++
			if g.flapCounters[routeKey] > g.maxFlaps {
				return false, ErrSignalingStormDampened
			}
		} else {
			g.flapCounters[routeKey] = 0
		}
	}
	g.lastFlapTime[routeKey] = now

	return true, nil
}

// computeXORDistance calcula la métrica XOR de 256 bits para ordenamiento de Kleinberg
func computeXORDistance(a, b []byte) uint64 {
	var dist uint64
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen && i < 8; i++ {
		dist = (dist << 8) | uint64(a[i]^b[i])
	}
	return dist
}
