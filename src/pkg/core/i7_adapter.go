package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	ErrAdapterClosed      = errors.New("i7-adapter: conexión UDP cerrada")
	ErrCorruptFrame       = errors.New("i7-adapter: trama de contenedor incompleta o corrupta")
	ErrUnsupportedVersion = errors.New("i7-adapter: versión de protocolo I7 incompatible")
)

// I7UDPAdapter desacopla el Núcleo Mínimo I7 del socket físico de red
type I7UDPAdapter struct {
	mu     sync.RWMutex
	conn   *net.UDPConn
	closed bool
}

// NewI7UDPAdapter inicializa un adaptador con un socket UDP real
func NewI7UDPAdapter(conn *net.UDPConn) *I7UDPAdapter {
	return &I7UDPAdapter{
		conn: conn,
	}
}

// ListenI7UDP abre un socket UDP local real en la dirección especificada
func ListenI7UDP(address string) (*I7UDPAdapter, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("resolución de dirección falló: %w", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("apertura de socket UDP falló: %w", err)
	}
	return NewI7UDPAdapter(conn), nil
}

// LocalAddr retorna la dirección de escucha del socket
func (a *I7UDPAdapter) LocalAddr() net.Addr {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.conn == nil {
		return nil
	}
	return a.conn.LocalAddr()
}

// SendContainer serializa y emite un I7Container físicamente hacia la dirección remota
func (a *I7UDPAdapter) SendContainer(container *I7Container, remoteAddr *net.UDPAddr) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.closed || a.conn == nil {
		return ErrAdapterClosed
	}
	if container == nil {
		return errors.New("i7-adapter: contenedor nulo")
	}

	wireBytes, err := SerializeContainer(container)
	if err != nil {
		return err
	}

	if !ValidateMTU(len(wireBytes)) {
		return fmt.Errorf("%w: trama serializada de %d bytes", ErrPayloadTooLarge, len(wireBytes))
	}

	_, err = a.conn.WriteToUDP(wireBytes, remoteAddr)
	if err != nil {
		return fmt.Errorf("error de transmisión física UDP: %w", err)
	}
	return nil
}

// ReceiveContainer recibe y deserializa una trama I7 desde el socket UDP físico
func (a *I7UDPAdapter) ReceiveContainer(timeout time.Duration) (*I7Container, *net.UDPAddr, error) {
	a.mu.RLock()
	conn := a.conn
	closed := a.closed
	a.mu.RUnlock()

	if closed || conn == nil {
		return nil, nil, ErrAdapterClosed
	}

	buffer := make([]byte, I7StandardMTU+128)
	if timeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(timeout))
	} else {
		_ = conn.SetReadDeadline(time.Time{})
	}

	n, fromAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, nil, fmt.Errorf("error de recepción física UDP: %w", err)
	}

	container, err := DeserializeContainer(buffer[:n])
	if err != nil {
		return nil, fromAddr, err
	}
	return container, fromAddr, nil
}

// Close cierra limpiamente el socket físico de red
func (a *I7UDPAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return nil
	}
	a.closed = true
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// SerializeContainer codifica un I7Container en formato binario de alambre
// Estructura fija: [Version:1][ChannelID:4][TagLen:1][Tag:N][PayloadLen:2][Payload:M]
func SerializeContainer(c *I7Container) ([]byte, error) {
	if c.Version != I7Version {
		return nil, ErrUnsupportedVersion
	}
	tagLen := len(c.Tag)
	payloadLen := len(c.Payload)
	totalLen := 1 + 4 + 1 + tagLen + 2 + payloadLen
	if totalLen > I7StandardMTU {
		return nil, ErrPayloadTooLarge
	}

	wire := make([]byte, totalLen)
	wire[0] = c.Version
	binary.BigEndian.PutUint32(wire[1:5], c.ChannelID)
	wire[5] = byte(tagLen)
	offset := 6
	copy(wire[offset:offset+tagLen], c.Tag)
	offset += tagLen
	binary.BigEndian.PutUint16(wire[offset:offset+2], uint16(payloadLen))
	offset += 2
	copy(wire[offset:offset+payloadLen], c.Payload)

	return wire, nil
}

// DeserializeContainer decodifica bytes de alambre hacia un I7Container
func DeserializeContainer(data []byte) (*I7Container, error) {
	if len(data) < 8 {
		return nil, ErrCorruptFrame
	}
	version := data[0]
	if version != I7Version {
		return nil, ErrUnsupportedVersion
	}
	channelID := binary.BigEndian.Uint32(data[1:5])
	tagLen := int(data[5])
	offset := 6
	if len(data) < offset+tagLen+2 {
		return nil, ErrCorruptFrame
	}
	tag := make([]byte, tagLen)
	copy(tag, data[offset:offset+tagLen])
	offset += tagLen

	payloadLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	if len(data) < offset+payloadLen {
		return nil, ErrCorruptFrame
	}
	payload := make([]byte, payloadLen)
	copy(payload, data[offset:offset+payloadLen])

	return &I7Container{
		Version:   version,
		ChannelID: channelID,
		Payload:   payload,
		Tag:       tag,
	}, nil
}
