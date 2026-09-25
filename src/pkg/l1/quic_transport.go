package l1

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

// Magic identifier para tramas de transporte multiplexado
const (
	QUICFrameMagicStream uint8 = 0x01
	QUICFrameMagicACK    uint8 = 0x02
	QUICFrameMagicReset  uint8 = 0x03
	MaxPayloadPerStream  int   = 1200 // Respeta el MTU canónico de 1280B
)

var (
	ErrOversizedStreamPayload = errors.New("quic: payload excede el límite del MTU canónico")
	ErrStreamClosed           = errors.New("quic: flujo cerrado")
	ErrInvalidFrameHeader     = errors.New("quic: encabezado de trama binaria inválido")
)

// StreamFrame representa un fragmento de datos multiplexado dentro de un datagrama IPVN7
type StreamFrame struct {
	StreamID uint32
	Offset   uint64
	Fin      bool
	Data     []byte
}

// ACKFrame representa una confirmación selectiva de entrega (SACK)
type ACKFrame struct {
	StreamID     uint32
	LargestAcked uint64
	AckBitmask   uint64 // Ventana de 64 posiciones precedentes recibidas con éxito
}

// MultiplexedStream gestiona la recepción y reensamblaje ordenado de un stream individual
type MultiplexedStream struct {
	ID         uint32
	mu         sync.Mutex
	readOffset uint64
	buffer     map[uint64][]byte
	closed     bool
}

func newMultiplexedStream(id uint32) *MultiplexedStream {
	return &MultiplexedStream{
		ID:     id,
		buffer: make(map[uint64][]byte),
	}
}

// IngestChunk inserta un fragmento recibido y retorna los datos contiguos disponibles
func (s *MultiplexedStream) IngestChunk(offset uint64, data []byte) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	if offset < s.readOffset {
		// Fragmento ya consumido previamente (retransmisión redundante)
		return nil
	}

	// Almacenar fragmento en búfer de reordenamiento
	chunkCopy := make([]byte, len(data))
	copy(chunkCopy, data)
	s.buffer[offset] = chunkCopy

	// Drenar fragmentos contiguos en orden
	var assembled []byte
	for {
		chunk, exists := s.buffer[s.readOffset]
		if !exists {
			break
		}
		assembled = append(assembled, chunk...)
		delete(s.buffer, s.readOffset)
		s.readOffset += uint64(len(chunk))
	}
	return assembled
}

// QUICTransportEngine orquesta flujos multiplexados confiables sobre datagramas no confiables
type QUICTransportEngine struct {
	mu      sync.RWMutex
	streams map[uint32]*MultiplexedStream
}

// NewQUICTransportEngine instancia el motor de transporte multiplexado
func NewQUICTransportEngine() *QUICTransportEngine {
	return &QUICTransportEngine{
		streams: make(map[uint32]*MultiplexedStream),
	}
}

// EncodeStreamFrame serializa un fragmento en formato binario compacto zero-copy
func (e *QUICTransportEngine) EncodeStreamFrame(frame *StreamFrame) ([]byte, error) {
	if len(frame.Data) > MaxPayloadPerStream {
		return nil, ErrOversizedStreamPayload
	}

	// Encabezado binario: Magic(1B) + StreamID(4B) + Offset(8B) + Flags(1B) + Len(2B) = 16B
	buf := make([]byte, 16+len(frame.Data))
	buf[0] = QUICFrameMagicStream
	binary.BigEndian.PutUint32(buf[1:5], frame.StreamID)
	binary.BigEndian.PutUint64(buf[5:13], frame.Offset)
	if frame.Fin {
		buf[13] = 0x01
	} else {
		buf[13] = 0x00
	}
	binary.BigEndian.PutUint16(buf[14:16], uint16(len(frame.Data)))
	copy(buf[16:], frame.Data)
	return buf, nil
}

// DecodeFrame deserializa una trama binaria validando su integridad
func (e *QUICTransportEngine) DecodeFrame(raw []byte) (*StreamFrame, *ACKFrame, error) {
	if len(raw) < 1 {
		return nil, nil, ErrInvalidFrameHeader
	}

	switch raw[0] {
	case QUICFrameMagicStream:
		if len(raw) < 16 {
			return nil, nil, ErrInvalidFrameHeader
		}
		streamID := binary.BigEndian.Uint32(raw[1:5])
		offset := binary.BigEndian.Uint64(raw[5:13])
		fin := raw[13] == 0x01
		dataLen := int(binary.BigEndian.Uint16(raw[14:16]))
		if len(raw) < 16+dataLen {
			return nil, nil, ErrInvalidFrameHeader
		}
		frame := &StreamFrame{
			StreamID: streamID,
			Offset:   offset,
			Fin:      fin,
			Data:     raw[16 : 16+dataLen],
		}
		return frame, nil, nil

	case QUICFrameMagicACK:
		// Magic(1B) + StreamID(4B) + LargestAcked(8B) + AckBitmask(8B) = 21B
		if len(raw) < 21 {
			return nil, nil, ErrInvalidFrameHeader
		}
		streamID := binary.BigEndian.Uint32(raw[1:5])
		largestAcked := binary.BigEndian.Uint64(raw[5:13])
		ackBitmask := binary.BigEndian.Uint64(raw[13:21])
		ack := &ACKFrame{
			StreamID:     streamID,
			LargestAcked: largestAcked,
			AckBitmask:   ackBitmask,
		}
		return nil, ack, nil

	default:
		return nil, nil, fmt.Errorf("quic: tipo de trama desconocido: 0x%02x", raw[0])
	}
}

// EncodeACKFrame serializa un acuse de recibo selectivo (SACK)
func (e *QUICTransportEngine) EncodeACKFrame(ack *ACKFrame) []byte {
	buf := make([]byte, 21)
	buf[0] = QUICFrameMagicACK
	binary.BigEndian.PutUint32(buf[1:5], ack.StreamID)
	binary.BigEndian.PutUint64(buf[5:13], ack.LargestAcked)
	binary.BigEndian.PutUint64(buf[13:21], ack.AckBitmask)
	return buf
}

// ProcessIncomingStreamChunk procesa un fragmento entrante y reensambla flujos contiguos
func (e *QUICTransportEngine) ProcessIncomingStreamChunk(frame *StreamFrame) []byte {
	e.mu.Lock()
	stream, exists := e.streams[frame.StreamID]
	if !exists {
		stream = newMultiplexedStream(frame.StreamID)
		e.streams[frame.StreamID] = stream
	}
	e.mu.Unlock()

	return stream.IngestChunk(frame.Offset, frame.Data)
}

// CloseStream elimina un flujo completado liberando recursos de memoria
func (e *QUICTransportEngine) CloseStream(streamID uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.streams, streamID)
}
