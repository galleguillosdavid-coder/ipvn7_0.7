package l1

import (
	"fmt"
)

// SphinxPacket representa el paquete serializado de 1280 bytes exactos
type SphinxPacket struct {
	KeyHeader     [SphinxKeyHeaderSize]byte    `json:"-"`
	RoutingHeader [SphinxRoutingTotalSize]byte `json:"-"`
	Payload       [SphinxPayloadSize]byte      `json:"-"`
}

// PeelResult contiene el resultado de procesar un salto cebolla
type PeelResult struct {
	Action     string        `json:"action"` // FORWARD, DELIVER, DROP
	CircuitID  string        `json:"circuit_id"`
	NextHopDID string        `json:"next_hop_did"`
	NextPacket *SphinxPacket `json:"-"`
	RawPayload []byte        `json:"raw_payload,omitempty"`
	IsExit     bool          `json:"is_exit"`
	PeelTimeUs int64         `json:"peel_time_us"`
	TagHash    string        `json:"tag_hash"`
}

// Serialize convierte el paquete a exactamente 1280 bytes para el wire
func (p *SphinxPacket) Serialize() []byte {
	out := make([]byte, SphinxPacketSize)
	copy(out[0:SphinxKeyHeaderSize], p.KeyHeader[:])
	copy(out[SphinxKeyHeaderSize:SphinxKeyHeaderSize+SphinxRoutingTotalSize], p.RoutingHeader[:])
	copy(out[SphinxKeyHeaderSize+SphinxRoutingTotalSize:], p.Payload[:])
	return out
}

// DeserializeSphinxPacket reconstruye el paquete validando el límite estricto de 1280 bytes
func DeserializeSphinxPacket(wire []byte) (*SphinxPacket, error) {
	if len(wire) != SphinxPacketSize {
		return nil, fmt.Errorf("tamaño wire inválido (%d bytes): se requieren exactamente %d bytes deterministas", len(wire), SphinxPacketSize)
	}

	var p SphinxPacket
	copy(p.KeyHeader[:], wire[0:SphinxKeyHeaderSize])
	copy(p.RoutingHeader[:], wire[SphinxKeyHeaderSize:SphinxKeyHeaderSize+SphinxRoutingTotalSize])
	copy(p.Payload[:], wire[SphinxKeyHeaderSize+SphinxRoutingTotalSize:])
	return &p, nil
}
