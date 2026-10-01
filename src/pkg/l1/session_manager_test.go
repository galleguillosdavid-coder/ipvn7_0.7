package l1_test

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestPQCSessionManager_FullFlow(t *testing.T) {
	// 1. Crear identidades de Alice y Bob
	idAlice, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando id Alice: %v", err)
	}
	idBob, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando id Bob: %v", err)
	}

	keysAlice, err := l1.GenerateHybridKeyPair(idAlice.DID())
	if err != nil {
		t.Fatalf("Error generando claves híbridas Alice: %v", err)
	}
	keysBob, err := l1.GenerateHybridKeyPair(idBob.DID())
	if err != nil {
		t.Fatalf("Error generando claves híbridas Bob: %v", err)
	}

	fwAlice := l1.NewZTNAFirewall(true)
	fwBob := l1.NewZTNAFirewall(true)

	mgrAlice := l1.NewPQCSessionManager(idAlice, keysAlice, fwAlice)
	mgrBob := l1.NewPQCSessionManager(idBob, keysBob, fwBob)

	// Antes del handshake, ZTNA debe denegar tráfico entrante
	decisionA, _ := fwAlice.EvaluateInbound(idBob.DID(), 7777)
	if decisionA == l1.DecisionAccept {
		t.Fatalf("ZTNA Alice no debe aceptar tráfico previo a autenticación")
	}

	// 2. Alice crea datagrama de inicio 1-RTT
	initPkt, err := mgrAlice.CreateHandshakeInitPacket(idBob.DID(), keysBob.ClassicalKEMPub, keysBob.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Error creando handshake init: %v", err)
	}

	// 3. Bob procesa handshake init y responde
	respPkt, err := mgrBob.HandleHandshakeInitPacket(initPkt)
	if err != nil {
		t.Fatalf("Error procesando handshake init en Bob: %v", err)
	}

	// Bob ya debe tener la sesión activa y haber autorizado a Alice en su ZTNA
	if !mgrBob.HasSession(idAlice.DID()) {
		t.Fatalf("Bob debe tener sesión activa tras responder init")
	}
	decisionB, _ := fwBob.EvaluateInbound(idAlice.DID(), 7777)
	if decisionB != l1.DecisionAccept {
		t.Fatalf("ZTNA Bob debe autorizar a Alice tras el handshake")
	}

	// 4. Alice procesa respuesta de Bob
	if err := mgrAlice.HandleHandshakeRespPacket(respPkt); err != nil {
		t.Fatalf("Error procesando handshake resp en Alice: %v", err)
	}

	// Alice ya debe tener la sesión activa y haber autorizado a Bob en su ZTNA
	if !mgrAlice.HasSession(idBob.DID()) {
		t.Fatalf("Alice debe tener sesión activa tras finalizar 1-RTT")
	}
	decisionA2, _ := fwAlice.EvaluateInbound(idBob.DID(), 7777)
	if decisionA2 != l1.DecisionAccept {
		t.Fatalf("ZTNA Alice debe autorizar a Bob tras finalizar el handshake")
	}

	// 5. Cifrado y descifrado de datos en el datapath principal
	plaintext := []byte("PAQUETE_DATAPATH_PQC_IPVN7_CONFIDENCIAL")
	dataPkt, err := mgrAlice.EncryptDataPacket(idBob.DID(), plaintext)
	if err != nil {
		t.Fatalf("Error cifrando paquete de datos: %v", err)
	}

	if bytes.Equal(dataPkt.Payload, plaintext) {
		t.Fatalf("El payload del paquete de datos debe estar cifrado, no en texto plano")
	}

	decrypted, err := mgrBob.DecryptDataPacket(dataPkt)
	if err != nil {
		t.Fatalf("Error descifrando paquete de datos en Bob: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Texto descifrado (%s) no coincide con original (%s)", string(decrypted), string(plaintext))
	}
}

func TestPQCDatapath_PhysicalUDP_ZTNA_AntiReplay(t *testing.T) {
	connA, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Error abriendo UDP A: %v", err)
	}
	defer connA.Close()

	connB, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Error abriendo UDP B: %v", err)
	}
	defer connB.Close()

	addrA := connA.LocalAddr().(*net.UDPAddr)
	addrB := connB.LocalAddr().(*net.UDPAddr)

	idA, _ := l0.GenerateIdentity()
	idB, _ := l0.GenerateIdentity()
	idC, _ := l0.GenerateIdentity() // Atacante no autorizado

	keysA, _ := l1.GenerateHybridKeyPair(idA.DID())
	keysB, _ := l1.GenerateHybridKeyPair(idB.DID())

	fwA := l1.NewZTNAFirewall(true)
	fwB := l1.NewZTNAFirewall(true)

	mgrA := l1.NewPQCSessionManager(idA, keysA, fwA)
	mgrB := l1.NewPQCSessionManager(idB, keysB, fwB)
	antiReplayB := l1.NewAntiReplayFilter(nil)

	// 1. Handshake Init A -> B por UDP físico
	initPkt, err := mgrA.CreateHandshakeInitPacket(idB.DID(), keysB.ClassicalKEMPub, keysB.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Error creando init: %v", err)
	}
	rawInit, err := initPkt.Encode()
	if err != nil {
		t.Fatalf("Error codificando init: %v", err)
	}
	if _, err := connA.WriteToUDP(rawInit, addrB); err != nil {
		t.Fatalf("Error enviando init: %v", err)
	}

	buf := make([]byte, 2048)
	_ = connB.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := connB.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Error leyendo init en B: %v", err)
	}
	recInitPkt, err := l0.DecodePacket(buf[:n])
	if err != nil {
		t.Fatalf("Error decodificando init en B: %v", err)
	}

	// 2. B procesa Init y responde a A por UDP físico
	respPkt, err := mgrB.HandleHandshakeInitPacket(recInitPkt)
	if err != nil {
		t.Fatalf("Error en B procesando init: %v", err)
	}
	rawResp, err := respPkt.Encode()
	if err != nil {
		t.Fatalf("Error codificando resp: %v", err)
	}
	if _, err := connB.WriteToUDP(rawResp, addrA); err != nil {
		t.Fatalf("Error enviando resp: %v", err)
	}

	_ = connA.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err = connA.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Error leyendo resp en A: %v", err)
	}
	recRespPkt, err := l0.DecodePacket(buf[:n])
	if err != nil {
		t.Fatalf("Error decodificando resp en A: %v", err)
	}

	if err := mgrA.HandleHandshakeRespPacket(recRespPkt); err != nil {
		t.Fatalf("Error en A procesando resp: %v", err)
	}

	// 3. A envía paquete de datos cifrado con PQC a B
	secretPayload := []byte("DATAGRAMA_PQC_SOBERANO_EN_SOCKET_UDP_FISICO")
	dataPkt, err := mgrA.EncryptDataPacket(idB.DID(), secretPayload)
	if err != nil {
		t.Fatalf("Error cifrando datos en A: %v", err)
	}
	rawData, err := dataPkt.Encode()
	if err != nil {
		t.Fatalf("Error codificando paquete de datos: %v", err)
	}
	if _, err := connA.WriteToUDP(rawData, addrB); err != nil {
		t.Fatalf("Error enviando datos: %v", err)
	}

	// 4. B recibe por socket UDP y evalúa el pipeline completo:
	// Decode -> AntiReplay -> ZTNA -> Session -> Decrypt
	_ = connB.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err = connB.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Error leyendo datos en B: %v", err)
	}
	recDataPkt, err := l0.DecodePacket(buf[:n])
	if err != nil {
		t.Fatalf("Error decodificando datos: %v", err)
	}

	// Anti-Replay L1
	sKeysB, _ := mgrB.GetSession(recDataPkt.SourceDID)
	sessIDB := binary.BigEndian.Uint64(sKeysB.SessionID[:8])
	if !antiReplayB.Accept(recDataPkt.SourceDID, sessIDB, recDataPkt.Sequence, recDataPkt.Timestamp) {
		t.Fatalf("Anti-Replay debió aceptar el primer paquete con secuencia %d", recDataPkt.Sequence)
	}

	// ZTNA Default-Deny
	decision, _ := fwB.EvaluateInbound(recDataPkt.SourceDID, uint16(addrB.Port))
	if decision != l1.DecisionAccept {
		t.Fatalf("ZTNA debió aceptar al par autenticado %s", recDataPkt.SourceDID)
	}

	// Descifrado AEAD ChaCha20-Poly1305
	decrypted, err := mgrB.DecryptDataPacket(recDataPkt)
	if err != nil {
		t.Fatalf("Error descifrando payload: %v", err)
	}
	if !bytes.Equal(decrypted, secretPayload) {
		t.Fatalf("Payload descifrado no coincide con el original")
	}

	// 5. Test de Falsabilidad 1: Replay attack debe ser descartado
	if antiReplayB.Accept(recDataPkt.SourceDID, sessIDB, recDataPkt.Sequence, recDataPkt.Timestamp) {
		t.Fatalf("Anti-Replay debió RECHAZAR el paquete repetido!")
	}

	// 6. Test de Falsabilidad 2: Atacante C (no autenticado) debe ser bloqueado por ZTNA Default-Deny
	attackPkt := l0.NewPacket(l0.MsgTypeData, idC.DID(), idB.DID(), 999, nil, []byte("MALICIOUS"))
	decisionC, _ := fwB.EvaluateInbound(attackPkt.SourceDID, uint16(addrB.Port))
	if decisionC == l1.DecisionAccept {
		t.Fatalf("ZTNA debió BLOQUEAR por Default-Deny al atacante no autenticado C!")
	}
}
