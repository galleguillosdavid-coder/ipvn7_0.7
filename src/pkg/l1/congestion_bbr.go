// Package l1 implementa el controlador de congestión BBR (Bottleneck Bandwidth and RTT)
// para datagramas UDP en la malla ipvn7 v0.7.
// Modela el cuello de botella físico de la ruta sin depender de la pérdida de paquetes.
package l1

import (
	"math"
	"sync"
	"time"
)

// BBRState estados operativos del controlador de congestión
type BBRState string

const (
	BBRStateStartup  BBRState = "STARTUP"
	BBRStateDrain    BBRState = "DRAIN"
	BBRStateProbeBW  BBRState = "PROBE_BW"
	BBRStateProbeRTT BBRState = "PROBE_RTT"
)

// BBRController modela el ancho de banda y RTT físico de la ruta
type BBRController struct {
	mu           sync.RWMutex
	state        BBRState
	btlBw        float64       // Máxima tasa de entrega observada (bytes/segundo)
	rtProp       time.Duration // Mínimo retardo de ida y vuelta observado
	rtPropStamp  time.Time     // Marca temporal del último mínimo rtProp
	pacingGain   float64       // Factor de amplificación de despacho
	cwndGain     float64       // Factor de amplificación de ventana de congestión
	pacingRate   float64       // Bytes por segundo despachados
	cwnd         int           // Ventana de congestión en bytes
	cycleIndex   int
	lastCycle    time.Time
}

var (
	// Ciclo canónico de ganancias para exploración de ancho de banda en PROBE_BW
	probeBWGains = []float64{1.25, 0.75, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
)

// NewBBRController inicializa el controlador BBR en modo STARTUP
func NewBBRController(initialBwBps float64, initialRTT time.Duration) *BBRController {
	if initialBwBps <= 0 {
		initialBwBps = 1000000 // 1 MB/s por defecto
	}
	if initialRTT <= 0 {
		initialRTT = 20 * time.Millisecond
	}

	b := &BBRController{
		state:       BBRStateStartup,
		btlBw:       initialBwBps,
		rtProp:      initialRTT,
		rtPropStamp: time.Now(),
		pacingGain:  2.885, // 2/ln(2) aceleración estándar BBR
		cwndGain:    2.0,
		lastCycle:   time.Now(),
	}
	b.recalculatePacingAndCWND()
	return b
}

// OnAck actualiza el modelo BBR al recibir confirmación de un datagrama o bloque
func (b *BBRController) OnAck(deliveryRateBps float64, rtt time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	// 1. Actualizar Bottleneck Bandwidth (filtro de máximo)
	if deliveryRateBps > b.btlBw {
		b.btlBw = deliveryRateBps
	}

	// 2. Actualizar RTprop (filtro de mínimo con ventana de 10s)
	if rtt > 0 && (rtt < b.rtProp || now.Sub(b.rtPropStamp) > 10*time.Second) {
		b.rtProp = rtt
		b.rtPropStamp = now
	}

	// 3. Transición de estados FSM
	b.updateState(now)

	// 4. Recalcular pacing y cwnd
	b.recalculatePacingAndCWND()
}

func (b *BBRController) updateState(now time.Time) {
	switch b.state {
	case BBRStateStartup:
		// Transitar a DRAIN si el ancho de banda se estabiliza (o simulado tras ráfaga)
		if now.Sub(b.lastCycle) > 500*time.Millisecond {
			b.state = BBRStateDrain
			b.pacingGain = 1.0 / 2.885
			b.cwndGain = 2.0
			b.lastCycle = now
		}
	case BBRStateDrain:
		// Transitar a PROBE_BW cuando la cola se haya vaciado
		if now.Sub(b.lastCycle) > 300*time.Millisecond {
			b.state = BBRStateProbeBW
			b.cycleIndex = 0
			b.pacingGain = probeBWGains[b.cycleIndex]
			b.cwndGain = 2.0
			b.lastCycle = now
		}
	case BBRStateProbeBW:
		// Rotar ganancias del ciclo cada RTprop o intervalo de 100ms
		interval := b.rtProp
		if interval < 50*time.Millisecond {
			interval = 50 * time.Millisecond
		}
		if now.Sub(b.lastCycle) >= interval {
			b.cycleIndex = (b.cycleIndex + 1) % len(probeBWGains)
			b.pacingGain = probeBWGains[b.cycleIndex]
			b.lastCycle = now
		}

		// Si pasaron más de 10 segundos sin actualizar RTprop, entrar a PROBE_RTT
		if now.Sub(b.rtPropStamp) > 10*time.Second {
			b.state = BBRStateProbeRTT
			b.pacingGain = 1.0
			b.cwndGain = 1.0
			b.lastCycle = now
		}
	case BBRStateProbeRTT:
		if now.Sub(b.lastCycle) > 200*time.Millisecond {
			b.state = BBRStateProbeBW
			b.cycleIndex = 0
			b.pacingGain = probeBWGains[0]
			b.cwndGain = 2.0
			b.rtPropStamp = now
			b.lastCycle = now
		}
	}
}

func (b *BBRController) recalculatePacingAndCWND() {
	// BDP = BtlBw * RTprop
	rttSec := b.rtProp.Seconds()
	bdp := b.btlBw * rttSec

	// CWND = cwndGain * BDP (mínimo 4 datagramas MTU 1280B = 5120 bytes)
	targetCWND := int(math.Round(b.cwndGain * bdp))
	if targetCWND < 5120 {
		targetCWND = 5120
	}
	b.cwnd = targetCWND

	// PacingRate = pacingGain * BtlBw
	b.pacingRate = b.pacingGain * b.btlBw
}

// CanSend determina si hay presupuesto de ventana para emitir otro datagrama
func (b *BBRController) CanSend(inflightBytes int) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return inflightBytes < b.cwnd
}

// Metrics retorna instantánea de las métricas físicas estimadas
func (b *BBRController) Metrics() (BBRState, float64, time.Duration, float64, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state, b.btlBw, b.rtProp, b.pacingRate, b.cwnd
}
