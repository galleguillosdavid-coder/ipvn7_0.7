// Package l1 implementa la ventana de protección Anti-Replay con aislamiento Cross-Session
// rescatada de ip7uin_MVP_Spec_v1.1.docx (Secciones 3, 4 y 7).
// Valida exhaustivamente 4 vectores de ataque:
// 1. Replay Inmediato (duplicado en la misma ventana de secuencia).
// 2. Replay Tardío (paquete obsoleto que cayó fuera del horizonte de la ventana).
// 3. Replay Cross-Session (secuencia duplicada o inyectada en sesiones distintas).
// 4. Reorden con Jitter (paquetes legítimos que llegan desordenados dentro del umbral).
package l1

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	AntiReplayWindowSize = 1024 // Tamaño de la ventana deslizante por sesión
	MaxAllowedJitterSec  = 60   // Máximo retraso temporal permitido (60 segundos)
)

// SessionReplayTracker mantiene el estado de una sesión individual
type SessionReplayTracker struct {
	SessionID uint64
	MaxSeq    uint64
	SeenBits  []uint64 // Mapa de bits para ventana deslizante (1024 bits = 16 x uint64)
	LastSeen  int64    // Timestamp UNIX
}

// AntiReplayStats contiene métricas de paquetes evaluados y descartados
type AntiReplayStats struct {
	EvaluatedPackets  uint64 `json:"evaluated_packets"`
	AcceptedPackets   uint64 `json:"accepted_packets"`
	ImmediateReplays  uint64 `json:"immediate_replays"`
	StaleReplays      uint64 `json:"stale_replays"`
	CrossSessionDrops uint64 `json:"cross_session_drops"`
	TimestampDrops    uint64 `json:"timestamp_drops"`
	ActiveSessions    int    `json:"active_sessions"`
}

// AntiReplayFilter gestiona las ventanas de validación para todos los pares y sesiones
type AntiReplayFilter struct {
	mu            sync.RWMutex
	sessions      map[string]*SessionReplayTracker // Clave: originDID:sessionID
	latestSeen    map[string]int64                 // Clave: originDID -> último timestamp UNIX registrado
	latestSession map[string]uint64                // Clave: originDID -> mayor SessionID vista
	arbiter       *GlobalMemoryArbiter
	stats         AntiReplayStats
}

// NewAntiReplayFilter inicializa el filtro con enlace opcional al árbitro de memoria
func NewAntiReplayFilter(arbiter *GlobalMemoryArbiter) *AntiReplayFilter {
	return &AntiReplayFilter{
		sessions:      make(map[string]*SessionReplayTracker),
		latestSeen:    make(map[string]int64),
		latestSession: make(map[string]uint64),
		arbiter:       arbiter,
	}
}

// Accept evalúa un paquete y determina si es legítimo o si constituye un ataque de replay.
// Retorna true si es aceptado (y actualiza el estado de la ventana), o false si es descartado.
func (f *AntiReplayFilter) Accept(originDID string, sessionID uint64, seq uint64, ts int64) bool {
	atomic.AddUint64(&f.stats.EvaluatedPackets, 1)

	now := time.Now().Unix()

	// Normalizar timestamp si viene en nanosegundos o milisegundos
	tsSec := ts
	if tsSec > 1e14 {
		tsSec = tsSec / 1e9
	} else if tsSec > 1e11 {
		tsSec = tsSec / 1e3
	}

	// 1. Validación de Timestamp (previene paquetes congelados o del futuro distante)
	diff := now - tsSec
	if diff > MaxAllowedJitterSec || diff < -10 {
		atomic.AddUint64(&f.stats.TimestampDrops, 1)
		return false
	}

	sessionKey := fmt.Sprintf("%s:%d", originDID, sessionID)

	f.mu.Lock()
	defer f.mu.Unlock()

	lastSeenTs := f.latestSeen[originDID]
	lastSessID := f.latestSession[originDID]

	// 2. Validación Cross-Session contra sesiones obsoletas o inyecciones pasadas
	if lastSessID > 0 && sessionID < lastSessID && tsSec < lastSeenTs {
		atomic.AddUint64(&f.stats.CrossSessionDrops, 1)
		return false
	}

	tracker, exists := f.sessions[sessionKey]
	if !exists {
		// Nueva sesión: reservar memoria en el árbitro
		if f.arbiter != nil && !f.arbiter.Reserve(BudgetReplay, 256) {
			// Rechazo preventivo por saturación de cuota de memoria
			return false
		}

		tracker = &SessionReplayTracker{
			SessionID: sessionID,
			MaxSeq:    seq,
			SeenBits:  make([]uint64, AntiReplayWindowSize/64),
			LastSeen:  now,
		}
		tracker.markBit(0) // Marcar la secuencia actual
		f.sessions[sessionKey] = tracker
		if tsSec > lastSeenTs {
			f.latestSeen[originDID] = tsSec
		}
		if sessionID > lastSessID {
			f.latestSession[originDID] = sessionID
		}

		atomic.AddUint64(&f.stats.AcceptedPackets, 1)
		return true
	}

	tracker.LastSeen = now
	if tsSec > lastSeenTs {
		f.latestSeen[originDID] = tsSec
	}
	if sessionID > lastSessID {
		f.latestSession[originDID] = sessionID
	}

	// 3. Paquete que avanza la secuencia (nuevo máximo)
	if seq > tracker.MaxSeq {
		shift := seq - tracker.MaxSeq
		if shift >= AntiReplayWindowSize {
			// Ráfaga que sobrepasa toda la ventana: reiniciar mapa de bits
			tracker.SeenBits = make([]uint64, AntiReplayWindowSize/64)
		} else {
			tracker.shiftWindow(shift)
		}
		tracker.MaxSeq = seq
		tracker.markBit(0)

		atomic.AddUint64(&f.stats.AcceptedPackets, 1)
		return true
	}

	// 4. Paquete retrasado (seq <= tracker.MaxSeq)
	offset := tracker.MaxSeq - seq

	// Replay Tardío (fuera de la ventana deslizante)
	if offset >= AntiReplayWindowSize {
		atomic.AddUint64(&f.stats.StaleReplays, 1)
		return false
	}

	// Replay Inmediato (ya visto en la ventana actual)
	if tracker.isBitSet(offset) {
		atomic.AddUint64(&f.stats.ImmediateReplays, 1)
		return false
	}

	// Reorden con Jitter legítimo dentro de la ventana
	tracker.markBit(offset)
	atomic.AddUint64(&f.stats.AcceptedPackets, 1)
	return true
}

func (t *SessionReplayTracker) markBit(offset uint64) {
	wordIdx := offset / 64
	bitIdx := offset % 64
	if int(wordIdx) < len(t.SeenBits) {
		t.SeenBits[wordIdx] |= (1 << bitIdx)
	}
}

func (t *SessionReplayTracker) isBitSet(offset uint64) bool {
	wordIdx := offset / 64
	bitIdx := offset % 64
	if int(wordIdx) < len(t.SeenBits) {
		return (t.SeenBits[wordIdx] & (1 << bitIdx)) != 0
	}
	return false
}

func (t *SessionReplayTracker) shiftWindow(shift uint64) {
	if shift >= AntiReplayWindowSize {
		for i := range t.SeenBits {
			t.SeenBits[i] = 0
		}
		return
	}

	// Desplazar bits hacia la derecha según el avance
	wordShift := int(shift / 64)
	bitShift := shift % 64

	numWords := len(t.SeenBits)
	newBits := make([]uint64, numWords)

	for i := 0; i < numWords; i++ {
		srcIdx := i - wordShift
		if srcIdx >= 0 {
			newBits[i] |= (t.SeenBits[srcIdx] << bitShift)
		}
		if bitShift > 0 && srcIdx-1 >= 0 {
			newBits[i] |= (t.SeenBits[srcIdx-1] >> (64 - bitShift))
		}
	}

	t.SeenBits = newBits
}

// GetStats retorna las métricas de filtrado anti-replay
func (f *AntiReplayFilter) GetStats() AntiReplayStats {
	f.mu.RLock()
	defer f.mu.RUnlock()

	s := f.stats
	s.ActiveSessions = len(f.sessions)
	return s
}
