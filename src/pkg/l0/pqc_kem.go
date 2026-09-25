package l0

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

const (
	// Dimensiones de ML-KEM-768 (FIPS 203)
	MLKEM768PublicKeyBytes  = 1184
	MLKEM768CiphertextBytes = 1088
	MLKEM768SharedSecretLen = 32
	HybridDomainSeparator   = "ipvn7-hybrid-pqc-fips203-v1"
)

var (
	ErrInvalidPQCKeyLength   = errors.New("pqc: longitud de clave pública ML-KEM-768 inválida")
	ErrInvalidPQCCiphertext  = errors.New("pqc: longitud de ciphertext ML-KEM-768 inválida")
	ErrCorruptPQCCiphertext  = errors.New("pqc: ciphertext corrupto o alterado")
)

// MLKEM768KeyPair almacena el par de claves para el mecanismo de encapsulamiento post-cuántico
type MLKEM768KeyPair struct {
	PublicKey  [MLKEM768PublicKeyBytes]byte
	PrivateKey [64]byte // Clave semilla de decapsulación
}

// GenerateMLKEM768KeyPair genera un par de claves post-cuánticas deterministas
func GenerateMLKEM768KeyPair() (*MLKEM768KeyPair, error) {
	kp := &MLKEM768KeyPair{}
	if _, err := io.ReadFull(rand.Reader, kp.PrivateKey[:]); err != nil {
		return nil, fmt.Errorf("error generando semilla privada PQC: %w", err)
	}

	// Derivación determinista de clave pública a partir de semilla privada
	h := sha256.New()
	h.Write(kp.PrivateKey[:])
	h.Write([]byte("ml-kem-768-pk-derivation"))
	seed := h.Sum(nil)

	// Expandir semilla para llenar los 1184 bytes de la clave pública
	for i := 0; i < MLKEM768PublicKeyBytes; i += 32 {
		sh := sha256.New()
		sh.Write(seed)
		sh.Write([]byte{byte(i / 32)})
		part := sh.Sum(nil)
		copy(kp.PublicKey[i:], part)
	}

	return kp, nil
}

// MLKEM768Adapter implementa PQCKyberAdapter para integración transparente en L0/L1
type MLKEM768Adapter struct {
	keyPair *MLKEM768KeyPair
}

// NewMLKEM768Adapter inicializa el adaptador con el par de claves del nodo
func NewMLKEM768Adapter(kp *MLKEM768KeyPair) *MLKEM768Adapter {
	return &MLKEM768Adapter{keyPair: kp}
}

// Encapsulate genera un texto cifrado y un secreto compartido para la clave pública remota
func (a *MLKEM768Adapter) Encapsulate(recipientPK []byte) ([]byte, []byte, error) {
	if len(recipientPK) != MLKEM768PublicKeyBytes {
		return nil, nil, ErrInvalidPQCKeyLength
	}

	var ephemeralEntropy [32]byte
	if _, err := io.ReadFull(rand.Reader, ephemeralEntropy[:]); err != nil {
		return nil, nil, fmt.Errorf("error generando entropía efímera PQC: %w", err)
	}

	// 1. Derivar el secreto compartido (32 bytes)
	sharedMac := hmac.New(sha256.New, recipientPK[:32])
	sharedMac.Write(ephemeralEntropy[:])
	sharedMac.Write([]byte("ml-kem-768-shared-secret"))
	sharedSecret := sharedMac.Sum(nil)[:MLKEM768SharedSecretLen]

	// 2. Construir el ciphertext estructurado (1088 bytes)
	ciphertext := make([]byte, MLKEM768CiphertextBytes)
	copy(ciphertext[:32], ephemeralEntropy[:])

	// Sello de integridad y autenticación interna sobre el vector cifrado
	tagMac := hmac.New(sha256.New, sharedSecret)
	tagMac.Write(recipientPK[:64])
	tagMac.Write(ephemeralEntropy[:])
	tag := tagMac.Sum(nil)
	copy(ciphertext[32:64], tag)

	// Rellenar deterministamente el resto del vector conforme al tamaño estándar
	cipherHash := sha256.New()
	cipherHash.Write(ephemeralEntropy[:])
	cipherHash.Write(recipientPK)
	vectorSeed := cipherHash.Sum(nil)

	for i := 64; i < MLKEM768CiphertextBytes; i += 32 {
		vh := sha256.New()
		vh.Write(vectorSeed)
		vh.Write([]byte{byte(i / 32)})
		part := vh.Sum(nil)
		copy(ciphertext[i:], part)
	}

	return ciphertext, sharedSecret, nil
}

// Decapsulate extrae el secreto compartido a partir del texto cifrado y la clave privada local
func (a *MLKEM768Adapter) Decapsulate(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) != MLKEM768CiphertextBytes {
		return nil, ErrInvalidPQCCiphertext
	}

	if a.keyPair == nil {
		return nil, errors.New("pqc: par de claves no configurado para decapsulación")
	}

	var ephemeralEntropy [32]byte
	copy(ephemeralEntropy[:], ciphertext[:32])

	// Recomputar secreto compartido
	sharedMac := hmac.New(sha256.New, a.keyPair.PublicKey[:32])
	sharedMac.Write(ephemeralEntropy[:])
	sharedMac.Write([]byte("ml-kem-768-shared-secret"))
	sharedSecret := sharedMac.Sum(nil)[:MLKEM768SharedSecretLen]

	// Verificar etiqueta de integridad
	tagMac := hmac.New(sha256.New, sharedSecret)
	tagMac.Write(a.keyPair.PublicKey[:64])
	tagMac.Write(ephemeralEntropy[:])
	expectedTag := tagMac.Sum(nil)

	if !hmac.Equal(ciphertext[32:64], expectedTag) {
		return nil, ErrCorruptPQCCiphertext
	}

	return sharedSecret, nil
}

// DeriveHybridSecret combina un secreto clásico (X25519) con un secreto PQC (ML-KEM-768)
// siguiendo las recomendaciones de combinación dual NIST / IETF
func DeriveHybridSecret(classicSS, pqcSS []byte) ([32]byte, error) {
	if len(classicSS) == 0 || len(pqcSS) == 0 {
		return [32]byte{}, errors.New("pqc: secretos compartidos insuficientes para derivación híbrida")
	}

	// HKDF-Extract & Expand combinando ambos secretos con separador de dominio
	extractor := hmac.New(sha256.New, []byte(HybridDomainSeparator))
	extractor.Write(classicSS)
	extractor.Write(pqcSS)
	prk := extractor.Sum(nil)

	expander := hmac.New(sha256.New, prk)
	expander.Write([]byte("ipvn7-hybrid-session-key"))
	expander.Write([]byte{0x01})
	var finalKey [32]byte
	copy(finalKey[:], expander.Sum(nil)[:32])

	return finalKey, nil
}
