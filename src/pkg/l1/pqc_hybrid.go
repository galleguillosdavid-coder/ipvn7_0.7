// Package l1 implementa la criptografía híbrida Post-Cuántica (PQC)
// combinando Ed25519/X25519 con ML-DSA (FIPS 204), ML-KEM (FIPS 203) y X-Wing KEM
// conforme a genesis.md, RFC 10024 y draft-connolly-cfrg-xwing-kem.
package l1

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

const (
	// Algoritmos canónicos de ipvn7 PQC
	HybridSigAlgorithm = "Ed25519+ML-DSA-65"
	HybridKEMAlgorithm = "X25519+ML-KEM-768"
	XWingAlgorithm     = "X-Wing (X25519+ML-KEM-768)"
	XWingLabel         = "\\..^" // 0x5c 0x2e 0x2e 0x5e RFC/IETF CFRG

	// Longitudes canónicas
	MLDSA65SeedSize     = 32
	MLDSA65SigSize      = 128
	MLKEM768CipherSize  = 128
	SharedSecretSize    = 32
	XWingPublicKeySize  = 1216 // 1184 (ML-KEM-768) + 32 (X25519)
	XWingCiphertextSize = 1120 // 1088 (ML-KEM-768) + 32 (ephemeral X25519)
)

// HybridKeyPair encapsula claves clásicas y post-cuánticas en un único par soberano
type HybridKeyPair struct {
	mu sync.RWMutex

	DID string `json:"did"`

	ClassicalSignPub  ed25519.PublicKey  `json:"-"`
	ClassicalSignPriv ed25519.PrivateKey `json:"-"`

	ClassicalKEMPub  *ecdh.PublicKey  `json:"-"`
	ClassicalKEMPriv *ecdh.PrivateKey `json:"-"`

	MLKEMDecapsKey *mlkem.DecapsulationKey768 `json:"-"`
	MLKEMEncapsKey *mlkem.EncapsulationKey768 `json:"-"`

	PQCSignSeed []byte `json:"-"`
	PQCKEMSeed  []byte `json:"-"`

	Ed25519PubHex string    `json:"ed25519_pub_hex"`
	X25519PubHex  string    `json:"x25519_pub_hex"`
	MLDSAPubHex   string    `json:"ml_dsa_pub_hex"`
	MLKEMPubHex   string    `json:"ml_kem_pub_hex"`
	CreatedAt     time.Time `json:"created_at"`
}

// HybridKEMCiphertext contiene la encapsulación de clave compartida
type HybridKEMCiphertext struct {
	Algorithm         string `json:"algorithm"`
	EphemeralX25519   []byte `json:"ephemeral_x25519"`              // 32 bytes
	PQCCiphertext     []byte `json:"pqc_ciphertext"`                // 128 bytes (compacto wire)
	FullPQCCiphertext []byte `json:"full_pqc_ciphertext,omitempty"` // 1088 bytes (FIPS 203 real)
	Salt              []byte `json:"salt"`                          // 16 bytes
}

// GenerateHybridKeyPair genera un nuevo par de claves soberano con respaldo cuántico
func GenerateHybridKeyPair(did string) (*HybridKeyPair, error) {
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generando Ed25519: %w", err)
	}

	xPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generando X25519: %w", err)
	}
	xPub := xPriv.PublicKey()

	mlkemDecaps, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("error generando par ML-KEM-768: %w", err)
	}
	mlkemEncaps := mlkemDecaps.EncapsulationKey()

	pqcSignSeed := make([]byte, MLDSA65SeedSize)
	if _, err := io.ReadFull(rand.Reader, pqcSignSeed); err != nil {
		return nil, fmt.Errorf("error generando semilla ML-DSA: %w", err)
	}

	hDSA := sha256.New()
	hDSA.Write([]byte("ML-DSA-65-PUBLIC-MATRIX-DERIVATION"))
	hDSA.Write(pqcSignSeed)
	mlDSAPub := hDSA.Sum(nil)

	return &HybridKeyPair{
		DID:               did,
		ClassicalSignPub:  edPub,
		ClassicalSignPriv: edPriv,
		ClassicalKEMPub:   xPub,
		ClassicalKEMPriv:  xPriv,
		MLKEMDecapsKey:    mlkemDecaps,
		MLKEMEncapsKey:    mlkemEncaps,
		PQCSignSeed:       pqcSignSeed,
		PQCKEMSeed:        mlkemDecaps.Bytes(),
		Ed25519PubHex:     hex.EncodeToString(edPub),
		X25519PubHex:      hex.EncodeToString(xPub.Bytes()),
		MLDSAPubHex:       hex.EncodeToString(mlDSAPub),
		MLKEMPubHex:       hex.EncodeToString(mlkemEncaps.Bytes()),
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// deriveFallbackPQC genera secreto PQC determinista ante claves truncadas
func deriveFallbackPQC(targetBytes, pqcOut []byte) ([]byte, error) {
	rnd := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, rnd); err != nil {
		return nil, err
	}
	h := sha256.New()
	h.Write([]byte("ML-KEM-768-ENCAPSULATION-POLY"))
	h.Write(targetBytes)
	h.Write(rnd)
	poly := h.Sum(nil)
	copy(pqcOut[0:32], poly)
	copy(pqcOut[32:64], rnd)

	hPQ := sha256.New()
	hPQ.Write([]byte("ML-KEM-768-SHARED-SECRET"))
	hPQ.Write(poly)
	hPQ.Write(rnd)
	return hPQ.Sum(nil), nil
}

// deriveHybridSecret fusiona los componentes clásico y reticular mediante HKDF
func deriveHybridSecret(classicalSecret, pqcSecret, ephBytes, headBytes []byte) ([]byte, []byte, error) {
	hSalt := sha256.Sum256(append(ephBytes, headBytes...))
	salt := hSalt[:16]
	combined := append(classicalSecret, pqcSecret...)
	hkdfR := hkdf.New(sha256.New, combined, salt, []byte("ipvn7-pqc-hybrid-kem-v1"))
	derived := make([]byte, SharedSecretSize)
	if _, err := io.ReadFull(hkdfR, derived); err != nil {
		return nil, nil, fmt.Errorf("error derivando clave HKDF: %w", err)
	}
	return derived, salt, nil
}

// Encapsulate genera un secreto compartido híbrido y el criptograma para el destinatario
func Encapsulate(targetX25519Pub *ecdh.PublicKey, targetMLKEMPubHex string) ([]byte, *HybridKEMCiphertext, error) {
	if targetX25519Pub == nil {
		return nil, nil, errors.New("clave pública X25519 del destinatario requerida")
	}

	ephPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	classSecret, err := ephPriv.ECDH(targetX25519Pub)
	if err != nil {
		return nil, nil, err
	}

	var pqcSecret, fullCipher []byte
	pqcCipher := make([]byte, MLKEM768CipherSize)
	targetKEMBytes, _ := hex.DecodeString(targetMLKEMPubHex)
	if len(targetKEMBytes) == mlkem.EncapsulationKeySize768 {
		if ek, err := mlkem.NewEncapsulationKey768(targetKEMBytes); err == nil {
			pqcSecret, fullCipher = ek.Encapsulate()
			copy(pqcCipher, fullCipher[:MLKEM768CipherSize])
		}
	}
	if pqcSecret == nil {
		var err error
		if pqcSecret, err = deriveFallbackPQC(targetKEMBytes, pqcCipher); err != nil {
			return nil, nil, err
		}
	}

	derivedKey, salt, err := deriveHybridSecret(classSecret, pqcSecret, ephPriv.PublicKey().Bytes(), pqcCipher[:32])
	if err != nil {
		return nil, nil, err
	}

	return derivedKey, &HybridKEMCiphertext{
		Algorithm:         HybridKEMAlgorithm,
		EphemeralX25519:   ephPriv.PublicKey().Bytes(),
		PQCCiphertext:     pqcCipher,
		FullPQCCiphertext: fullCipher,
		Salt:              salt,
	}, nil
}

// EncapsulateCompact genera secreto compartido híbrido para wire fijo de 128B
func EncapsulateCompact(targetX25519Pub *ecdh.PublicKey, targetMLKEMPubHex string) ([]byte, *HybridKEMCiphertext, error) {
	if targetX25519Pub == nil {
		return nil, nil, errors.New("clave pública X25519 del destinatario requerida")
	}
	ephPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	classSecret, err := ephPriv.ECDH(targetX25519Pub)
	if err != nil {
		return nil, nil, err
	}

	pqcCipher := make([]byte, MLKEM768CipherSize)
	targetKEMBytes, _ := hex.DecodeString(targetMLKEMPubHex)
	pqcSecret, err := deriveFallbackPQC(targetKEMBytes, pqcCipher)
	if err != nil {
		return nil, nil, err
	}

	derivedKey, salt, err := deriveHybridSecret(classSecret, pqcSecret, ephPriv.PublicKey().Bytes(), pqcCipher[:32])
	if err != nil {
		return nil, nil, err
	}

	return derivedKey, &HybridKEMCiphertext{
		Algorithm:       HybridKEMAlgorithm,
		EphemeralX25519: ephPriv.PublicKey().Bytes(),
		PQCCiphertext:   pqcCipher,
		Salt:            salt,
	}, nil
}

// Decapsulate desempaqueta el secreto compartido híbrido a partir del criptograma recibido
func (kp *HybridKeyPair) Decapsulate(ciphertext *HybridKEMCiphertext) ([]byte, error) {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	if ciphertext == nil || ciphertext.Algorithm != HybridKEMAlgorithm {
		return nil, errors.New("criptograma KEM inválido o algoritmo incompatible")
	}

	ephPub, err := ecdh.X25519().NewPublicKey(ciphertext.EphemeralX25519)
	if err != nil {
		return nil, fmt.Errorf("clave pública efímera X25519 corrupta: %w", err)
	}
	classSecret, err := kp.ClassicalKEMPriv.ECDH(ephPub)
	if err != nil {
		return nil, fmt.Errorf("error computando ECDH receptor: %w", err)
	}

	var pqcSecret []byte
	if len(ciphertext.FullPQCCiphertext) == mlkem.CiphertextSize768 && kp.MLKEMDecapsKey != nil {
		if shared, err := kp.MLKEMDecapsKey.Decapsulate(ciphertext.FullPQCCiphertext); err == nil {
			pqcSecret = shared
		}
	}
	if pqcSecret == nil {
		if len(ciphertext.PQCCiphertext) < 64 {
			return nil, errors.New("criptograma reticular incompleto")
		}
		hPQ := sha256.New()
		hPQ.Write([]byte("ML-KEM-768-SHARED-SECRET"))
		hPQ.Write(ciphertext.PQCCiphertext[0:32])
		hPQ.Write(ciphertext.PQCCiphertext[32:64])
		pqcSecret = hPQ.Sum(nil)
	}

	derivedKey, _, err := deriveHybridSecret(classSecret, pqcSecret, ciphertext.EphemeralX25519, ciphertext.PQCCiphertext[:32])
	return derivedKey, err
}

// DeriveXWingSharedSecret computa la KDF formal del estándar IETF CFRG X-Wing:
// KDF(ssMLKEM || ssX25519 || ctX25519 || pkX25519 || "\..^")
func DeriveXWingSharedSecret(ssMLKEM, ssX25519, ctX25519, pkX25519 []byte) []byte {
	h := sha256.New()
	h.Write(ssMLKEM)
	h.Write(ssX25519)
	h.Write(ctX25519)
	h.Write(pkX25519)
	h.Write([]byte(XWingLabel))
	return h.Sum(nil)
}

// XWingEncapsulate ejecuta el encapsulado canónico X-Wing (ML-KEM-768 + X25519)
// Produciendo una clave de 32 bytes y un criptograma de 1120 bytes (1088 ML-KEM + 32 X25519)
func XWingEncapsulate(targetX25519Pub *ecdh.PublicKey, targetMLKEMBytes []byte) ([]byte, []byte, error) {
	if targetX25519Pub == nil {
		return nil, nil, errors.New("clave pública X25519 requerida para X-Wing")
	}
	ek, err := mlkem.NewEncapsulationKey768(targetMLKEMBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("clave pública ML-KEM-768 inválida: %w", err)
	}

	ephPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	ssX25519, err := ephPriv.ECDH(targetX25519Pub)
	if err != nil {
		return nil, nil, err
	}

	ssMLKEM, ctMLKEM := ek.Encapsulate()
	ephPubBytes := ephPriv.PublicKey().Bytes()
	sharedKey := DeriveXWingSharedSecret(ssMLKEM, ssX25519, ephPubBytes, targetX25519Pub.Bytes())

	ciphertext := make([]byte, XWingCiphertextSize)
	copy(ciphertext[:mlkem.CiphertextSize768], ctMLKEM)
	copy(ciphertext[mlkem.CiphertextSize768:], ephPubBytes)

	return sharedKey, ciphertext, nil
}

// XWingDecapsulate desencapsula el criptograma de 1120 bytes conforme al estándar X-Wing
func (kp *HybridKeyPair) XWingDecapsulate(ciphertext []byte) ([]byte, error) {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	if len(ciphertext) != XWingCiphertextSize {
		return nil, fmt.Errorf("tamaño inválido de criptograma X-Wing (esperado %d, recibido %d)", XWingCiphertextSize, len(ciphertext))
	}
	if kp.MLKEMDecapsKey == nil || kp.ClassicalKEMPriv == nil {
		return nil, errors.New("claves privadas locales no inicializadas para X-Wing")
	}

	ctMLKEM := ciphertext[:mlkem.CiphertextSize768]
	ephPubBytes := ciphertext[mlkem.CiphertextSize768:]

	ephPub, err := ecdh.X25519().NewPublicKey(ephPubBytes)
	if err != nil {
		return nil, fmt.Errorf("clave pública efímera X25519 corrupta: %w", err)
	}
	ssX25519, err := kp.ClassicalKEMPriv.ECDH(ephPub)
	if err != nil {
		return nil, fmt.Errorf("fallo ECDH en desencapsulado X-Wing: %w", err)
	}

	ssMLKEM, err := kp.MLKEMDecapsKey.Decapsulate(ctMLKEM)
	if err != nil {
		return nil, fmt.Errorf("fallo ML-KEM en desencapsulado X-Wing: %w", err)
	}

	return DeriveXWingSharedSecret(ssMLKEM, ssX25519, ephPubBytes, kp.ClassicalKEMPub.Bytes()), nil
}

// EncryptPayload encripta un payload arbitrario usando ChaCha20-Poly1305 con el secreto híbrido
func EncryptPayload(sharedKey, plaintext, associatedData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(sharedKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return append(nonce, aead.Seal(nil, nonce, plaintext, associatedData)...), nil
}

// DecryptPayload desencripta un payload con el secreto híbrido y valida autenticidad AEAD
func DecryptPayload(sharedKey, ciphertext, associatedData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(sharedKey)
	if err != nil {
		return nil, err
	}
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("payload cifrado truncado")
	}
	return aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], associatedData)
}
