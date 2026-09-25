package l0

import (
	"sync"
)

// ReplayWindowSize1024 define el tamaño canónico de la ventana anti-repetición (16 * 64 = 1024 bits).
const ReplayWindowSize1024 = 1024

// AntiReplayFilter implementa una ventana deslizante determinista de 1024 posiciones (RFC 4303 compatible).
type AntiReplayFilter struct {
	mu      sync.Mutex
	lastSeq uint64
	bitmap  [16]uint64
}

// NewAntiReplayFilter inicializa el filtro anti-repetición con ventana de 1024 bits.
func NewAntiReplayFilter() *AntiReplayFilter {
	return &AntiReplayFilter{}
}

// shiftLeft desplaza el bitmap de 1024 bits hacia la izquierda por n bits.
func (f *AntiReplayFilter) shiftLeft(diff uint64) {
	words := int(diff / 64)
	bits := uint(diff % 64)

	if words >= 16 {
		f.bitmap = [16]uint64{}
		return
	}

	if words > 0 {
		for i := 15; i >= words; i-- {
			f.bitmap[i] = f.bitmap[i-words]
		}
		for i := 0; i < words; i++ {
			f.bitmap[i] = 0
		}
	}

	if bits > 0 {
		carryShift := 64 - bits
		for i := 15; i > 0; i-- {
			f.bitmap[i] = (f.bitmap[i] << bits) | (f.bitmap[i-1] >> carryShift)
		}
		f.bitmap[0] <<= bits
	}
}

// ValidateAndUpdate verifica si un número de secuencia es válido y no repetido dentro de la ventana de 1024 bits.
func (f *AntiReplayFilter) ValidateAndUpdate(seq uint64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if seq > f.lastSeq {
		diff := seq - f.lastSeq
		if diff >= ReplayWindowSize1024 {
			f.bitmap = [16]uint64{}
			f.bitmap[0] = 1
		} else {
			f.shiftLeft(diff)
			f.bitmap[0] |= 1
		}
		f.lastSeq = seq
		return true
	}

	diff := f.lastSeq - seq
	if diff >= ReplayWindowSize1024 {
		return false // Fuera de ventana
	}

	wordIdx := diff / 64
	bitIdx := diff % 64
	mask := uint64(1) << bitIdx

	if (f.bitmap[wordIdx] & mask) != 0 {
		return false // Repetición detectada
	}

	f.bitmap[wordIdx] |= mask
	return true
}
