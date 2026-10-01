package l0_test

import (
	"crypto/rand"
	"testing"

	"ipvn7/pkg/l0"
)

// TestFuzz_PacketBoundaryAndTruncation evalua el rechazo seguro de tramas truncadas o de tamano invalido
func TestFuzz_PacketBoundaryAndTruncation(t *testing.T) {
	// 1. Buffers de tamano nulo o ultra-reducido
	for size := 0; size < 16; size++ {
		dummy := make([]byte, size)
		_, _ = rand.Read(dummy)
		_, err := l0.DecodePacket(dummy)
		if err == nil {
			t.Errorf("DecodePacket debio fallar con buffer de %d bytes", size)
		}
	}

	// 2. Trama desbordando MTU canónico de 1280B
	oversized := make([]byte, l0.MaxPacketSize+1)
	_, _ = rand.Read(oversized)
	_, err := l0.DecodePacket(oversized)
	if err == nil {
		t.Errorf("DecodePacket debio rechazar trama que excede MaxPacketSize (1281 bytes)")
	}
}

// TestFuzz_ChecksumTamperingBitFlips verifica que la alteracion de 1 solo bit sea descartada deterministamente
func TestFuzz_ChecksumTamperingBitFlips(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}

	nonce, _ := l0.GenerateNonce()
	pkt := l0.NewPacket(l0.MsgTypeData, id.DID(), "did:ipvn7:peer_target", 1, nonce, []byte("Payload seguro"))
	_ = pkt.SignPacket(id)
	enc, err := pkt.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	expectedChecksum := l0.FastPacketChecksum(enc)
	if !l0.VerifyPacketChecksum(enc, expectedChecksum) {
		t.Fatalf("Checksum original deberia ser valido")
	}

	// Probar voltear un bit en cada posicion de los primeros 64 bytes
	testLimit := len(enc)
	if testLimit > 64 {
		testLimit = 64
	}

	for i := 0; i < testLimit; i++ {
		corrupted := make([]byte, len(enc))
		copy(corrupted, enc)
		corrupted[i] ^= 0x01 // Bitflip

		if l0.VerifyPacketChecksum(corrupted, expectedChecksum) {
			t.Errorf("Colision o falla en tiempo constante: bitflip en byte %d paso checksum", i)
		}
	}
}

// TestFuzz_MalformedPQCEncapsulation verifica que ciphertexts corruptos no provoquen panic en ML-KEM
func TestFuzz_MalformedPQCEncapsulation(t *testing.T) {
	kp, err := l0.GenerateMLKEM768KeyPair()
	if err != nil {
		t.Fatalf("GenerateMLKEM768KeyPair falló: %v", err)
	}
	kem := l0.NewMLKEM768Adapter(kp)

	// Ciphertext truncado o vacio debe retornar error
	_, err = kem.Decapsulate([]byte("invalido"))
	if err == nil {
		t.Errorf("Decapsulate debio fallar con longitud de ciphertext invalida")
	}

	// Ciphertext de tamano correcto (1088 bytes) pero con datos aleatorios corruptos:
	// NIST FIPS 203 especifica rechazo implicito (deriva clave pseudoaleatoria sin panics)
	corruptCT := make([]byte, 1088)
	_, _ = rand.Read(corruptCT)
	ss, err := kem.Decapsulate(corruptCT)
	if err != nil {
		t.Errorf("FIPS 203 decapsulate no debio fallar con error en rechazo implicito: %v", err)
	}
	if len(ss) != 32 {
		t.Errorf("se esperaba secreto de 32 bytes en rechazo implicito FIPS 203")
	}
}

// TestFuzz_TamperedSignatureRejection valida que la modificacion de campos anula la firma sin panics
func TestFuzz_TamperedSignatureRejection(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	nonce, _ := l0.GenerateNonce()
	pkt := l0.NewPacket(l0.MsgTypeData, id.DID(), "did:ipvn7:target", 42, nonce, []byte("Autentico"))
	_ = pkt.SignPacket(id)

	// Alterar el payload tras firmar
	pkt.Payload = []byte("Inyectado")
	valid, err := pkt.VerifyPacketSignature()
	if err != nil || valid {
		t.Errorf("Firma alterada no fue rechazada: valid=%v, err=%v", valid, err)
	}
}

// TestMTUBoundary_ExactAndExcess verifica formalmente el comportamiento estricto en la frontera de MTU (Fase 12):
// - Paquete <= 1280 bytes: procesado correctamente.
// - Paquete == 1281 bytes: rechazado inmediatamente sin desbordamiento ni pánico.
func TestMTUBoundary_ExactAndExcess(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}

	// 1. Validar rechazo tajante de exactamente 1281 bytes (MTU + 1)
	excessData := make([]byte, l0.MaxPacketSize+1)
	_, err = l0.DecodePacket(excessData)
	if err == nil {
		t.Fatalf("DecodePacket debió rechazar paquete de 1281 bytes")
	}

	// 2. Construir paquete con payload calibrado para alcanzar exactamente la frontera
	nonce, _ := l0.GenerateNonce()
	basePkt := l0.NewPacket(l0.MsgTypeData, id.DID(), "did:ipvn7:destination", 100, nonce, nil)
	_ = basePkt.SignPacket(id)
	baseEnc, err := basePkt.Encode()
	if err != nil {
		t.Fatalf("base Encode: %v", err)
	}

	remainingBytes := l0.MaxPacketSize - len(baseEnc)
	if remainingBytes > 0 {
		fillPayload := make([]byte, remainingBytes-10) // Margen para overhead CBOR
		pktBoundary := l0.NewPacket(l0.MsgTypeData, id.DID(), "did:ipvn7:destination", 101, nonce, fillPayload)
		_ = pktBoundary.SignPacket(id)
		encBoundary, err := pktBoundary.Encode()
		if err != nil {
			t.Fatalf("encBoundary: %v", err)
		}
		if len(encBoundary) > l0.MaxPacketSize {
			t.Fatalf("Paquete excedió MTU determinista: %d > %d", len(encBoundary), l0.MaxPacketSize)
		}
		decoded, err := l0.DecodePacket(encBoundary)
		if err != nil {
			t.Fatalf("DecodePacket falló en frontera válida de MTU: %v", err)
		}
		if decoded == nil {
			t.Fatalf("Paquete decodificado es nil")
		}
	}
}
