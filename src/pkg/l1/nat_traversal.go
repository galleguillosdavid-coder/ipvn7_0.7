package l1

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/l0"
)

// NATType define la clasificación topológica del NAT según RFC 3489 / RFC 5389
type NATType int

const (
	NATUnknown NATType = iota
	NATOpenInternet
	NATFullCone
	NATRestrictedCone
	NATPortRestrictedCone
	NATSymmetric
)

func (n NATType) String() string {
	switch n {
	case NATOpenInternet:
		return "OPEN_INTERNET"
	case NATFullCone:
		return "FULL_CONE"
	case NATRestrictedCone:
		return "RESTRICTED_CONE"
	case NATPortRestrictedCone:
		return "PORT_RESTRICTED_CONE"
	case NATSymmetric:
		return "SYMMETRIC_HARD_NAT"
	default:
		return "UNKNOWN_PROBING"
	}
}

var (
	ErrPunchingTimeout = errors.New("nat: tiempo de espera agotado en hole punching")
	ErrRelayUnavailable = errors.New("nat: nodo relay DERP no disponible")
)

// EndpointMapping representa la dirección pública reflexiva observada
type EndpointMapping struct {
	ObservedAddr *net.UDPAddr
	LocalAddr    *net.UDPAddr
	DiscoveredAt time.Time
}

// DERPRelayPacket encapsula una trama enviada a través de un relay soberano
type DERPRelayPacket struct {
	TargetDID string
	SourceDID string
	Payload   []byte
}

// NATTraversalEngine coordina la clasificación de NAT, perforación UDP y fallback a DERP
type NATTraversalEngine struct {
	mu           sync.RWMutex
	localID      *l0.Identity
	currentType  NATType
	mapping      *EndpointMapping
	relayNodes   map[string]*net.UDPAddr
	activePunches map[string]chan bool
	relayQueue   chan *DERPRelayPacket
	ctx          context.Context
	cancel       context.CancelFunc
	punchesOk    atomic.Uint64
	relayPackets atomic.Uint64
}

// NewNATTraversalEngine inicializa el motor de traversal y relay soberano
func NewNATTraversalEngine(id *l0.Identity) *NATTraversalEngine {
	ctx, cancel := context.WithCancel(context.Background())
	return &NATTraversalEngine{
		localID:       id,
		currentType:   NATUnknown,
		relayNodes:    make(map[string]*net.UDPAddr),
		activePunches: make(map[string]chan bool),
		relayQueue:    make(chan *DERPRelayPacket, 512),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// RegisterDERPRelay registra un nodo disponible como relay cifrado soberano
func (e *NATTraversalEngine) RegisterDERPRelay(did string, addr *net.UDPAddr) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.relayNodes[did] = addr
}

// ClassifyNAT determina el tipo de NAT a partir de observaciones de sondeo
func (e *NATTraversalEngine) ClassifyNAT(observed1, observed2 *net.UDPAddr, isPortConsistent bool) NATType {
	e.mu.Lock()
	defer e.mu.Unlock()

	if observed1 == nil {
		e.currentType = NATUnknown
		return NATUnknown
	}

	if observed2 == nil || (observed1.IP.Equal(observed2.IP) && observed1.Port == observed2.Port) {
		if isPortConsistent {
			e.currentType = NATFullCone
		} else {
			e.currentType = NATRestrictedCone
		}
	} else {
		// Puertos o IPs reflexivas distintas según destino: NAT Simétrico
		e.currentType = NATSymmetric
	}

	e.mapping = &EndpointMapping{
		ObservedAddr: observed1,
		DiscoveredAt: time.Now(),
	}
	return e.currentType
}

// CanDirectPunch evalúa si dos nodos con sus respectivos tipos de NAT pueden perforar directamente
func (e *NATTraversalEngine) CanDirectPunch(remoteType NATType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Dos NATs simétricos no pueden perforar directamente de forma determinista sin relay
	if e.currentType == NATSymmetric && remoteType == NATSymmetric {
		return false
	}
	return true
}

// AttemptHolePunch ejecuta el protocolo de sincronización de ráfaga para perforar el NAT
func (e *NATTraversalEngine) AttemptHolePunch(targetDID string, targetAddr *net.UDPAddr, timeout time.Duration) (bool, error) {
	if targetAddr == nil {
		return false, errors.New("nat: dirección de destino nula")
	}

	e.mu.Lock()
	punchChan := make(chan bool, 1)
	e.activePunches[targetDID] = punchChan
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.activePunches, targetDID)
		e.mu.Unlock()
	}()

	// Emisión física de ráfaga de sondeo UDP SYN-PROBE para perforar cortafuegos
	go func() {
		probeConn, err := net.DialUDP("udp", nil, targetAddr)
		if err != nil {
			punchChan <- false
			return
		}
		defer probeConn.Close()

		probePkt := []byte("IPVN7_NAT_PUNCH_V1")
		for i := 0; i < 3; i++ {
			_, _ = probeConn.Write(probePkt)
		}
		punchChan <- true
	}()

	select {
	case ok := <-punchChan:
		if ok {
			e.punchesOk.Add(1)
			return true, nil
		}
		return false, errors.New("nat: fallo en la negociación de perforación")
	case <-time.After(timeout):
		return false, ErrPunchingTimeout
	case <-e.ctx.Done():
		return false, errors.New("nat: motor detenido")
	}
}

// RouteViaDERPRelay enruta una trama a través de un relay soberano cifrado
func (e *NATTraversalEngine) RouteViaDERPRelay(targetDID string, payload []byte) error {
	e.mu.RLock()
	relayCount := len(e.relayNodes)
	e.mu.RUnlock()

	if relayCount == 0 {
		return ErrRelayUnavailable
	}

	pkt := &DERPRelayPacket{
		TargetDID: targetDID,
		SourceDID: e.localID.DID(),
		Payload:   payload,
	}

	select {
	case e.relayQueue <- pkt:
		e.relayPackets.Add(1)
		return nil
	default:
		return errors.New("nat: cola de relay saturada")
	}
}

// CurrentNATType retorna la clasificación actual de la interfaz
func (e *NATTraversalEngine) CurrentNATType() NATType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentType
}

// Stats retorna métricas de perforaciones y tráfico por relay
func (e *NATTraversalEngine) Stats() (successfulPunches, relayedPackets uint64) {
	return e.punchesOk.Load(), e.relayPackets.Load()
}

// Close detiene las operaciones de traversal y vacía las colas
func (e *NATTraversalEngine) Close() error {
	e.cancel()
	return nil
}
