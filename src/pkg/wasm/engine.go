package wasm

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"ipvn7/pkg/l0"
)

// WasmIdentity representa un par de identidad soberana exportable a entornos WASM/JS
type WasmIdentity struct {
	DID        string `json:"did"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// WasmEngine provee primitivas criptográficas y de red optimizadas para ejecución en WebAssembly
type WasmEngine struct {
	version string
}

// NewWasmEngine inicializa el motor WASM soberano
func NewWasmEngine() *WasmEngine {
	return &WasmEngine{version: "0.7.0-wasm"}
}

// Version retorna la versión actual del motor WASM
func (e *WasmEngine) Version() string {
	return e.version
}

// GenerateIdentity crea una identidad criptográfica soberana en el sandbox de WebAssembly
func (e *WasmEngine) GenerateIdentity() (*WasmIdentity, error) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		return nil, fmt.Errorf("wasm_engine: error generando identidad: %w", err)
	}
	return &WasmIdentity{
		DID:        id.DID(),
		PublicKey:  hex.EncodeToString(id.PublicKey),
		PrivateKey: hex.EncodeToString(id.PrivateKey),
	}, nil
}

// SignMessage firma un mensaje arbitrario con una clave privada Ed25519 en formato hex
func (e *WasmEngine) SignMessage(privHex string, msg []byte) (string, error) {
	privBytes, err := hex.DecodeString(privHex)
	if err != nil || len(privBytes) != ed25519.PrivateKeySize {
		return "", errors.New("wasm_engine: clave privada invalida (debe ser 64 bytes hex)")
	}
	sig := ed25519.Sign(ed25519.PrivateKey(privBytes), msg)
	return hex.EncodeToString(sig), nil
}

// VerifySignature verifica una firma Ed25519 contra una clave pública en formato hex
func (e *WasmEngine) VerifySignature(pubHex string, msg []byte, sigHex string) bool {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		return false
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(pubBytes), msg, sigBytes)
}

// ComputeProofOfWork resuelve un Micro-PoW para arranque en frío o mitigación Sybil en el navegador
func (e *WasmEngine) ComputeProofOfWork(did string, seq uint64, difficultyBits int) (uint64, error) {
	if difficultyBits <= 0 || difficultyBits > 16 {
		difficultyBits = 8 // Default 8 bits (~256 iteraciones)
	}
	mask := uint8(0xFF << (8 - (difficultyBits % 9)))
	if difficultyBits > 8 {
		mask = 0xFF
	}

	var nonce uint64
	var seqBytes [8]byte
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(seqBytes[:], seq)

	maxAttempts := uint64(1000000)
	for nonce = 1; nonce < maxAttempts; nonce++ {
		binary.BigEndian.PutUint64(nonceBytes[:], nonce)
		h := sha256.New()
		h.Write([]byte(did))
		h.Write(seqBytes[:])
		h.Write(nonceBytes[:])
		digest := h.Sum(nil)

		if (digest[0] & mask) == 0x00 {
			if difficultyBits <= 8 || (digest[1]&0xF0) == 0x00 {
				return nonce, nil
			}
		}
	}
	return 0, errors.New("wasm_engine: limite de iteraciones PoW excedido")
}

// ValidatePacketFrame valida estructuralmente una trama canónica de 1280B de IPVN7
func (e *WasmEngine) ValidatePacketFrame(packetHex string) (bool, error) {
	raw, err := hex.DecodeString(packetHex)
	if err != nil {
		return false, errors.New("wasm_engine: formato hexadecimal corrupto")
	}
	if len(raw) != 1280 {
		return false, fmt.Errorf("wasm_engine: longitud de trama invalida (%d != 1280)", len(raw))
	}
	// Verificación de número mágico IPVN7 o cabecera canónica
	if raw[0] != 0x07 && raw[0] != 0x17 { // 0x07 = IPVN7 Raw, 0x17 = TLS Masquerade
		return false, errors.New("wasm_engine: cabecera o numero magico desconocido")
	}
	return true, nil
}

// TimestampNow retorna la marca de tiempo Unix actual para sincronización de época
func (e *WasmEngine) TimestampNow() int64 {
	return time.Now().Unix()
}
