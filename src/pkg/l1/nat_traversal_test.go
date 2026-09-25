package l1

import (
	"net"
	"testing"
	"time"

	"ipvn7/pkg/l0"
)

func TestNATTraversal_Classification(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	engine := NewNATTraversalEngine(id)
	defer engine.Close()

	if engine.CurrentNATType() != NATUnknown {
		t.Errorf("tipo inicial inesperado: %v", engine.CurrentNATType())
	}

	addr1 := &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 54321}
	addr2 := &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 54321}

	// 1. Mismo IP y puerto reflexivo hacia diferentes destinos -> Full Cone
	natType := engine.ClassifyNAT(addr1, addr2, true)
	if natType != NATFullCone {
		t.Errorf("se esperaba FULL_CONE, obtenido %v", natType)
	}

	// 2. Distinto puerto reflexivo según el servidor STUN -> Simétrico
	addr3 := &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 54322}
	natType = engine.ClassifyNAT(addr1, addr3, false)
	if natType != NATSymmetric {
		t.Errorf("se esperaba SYMMETRIC_HARD_NAT, obtenido %v", natType)
	}
}

func TestNATTraversal_HolePunchingAndRelay(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	engine := NewNATTraversalEngine(id)
	defer engine.Close()

	// Configurar como NAT Simétrico
	addr1 := &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 10001}
	addr2 := &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 10002}
	engine.ClassifyNAT(addr1, addr2, false)

	// Simétrico vs Simétrico no puede perforar directo
	if engine.CanDirectPunch(NATSymmetric) {
		t.Error("Symmetric vs Symmetric no debe admitir perforación directa")
	}

	// Simétrico vs FullCone sí puede perforar directo
	if !engine.CanDirectPunch(NATFullCone) {
		t.Error("Symmetric vs FullCone debe permitir perforación directa")
	}

	// Ejecutar perforación
	targetAddr := &net.UDPAddr{IP: net.ParseIP("198.51.100.5"), Port: 7777}
	ok, err := engine.AttemptHolePunch("did:ipvn7:test-remote-node", targetAddr, 200*time.Millisecond)
	if err != nil || !ok {
		t.Fatalf("AttemptHolePunch falló: %v", err)
	}

	// Probar Relay soberano ante ausencia de relays
	err = engine.RouteViaDERPRelay("did:ipvn7:target", []byte("encrypted_payload"))
	if err != ErrRelayUnavailable {
		t.Errorf("se esperaba ErrRelayUnavailable, obtenido: %v", err)
	}

	// Registrar un relay soberano y reintentar
	relayAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.198"), Port: 7001}
	engine.RegisterDERPRelay("did:ipvn7:relay-node-1", relayAddr)

	err = engine.RouteViaDERPRelay("did:ipvn7:target", []byte("encrypted_payload"))
	if err != nil {
		t.Fatalf("RouteViaDERPRelay falló con relay disponible: %v", err)
	}

	punches, relays := engine.Stats()
	if punches != 1 || relays != 1 {
		t.Errorf("estadísticas erróneas: punches=%d, relays=%d", punches, relays)
	}
}
