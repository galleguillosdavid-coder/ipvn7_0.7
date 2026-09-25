package l1

import (
	"context"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ProbeMagicNumber uint32 = 0x49503750 // "IP7P" (IPVN7 Probe)
	ProbeWireSize           = 32
	DefaultProbeFreq        = 2 * time.Second
	DefaultProbeTimeout     = 800 * time.Millisecond
)

// ProbePacket estructura la trama mínima para medición de RTT
type ProbePacket struct {
	Magic     uint32
	Sequence  uint32
	Timestamp int64
}

// PeerProbeStats almacena la telemetría viva de sondeo por par
type PeerProbeStats struct {
	DID          string
	Endpoint     string
	LastRTTMs    float64
	MinRTTMs     float64
	MaxRTTMs     float64
	JitterMs     float64
	ProbesSent   uint64
	ProbesRecv   uint64
	ProbesLost   uint64
	ConsecLosses int
	LastUpdated  time.Time
}

// WANActiveProber orquesta el sondeo activo de enlaces WAN para auto-reparación
type WANActiveProber struct {
	mu           sync.RWMutex
	router       *KleinbergRouter
	healing      *LinkHealingEngine
	conn         net.PacketConn
	stats        map[string]*PeerProbeStats
	running      atomic.Bool
	stopChan     chan struct{}
	seqCounter   atomic.Uint32
	probeFreq    time.Duration
	probeTimeout time.Duration
}

// NewWANActiveProber crea una nueva instancia del motor de sondeo activo
func NewWANActiveProber(router *KleinbergRouter, healing *LinkHealingEngine, conn net.PacketConn) *WANActiveProber {
	return &WANActiveProber{
		router:       router,
		healing:      healing,
		conn:         conn,
		stats:        make(map[string]*PeerProbeStats),
		stopChan:     make(chan struct{}),
		probeFreq:    DefaultProbeFreq,
		probeTimeout: DefaultProbeTimeout,
	}
}

// SetFrequency ajusta la cadencia de emisión de sondas
func (p *WANActiveProber) SetFrequency(freq time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.probeFreq = freq
}

// Start inicia el ciclo de sondeo periódico en segundo plano
func (p *WANActiveProber) Start(ctx context.Context) {
	if !p.running.CompareAndSwap(false, true) {
		return
	}

	go p.probeLoop(ctx)
}

// Stop detiene el motor de sondeo
func (p *WANActiveProber) Stop() {
	if p.running.CompareAndSwap(true, false) {
		close(p.stopChan)
	}
}

// probeLoop ejecuta el barrido periódico sobre los pares activos en los anillos
func (p *WANActiveProber) probeLoop(ctx context.Context) {
	ticker := time.NewTicker(p.probeFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.ProbeAllPeers()
		}
	}
}

// ProbeAllPeers ejecuta una pasada de sondeo sobre todos los pares del enrutador
func (p *WANActiveProber) ProbeAllPeers() {
	if p.router == nil || p.conn == nil {
		return
	}

	peers := p.router.GetAllPeers()
	for _, peer := range peers {
		if peer.Locator.PhysicalAddr == nil {
			continue
		}
		endpoint := peer.Locator.PhysicalAddr.String()
		p.ProbeSinglePeer(peer.DID, endpoint)
	}
}

// ProbeSinglePeer emite un datagrama de sondeo hacia un par específico
func (p *WANActiveProber) ProbeSinglePeer(did, endpoint string) {
	rAddr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		p.recordLoss(did, endpoint)
		return
	}

	seq := p.seqCounter.Add(1)
	nowNano := time.Now().UnixNano()

	buf := make([]byte, ProbeWireSize)
	binary.BigEndian.PutUint32(buf[0:4], ProbeMagicNumber)
	binary.BigEndian.PutUint32(buf[4:8], seq)
	binary.BigEndian.PutUint64(buf[8:16], uint64(nowNano))

	p.mu.Lock()
	stat, exists := p.stats[did]
	if !exists {
		stat = &PeerProbeStats{
			DID:      did,
			Endpoint: endpoint,
		}
		p.stats[did] = stat
	}
	stat.ProbesSent++
	p.mu.Unlock()

	_, err = p.conn.WriteTo(buf, rAddr)
	if err != nil {
		p.recordLoss(did, endpoint)
	}
}

// HandleIncomingProbe procesa un eco o datagrama de sondeo entrante
func (p *WANActiveProber) HandleIncomingProbe(buf []byte, from net.Addr, isResponse bool) bool {
	if len(buf) < ProbeWireSize {
		return false
	}
	magic := binary.BigEndian.Uint32(buf[0:4])
	if magic != ProbeMagicNumber {
		return false
	}

	sentNano := int64(binary.BigEndian.Uint64(buf[8:16]))
	nowNano := time.Now().UnixNano()
	rttMs := float64(nowNano-sentNano) / 1e6

	if rttMs < 0 || rttMs > 10000 {
		return false
	}

	// Si es una solicitud de sondeo entrante, devolvemos el eco inmediatamente
	if !isResponse && p.conn != nil {
		echoBuf := make([]byte, ProbeWireSize)
		copy(echoBuf, buf[:ProbeWireSize])
		_, _ = p.conn.WriteTo(echoBuf, from)
		return true
	}

	// Si es una respuesta de eco, actualizamos estadísticas
	p.recordSuccess(from.String(), rttMs)
	return true
}

func (p *WANActiveProber) recordSuccess(endpoint string, rttMs float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for did, stat := range p.stats {
		if stat.Endpoint == endpoint {
			stat.ProbesRecv++
			stat.ConsecLosses = 0
			stat.LastRTTMs = rttMs
			stat.LastUpdated = time.Now()

			if stat.MinRTTMs == 0 || rttMs < stat.MinRTTMs {
				stat.MinRTTMs = rttMs
			}
			if rttMs > stat.MaxRTTMs {
				stat.MaxRTTMs = rttMs
			}

			// Jitter RFC 3550 aproximado
			diff := rttMs - stat.LastRTTMs
			if diff < 0 {
				diff = -diff
			}
			stat.JitterMs = stat.JitterMs + (diff-stat.JitterMs)/16.0

			// Retroalimentación a centinela de auto-reparación
			if p.healing != nil {
				p.healing.RecordProbeResult(did, true, rttMs)
			}
			return
		}
	}
}

func (p *WANActiveProber) recordLoss(did, endpoint string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	stat, exists := p.stats[did]
	if !exists {
		stat = &PeerProbeStats{
			DID:      did,
			Endpoint: endpoint,
		}
		p.stats[did] = stat
	}

	stat.ProbesLost++
	stat.ConsecLosses++
	stat.LastUpdated = time.Now()

	if p.healing != nil {
		p.healing.RecordProbeResult(did, false, 0)
	}
}

// RecordLoss registra explícitamente un fallo de entrega o timeout de sondeo
func (p *WANActiveProber) RecordLoss(did, endpoint string) {
	p.recordLoss(did, endpoint)
}

// GetPeerStats retorna una copia de las métricas de sondeo para un par
func (p *WANActiveProber) GetPeerStats(did string) (PeerProbeStats, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stat, exists := p.stats[did]
	if !exists {
		return PeerProbeStats{}, false
	}
	return *stat, true
}
