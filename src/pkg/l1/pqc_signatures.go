// Package l1 implementa la firma híbrida dual Ed25519 (estándar RFC 8032) + vector experimental determinista.
// ESTADO TAXONÓMICO: DEMOSTRADO FÍSICAMENTE (Ed25519 estándar de producción) / EXPERIMENTAL (Vector reticular HMAC).
// NOTA TÉCNICA OBLIGATORIA (Auditoría Externa): El componente reticular adjunto es un compromiso
// determinista derivado por semilla; NO constituye una implementación formal de NIST FIPS 204.
// ML-DSA NO ESTÁ IMPLEMENTADO en IPVN7 v0.7.0. La garantía de no-repudio y autenticidad
// en producción descansa estrictamente en Ed25519.
package l1

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"time"
)

// HybridSignature representa una firma dual clásica + reticular experimental
type HybridSignature struct {
	Algorithm    string `json:"algorithm"`
	ClassicalSig []byte `json:"classical_sig"` // 64 bytes (Ed25519)
	PQCSig       []byte `json:"pqc_sig"`       // Vector reticular experimental
	Timestamp    int64  `json:"timestamp"`
}

// Sign genera una firma híbrida dual: Ed25519 (producción) + Vector Experimental
func (kp *HybridKeyPair) Sign(message []byte) (*HybridSignature, error) {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	// Firma clásica Ed25519
	classicalSig := ed25519.Sign(kp.ClassicalSignPriv, message)

	// Vector reticular experimental:
	// Deterministic Lattice commitment usando HMAC-SHA256 con semilla y digest del mensaje
	pqcSig := make([]byte, ExperimentalSigSize)
	mac := hmac.New(sha256.New, kp.PQCSignSeed)
	mac.Write([]byte("EXPERIMENTAL-SIG-LATTICE-VECTOR"))
	mac.Write(message)
	digest1 := mac.Sum(nil)

	mac2 := hmac.New(sha256.New, digest1)
	mac2.Write([]byte("EXPERIMENTAL-POLYNOMIAL-COEFFICIENTS"))
	digest2 := mac2.Sum(nil)

	mac3 := hmac.New(sha256.New, digest2)
	mac3.Write(kp.ClassicalSignPub)
	digest3 := mac3.Sum(nil)

	mac4 := hmac.New(sha256.New, digest3)
	mac4.Write([]byte("EXPERIMENTAL-FINAL-VECTOR"))
	digest4 := mac4.Sum(nil)

	copy(pqcSig[0:32], digest1)
	copy(pqcSig[32:64], digest2)
	copy(pqcSig[64:96], digest3)
	copy(pqcSig[96:128], digest4)

	return &HybridSignature{
		Algorithm:    HybridSigAlgorithm,
		ClassicalSig: classicalSig,
		PQCSig:       pqcSig,
		Timestamp:    time.Now().UTC().Unix(),
	}, nil
}

// Verify valida exhaustivamente que la firma Ed25519 y el vector reticular experimental sean correctos
func (kp *HybridKeyPair) Verify(message []byte, sig *HybridSignature) bool {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	if sig == nil || sig.Algorithm != HybridSigAlgorithm {
		return false
	}

	// 1. Verificación Clásica Ed25519
	if len(sig.ClassicalSig) != ed25519.SignatureSize {
		return false
	}
	if !ed25519.Verify(kp.ClassicalSignPub, message, sig.ClassicalSig) {
		return false
	}

	// 2. Verificación de vector reticular experimental
	if len(sig.PQCSig) != ExperimentalSigSize {
		return false
	}

	expectedSig := make([]byte, ExperimentalSigSize)
	mac := hmac.New(sha256.New, kp.PQCSignSeed)
	mac.Write([]byte("EXPERIMENTAL-SIG-LATTICE-VECTOR"))
	mac.Write(message)
	digest1 := mac.Sum(nil)

	mac2 := hmac.New(sha256.New, digest1)
	mac2.Write([]byte("EXPERIMENTAL-POLYNOMIAL-COEFFICIENTS"))
	digest2 := mac2.Sum(nil)

	mac3 := hmac.New(sha256.New, digest2)
	mac3.Write(kp.ClassicalSignPub)
	digest3 := mac3.Sum(nil)

	mac4 := hmac.New(sha256.New, digest3)
	mac4.Write([]byte("EXPERIMENTAL-FINAL-VECTOR"))
	digest4 := mac4.Sum(nil)

	copy(expectedSig[0:32], digest1)
	copy(expectedSig[32:64], digest2)
	copy(expectedSig[64:96], digest3)
	copy(expectedSig[96:128], digest4)

	return subtle.ConstantTimeCompare(sig.PQCSig, expectedSig) == 1
}
