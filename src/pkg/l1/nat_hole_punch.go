package l1

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	HolePunchMagic   uint16 = 0x4850 // "HP"
	HolePunchProbeOp uint8  = 0x01
	HolePunchAckOp   uint8  = 0x02
	DefaultPunchRetries     = 5
	DefaultPunchGap         = 30 * time.Millisecond
)

// CandidateEndpoint almacena los candidatos de transporte locales y reflexivos (ICE-Lite)
type CandidateEndpoint struct {
	LocalAddr     *net.UDPAddr `json:"local_addr"`
	ReflexiveAddr *net.UDPAddr `json:"reflexive_addr"`
	RelayAddr     *net.UDPAddr `json:"relay_addr,omitempty"`
}

// HolePunchMessage modela la trama binaria ultraligera de perforación NAT (Zero-Overhead)
type HolePunchMessage struct {
	Magic     uint16   `json:"magic"`
	OpCode    uint8    `json:"op_code"`
	SessionID [16]byte `json:"session_id"`
	Sequence  uint32   `json:"sequence"`
}

// Encode serializa la trama de perforación en exactamente 23 bytes deterministas
func (m *HolePunchMessage) Encode() []byte {
	buf := make([]byte, 23)
	binary.BigEndian.PutUint16(buf[0:2], m.Magic)
	buf[2] = m.OpCode
	copy(buf[3:19], m.SessionID[:])
	binary.BigEndian.PutUint32(buf[19:23], m.Sequence)
	return buf
}

// DecodeHolePunchMessage deserializa y valida la trama de perforación NAT
func DecodeHolePunchMessage(b []byte) (*HolePunchMessage, error) {
	if len(b) < 23 {
		return nil, errors.New("trama de perforación truncada (<23 bytes)")
	}
	magic := binary.BigEndian.Uint16(b[0:2])
	if magic != HolePunchMagic {
		return nil, fmt.Errorf("magic inválido en trama de perforación: 0x%04x", magic)
	}

	var sessionID [16]byte
	copy(sessionID[:], b[3:19])

	return &HolePunchMessage{
		Magic:     magic,
		OpCode:    b[2],
		SessionID: sessionID,
		Sequence:  binary.BigEndian.Uint32(b[19:23]),
	}, nil
}

// NATHolePunchEngine orquesta la perforación simétrica UDP entre pares
type NATHolePunchEngine struct {
	mu       sync.RWMutex
	conn     *net.UDPConn
	sessions map[[16]byte]chan *net.UDPAddr
}

// NewNATHolePunchEngine inicializa el motor de perforación de puertos
func NewNATHolePunchEngine(conn *net.UDPConn) *NATHolePunchEngine {
	return &NATHolePunchEngine{
		conn:     conn,
		sessions: make(map[[16]byte]chan *net.UDPAddr),
	}
}

// ExecutePunch envía ráfagas coordinadas hacia el par remoto hasta perforar el cono NAT
func (e *NATHolePunchEngine) ExecutePunch(
	targetAddr *net.UDPAddr,
	timeout time.Duration,
) (*net.UDPAddr, error) {
	if targetAddr == nil {
		return nil, errors.New("dirección remota de perforación no puede ser nula")
	}

	var sessionID [16]byte
	if _, err := rand.Read(sessionID[:]); err != nil {
		return nil, err
	}

	doneChan := make(chan *net.UDPAddr, 1)
	e.mu.Lock()
	e.sessions[sessionID] = doneChan
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.sessions, sessionID)
		e.mu.Unlock()
	}()

	probeMsg := &HolePunchMessage{
		Magic:     HolePunchMagic,
		OpCode:    HolePunchProbeOp,
		SessionID: sessionID,
		Sequence:  1,
	}
	wire := probeMsg.Encode()

	ticker := time.NewTicker(DefaultPunchGap)
	defer ticker.Stop()

	deadline := time.After(timeout)
	attempts := 0

	for {
		select {
		case remoteSuccess := <-doneChan:
			// Enviar confirmación final (ACK)
			ackMsg := &HolePunchMessage{
				Magic:     HolePunchMagic,
				OpCode:    HolePunchAckOp,
				SessionID: sessionID,
				Sequence:  uint32(attempts + 1),
			}
			_, _ = e.conn.WriteToUDP(ackMsg.Encode(), remoteSuccess)
			return remoteSuccess, nil

		case <-deadline:
			return nil, fmt.Errorf("timeout perforando NAT hacia %s tras %d intentos", targetAddr, attempts)

		case <-ticker.C:
			if attempts >= DefaultPunchRetries*4 {
				return nil, fmt.Errorf("límite de ráfagas alcanzado hacia %s", targetAddr)
			}
			attempts++
			probeMsg.Sequence = uint32(attempts)
			_, _ = e.conn.WriteToUDP(wire, targetAddr)
		}
	}
}

// HandleIncomingPacket procesa paquetes de perforación NAT entrantes
func (e *NATHolePunchEngine) HandleIncomingPacket(data []byte, from *net.UDPAddr) bool {
	if len(data) < 23 || binary.BigEndian.Uint16(data[0:2]) != HolePunchMagic {
		return false
	}

	msg, err := DecodeHolePunchMessage(data)
	if err != nil {
		return false
	}

	e.mu.RLock()
	ch, exists := e.sessions[msg.SessionID]
	e.mu.RUnlock()

	if msg.OpCode == HolePunchProbeOp {
		// Responder inmediatamente con ACK al origen del probe
		ackMsg := &HolePunchMessage{
			Magic:     HolePunchMagic,
			OpCode:    HolePunchAckOp,
			SessionID: msg.SessionID,
			Sequence:  msg.Sequence,
		}
		_, _ = e.conn.WriteToUDP(ackMsg.Encode(), from)
	}

	if exists && (msg.OpCode == HolePunchAckOp || msg.OpCode == HolePunchProbeOp) {
		select {
		case ch <- from:
		default:
		}
	}

	return true
}
