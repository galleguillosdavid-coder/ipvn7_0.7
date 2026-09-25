package l1

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"ipvn7/pkg/l0"
)

type mockMeshForwarder struct {
	mu        sync.Mutex
	forwarded map[string][][]byte
}

func newMockMeshForwarder() *mockMeshForwarder {
	return &mockMeshForwarder{forwarded: make(map[string][][]byte)}
}

func (m *mockMeshForwarder) ForwardToMesh(targetDID string, ipPacket []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]byte, len(ipPacket))
	copy(copied, ipPacket)
	m.forwarded[targetDID] = append(m.forwarded[targetDID], copied)
	return nil
}

func (m *mockMeshForwarder) Count(targetDID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.forwarded[targetDID])
}

// buildSyntheticIPPacket construye un datagrama IPv4 canónico para pruebas
func buildSyntheticIPPacket(srcIP, dstIP net.IP, payload []byte) []byte {
	src := srcIP.To4()
	dst := dstIP.To4()
	totalLen := 20 + len(payload)
	pkt := make([]byte, totalLen)
	pkt[0] = 0x45
	pkt[2] = byte(totalLen >> 8)
	pkt[3] = byte(totalLen)
	pkt[8] = 64
	pkt[9] = 17 // UDP
	copy(pkt[12:16], src)
	copy(pkt[16:20], dst)
	copy(pkt[20:], payload)
	return pkt
}

func TestTUNRouter_LifecycleAndForwarding(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	adapter := NewUserspaceVirtualAdapter(id)
	forwarder := newMockMeshForwarder()
	router := NewTUNRouter(adapter, forwarder)
	defer router.Close()

	srcIP := net.ParseIP("10.7.0.1")
	dstIP := net.ParseIP("10.7.0.2")
	expectedDID := ResolveIPToDID(dstIP)
	payload := []byte("ping_ipvn7_virtual_tun_router")

	rawPacket := buildSyntheticIPPacket(srcIP, dstIP, payload)

	// Simular inyección desde el host a través del canal outbound del adaptador
	if err := adapter.InjectPacket(rawPacket); err != nil {
		t.Fatalf("error inyectando paquete en adaptador: %v", err)
	}

	// Esperar que el loop de lectura lo extraiga y lo remita al forwarder
	time.Sleep(50 * time.Millisecond)
	if forwarder.Count(expectedDID) != 1 {
		t.Fatalf("se esperaba 1 paquete hacia %s, encontrados %d", expectedDID, forwarder.Count(expectedDID))
	}

	rxPkts, _, rxBytes, _ := router.Stats()
	if rxPkts != 1 || rxBytes != uint64(len(rawPacket)) {
		t.Errorf("métricas erróneas: rxPkts=%d rxBytes=%d", rxPkts, rxBytes)
	}

	// Probar inyección desde la malla hacia el adaptador
	if err := router.InjectFromMesh(rawPacket); err != nil {
		t.Fatalf("error inyectando desde la malla: %v", err)
	}

	_, txPkts, _, txBytes := router.Stats()
	if txPkts != 1 || txBytes != uint64(len(rawPacket)) {
		t.Errorf("métricas de Tx erróneas: txPkts=%d txBytes=%d", txPkts, txBytes)
	}
}

func TestResolveIPToDID(t *testing.T) {
	ip4 := net.ParseIP("10.7.0.100")
	did4 := ResolveIPToDID(ip4)
	if did4 != "did:ipvn7:node-v4-10-7-0-100" {
		t.Errorf("DID IPv4 no esperado: %s", did4)
	}

	ip6 := net.ParseIP("fd07::1")
	did6 := ResolveIPToDID(ip6)
	if len(did6) == 0 || !bytes.HasPrefix([]byte(did6), []byte("did:ipvn7:node-v6-")) {
		t.Errorf("DID IPv6 inválido: %s", did6)
	}
}
