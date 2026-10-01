package l1

import (
	"testing"
	"time"
)

func TestZTNAFirewall(t *testing.T) {
	fw := NewZTNAFirewall(true) // Default-Deny activado

	didA := "did:ipvn7:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	didB := "did:ipvn7:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	// 1. DID no registrado debe ser rechazado por Default-Deny
	dec, _ := fw.EvaluateInbound(didA, 8080)
	if dec != DecisionDropDefaultDeny {
		t.Fatalf("Esperado DecisionDropDefaultDeny, obtenido %v", dec)
	}

	// 2. Autorizar didA únicamente para puerto virtual 7001
	fw.AuthorizeDID(&DIDPolicy{
		DID:          didA,
		AllowInbound: true,
		AllowedPorts: []uint16{7001},
	})

	// Puerto 7001 debe ser aceptado
	dec, _ = fw.EvaluateInbound(didA, 7001)
	if dec != DecisionAccept {
		t.Fatalf("Esperado DecisionAccept en puerto 7001, obtenido %v", dec)
	}

	// Puerto 8080 para didA debe ser rechazado por ACL
	dec, _ = fw.EvaluateInbound(didA, 8080)
	if dec != DecisionDropACL {
		t.Fatalf("Esperado DecisionDropACL en puerto 8080, obtenido %v", dec)
	}

	// 3. Probar política expirada
	fw.AuthorizeDID(&DIDPolicy{
		DID:          didB,
		AllowInbound: true,
		ExpiresAt:    time.Now().Add(-1 * time.Minute), // ya expiró
	})
	dec, _ = fw.EvaluateInbound(didB, 7001)
	if dec != DecisionDropACL {
		t.Fatalf("Esperado DecisionDropACL por expiración, obtenido %v", dec)
	}

	// 4. Revocar didA
	fw.RevokeDID(didA)
	dec, _ = fw.EvaluateInbound(didA, 7001)
	if dec != DecisionDropDefaultDeny {
		t.Fatalf("Esperado DecisionDropDefaultDeny tras revocación, obtenido %v", dec)
	}

	// 5. Probar métodos de interfaz ZTNAFirewall (AllowPeer, IsAllowed, RevokePeer)
	didC := "did:ipvn7:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	if fw.IsAllowed(didC) {
		t.Fatalf("Esperado IsAllowed=false para didC no autorizado")
	}
	fw.AllowPeer(didC, []uint16{7001})
	if !fw.IsAllowed(didC) {
		t.Fatalf("Esperado IsAllowed=true para didC autorizado con AllowPeer")
	}
	dec, _ = fw.EvaluateInbound(didC, 7001)
	if dec != DecisionAccept {
		t.Fatalf("Esperado DecisionAccept para didC en puerto 7001, obtenido %v", dec)
	}
	fw.RevokePeer(didC)
	if fw.IsAllowed(didC) {
		t.Fatalf("Esperado IsAllowed=false tras RevokePeer")
	}
}
