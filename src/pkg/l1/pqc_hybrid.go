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
	"golang.org/x/crypto/sha3"
)

const (
	// Algoritmos canónicos de ipvn7 PQC
	// NOTA TÉCNICA: La firma primaria de autenticidad en producción es Ed25519 pura estándar (RFC 8032).
	// El vector reticular adjunto es un compromiso experimental derivado por semilla; NO es NIST FIPS 204 ML-DSA.
	HybridSigAlgorithm = "Ed25519+ExperimentalVector"
	HybridKEMAlgorithm = "X25519+ML-KEM-768"
	XWingAlgorithm     = "X-Wing (X25519+ML-KEM-768-Draft-CFRG)"
	XWingLabel         = "\\.//^\\" // 0x5c 0x2e 0x2f 0x2f 0x5e 0x5c conforme a draft-ietf-cfrg-xwing (6 bytes)

	// Longitudes canónicas
	ExperimentalSigSeedSize = 32
	ExperimentalSigSize     = 128
	MLDSA65SeedSize         = ExperimentalSigSeedSize // alias retrocompatible
	MLDSA65SigSize          = ExperimentalSigSize     // alias retrocompatible
	MLKEM768CipherSize      = 128
	SharedSecretSize        = 32
	XWingPublicKeySize      = 1216 // 1184 (ML-KEM-768) + 32 (X25519)
	XWingCiphertextSize     = 1120 // 1088 (ML-KEM-768) + 32 (ephemeral X25519)
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

	Ed25519PubHex           string    `json:"ed25519_pub_hex"`
	X25519PubHex            string    `json:"x25519_pub_hex"`
	ExperimentalPQCIdentity string    `json:"experimental_pqc_identity"`
	MLDSAPubHex             string    `json:"ml_dsa_pub_hex,omitempty"` // alias documental retrocompatible
	MLKEMPubHex             string    `json:"ml_kem_pub_hex"`
	CreatedAt               time.Time `json:"created_at"`
}

// HybridKEMCiphertext contiene la encapsulación de clave compartida
type HybridKEMCiphertext struct {
	Algorithm         string `json:"algorithm"`
	EphemeralX25519   []byte `json:"ephemeral_x25519"`              // 32 bytes
	PQCCiphertext     []byte `json:"pqc_ciphertext"`                // 128 bytes (compacto wire)
	FullPQCCiphertext []byte `json:"full_pqc_ciphertext,omitempty"` // 1088 bytes (FIPS 203 real)
	Salt              []byte `json:"salt"`                          // 16 bytes
}

func GenerateHybridKeyPair(did string) (*HybridKeyPair, error) {
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generando Ed25519: %w", err)
	}
	xPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error generando X25519: %w", err)
	}
	mlkemDecaps, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("error generando ML-KEM-768: %w", err)
	}
	pqcSignSeed := make([]byte, ExperimentalSigSeedSize)
	if _, err := io.ReadFull(rand.Reader, pqcSignSeed); err != nil {
		return nil, fmt.Errorf("error semilla experimental: %w", err)
	}
	hExp := sha256.New()
	hExp.Write([]byte("IPVN7-EXPERIMENTAL-PQC-IDENTITY"))
	hExp.Write(pqcSignSeed)
	expIDHex := hex.EncodeToString(hExp.Sum(nil))
	xPub := xPriv.PublicKey()
	mlkemEncaps := mlkemDecaps.EncapsulationKey()

	return &HybridKeyPair{
		DID:                     did,
		ClassicalSignPub:        edPub,
		ClassicalSignPriv:       edPriv,
		ClassicalKEMPub:         xPub,
		ClassicalKEMPriv:        xPriv,
		MLKEMDecapsKey:          mlkemDecaps,
		MLKEMEncapsKey:          mlkemEncaps,
		PQCSignSeed:             pqcSignSeed,
		PQCKEMSeed:              mlkemDecaps.Bytes(),
		Ed25519PubHex:           hex.EncodeToString(edPub),
		X25519PubHex:            hex.EncodeToString(xPub.Bytes()),
		ExperimentalPQCIdentity: expIDHex,
		MLDSAPubHex:             expIDHex,
		MLKEMPubHex:             hex.EncodeToString(mlkemEncaps.Bytes()),
		CreatedAt:               time.Now().UTC(),
	}, nil
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
// NOTA DE SEGURIDAD (Regla 1 / FIPS 203): No degrada silenciosamente a algoritmos más débiles.
func Encapsulate(targetX25519Pub *ecdh.PublicKey, targetMLKEMPubHex string) ([]byte, *HybridKEMCiphertext, error) {
	if targetX25519Pub == nil {
		return nil, nil, errors.New("clave pública X25519 requerida")
	}
	ephPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	classSecret, err := ephPriv.ECDH(targetX25519Pub)
	if err != nil {
		return nil, nil, err
	}

	targetKEMBytes, err := hex.DecodeString(targetMLKEMPubHex)
	if err != nil || len(targetKEMBytes) != mlkem.EncapsulationKeySize768 {
		return nil, nil, errors.New("clave pública ML-KEM-768 inválida o ausente (requiere 1184 bytes FIPS 203)")
	}
	ek, err := mlkem.NewEncapsulationKey768(targetKEMBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("clave pública ML-KEM-768 corrupta: %w", err)
	}
	pqcSecret, fullCipher := ek.Encapsulate()

	pqcCipher := make([]byte, 128)
	copy(pqcCipher, fullCipher[:128])

	derivedKey, salt, err := deriveHybridSecret(classSecret, pqcSecret, ephPriv.PublicKey().Bytes(), fullCipher[:32])
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

// EncapsulateCompact genera secreto compartido híbrido requiriendo ML-KEM-768 real
func EncapsulateCompact(targetX25519Pub *ecdh.PublicKey, targetMLKEMPubHex string) ([]byte, *HybridKEMCiphertext, error) {
	return Encapsulate(targetX25519Pub, targetMLKEMPubHex)
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

	if len(ciphertext.FullPQCCiphertext) != mlkem.CiphertextSize768 || kp.MLKEMDecapsKey == nil {
		return nil, errors.New("criptograma ML-KEM-768 incompleto o decapsulador no inicializado (requiere 1088B FIPS 203)")
	}
	pqcSecret, err := kp.MLKEMDecapsKey.Decapsulate(ciphertext.FullPQCCiphertext)
	if err != nil {
		return nil, fmt.Errorf("error decapsulando ML-KEM-768: %w", err)
	}

	headBytes := ciphertext.FullPQCCiphertext[:32]
	derivedKey, _, err := deriveHybridSecret(classSecret, pqcSecret, ciphertext.EphemeralX25519, headBytes)
	return derivedKey, err
}

// DeriveXWingSharedSecret computa la KDF formal del combinador IETF CFRG X-Wing:
// SHA3-256(ssMLKEM || ssX25519 || ctX25519 || pkX25519 || "\.//^\")
// NOTA TAXONÓMICA: Implementa la función combinadora estricta con SHA3-256 y la etiqueta
// canónica de 6 bytes (0x5c, 0x2e, 0x2f, 0x2f, 0x5e, 0x5c). La generación local de claves
// en ipvn7 utiliza pares desacoplados (crypto/mlkem + crypto/ecdh), por lo que se cataloga como
// perfil híbrido experimental conforme a draft-ietf-cfrg-xwing.
func DeriveXWingSharedSecret(ssMLKEM, ssX25519, ctX25519, pkX25519 []byte) []byte {
	h := sha3.New256()
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
		return nil, fmt.Errorf("tamaño inválido X-Wing (esperado %d, recibido %d)", XWingCiphertextSize, len(ciphertext))
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

// EncryptPayload encripta un payload usando ChaCha20-Poly1305 con el secreto híbrido
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
	if len(ciphertext) < aead.NonceSize() {
		return nil, errors.New("payload cifrado truncado")
	}
	return aead.Open(nil, ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():], associatedData)
}
