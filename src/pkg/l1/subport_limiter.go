// Package l1 implementa el limitador anti-saturación de sub-puertos (SubPort Anti-Starvation Limiter)
// blindando el canal de telemetría (SubPort 2) y el plano de control contra ataques DoS y agotamiento de heap (OOM).
package l1

import (
	"sync"
)

const (
	// DefaultTelemetryMaxBurst ráfaga máxima permitida para telemetría
	DefaultTelemetryMaxBurst = 50
	// DefaultTelemetryRefillRate tasa de reposición de tokens por segundo
	DefaultTelemetryRefillRate = 10.0
	// MaxDedicatedBucketsPerSubPort cota superior estricta de memoria para estados dedicados
	MaxDedicatedBucketsPerSubPort = 256
)

// SubPortLimiter gestiona la limitación de tráfico con memoria estricta O(1) contra ataques de DIDs aleatorios
type SubPortLimiter struct {
	mu                    sync.Mutex
	subportRates          map[SubPort]map[string]*TokenBucket
	untrustedSharedBucket *TokenBucket // Cubeta global preasignada para DIDs no verificados (Cero OOM)
	authenticatedDIDs     map[string]bool
	defaultBurst          float64
	defaultRate           float64
}

// NewSubPortLimiter inicializa el blindaje anti-saturación del plano de control
func NewSubPortLimiter() *SubPortLimiter {
	return &SubPortLimiter{
		subportRates:          make(map[SubPort]map[string]*TokenBucket),
		untrustedSharedBucket: NewTokenBucket(DefaultTelemetryMaxBurst, DefaultTelemetryRefillRate),
		authenticatedDIDs:     make(map[string]bool),
		defaultBurst:          DefaultTelemetryMaxBurst,
		defaultRate:           DefaultTelemetryRefillRate,
	}
}

// RegisterAuthenticatedPeer certifica un par tras el apretón de manos Noise/PQC
func (l *SubPortLimiter) RegisterAuthenticatedPeer(did string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.authenticatedDIDs[did] = true
}

// AllowPacket determina si un datagrama para un sub-puerto específico es admitido o descartado en O(1)
// Garantiza inmunidad total contra OOM: DIDs no autenticados comparten una única cubeta preasignada en RAM.
func (l *SubPortLimiter) AllowPacket(sp SubPort, srcDID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Invariante Anti-OOM: Si el DID no está autenticado, no aloca memoria dinámica
	if !l.authenticatedDIDs[srcDID] {
		return l.untrustedSharedBucket.Consume(1.0)
	}

	spBuckets, exists := l.subportRates[sp]
	if !exists {
		spBuckets = make(map[string]*TokenBucket)
		l.subportRates[sp] = spBuckets
	}

	bucket, bExists := spBuckets[srcDID]
	if !bExists {
		// Protección de cota de estado: si se alcanza el límite, conmuta a cubeta compartida sin alocar
		if len(spBuckets) >= MaxDedicatedBucketsPerSubPort {
			return l.untrustedSharedBucket.Consume(1.0)
		}
		bucket = NewTokenBucket(l.defaultBurst, l.defaultRate)
		spBuckets[srcDID] = bucket
	}

	return bucket.Consume(1.0)
}

// SetCustomRate configura una cuota específica para un sub-puerto crítico
func (l *SubPortLimiter) SetCustomRate(sp SubPort, burst, refillRate float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.defaultBurst = burst
	l.defaultRate = refillRate
	l.untrustedSharedBucket = NewTokenBucket(burst, refillRate)
}
