// Package l1 implementa el túnel MASQUE (RFC 9298 CONNECT-UDP) para transporte
// de datagramas soberanos ipvn7 sobre HTTPS estándar (puerto 443) con Capsule Protocol (RFC 9297).
// Permite que nodos en entornos corporativos hostiles con UDP bloqueado mantengan la malla.
package l1

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

const (
	// MASQUECapsuleDatagram identifica el tipo de cápsula para datagramas según RFC 9298
	MASQUECapsuleDatagram = 0x00

	// MASQUEDefaultContextID define el ID de contexto para datagramas directos sin compresión
	MASQUEDefaultContextID = 0x00

	// Encabezado Capsule-Protocol requerido por RFC 9297
	MASQUECapsuleProtocolHeader = "Capsule-Protocol: ?1"
)

var (
	ErrCapsuleTruncated = errors.New("cápsula MASQUE truncada o incompleta")
	ErrInvalidContextID = errors.New("id de contexto MASQUE no soportado")
)

// MASQUETunnelConfig define los parámetros del túnel CONNECT-UDP
type MASQUETunnelConfig struct {
	TargetHost       string `json:"target_host"`
	TargetPort       int    `json:"target_port"`
	MASQUEProxyAddr  string `json:"masque_proxy_addr"`
	UseCapsuleProto  bool   `json:"use_capsule_proto"`
}

// MASQUETunnelSession gestiona una sesión de transporte CONNECT-UDP
type MASQUETunnelSession struct {
	mu           sync.RWMutex
	Config       MASQUETunnelConfig
	isConnected  atomic.Bool
	bytesSent    atomic.Uint64
	bytesRecv    atomic.Uint64
	framesSent   atomic.Uint64
	framesRecv   atomic.Uint64
}

// NewMASQUETunnelSession crea una nueva sesión de túnel MASQUE
func NewMASQUETunnelSession(cfg MASQUETunnelConfig) *MASQUETunnelSession {
	return &MASQUETunnelSession{
		Config: cfg,
	}
}

// BuildConnectUDPRequest genera la petición HTTP estándar conforme a RFC 9298
func BuildConnectUDPRequest(targetHost string, targetPort int, proxyHost string) string {
	return fmt.Sprintf("CONNECT-UDP /well-known/masque/udp/%s/%d/ HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Capsule-Protocol: ?1\r\n"+
		"User-Agent: ipvn7-masque-tunnel/1.0\r\n\r\n", targetHost, targetPort, proxyHost)
}

// EncodeVarintRFC9000 codifica un entero en formato variable RFC 9000 / RFC 9297
func EncodeVarintRFC9000(val uint64) []byte {
	if val <= 63 {
		return []byte{byte(val)}
	} else if val <= 16383 {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(val)|0x4000)
		return buf
	} else if val <= 1073741823 {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(val)|0x80000000)
		return buf
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, val|0xC000000000000000)
	return buf
}

// DecodeVarintRFC9000 decodifica un entero variable RFC 9000 / RFC 9297 retornando bytes leídos
func DecodeVarintRFC9000(data []byte) (uint64, int, error) {
	if len(data) == 0 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	first := data[0]
	prefix := first >> 6
	switch prefix {
	case 0:
		return uint64(first & 0x3F), 1, nil
	case 1:
		if len(data) < 2 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val := binary.BigEndian.Uint16(data[:2]) & 0x3FFF
		return uint64(val), 2, nil
	case 2:
		if len(data) < 4 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val := binary.BigEndian.Uint32(data[:4]) & 0x3FFFFFFF
		return uint64(val), 4, nil
	case 3:
		if len(data) < 8 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val := binary.BigEndian.Uint64(data[:8]) & 0x3FFFFFFFFFFFFFFF
		return val, 8, nil
	}
	return 0, 0, errors.New("formato varint inválido")
}

// WrapDatagramCapsule encapsula un datagrama UDP en formato Capsule HTTP Datagram (RFC 9298)
// Formato: [Context ID: Varint] + [Payload: []byte]
func WrapDatagramCapsule(payload []byte) []byte {
	ctxIDBytes := EncodeVarintRFC9000(MASQUEDefaultContextID)
	out := make([]byte, len(ctxIDBytes)+len(payload))
	copy(out, ctxIDBytes)
	copy(out[len(ctxIDBytes):], payload)
	return out
}

// UnwrapDatagramCapsule extrae la carga útil de una cápsula HTTP Datagram (RFC 9298)
func UnwrapDatagramCapsule(data []byte) ([]byte, error) {
	ctxID, n, err := DecodeVarintRFC9000(data)
	if err != nil {
		return nil, ErrCapsuleTruncated
	}
	if ctxID != MASQUEDefaultContextID {
		return nil, fmt.Errorf("%w: recibido %d", ErrInvalidContextID, ctxID)
	}
	return data[n:], nil
}

// SendDatagram procesa y contabiliza la emisión de una trama sobre el túnel MASQUE
func (s *MASQUETunnelSession) SendDatagram(payload []byte) []byte {
	enc := WrapDatagramCapsule(payload)
	s.bytesSent.Add(uint64(len(enc)))
	s.framesSent.Add(1)
	return enc
}

// ReceiveDatagram desempaqueta y contabiliza una trama recibida sobre el túnel MASQUE
func (s *MASQUETunnelSession) ReceiveDatagram(data []byte) ([]byte, error) {
	raw, err := UnwrapDatagramCapsule(data)
	if err != nil {
		return nil, err
	}
	s.bytesRecv.Add(uint64(len(data)))
	s.framesRecv.Add(1)
	return raw, nil
}

// GetStats retorna las métricas de transmisión de la sesión MASQUE
func (s *MASQUETunnelSession) GetStats() (sentBytes, recvBytes, sentFrames, recvFrames uint64) {
	return s.bytesSent.Load(), s.bytesRecv.Load(), s.framesSent.Load(), s.framesRecv.Load()
}
