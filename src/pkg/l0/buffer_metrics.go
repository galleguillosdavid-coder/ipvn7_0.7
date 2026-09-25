// Package l0 implementa el monitor de salud y telemetría de memoria en tiempo real
// para anillos de buffers de datagramas de 1280 bytes en la malla ipvn7 v0.7.
package l0

import (
	"sync/atomic"
)

const (
	// DefaultMaxBufferBudget límite estricto de buffers concurrentes en tránsito (10,000 * 1280B = 12.8 MB)
	DefaultMaxBufferBudget uint64 = 10000
)

// BufferHealthSnapshot captura el estado instantáneo de la memoria de datagramas
type BufferHealthSnapshot struct {
	BuffersAllocated uint64 `json:"buffers_allocated"`
	BuffersRecycled  uint64 `json:"buffers_recycled"`
	BuffersInUse     uint64 `json:"buffers_in_use"`
	HighWaterMark    uint64 `json:"high_water_mark"`
	MaxBudget        uint64 `json:"max_budget"`
	LeakAlert        bool   `json:"leak_alert"`
}

// BufferHealthMonitor rastrea atómicamente el ciclo de vida de los buffers sin overhead de bloqueos
type BufferHealthMonitor struct {
	allocated uint64
	recycled  uint64
	hwm       uint64
	maxBudget uint64
}

// NewBufferHealthMonitor inicializa un nuevo monitor con el presupuesto especificado
func NewBufferHealthMonitor(maxBudget uint64) *BufferHealthMonitor {
	if maxBudget == 0 {
		maxBudget = DefaultMaxBufferBudget
	}
	return &BufferHealthMonitor{
		maxBudget: maxBudget,
	}
}

// TrackAlloc registra la asignación de un buffer del pool
func (m *BufferHealthMonitor) TrackAlloc() {
	alloc := atomic.AddUint64(&m.allocated, 1)
	recycled := atomic.LoadUint64(&m.recycled)
	if alloc >= recycled {
		inUse := alloc - recycled
		// Actualizar High Water Mark si es necesario
		for {
			currentHWM := atomic.LoadUint64(&m.hwm)
			if inUse <= currentHWM {
				break
			}
			if atomic.CompareAndSwapUint64(&m.hwm, currentHWM, inUse) {
				break
			}
		}
	}
}

// TrackRecycle registra la devolución exitosa de un buffer al pool
func (m *BufferHealthMonitor) TrackRecycle() {
	atomic.AddUint64(&m.recycled, 1)
}

// Snapshot genera un reporte inmutable del estado del pool de memoria
func (m *BufferHealthMonitor) Snapshot() BufferHealthSnapshot {
	alloc := atomic.LoadUint64(&m.allocated)
	recycled := atomic.LoadUint64(&m.recycled)

	var inUse uint64
	if alloc >= recycled {
		inUse = alloc - recycled
	}

	hwm := atomic.LoadUint64(&m.hwm)
	budget := atomic.LoadUint64(&m.maxBudget)

	return BufferHealthSnapshot{
		BuffersAllocated: alloc,
		BuffersRecycled:  recycled,
		BuffersInUse:     inUse,
		HighWaterMark:    hwm,
		MaxBudget:        budget,
		LeakAlert:        inUse > budget,
	}
}

// Reset reinicia los contadores del monitor
func (m *BufferHealthMonitor) Reset() {
	atomic.StoreUint64(&m.allocated, 0)
	atomic.StoreUint64(&m.recycled, 0)
	atomic.StoreUint64(&m.hwm, 0)
}
