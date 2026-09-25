package wasm

import (
	"encoding/hex"
	"testing"
)

func TestWasmEngine_GenerateIdentityAndSignVerify(t *testing.T) {
	engine := NewWasmEngine()
	if engine.Version() != "0.7.0-wasm" {
		t.Fatalf("Version inesperada: %s", engine.Version())
	}

	id, err := engine.GenerateIdentity()
	if err != nil {
		t.Fatalf("Fallo generando identidad WASM: %v", err)
	}
	if len(id.DID) == 0 || len(id.PublicKey) != 64 || len(id.PrivateKey) != 128 {
		t.Fatalf("Campos de identidad malformados: %+v", id)
	}

	msg := []byte("IPVN7 Sovereign Datagram Payload for WASM")
	sigHex, err := engine.SignMessage(id.PrivateKey, msg)
	if err != nil {
		t.Fatalf("Fallo firmando mensaje: %v", err)
	}
	if len(sigHex) != 128 {
		t.Fatalf("Longitud de firma invalida: %d", len(sigHex))
	}

	valid := engine.VerifySignature(id.PublicKey, msg, sigHex)
	if !valid {
		t.Fatalf("Firma debio ser valida")
	}

	tamperedMsg := []byte("IPVN7 Tampered Datagram Payload")
	if engine.VerifySignature(id.PublicKey, tamperedMsg, sigHex) {
		t.Fatalf("Firma sobre mensaje alterado no debio ser valida")
	}
}

func TestWasmEngine_ComputeProofOfWork(t *testing.T) {
	engine := NewWasmEngine()
	did := "did:ipvn7:wasm_test_client_01"
	seq := uint64(100500)

	nonce, err := engine.ComputeProofOfWork(did, seq, 8)
	if err != nil {
		t.Fatalf("Error calculando PoW en WASM: %v", err)
	}
	if nonce == 0 {
		t.Fatalf("Nonce no debio ser cero")
	}
}

func TestWasmEngine_ValidatePacketFrame(t *testing.T) {
	engine := NewWasmEngine()

	// 1. Trama canónica válida (1280B, prefijo 0x07)
	validFrame := make([]byte, 1280)
	validFrame[0] = 0x07
	ok, err := engine.ValidatePacketFrame(hex.EncodeToString(validFrame))
	if !ok || err != nil {
		t.Fatalf("Trama valida rechazada: %v", err)
	}

	// 2. Trama corta rechazada
	shortFrame := make([]byte, 100)
	shortFrame[0] = 0x07
	okShort, errShort := engine.ValidatePacketFrame(hex.EncodeToString(shortFrame))
	if okShort || errShort == nil {
		t.Fatalf("Trama corta debio ser rechazada")
	}

	// 3. Trama con número mágico corrupto rechazada
	badMagic := make([]byte, 1280)
	badMagic[0] = 0xFF
	okBad, errBad := engine.ValidatePacketFrame(hex.EncodeToString(badMagic))
	if okBad || errBad == nil {
		t.Fatalf("Trama con numero magico invalido debio ser rechazada")
	}
}
