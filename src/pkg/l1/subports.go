// Package l1 implementa la multiplexación de sub-puertos lógicos virtuales (SubPort uint16)
// rescatada de Ipv7-8 (transport/subports.go) y formalizada para ipvn7 v0.7.
// Permite multiplexar 65,536 canales virtuales sobre el puerto físico UDP de malla (7777 / 7001).
package l1

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

// SubPort representa un canal lógico virtual dentro del protocolo ipvn7
type SubPort uint16

const (
	// Canales lógicos canónicos
	SubPortDefault   SubPort = 0
	SubPortChat      SubPort = 1 // Mensajería y chat E2EE soberano
	SubPortTelemetry SubPort = 2 // Telemetría y diagnóstico en vivo
	SubPortDAGStore  SubPort = 3 // Almacén inmutable DAG y DTN
	SubPortBenchmark SubPort = 4 // Pruebas empíricas de saturación y jitter RFC 3550
	SubPortVPNProxy  SubPort = 5 // Túnel de navegación web y SOCKS5

	// Rango para microservicios y satélites dinámicos del Smart Gateway
	SubPortDynamicMin SubPort = 1000
	SubPortDynamicMax SubPort = 65535
)

// SubPortHandler define la función receptora para un sub-puerto específico
type SubPortHandler func(srcDID string, payload []byte)

// SubPortMux orquesta la entrega de datagramas a los manejadores registrados
type SubPortMux struct {
	mu       sync.RWMutex
	handlers map[SubPort]SubPortHandler
}

// NewSubPortMux inicializa el multiplexor de canales lógicos
func NewSubPortMux() *SubPortMux {
	return &SubPortMux{
		handlers: make(map[SubPort]SubPortHandler),
	}
}

// Register asocia un manejador a un sub-puerto específico
func (m *SubPortMux) Register(sp SubPort, handler SubPortHandler) error {
	if handler == nil {
		return errors.New("subport: handler no puede ser nulo")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[sp] = handler
	return nil
}

// Unregister elimina la suscripción a un sub-puerto
func (m *SubPortMux) Unregister(sp SubPort) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.handlers, sp)
}

// Dispatch entrega el payload al manejador suscrito devolviendo true si fue atendido
func (m *SubPortMux) Dispatch(sp SubPort, srcDID string, payload []byte) bool {
	m.mu.RLock()
	handler, exists := m.handlers[sp]
	m.mu.RUnlock()

	if !exists || handler == nil {
		return false
	}

	handler(srcDID, payload)
	return true
}

// EncodeSubPortFrame antepone la cabecera canónica de sub-puerto de 2 bytes
func EncodeSubPortFrame(sp SubPort, payload []byte) []byte {
	frame := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(frame[0:2], uint16(sp))
	copy(frame[2:], payload)
	return frame
}

// DecodeSubPortFrame extrae el sub-puerto y la carga útil de la trama
func DecodeSubPortFrame(frame []byte) (SubPort, []byte, error) {
	if len(frame) < 2 {
		return 0, nil, fmt.Errorf("subport: trama demasiado corta (%d bytes, min 2)", len(frame))
	}
	sp := SubPort(binary.BigEndian.Uint16(frame[0:2]))
	return sp, frame[2:], nil
}

// RegisteredSubPorts lista todos los canales virtuales actualmente activos
func (m *SubPortMux) RegisteredSubPorts() []SubPort {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]SubPort, 0, len(m.handlers))
	for sp := range m.handlers {
		res = append(res, sp)
	}
	return res
}
