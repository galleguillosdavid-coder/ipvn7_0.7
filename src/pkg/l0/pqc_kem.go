package l0

import (
	"crypto/hmac"
	"crypto/mlkem"
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	// Dimensiones de ML-KEM-768 (NIST FIPS 203 estándar oficial)
	MLKEM768PublicKeyBytes  = 1184
	MLKEM768CiphertextBytes = 1088
	MLKEM768SharedSecretLen = 32
	HybridDomainSeparator   = "ipvn7-hybrid-pqc-fips203-v1"
)

var (
	ErrInvalidPQCKeyLength  = errors.New("pqc: longitud de clave pública ML-KEM-768 inválida")
	ErrInvalidPQCCiphertext = errors.New("pqc: longitud de ciphertext ML-KEM-768 inválida")
	ErrCorruptPQCCiphertext = errors.New("pqc: ciphertext corrupto o alterado")
)

// MLKEM768KeyPair almacena el par de claves ML-KEM-768 (FIPS 203)
type MLKEM768KeyPair struct {
	PublicKey  [MLKEM768PublicKeyBytes]byte
	PrivateKey [64]byte // Semilla canónica de decapsulación de 64 bytes
	decKey     *mlkem.DecapsulationKey768
}

// GenerateMLKEM768KeyPair genera un par de claves ML-KEM-768 real con crypto/mlkem
func GenerateMLKEM768KeyPair() (*MLKEM768KeyPair, error) {
	dk, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, fmt.Errorf("error generando clave ML-KEM-768 FIPS 203: %w", err)
	}

	kp := &MLKEM768KeyPair{decKey: dk}
	copy(kp.PublicKey[:], dk.EncapsulationKey().Bytes())
	copy(kp.PrivateKey[:], dk.Bytes())
	return kp, nil
}

// MLKEM768Adapter implementa la interfaz PQC usando el estándar NIST FIPS 203 real
type MLKEM768Adapter struct {
	keyPair *MLKEM768KeyPair
}

// NewMLKEM768Adapter inicializa el adaptador con el par de claves del nodo
func NewMLKEM768Adapter(kp *MLKEM768KeyPair) *MLKEM768Adapter {
	if kp != nil && kp.decKey == nil {
		// Reconstruir decapsulation key desde la semilla si fue cargada de almacenamiento
		if dk, err := mlkem.NewDecapsulationKey768(kp.PrivateKey[:]); err == nil {
			kp.decKey = dk
		}
	}
	return &MLKEM768Adapter{keyPair: kp}
}

// Encapsulate genera el ciphertext estándar FIPS 203 y el secreto compartido
func (a *MLKEM768Adapter) Encapsulate(recipientPK []byte) ([]byte, []byte, error) {
	if len(recipientPK) != MLKEM768PublicKeyBytes {
		return nil, nil, ErrInvalidPQCKeyLength
	}

	ek, err := mlkem.NewEncapsulationKey768(recipientPK)
	if err != nil {
		return nil, nil, fmt.Errorf("error parseando clave de encapsulamiento ML-KEM-768: %w", err)
	}

	sharedSecret, ciphertext := ek.Encapsulate()
	return ciphertext, sharedSecret, nil
}

// Decapsulate decapsula el ciphertext FIPS 203 recuperando el secreto compartido
func (a *MLKEM768Adapter) Decapsulate(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) != MLKEM768CiphertextBytes {
		return nil, ErrInvalidPQCCiphertext
	}

	if a.keyPair == nil || a.keyPair.decKey == nil {
		return nil, errors.New("pqc: par de claves no configurado para decapsulación")
	}

	sharedSecret, err := a.keyPair.decKey.Decapsulate(ciphertext)
	if err != nil {
		return nil, ErrCorruptPQCCiphertext
	}

	return sharedSecret, nil
}

// DeriveHybridSecret combina un secreto clásico (X25519) con el secreto PQC (ML-KEM-768 FIPS 203)
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
