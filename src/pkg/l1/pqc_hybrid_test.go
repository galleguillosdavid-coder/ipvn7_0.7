package l1

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestHybridPQCKeygenSignAndVerify(t *testing.T) {
	did := "did:ipvn7:test-node-pqc"
	kp, err := GenerateHybridKeyPair(did)
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair failed: %v", err)
	}

	if kp.Ed25519PubHex == "" || kp.X25519PubHex == "" || kp.MLDSAPubHex == "" || kp.MLKEMPubHex == "" {
		t.Fatalf("Claves públicas incompletas en HybridKeyPair: %+v", kp)
	}

	msg := []byte("Axiom: Zero-PII y Core Freeze inalterable en malla cuántica.")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if sig.Algorithm != HybridSigAlgorithm {
		t.Errorf("Algoritmo inesperado: %s", sig.Algorithm)
	}

	if !kp.Verify(msg, sig) {
		t.Fatalf("Verify rechazó firma híbrida válida!")
	}

	// Probar alteración de mensaje
	tamperedMsg := []byte("Axiom: Alteración maliciosa")
	if kp.Verify(tamperedMsg, sig) {
		t.Fatalf("Verify aceptó mensaje alterado!")
	}

	// Probar alteración de firma clásica
	sigCorrupt := *sig
	sigCorrupt.ClassicalSig = make([]byte, len(sig.ClassicalSig))
	copy(sigCorrupt.ClassicalSig, sig.ClassicalSig)
	sigCorrupt.ClassicalSig[0] ^= 0xFF
	if kp.Verify(msg, &sigCorrupt) {
		t.Fatalf("Verify aceptó firma clásica corrompida!")
	}

	// Probar alteración de firma cuántica
	sigCorruptPQC := *sig
	sigCorruptPQC.PQCSig = make([]byte, len(sig.PQCSig))
	copy(sigCorruptPQC.PQCSig, sig.PQCSig)
	sigCorruptPQC.PQCSig[0] ^= 0xFF
	if kp.Verify(msg, &sigCorruptPQC) {
		t.Fatalf("Verify aceptó firma cuántica corrompida!")
	}
}

func TestHybridPQCKEMAndAEAD(t *testing.T) {
	// Receptor (Bob)
	bobKP, err := GenerateHybridKeyPair("did:ipvn7:bob")
	if err != nil {
		t.Fatalf("Error generando clave de Bob: %v", err)
	}

	// Emisor (Alice) encapsula para Bob
	aliceKey, ciphertext, err := Encapsulate(bobKP.ClassicalKEMPub, bobKP.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Encapsulate failed: %v", err)
	}

	if len(aliceKey) != SharedSecretSize {
		t.Fatalf("Tamaño de clave compartida inválido: %d", len(aliceKey))
	}

	// Bob desencapsula
	bobKey, err := bobKP.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate failed: %v", err)
	}

	if !bytes.Equal(aliceKey, bobKey) {
		t.Fatalf("Claves compartidas no coinciden! Alice: %x, Bob: %x", aliceKey, bobKey)
	}

	// Probar cifrado y descifrado de carga útil con ChaCha20-Poly1305
	plaintext := []byte("Datagrama determinista ipvn7 de 1280 bytes protegido cuánticamente")
	ad := []byte("did:ipvn7:bob")

	encrypted, err := EncryptPayload(aliceKey, plaintext, ad)
	if err != nil {
		t.Fatalf("EncryptPayload failed: %v", err)
	}

	decrypted, err := DecryptPayload(bobKey, encrypted, ad)
	if err != nil {
		t.Fatalf("DecryptPayload failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("Descifrado no coincide con el texto plano original")
	}

	// Probar alteración de datos asociados (AD)
	_, err = DecryptPayload(bobKey, encrypted, []byte("did:ipvn7:mallory"))
	if err == nil {
		t.Fatalf("DecryptPayload debió fallar por AD alterado!")
	}
}

func TestRealFIPS203MLKEMIntegration(t *testing.T) {
	nodeB, err := GenerateHybridKeyPair("did:ipvn7:bob")
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair B failed: %v", err)
	}

	if nodeB.MLKEMDecapsKey == nil || nodeB.MLKEMEncapsKey == nil {
		t.Fatalf("Claves nativas FIPS 203 ML-KEM no inicializadas!")
	}

	keyA, cipher, err := Encapsulate(nodeB.ClassicalKEMPub, nodeB.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Encapsulate failed: %v", err)
	}

	if len(cipher.FullPQCCiphertext) != 1088 {
		t.Fatalf("Ciphertext FIPS 203 no tiene tamaño canónico (1088 bytes): %d", len(cipher.FullPQCCiphertext))
	}

	keyB, err := nodeB.Decapsulate(cipher)
	if err != nil {
		t.Fatalf("Decapsulate failed: %v", err)
	}

	if !bytes.Equal(keyA, keyB) {
		t.Fatalf("Secreto compartido post-cuántico FIPS 203 no coincide!")
	}
}

func TestXWingKEM_ConformityAndRoundtrip(t *testing.T) {
	nodeB, err := GenerateHybridKeyPair("did:ipvn7:bob-xwing")
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair failed: %v", err)
	}

	mlkemBytes := nodeB.MLKEMEncapsKey.Bytes()
	aliceShared, cipher, err := XWingEncapsulate(nodeB.ClassicalKEMPub, mlkemBytes)
	if err != nil {
		t.Fatalf("XWingEncapsulate failed: %v", err)
	}

	if len(cipher) != XWingCiphertextSize {
		t.Fatalf("Tamaño inválido de criptograma X-Wing: esperado %d, obtenido %d", XWingCiphertextSize, len(cipher))
	}

	bobShared, err := nodeB.XWingDecapsulate(cipher)
	if err != nil {
		t.Fatalf("XWingDecapsulate failed: %v", err)
	}

	if !bytes.Equal(aliceShared, bobShared) {
		t.Fatalf("Secreto compartido X-Wing no coincide entre emisor y receptor!")
	}

	// Probar fallo o rechazo implícito (IND-CCA2) ante criptograma corrupto
	corruptCipher := make([]byte, len(cipher))
	copy(corruptCipher, cipher)
	corruptCipher[0] ^= 0xFF
	corruptKey, err := nodeB.XWingDecapsulate(corruptCipher)
	if err == nil && bytes.Equal(aliceShared, corruptKey) {
		t.Fatalf("XWingDecapsulate no debió derivar la misma clave con un criptograma corrupto!")
	}
}

// TestPQC_PhysicalUDPLoopback_Bidirectional verifica el cierre de bucle físico completo
// (Loop Closure) entre dos sockets UDP reales del sistema operativo utilizando criptografía
// post-cuántica ML-KEM-768 real y AEAD ChaCha20-Poly1305.
func TestPQC_PhysicalUDPLoopback_Bidirectional(t *testing.T) {
	// 1. Iniciar sockets UDP reales en loopback
	connA, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Fallo abriendo UDP A: %v", err)
	}
	defer connA.Close()

	connB, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Fallo abriendo UDP B: %v", err)
	}
	defer connB.Close()

	addrB := connB.LocalAddr().(*net.UDPAddr)
	addrA := connA.LocalAddr().(*net.UDPAddr)

	// 2. Bob genera par de claves con FIPS 203 ML-KEM-768
	bobKP, err := GenerateHybridKeyPair("did:ipvn7:bob-loopback")
	if err != nil {
		t.Fatalf("GenerateHybridKeyPair Bob failed: %v", err)
	}

	// 3. Alice encapsula para Bob y cifra payload
	sharedKeyA, cipher, err := Encapsulate(bobKP.ClassicalKEMPub, bobKP.MLKEMPubHex)
	if err != nil {
		t.Fatalf("Encapsulate failed: %v", err)
	}

	rawPayloadA := []byte("DATAGRAMA_PQC_SOBERANO_A_TO_B_FIPS203")
	adA := []byte("ipvn7-session-a-b")
	encryptedA, err := EncryptPayload(sharedKeyA, rawPayloadA, adA)
	if err != nil {
		t.Fatalf("EncryptPayload failed: %v", err)
	}

	// 4. Alice envía criptograma + payload a Bob por UDP físico
	cipherBytes, err := json.Marshal(cipher)
	if err != nil {
		t.Fatalf("json.Marshal cipher failed: %v", err)
	}

	var wireBuf bytes.Buffer
	headerLen := uint16(len(cipherBytes))
	wireBuf.WriteByte(byte(headerLen >> 8))
	wireBuf.WriteByte(byte(headerLen & 0xFF))
	wireBuf.Write(cipherBytes)
	wireBuf.Write(encryptedA)

	if _, err := connA.WriteToUDP(wireBuf.Bytes(), addrB); err != nil {
		t.Fatalf("WriteToUDP A->B failed: %v", err)
	}

	// 5. Bob recibe en socket UDP físico
	bufB := make([]byte, 4096)
	_ = connB.SetReadDeadline(time.Now().Add(2 * time.Second))
	nB, _, err := connB.ReadFromUDP(bufB)
	if err != nil {
		t.Fatalf("ReadFromUDP B failed: %v", err)
	}

	// Deserializar en Bob
	r := bytes.NewReader(bufB[:nB])
	hLenH, _ := r.ReadByte()
	hLenL, _ := r.ReadByte()
	hLen := (int(hLenH) << 8) | int(hLenL)
	rawHeader := make([]byte, hLen)
	_, _ = r.Read(rawHeader)
	encPayload := make([]byte, r.Len())
	_, _ = r.Read(encPayload)

	var receivedCipher HybridKEMCiphertext
	if err := json.Unmarshal(rawHeader, &receivedCipher); err != nil {
		t.Fatalf("Bob json.Unmarshal failed: %v", err)
	}

	// Bob desencapsula
	sharedKeyB, err := bobKP.Decapsulate(&receivedCipher)
	if err != nil {
		t.Fatalf("Bob Decapsulate failed: %v", err)
	}

	if !bytes.Equal(sharedKeyA, sharedKeyB) {
		t.Fatalf("Claves compartidas discrepan tras transmisión física UDP")
	}

	// Bob descifra
	decryptedB, err := DecryptPayload(sharedKeyB, encPayload, adA)
	if err != nil {
		t.Fatalf("Bob DecryptPayload failed: %v", err)
	}

	if !bytes.Equal(rawPayloadA, decryptedB) {
		t.Fatalf("Carga útil descifrada por Bob no coincide: %s vs %s", string(decryptedB), string(rawPayloadA))
	}

	// 6. Retorno bidireccional B -> A por UDP físico con la clave compartida establecida
	rawPayloadB := []byte("DATAGRAMA_PQC_SOBERANO_B_TO_A_CONFIRMADO")
	adB := []byte("ipvn7-session-b-a")
	encryptedB, err := EncryptPayload(sharedKeyB, rawPayloadB, adB)
	if err != nil {
		t.Fatalf("EncryptPayload B->A failed: %v", err)
	}

	if _, err := connB.WriteToUDP(encryptedB, addrA); err != nil {
		t.Fatalf("WriteToUDP B->A failed: %v", err)
	}

	// Alice recibe en su socket
	bufA := make([]byte, 2048)
	_ = connA.SetReadDeadline(time.Now().Add(2 * time.Second))
	nA, _, err := connA.ReadFromUDP(bufA)
	if err != nil {
		t.Fatalf("ReadFromUDP A failed: %v", err)
	}

	decryptedA, err := DecryptPayload(sharedKeyA, bufA[:nA], adB)
	if err != nil {
		t.Fatalf("Alice DecryptPayload failed: %v", err)
	}

	if !bytes.Equal(rawPayloadB, decryptedA) {
		t.Fatalf("Carga útil descifrada por Alice no coincide: %s vs %s", string(decryptedA), string(rawPayloadB))
	}
}
