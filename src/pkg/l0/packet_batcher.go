// Package l0 implementa el procesamiento por lotes (batching) y el anillo de buffers
// zero-copy para maximizar el throughput de datagramas de 1280 bytes en la malla ipvn7.
package l0

import (
	"errors"
	"sync"
	"time"
)

const (
	// CanonicalPacketSize tamaño MTU inmutable canónico de datagrama ipvn7
	CanonicalPacketSize = 1280
	// DefaultMaxBatchSize número máximo de paquetes por lote para syscalls optimizadas
	DefaultMaxBatchSize = 32
)

// PacketBuffer representa un vector de memoria Scatter-Gather (iovec nativo sin memcpy)
type PacketBuffer struct {
	Raw    []byte                    // Puntero/slice directo a la memoria original (Zero-Copy)
	Data   [CanonicalPacketSize]byte // Búfer de respaldo estático
	Length int
}

// Bytes retorna la porción de memoria virtual sin duplicación de bytes
func (pb *PacketBuffer) Bytes() []byte {
	if pb.Raw != nil {
		return pb.Raw
	}
	return pb.Data[:pb.Length]
}

// BatchBufferPool gestiona el reciclaje de buffers sin alocación dinámica en el heap
type BatchBufferPool struct {
	pool sync.Pool
}

// NewBatchBufferPool inicializa el pool de buffers reutilizables
func NewBatchBufferPool() *BatchBufferPool {
	return &BatchBufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &PacketBuffer{}
			},
		},
	}
}

// Get obtiene un buffer limpio del pool
func (p *BatchBufferPool) Get() *PacketBuffer {
	buf := p.pool.Get().(*PacketBuffer)
	buf.Raw = nil
	buf.Length = 0
	return buf
}

// Put retorna el buffer al pool para su reutilización inmediata
func (p *BatchBufferPool) Put(buf *PacketBuffer) {
	if buf != nil {
		buf.Raw = nil
		buf.Length = 0
		p.pool.Put(buf)
	}
}

// BatchHandler función callback receptora cuando un lote se completa o se vacía
type BatchHandler func(batch []*PacketBuffer) error

// PacketBatcher acumula paquetes entrantes y los despacha en ráfagas de alta eficiencia
type PacketBatcher struct {
	mu           sync.Mutex
	maxBatchSize int
	flushTimeout time.Duration
	pool         *BatchBufferPool
	batch        []*PacketBuffer
	handler      BatchHandler
	closed       bool
}

// NewPacketBatcher inicializa un nuevo agrupador por lotes
func NewPacketBatcher(maxBatchSize int, flushTimeout time.Duration, pool *BatchBufferPool, handler BatchHandler) *PacketBatcher {
	if maxBatchSize <= 0 || maxBatchSize > DefaultMaxBatchSize {
		maxBatchSize = DefaultMaxBatchSize
	}
	if flushTimeout <= 0 {
		flushTimeout = 5 * time.Millisecond
	}
	if pool == nil {
		pool = NewBatchBufferPool()
	}

	return &PacketBatcher{
		maxBatchSize: maxBatchSize,
		flushTimeout: flushTimeout,
		pool:         pool,
		batch:        make([]*PacketBuffer, 0, maxBatchSize),
		handler:      handler,
	}
}

// Push añade un paquete al lote actual. Si el lote alcanza la capacidad máxima, se vacía automáticamente.
func (b *PacketBatcher) Push(data []byte) error {
	if len(data) > CanonicalPacketSize {
		return errors.New("batcher: tamaño de paquete excede 1280 bytes MTU")
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return errors.New("batcher: ya cerrado")
	}

	buf := b.pool.Get()
	buf.Raw = data // Scatter-Gather I/O: asignación directa de vector sin memcpy
	buf.Length = len(data)

	b.batch = append(b.batch, buf)

	if len(b.batch) >= b.maxBatchSize {
		toFlush := b.batch
		b.batch = make([]*PacketBuffer, 0, b.maxBatchSize)
		b.mu.Unlock()

		return b.dispatch(toFlush)
	}

	b.mu.Unlock()
	return nil
}

// Flush vacía manualmente cualquier paquete remanente en el lote actual
func (b *PacketBatcher) Flush() error {
	b.mu.Lock()
	if len(b.batch) == 0 {
		b.mu.Unlock()
		return nil
	}

	toFlush := b.batch
	b.batch = make([]*PacketBuffer, 0, b.maxBatchSize)
	b.mu.Unlock()

	return b.dispatch(toFlush)
}

func (b *PacketBatcher) dispatch(batch []*PacketBuffer) error {
	if len(batch) == 0 {
		return nil
	}

	var err error
	if b.handler != nil {
		err = b.handler(batch)
	}

	// Liberar de vuelta al pool
	for _, buf := range batch {
		b.pool.Put(buf)
	}

	return err
}

// Close vacía los paquetes pendientes y marca el batcher como cerrado
func (b *PacketBatcher) Close() error {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()

	return b.Flush()
}
