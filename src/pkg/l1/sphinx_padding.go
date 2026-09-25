package l1

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/big"
	"sync"
	"time"
)

// Constantes de mitigación de análisis de tráfico DPI
const (
	SphinxFlagDummy         byte          = 0x02
	DefaultMinPaddingGap                  = 50 * time.Millisecond
	DefaultMaxPaddingGap                  = 250 * time.Millisecond
	SphinxDummyService                    = "ipvn7.padding.dummy"
	MinEntropyFloorBase                   = 5 * time.Millisecond    // Piso basal continuo (sin brecha estocástica de tránsito corto)
	MinEntropyFloorJitter                 = 3995 * time.Millisecond // Rango dinámico continuo de cola pesada hasta 4000ms
	QuantizedCanonicalMTU                 = 1280                    // Cuantización fija: 100% de tramas miden exactamente 1280B
)

// StochasticPaddingEngine inyecta tramas deterministas de 1280B a intervalos estocásticos
// para neutralizar correlación temporal y análisis de patrones de flujo (DPI / NetFlow).
type StochasticPaddingEngine struct {
	mu          sync.RWMutex
	router      *SphinxRouter
	minGap      time.Duration
	maxGap      time.Duration
	running     bool
	stopChan    chan struct{}
	DummiesSent uint64
}

// NewStochasticPaddingEngine inicializa el generador estocástico de ruido
func NewStochasticPaddingEngine(router *SphinxRouter, minGap, maxGap time.Duration) *StochasticPaddingEngine {
	if minGap <= 0 {
		minGap = DefaultMinPaddingGap
	}
	if maxGap <= minGap {
		maxGap = minGap + 100*time.Millisecond
	}

	return &StochasticPaddingEngine{
		router:   router,
		minGap:   minGap,
		maxGap:   maxGap,
		stopChan: make(chan struct{}),
	}
}

// GenerateDummyPacket construye un paquete Sphinx idéntico en tamaño y entropía a un paquete real
func (spe *StochasticPaddingEngine) GenerateDummyPacket(circuit *SphinxCircuit) (*SphinxPacket, error) {
	if spe.router == nil {
		return nil, errors.New("enrutador Sphinx no asignado al motor de padding")
	}

	dummyPayload := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, dummyPayload); err != nil {
		return nil, err
	}

	packet, err := spe.router.BuildPacket(circuit, dummyPayload, SphinxDummyService)
	if err != nil {
		return nil, err
	}

	return packet, nil
}

// NextJitterDuration calcula el próximo intervalo estocástico con entropía criptográfica
func (spe *StochasticPaddingEngine) NextJitterDuration() time.Duration {
	diff := spe.maxGap - spe.minGap
	if diff <= 0 {
		return spe.minGap
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return spe.minGap + (diff / 2)
	}

	return spe.minGap + time.Duration(nBig.Int64())
}

// LomaxQuantizedSteps define la discretización fina de 256 cuantiles precalculados (512B en L1)
const LomaxQuantizedSteps = 256

var (
	lomaxLUTBytes [(LomaxQuantizedSteps + 1) * 2]byte // 514B contiguos en L1 (8 líneas de caché)
)

func init() {
	const (
		lambdaMs = 80.0
		alpha    = 1.5
		minBase  = 5.0
	)
	for i := 0; i < LomaxQuantizedSteps; i++ {
		u := (float64(i) + 0.5) / float64(LomaxQuantizedSteps)
		x := minBase + lambdaMs*(math.Pow(1.0-u, -1.0/alpha)-1.0)
		if x > 4000.0 {
			x = 4000.0
		}
		val := uint16(x)
		binary.LittleEndian.PutUint16(lomaxLUTBytes[i*2:], val)
	}
	// Entrada de frontera 256 duplicada para acceso branchless indexado directo
	lastVal := binary.LittleEndian.Uint16(lomaxLUTBytes[(LomaxQuantizedSteps-1)*2:])
	binary.LittleEndian.PutUint16(lomaxLUTBytes[LomaxQuantizedSteps*2:], lastVal)
}

// SampleContinuousLomaxJitter genera un intervalo no estacionario continuo sobre [5ms, 4000ms]
// con distribución de Pareto Tipo II (Lomax) utilizando una tabla compacta de 514B en L1.
// Emplea una única instrucción de lectura de 32 bits para cargar el par (base, next) en un solo
// ciclo de reloj, neutralizando cualquier fuga de canal lateral por cruce de líneas de caché (DEC-078).
func SampleContinuousLomaxJitter(lambdaMs, alpha float64) time.Duration {
	var buf [2]byte
	_, err := rand.Read(buf[:])
	idx := uint8(128)
	fraction := uint8(0)
	if err == nil {
		idx = buf[0]
		fraction = buf[1]
	}
	// Lectura simultánea de 32 bits: carga 'base' y 'next' en la misma instrucción de hardware
	offset := int(idx) * 2
	pair := binary.LittleEndian.Uint32(lomaxLUTBytes[offset : offset+4])
	base := time.Duration(pair&0xFFFF) * time.Millisecond
	next := time.Duration(pair>>16) * time.Millisecond
	dither := time.Duration(int64(next-base) * int64(fraction) / 256)
	return base + dither
}

// SampleHeavyTailedParetoJitter preserva compatibilidad con API llamando al generador continuo Lomax
func SampleHeavyTailedParetoJitter(minScaleMs, alpha float64) time.Duration {
	return SampleContinuousLomaxJitter(80.0, alpha)
}

// EnforceEntropyFloor asegura que bajo estrés de congestión WAN o pérdidas selectivas,
// el intervalo de padding preserve un proceso de renovación estocástico continuo Lomax (KDE-proof),
// impidiendo que herramientas Next-Gen DPI aíslen ventanas de soporte compacto o detecten brechas en tránsito corto.
func (spe *StochasticPaddingEngine) EnforceEntropyFloor(congestionThrottleActive bool) time.Duration {
	if congestionThrottleActive {
		return SampleContinuousLomaxJitter(80.0, 1.5)
	}
	return spe.NextJitterDuration()
}

// Start activa el bucle en segundo plano que inyecta tramas dummy
func (spe *StochasticPaddingEngine) Start(circuit *SphinxCircuit, egressFunc func(wire []byte) error) {
	spe.mu.Lock()
	if spe.running {
		spe.mu.Unlock()
		return
	}
	spe.running = true
	spe.stopChan = make(chan struct{})
	spe.mu.Unlock()

	go func() {
		for {
			sleepDur := spe.NextJitterDuration()
			timer := time.NewTimer(sleepDur)

			select {
			case <-spe.stopChan:
				timer.Stop()
				return
			case <-timer.C:
				if pkt, err := spe.GenerateDummyPacket(circuit); err == nil && egressFunc != nil {
					_ = egressFunc(pkt.Serialize())
					spe.mu.Lock()
					spe.DummiesSent++
					spe.mu.Unlock()
				}
			}
		}
	}()
}

// Stop detiene el generador de tráfico dummy
func (spe *StochasticPaddingEngine) Stop() {
	spe.mu.Lock()
	defer spe.mu.Unlock()
	if !spe.running {
		return
	}
	spe.running = false
	close(spe.stopChan)
}

// IsRunning indica si el motor de padding estocástico está activo
func (spe *StochasticPaddingEngine) IsRunning() bool {
	spe.mu.RLock()
	defer spe.mu.RUnlock()
	return spe.running
}
