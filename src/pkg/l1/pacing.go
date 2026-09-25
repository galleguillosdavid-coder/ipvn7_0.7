package l1

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// PacerConfig define los parámetros operativos del marcapasos de paquetes
type PacerConfig struct {
	BottleneckRateBytesPerSec uint64        // Capacidad estimada del cuello de botella
	SafetyMarginPercent       uint8         // Margen de seguridad (ej. 10%)
	MaxBurstPackets           int           // Límite estricto de ráfaga para evitar bufferbloat
	MinInterval               time.Duration // Intervalo mínimo entre paquetes consecutivos
}

// DefaultPacerConfig retorna la configuración estándar para enlaces heterogéneos
func DefaultPacerConfig() PacerConfig {
	return PacerConfig{
		BottleneckRateBytesPerSec: 10 * 1024 * 1024 / 8, // 10 Mbps en bytes/seg por defecto
		SafetyMarginPercent:       10,                  // 10% margen de seguridad
		MaxBurstPackets:           1,                   // Ráfaga unitaria (flujo constante reloj-estricto)
		MinInterval:               100 * time.Microsecond,
	}
}

// PacketPacer implementa el marcapasos de red mediante Token Bucket oficial (golang.org/x/time/rate)
type PacketPacer struct {
	mu            sync.RWMutex
	config        PacerConfig
	limiter       *rate.Limiter
	effectiveRate uint64
	packetsSent   atomic.Uint64
	bytesSent     atomic.Uint64
	throttledCount atomic.Uint64
}

// NewPacketPacer inicializa el marcapasos de red subordinado al cuello de botella
func NewPacketPacer(cfg PacerConfig) *PacketPacer {
	if cfg.MaxBurstPackets <= 0 {
		cfg.MaxBurstPackets = 1
	}
	burstBytes := cfg.MaxBurstPackets * 1280

	p := &PacketPacer{
		config: cfg,
	}
	p.recalculateRate(cfg.BottleneckRateBytesPerSec, burstBytes)
	return p
}

// UpdateBottleneckRate actualiza la capacidad conocida de la ruta ante congestión (Backpressure)
func (p *PacketPacer) UpdateBottleneckRate(rateBytesPerSec uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.BottleneckRateBytesPerSec = rateBytesPerSec
	burstBytes := p.config.MaxBurstPackets * 1280
	p.recalculateRate(rateBytesPerSec, burstBytes)
}

func (p *PacketPacer) recalculateRate(rateBytesPerSec uint64, burstBytes int) {
	if rateBytesPerSec == 0 {
		rateBytesPerSec = 1024 * 128 // Mínimo 128 KB/s
	}
	margin := (rateBytesPerSec * uint64(p.config.SafetyMarginPercent)) / 100
	p.effectiveRate = rateBytesPerSec - margin

	limit := rate.Limit(p.effectiveRate)
	if p.limiter == nil {
		p.limiter = rate.NewLimiter(limit, burstBytes)
	} else {
		p.limiter.SetLimit(limit)
		p.limiter.SetBurst(burstBytes)
	}
}

// Pace regula el envío del paquete mediante el Token Bucket estándar de Go con context.Context
func (p *PacketPacer) Pace(ctx context.Context, packetSizeBytes int) error {
	if packetSizeBytes <= 0 {
		packetSizeBytes = 1280
	}

	p.mu.RLock()
	limiter := p.limiter
	p.mu.RUnlock()

	// Reserva atómica de tokens en el Token Bucket
	if err := limiter.WaitN(ctx, packetSizeBytes); err != nil {
		p.throttledCount.Add(1)
		return fmt.Errorf("envío cancelado en marcapasos: %w", err)
	}

	p.packetsSent.Add(1)
	p.bytesSent.Add(uint64(packetSizeBytes))
	return nil
}

// PacerStats retorna estadísticas de rendimiento del marcapasos
type PacerStats struct {
	EffectiveRateBytesPerSec uint64 `json:"effective_rate_bps"`
	PacketIntervalUs         int64  `json:"packet_interval_us"`
	PacketsSent              uint64 `json:"packets_sent"`
	BytesSent                uint64 `json:"bytes_sent"`
	ThrottledCount           uint64 `json:"throttled_count"`
}

func (p *PacketPacer) Stats() PacerStats {
	p.mu.RLock()
	rateVal := p.effectiveRate
	p.mu.RUnlock()

	var intervalUs int64
	if rateVal > 0 {
		packetsPerSec := float64(rateVal) / 1280.0
		if packetsPerSec > 0 {
			intervalUs = int64(1000000.0 / packetsPerSec)
		}
	}

	return PacerStats{
		EffectiveRateBytesPerSec: rateVal,
		PacketIntervalUs:         intervalUs,
		PacketsSent:              p.packetsSent.Load(),
		BytesSent:                p.bytesSent.Load(),
		ThrottledCount:           p.throttledCount.Load(),
	}
}

// String formatea las métricas del marcapasos
func (p *PacketPacer) String() string {
	s := p.Stats()
	return fmt.Sprintf("PacketPacer[Rate: %d B/s, Interval: %d µs, Sent: %d, Throttled: %d]",
		s.EffectiveRateBytesPerSec, s.PacketIntervalUs, s.PacketsSent, s.ThrottledCount)
}
