package l1_test

import (
	"net"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

// TestRoamingValidSignatureUnauthorizedDID prueba que una firma Ed25519 matemáticamente válida
// de un par no autorizado por política local ZTNA es descartada (DROP) y NO se auto-autoriza.
func TestRoamingValidSignatureUnauthorizedDID(t *testing.T) {
	idBob, _ := l0.GenerateIdentity()
	idMallory, _ := l0.GenerateIdentity()

	fwBob := l1.NewZTNAFirewall(true) // Default-Deny

	// Mallory genera y firma un paquete de roaming legítimo con su propia clave privada
	pkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, idMallory.DID(), idBob.DID(), 1, nil, nil)
	if err := pkt.SignPacket(idMallory); err != nil {
		t.Fatalf("error firmando paquete de Mallory: %v", err)
	}

	// 1. La firma de Mallory es matemáticamente válida (autenticidad demostrada)
	valid, err := pkt.VerifyPacketSignature()
	if err != nil || !valid {
		t.Fatalf("la firma de Mallory debió ser válida para su propio DID")
	}

	// 2. Sin embargo, la autorización ZTNA debe evaluar la política local y denegar
	decision, _ := fwBob.EvaluateInbound(pkt.SourceDID, 7777)
	if decision == l1.DecisionAccept {
		t.Fatalf("FALLA CRÍTICA ZTNA: Mallory fue autorizada automáticamente sin política previa")
	}
}

// TestRoamingValidSignatureAuthorizedDID prueba que un par previamente autorizado por política
// que envía un RoamingUpdate firmado válidamente es aceptado (ACCEPT) y su ruta actualizada.
func TestRoamingValidSignatureAuthorizedDID(t *testing.T) {
	idBob, _ := l0.GenerateIdentity()
	idAlice, _ := l0.GenerateIdentity()

	fwBob := l1.NewZTNAFirewall(true) // Default-Deny
	// Bob autoriza explícitamente a Alice por política administrativa
	fwBob.AuthorizeDID(&l1.DIDPolicy{
		DID:          idAlice.DID(),
		AllowInbound: true,
	})

	routerBob := l1.NewKleinbergRouter(idBob)

	pkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, idAlice.DID(), idBob.DID(), 1, nil, nil)
	if err := pkt.SignPacket(idAlice); err != nil {
		t.Fatalf("error firmando paquete de Alice: %v", err)
	}

	valid, err := pkt.VerifyPacketSignature()
	if err != nil || !valid {
		t.Fatalf("firma de Alice debe ser válida")
	}

	// ZTNA evalúa inbound
	decision, _ := fwBob.EvaluateInbound(pkt.SourceDID, 7777)
	if decision != l1.DecisionAccept {
		t.Fatalf("Alice debió ser aceptada por la política ZTNA previa")
	}

	newAddr, _ := net.ResolveUDPAddr("udp", "192.168.1.106:7777")
	if err := routerBob.HandleRoamingUpdate(pkt, newAddr); err != nil {
		t.Fatalf("error actualizando roaming de Alice: %v", err)
	}

	nextHop, err := routerBob.FindNextHop(idAlice.DID())
	if err != nil || nextHop.Locator.PhysicalAddr.String() != newAddr.String() {
		t.Fatalf("el router debió actualizar la ubicación de Alice a %s", newAddr.String())
	}
}

// TestRoamingInvalidSignature prueba que un paquete de roaming con firma corrupta o forjada es descartado (DROP).
func TestRoamingInvalidSignature(t *testing.T) {
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()

	pkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, idAlice.DID(), idBob.DID(), 1, nil, nil)
	_ = pkt.SignPacket(idAlice)
	// Corromper la firma
	pkt.Signature[0] ^= 0xFF

	valid, err := pkt.VerifyPacketSignature()
	if err == nil && valid {
		t.Fatalf("firma corrupta debió ser rechazada")
	}

	routerBob := l1.NewKleinbergRouter(idBob)
	newAddr, _ := net.ResolveUDPAddr("udp", "192.168.1.106:7777")
	if err := routerBob.HandleRoamingUpdate(pkt, newAddr); err == nil {
		t.Fatalf("HandleRoamingUpdate debió fallar ante firma inválida")
	}
}

// TestRoamingUnknownDID prueba que un paquete con un DID malformado o desconocido no verificable es descartado (DROP).
func TestRoamingUnknownDID(t *testing.T) {
	idBob, _ := l0.GenerateIdentity()
	routerBob := l1.NewKleinbergRouter(idBob)

	// DID sintácticamente inválido
	pkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, "did:ipvn7:malformed_unknown_did_123", idBob.DID(), 1, nil, nil)
	pkt.Signature = make([]byte, 64)

	valid, err := pkt.VerifyPacketSignature()
	if err == nil && valid {
		t.Fatalf("DID malformado debió fallar al verificar firma")
	}

	newAddr, _ := net.ResolveUDPAddr("udp", "192.168.1.106:7777")
	if err := routerBob.HandleRoamingUpdate(pkt, newAddr); err == nil {
		t.Fatalf("HandleRoamingUpdate debió fallar ante DID desconocido/malformado")
	}
}
