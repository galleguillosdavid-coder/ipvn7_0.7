package l1_test

import (
	"testing"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

// TestAdversarial_TestA_FakeSourceDID prueba que un atacante (Mallory) que suplanta el DID de Alice
// sin poseer su clave privada Ed25519 es rechazado de inmediato y Bob NO autoriza a Alice.
func TestAdversarial_TestA_FakeSourceDID(t *testing.T) {
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()
	idMallory, _ := l0.GenerateIdentity()

	keysBob, _ := l1.GenerateHybridKeyPair(idBob.DID())
	keysMallory, _ := l1.GenerateHybridKeyPair(idMallory.DID())

	fwBob := l1.NewZTNAFirewall(true)
	mgrBob := l1.NewPQCSessionManager(idBob, keysBob, fwBob)
	mgrMallory := l1.NewPQCSessionManager(idMallory, keysMallory, nil)

	pkt, err := mgrMallory.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Mallory falló creando init: %v", err)
	}

	// Mallory manipula el paquete asignando el SourceDID de Alice
	pkt.SourceDID = idAlice.DID()

	_, err = mgrBob.HandleHandshakeInitPacket(pkt)
	if err == nil {
		t.Fatal("Test A: Bob debió RECHAZAR el HandshakeInit con SourceDID suplantado")
	}

	// Verificar que Bob NO haya creado sesión con Alice ni autorizado a Alice en su firewall
	if mgrBob.HasSession(idAlice.DID()) {
		t.Fatal("Test A: Bob NO debió crear sesión para Alice")
	}
	decision, _ := fwBob.EvaluateInbound(idAlice.DID(), 7777)
	if decision == l1.DecisionAccept {
		t.Fatal("Test A: ZTNA de Bob NO debió autorizar al DID de Alice suplantado")
	}
}

// TestHandshake_DIDSpoofing implementa la prueba de ataque formal requerida por el checklist (Fase 1.4):
// Mallory declara SourceDID = Alice con KEM válido pero firmado por Mallory -> REJECT, NO SESSION, NO AUTHORIZATION.
func TestHandshake_DIDSpoofing(t *testing.T) {
	TestAdversarial_TestA_FakeSourceDID(t)
}

// TestAdversarial_TestB_InvalidSignature prueba que un paquete con SourceDID de Alice
// pero firmado con la clave privada de Mallory es rechazado.
func TestAdversarial_TestB_InvalidSignature(t *testing.T) {
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()
	idMallory, _ := l0.GenerateIdentity()

	keysBob, _ := l1.GenerateHybridKeyPair(idBob.DID())
	keysMallory, _ := l1.GenerateHybridKeyPair(idMallory.DID())

	fwBob := l1.NewZTNAFirewall(true)
	mgrBob := l1.NewPQCSessionManager(idBob, keysBob, fwBob)
	mgrMallory := l1.NewPQCSessionManager(idMallory, keysMallory, nil)

	pkt, err := mgrMallory.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Mallory falló creando init: %v", err)
	}

	pkt.SourceDID = idAlice.DID()
	// Mallory firma explícitamente el paquete con su propia identidad
	if err := pkt.SignPacket(idMallory); err != nil {
		t.Fatalf("Error firmando con Mallory: %v", err)
	}

	_, err = mgrBob.HandleHandshakeInitPacket(pkt)
	if err == nil {
		t.Fatal("Test B: Bob debió RECHAZAR firma de Mallory pretendiendo ser Alice")
	}
	if mgrBob.HasSession(idAlice.DID()) {
		t.Fatal("Test B: Bob NO debió registrar sesión para Alice")
	}
}

// TestAdversarial_TestC_CrossSession prueba que secuencias idénticas en sesiones distintas
// son aceptadas independientemente por el filtro AntiReplayFilter L1.
func TestAdversarial_TestC_CrossSession(t *testing.T) {
	filter := l1.NewAntiReplayFilter(nil)
	peer := "did:ipvn7:test_peer_alice"
	session1 := uint64(1001)
	session2 := uint64(1002)
	now := time.Now().Unix()

	if !filter.Accept(peer, session1, 10, now) {
		t.Fatal("Test C: session1 seq 10 debe ser aceptado")
	}
	if !filter.Accept(peer, session2, 10, now) {
		t.Fatal("Test C: session2 seq 10 debe ser aceptado (aislamiento cross-session)")
	}
}

// TestAdversarial_TestD_CrossPeer prueba que secuencias idénticas de pares distintos
// son aceptadas independientemente.
func TestAdversarial_TestD_CrossPeer(t *testing.T) {
	filter := l1.NewAntiReplayFilter(nil)
	peerAlice := "did:ipvn7:test_peer_alice"
	peerBob := "did:ipvn7:test_peer_bob"
	now := time.Now().Unix()

	if !filter.Accept(peerAlice, 1, 10, now) {
		t.Fatal("Test D: Alice seq 10 debe ser aceptado")
	}
	if !filter.Accept(peerBob, 1, 10, now) {
		t.Fatal("Test D: Bob seq 10 debe ser aceptado sin interferencia con Alice")
	}
}

// TestAdversarial_TestE_Replay prueba que la repetición del mismo paquete en la misma sesión es rechazada.
func TestAdversarial_TestE_Replay(t *testing.T) {
	filter := l1.NewAntiReplayFilter(nil)
	peer := "did:ipvn7:test_peer_alice"
	session := uint64(500)
	now := time.Now().Unix()

	if !filter.Accept(peer, session, 10, now) {
		t.Fatal("Test E: primer envío debe ser aceptado")
	}
	if filter.Accept(peer, session, 10, now) {
		t.Fatal("Test E: paquete repetido debe ser RECHAZADO")
	}
}

// TestAdversarial_TestF_HandshakeReplay prueba que el reenvío de un HandshakeInit capturado es rechazado
// por el filtro anti-replay de la sesión.
func TestAdversarial_TestF_HandshakeReplay(t *testing.T) {
	filter := l1.NewAntiReplayFilter(nil)
	peer := "did:ipvn7:test_peer_alice"
	handshakeSeq := uint64(1)
	now := time.Now().Unix()

	if !filter.Accept(peer, 0, handshakeSeq, now) {
		t.Fatal("Test F: primer HandshakeInit debe ser aceptado")
	}
	// Reenviar exactamente el mismo HandshakeInit
	if filter.Accept(peer, 0, handshakeSeq, now) {
		t.Fatal("Test F: HandshakeInit repetido debe ser RECHAZADO")
	}
}

// TestAdversarial_TestG_ResponseWrongDestination prueba que una respuesta dirigida a Carol
// presentada ante Alice es rechazada por DestDID mismatch.
func TestAdversarial_TestG_ResponseWrongDestination(t *testing.T) {
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()
	idCarol, _ := l0.GenerateIdentity()

	keysAlice, _ := l1.GenerateHybridKeyPair(idAlice.DID())
	keysBob, _ := l1.GenerateHybridKeyPair(idBob.DID())

	mgrAlice := l1.NewPQCSessionManager(idAlice, keysAlice, nil)

	// Alice inicia handshake hacia Bob
	_, err := mgrAlice.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Error creando init: %v", err)
	}

	// Construir paquete de respuesta fraudulento donde DestDID es Carol
	respPkt := l0.NewPacket(l0.MsgTypeHandshakeResp, idBob.DID(), idCarol.DID(), 1, nil, []byte("{}"))
	_ = respPkt.SignPacket(idBob)

	err = mgrAlice.HandleHandshakeRespPacket(respPkt)
	if err == nil {
		t.Fatal("Test G: Alice debió rechazar respuesta dirigida a Carol")
	}
}

// TestAdversarial_TestH_TamperedSourceDID prueba que si un atacante altera el SourceDID
// en tránsito después de que el paquete fue firmado, la firma falla.
func TestAdversarial_TestH_TamperedSourceDID(t *testing.T) {
	idAlice, _ := l0.GenerateIdentity()
	idBob, _ := l0.GenerateIdentity()
	keysBob, _ := l1.GenerateHybridKeyPair(idBob.DID())
	keysAlice, _ := l1.GenerateHybridKeyPair(idAlice.DID())

	mgrAlice := l1.NewPQCSessionManager(idAlice, keysAlice, nil)
	mgrBob := l1.NewPQCSessionManager(idBob, keysBob, nil)

	initPkt, err := mgrAlice.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Error creando init: %v", err)
	}

	// Atacante altera el último carácter del SourceDID
	tamperedSource := []byte(initPkt.SourceDID)
	if tamperedSource[len(tamperedSource)-1] == 'a' {
		tamperedSource[len(tamperedSource)-1] = 'b'
	} else {
		tamperedSource[len(tamperedSource)-1] = 'a'
	}
	initPkt.SourceDID = string(tamperedSource)

	_, err = mgrBob.HandleHandshakeInitPacket(initPkt)
	if err == nil {
		t.Fatal("Test H: Bob debió rechazar HandshakeInit con SourceDID alterado")
	}
}
