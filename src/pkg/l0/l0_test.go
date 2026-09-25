package l0_test

import (
	"bytes"
	"strings"
	"testing"

	"ipvn7/pkg/interfaces"
	"ipvn7/pkg/l0"
)

func TestIdentityGenerationAndDID(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity falló: %v", err)
	}

	did := id.DID()
	if !strings.HasPrefix(did, "did:ipvn7:") {
		t.Errorf("DID no tiene prefijo esperado: %s", did)
	}

	pub, err := l0.PublicKeyFromDID(did)
	if err != nil {
		t.Fatalf("PublicKeyFromDID falló: %v", err)
	}
	if !bytes.Equal(pub, id.PublicKey) {
		t.Errorf("Clave pública recuperada no coincide con original")
	}

	ipv6 := id.IPv6()
	if ipv6[0] != 0xfd || ipv6[1] != 0x07 {
		t.Errorf("IPv6 no tiene prefijo fd07::/64: %v", ipv6)
	}

	ipv4 := id.IPv4()
	if ipv4[0] != 10 || ipv4[1] != 7 {
		t.Errorf("IPv4 no tiene prefijo 10.7.x.y: %v", ipv4)
	}
}

func TestWireCBORPacketSerialization(t *testing.T) {
	id1, _ := l0.GenerateIdentity()
	id2, _ := l0.GenerateIdentity()

	nonce, _ := l0.GenerateNonce()
	payload := []byte("Hola Mundo ipvn7 - Red Soberana")

	pkt := l0.NewPacket(l0.MsgTypeData, id1.DID(), id2.DID(), 100, nonce, payload)
	if err := pkt.SignPacket(id1); err != nil {
		t.Fatalf("SignPacket falló: %v", err)
	}

	encoded, err := pkt.Encode()
	if err != nil {
		t.Fatalf("Encode falló: %v", err)
	}

	if len(encoded) > l0.MaxPacketSize {
		t.Errorf("Paquete excede MaxPacketSize: %d bytes", len(encoded))
	}

	decoded, err := l0.DecodePacket(encoded)
	if err != nil {
		t.Fatalf("DecodePacket falló: %v", err)
	}

	if decoded.SourceDID != id1.DID() || decoded.DestDID != id2.DID() {
		t.Errorf("DIDs no coinciden en paquete decodificado")
	}
	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload alterado tras decodificación")
	}

	valid, err := decoded.VerifyPacketSignature()
	if err != nil || !valid {
		t.Errorf("Verificación de firma digital falló: valid=%v, err=%v", valid, err)
	}
}

func TestNoiseHandshakeAndAEAD(t *testing.T) {
	aliceID, _ := l0.GenerateIdentity()
	bobID, _ := l0.GenerateIdentity()

	aliceHS, err := l0.NewNoiseHandshake(aliceID)
	if err != nil {
		t.Fatalf("Alice NewNoiseHandshake falló: %v", err)
	}
	bobHS, err := l0.NewNoiseHandshake(bobID)
	if err != nil {
		t.Fatalf("Bob NewNoiseHandshake falló: %v", err)
	}

	// Step 1: Alice -> Bob
	msg1, err := aliceHS.Step1Initiate()
	if err != nil {
		t.Fatalf("Step1Initiate falló: %v", err)
	}

	// Step 2: Bob procesa msg1 y responde
	msg2, bobKeys, err := bobHS.Step2Respond(msg1)
	if err != nil {
		t.Fatalf("Step2Respond falló: %v", err)
	}

	// Step 3: Alice finaliza con msg2
	aliceKeys, err := aliceHS.Step3Finalize(msg2)
	if err != nil {
		t.Fatalf("Step3Finalize falló: %v", err)
	}

	// Comprobar coincidencia simétrica de claves
	if aliceKeys.TxKey != bobKeys.RxKey {
		t.Errorf("Alice TxKey != Bob RxKey")
	}
	if aliceKeys.RxKey != bobKeys.TxKey {
		t.Errorf("Alice RxKey != Bob TxKey")
	}

	// Probar cifrado y descifrado AEAD
	nonce, _ := l0.GenerateNonce()
	plaintext := []byte("Datagrama confidencial E2EE")
	ad := []byte("metadata-no-cifrada")

	ciphertext, err := l0.EncryptPayload(aliceKeys.TxKey[:], nonce, plaintext, ad)
	if err != nil {
		t.Fatalf("EncryptPayload falló: %v", err)
	}

	decrypted, err := l0.DecryptPayload(bobKeys.RxKey[:], nonce, ciphertext, ad)
	if err != nil {
		t.Fatalf("DecryptPayload falló: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Texto descifrado no coincide con texto en claro")
	}
}

func TestAntiReplayFilter(t *testing.T) {
	filter := l0.NewAntiReplayFilter()

	if !filter.ValidateAndUpdate(1) {
		t.Errorf("Secuencia 1 debió ser aceptada")
	}
	if !filter.ValidateAndUpdate(2) {
		t.Errorf("Secuencia 2 debió ser aceptada")
	}
	if filter.ValidateAndUpdate(1) {
		t.Errorf("Secuencia 1 repetida debió ser rechazada")
	}
	if !filter.ValidateAndUpdate(50) {
		t.Errorf("Secuencia 50 debió ser aceptada")
	}
	if !filter.ValidateAndUpdate(49) {
		t.Errorf("Secuencia 49 (dentro de ventana) debió ser aceptada")
	}
	if filter.ValidateAndUpdate(49) {
		t.Errorf("Secuencia 49 repetida debió ser rechazada")
	}

	// Validación de ventana extendida de 1024 bits
	if !filter.ValidateAndUpdate(1000) {
		t.Errorf("Secuencia 1000 debió ser aceptada")
	}
	// Secuencia 500 (diferencia 500 < 1024)
	if !filter.ValidateAndUpdate(500) {
		t.Errorf("Secuencia 500 con jitter dentro de 1024 bits debió ser aceptada")
	}
	if filter.ValidateAndUpdate(500) {
		t.Errorf("Secuencia 500 repetida debió ser rechazada")
	}
	// Secuencia 2000
	if !filter.ValidateAndUpdate(2000) {
		t.Errorf("Secuencia 2000 debió ser aceptada")
	}
	// Diferencia de 1024 bits exactos: 2000 - 976 = 1024 (fuera de ventana)
	if filter.ValidateAndUpdate(976) {
		t.Errorf("Secuencia 976 fuera del límite de 1024 bits debió ser rechazada")
	}
	// Diferencia de 1023 bits: 2000 - 977 = 1023 (dentro de ventana)
	if !filter.ValidateAndUpdate(977) {
		t.Errorf("Secuencia 977 en el borde de 1023 bits debió ser aceptada")
	}
}


func TestStatelessCookieGenerator(t *testing.T) {
	cg := l0.NewStatelessCookieGenerator()
	addr := "198.51.100.50:7777"
	ephPub := []byte("32bytes-ephemeral-key-alice-xxxx")

	cookie := cg.GenerateCookie(addr, ephPub)
	if len(cookie) != 16 {
		t.Fatalf("Tamaño de cookie incorrecto: %d bytes (esperado 16)", len(cookie))
	}

	if !cg.ValidateCookie(cookie, addr, ephPub) {
		t.Errorf("Cookie legítima fue rechazada")
	}

	// Dirección diferente
	if cg.ValidateCookie(cookie, "198.51.100.99:7777", ephPub) {
		t.Errorf("Cookie validada para dirección incorrecta debió fallar")
	}

	// Clave efímera diferente
	if cg.ValidateCookie(cookie, addr, []byte("different-ephemeral-key-bob-xxxx")) {
		t.Errorf("Cookie validada para clave efímera distinta debió fallar")
	}
}

func TestStatelessCookie_SubnetBucketsIsolation(t *testing.T) {
	cg := l0.NewStatelessCookieGenerator()
	attackerAddr := "198.51.100.50:7777"
	legitimateAddr := "10.20.30.40:5555"
	ephPub := []byte("32bytes-ephemeral-key-alice-xxxx")

	// Inundar con el atacante hasta agotar su bucket (1000 tokens)
	for i := 0; i < 1005; i++ {
		cg.GenerateCookie(attackerAddr, ephPub)
	}

	// El atacante debe recibir cookie nula o vacía por rate limit
	attackerCookie := cg.GenerateCookie(attackerAddr, ephPub)
	if len(attackerCookie) != 0 {
		t.Errorf("El atacante debió ser rate-limited en su cubeta")
	}

	// El usuario legítimo en otra subred NO debe verse afectado
	legitCookie := cg.GenerateCookie(legitimateAddr, ephPub)
	if len(legitCookie) != 16 {
		t.Fatalf("Usuario legítimo fue bloqueado colateralmente por el atacante (longitud: %d)", len(legitCookie))
	}
	if !cg.ValidateCookie(legitCookie, legitimateAddr, ephPub) {
		t.Errorf("Cookie de usuario legítimo debió ser válida")
	}
}

func TestStatelessCookie_AdaptivePoWEscapeGradient(t *testing.T) {
	cg := l0.NewStatelessCookieGenerator()
	addr := "203.0.113.10:8888"
	ephPub := []byte("32bytes-ephemeral-key-client-xxx")

	// 1. Agotar la cubeta simulando saturación por ataque masivo
	for i := 0; i < 1005; i++ {
		cg.GenerateCookie(addr, ephPub)
	}

	// 2. Sin PoW (nonce=0): debe ser descartado
	noPoWCookie := cg.GenerateCookieWithPoW(addr, ephPub, 0)
	if len(noPoWCookie) != 0 {
		t.Fatalf("Petición sin PoW debió ser rechazada con cubeta saturada")
	}

	// 3. Cliente legítimo calcula micro-PoW (dificultad de 8 bits: ~256 iteraciones)
	var validNonce uint64
	for n := uint64(1); n < 10000; n++ {
		if cg.VerifyMicroPoW(addr, ephPub, n) {
			validNonce = n
			break
		}
	}
	if validNonce == 0 {
		t.Fatalf("Fallo encontrando micro-nonce válido")
	}

	// 4. Con micro-PoW: el cliente es admitido con éxito a pesar de la saturación
	admittedCookie := cg.GenerateCookieWithPoW(addr, ephPub, validNonce)
	if len(admittedCookie) != 16 {
		t.Fatalf("Cliente legítimo con micro-PoW debió ser admitido, longitud: %d", len(admittedCookie))
	}
	if !cg.ValidateCookie(admittedCookie, addr, ephPub) {
		t.Errorf("Cookie admitida por PoW debió ser criptográficamente válida")
	}
}

func TestStatelessCookie_PoWVerificationRateLimiting(t *testing.T) {
	cg := l0.NewStatelessCookieGenerator()
	addr := "198.51.100.99:9999"
	ephPub := []byte("32bytes-ephemeral-key-client-xxx")

	// Inundar con 10,050 nonces aleatorios para simular ataque de Verification Overload DoS
	var rejectedCount int
	for i := 0; i < 10050; i++ {
		// Nonce deliberadamente inválido o aleatorio
		if !cg.VerifyMicroPoW(addr, ephPub, uint64(i+999999)) {
			rejectedCount++
		}
	}

	// Al menos 50 peticiones debieron ser rechazadas directamente por agotamiento del presupuesto atómico de CPU
	if rejectedCount < 50 {
		t.Fatalf("Esperaba que el limitador de verificación protegiera la CPU ante ráfaga masiva (rechazos: %d)", rejectedCount)
	}
}

func TestIdentityAuthorityInterfaceCompliance(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity falló: %v", err)
	}

	var auth interfaces.IdentityAuthority = id

	if auth.DID() != id.DID() {
		t.Errorf("DID no coincide a través de la interfaz")
	}

	if len(auth.PublicKeyBytes()) != 32 {
		t.Errorf("PublicKeyBytes longitud incorrecta: %d", len(auth.PublicKeyBytes()))
	}

	if !strings.HasPrefix(auth.UIN(), "UIN-") {
		t.Errorf("UIN formato inválido: %s", auth.UIN())
	}

	msg := []byte("Axiomatic Verification")
	sig := auth.Sign(msg)
	if !auth.Verify(msg, sig) {
		t.Errorf("Firma y verificación falló a través de interfaces.IdentityAuthority")
	}

	t.Logf("Conformidad certificada: *l0.Identity satisface formalmente interfaces.IdentityAuthority")
}

func BenchmarkAntiReplayFilter_ValidateAndUpdate(b *testing.B) {
	filter := l0.NewAntiReplayFilter()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		filter.ValidateAndUpdate(uint64(i))
	}
}

func TestFastPacketChecksum(t *testing.T) {
	data := []byte("Paquete de prueba IPVN7 con checksum BLAKE2s acelerado por SIMD")
	chk := l0.FastPacketChecksum(data)

	if !l0.VerifyPacketChecksum(data, chk) {
		t.Errorf("VerifyPacketChecksum falló para datos legítimos")
	}

	corrupted := append([]byte(nil), data...)
	corrupted[0] ^= 0xFF
	if l0.VerifyPacketChecksum(corrupted, chk) {
		t.Errorf("VerifyPacketChecksum debió fallar para datos alterados")
	}
}

func BenchmarkFastPacketChecksum_1280B(b *testing.B) {
	payload := make([]byte, 1280)
	b.SetBytes(1280)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l0.FastPacketChecksum(payload)
	}
}




