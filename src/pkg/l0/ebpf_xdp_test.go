package l0

import (
	"encoding/binary"
	"testing"
)

func buildMockEthernetFrame(destPort uint16, magic uint32) []byte {
	// 14B Ethernet + 20B IPv4 + 8B UDP + 4B Magic
	frame := make([]byte, 46)

	// Ethernet
	binary.BigEndian.PutUint16(frame[12:14], 0x0800) // IPv4

	// IPv4
	frame[14] = 0x45 // IHL=5 (20B)
	frame[23] = 17   // UDP

	// UDP
	binary.BigEndian.PutUint16(frame[34:36], 12345)    // SrcPort
	binary.BigEndian.PutUint16(frame[36:38], destPort) // DstPort
	binary.BigEndian.PutUint16(frame[38:40], 12)       // Length (8B UDP + 4B payload)

	// Payload
	binary.BigEndian.PutUint32(frame[42:46], magic)
	return frame
}

func TestXDPFilter_ValidIPVN7Frame(t *testing.T) {
	sim := NewXDPKernelSimulator()
	validFrame := buildMockEthernetFrame(IPVN7UDPMeshPort, IPVN7MagicBytes)

	action, err := sim.ProcessFrame(validFrame)
	if err != nil {
		t.Fatalf("error procesando trama: %v", err)
	}
	if action != XDPActionPass {
		t.Errorf("se esperaba XDPActionPass, obtenido: %d", action)
	}

	processed, passed, dropped, _ := sim.Stats()
	if processed != 1 || passed != 1 || dropped != 0 {
		t.Errorf("estadísticas inesperadas: proc=%d, pass=%d, drop=%d", processed, passed, dropped)
	}
}

func TestXDPFilter_CorruptDrop(t *testing.T) {
	sim := NewXDPKernelSimulator()
	// Trama hacia puerto 7777 pero con magic corrupto o ataque DoS
	corruptFrame := buildMockEthernetFrame(IPVN7UDPMeshPort, 0xDEADBEEF)

	action, err := sim.ProcessFrame(corruptFrame)
	if err != nil {
		t.Fatalf("error procesando trama: %v", err)
	}
	if action != XDPActionDrop {
		t.Errorf("se esperaba XDPActionDrop contra trama corrupta, obtenido: %d", action)
	}

	_, _, dropped, _ := sim.Stats()
	if dropped != 1 {
		t.Errorf("conteo de paquetes descartados erróneo: %d != 1", dropped)
	}
}
