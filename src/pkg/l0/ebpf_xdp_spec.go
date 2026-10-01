package l0

import (
	"encoding/binary"
	"errors"
	"sync/atomic"
)

const (
	XDPActionAborted  uint32 = 0
	XDPActionDrop     uint32 = 1
	XDPActionPass     uint32 = 2
	XDPActionTX       uint32 = 3
	XDPActionRedirect uint32 = 4

	IPVN7MagicBytes uint32 = 0x49503756 // "IP7V"
	IPVN7UDPMeshPort       = 7777
)

var (
	ErrPacketTooShort   = errors.New("xdp: paquete inferior al tamaño mínimo de trama")
	ErrInvalidEtherType = errors.New("xdp: tipo de protocolo ethernet no admitido")
)

// XDPFilterStats registra las decisiones tomadas por el conmutador acelerado
type XDPFilterStats struct {
	TotalProcessed atomic.Uint64
	Passed         atomic.Uint64
	Dropped        atomic.Uint64
	Transmitted    atomic.Uint64
}

// XDPKernelSimulator emula la ejecución estricta del verificador y programa XDP en userspace
type XDPKernelSimulator struct {
	stats XDPFilterStats
}

// NewXDPKernelSimulator inicializa el simulador formal de instrucciones de conmutación
func NewXDPKernelSimulator() *XDPKernelSimulator {
	return &XDPKernelSimulator{}
}

// ProcessFrame evalúa una trama Ethernet completa replicando el verifier eBPF de Linux
func (s *XDPKernelSimulator) ProcessFrame(frame []byte) (uint32, error) {
	s.stats.TotalProcessed.Add(1)

	// Cabecera Ethernet mínima: 14 bytes
	if len(frame) < 14 {
		s.stats.Dropped.Add(1)
		return XDPActionDrop, ErrPacketTooShort
	}

	etherType := binary.BigEndian.Uint16(frame[12:14])
	if etherType != 0x0800 { // IPv4
		s.stats.Passed.Add(1)
		return XDPActionPass, nil
	}

	ipHeader := frame[14:]
	if len(ipHeader) < 20 {
		s.stats.Dropped.Add(1)
		return XDPActionDrop, ErrPacketTooShort
	}

	protocol := ipHeader[9]
	if protocol != 17 { // UDP
		s.stats.Passed.Add(1)
		return XDPActionPass, nil
	}

	ihl := int(ipHeader[0]&0x0F) * 4
	if len(ipHeader) < ihl+8 {
		s.stats.Dropped.Add(1)
		return XDPActionDrop, ErrPacketTooShort
	}

	udpHeader := ipHeader[ihl:]
	destPort := binary.BigEndian.Uint16(udpHeader[2:4])

	if destPort != IPVN7UDPMeshPort {
		s.stats.Passed.Add(1)
		return XDPActionPass, nil
	}

	payload := udpHeader[8:]
	if len(payload) >= 4 {
		magic := binary.BigEndian.Uint32(payload[:4])
		if magic == IPVN7MagicBytes {
			// Trama IPVN7 certificada
			s.stats.Passed.Add(1)
			return XDPActionPass, nil
		}
	}

	// Datagrama dirigido al puerto IPVN7 sin cabecera válida -> DROP inmediato
	s.stats.Dropped.Add(1)
	return XDPActionDrop, nil
}

// Stats retorna el desglose de decisiones de conmutación acelerada
func (s *XDPKernelSimulator) Stats() (processed, passed, dropped, tx uint64) {
	return s.stats.TotalProcessed.Load(),
		s.stats.Passed.Load(),
		s.stats.Dropped.Load(),
		s.stats.Transmitted.Load()
}
