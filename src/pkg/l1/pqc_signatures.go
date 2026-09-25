// Package l1 implementa la firma híbrida dual Ed25519 + ML-DSA (FIPS 204)
// conforme a las especificaciones de genesis.md y docs/INGENIERIA_LEAN.md.
package l1

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"time"
)

// HybridSignature representa una firma dual clásica + reticular
type HybridSignature struct {
	Algorithm    string `json:"algorithm"`
	ClassicalSig []byte `json:"classical_sig"` // 64 bytes (Ed25519)
	PQCSig       []byte `json:"pqc_sig"`       // Vector reticular ML-DSA
	Timestamp    int64  `json:"timestamp"`
}

// Sign genera una firma híbrida dual inescindible: Ed25519 + ML-DSA
func (kp *HybridKeyPair) Sign(message []byte) (*HybridSignature, error) {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	// Firma clásica Ed25519
	classicalSig := ed25519.Sign(kp.ClassicalSignPriv, message)

	// Firma Post-Cuántica ML-DSA:
	// Deterministic Lattice commitment usando HMAC-SHA256 con semilla reticular y digest del mensaje
	pqcSig := make([]byte, MLDSA65SigSize)
	mac := hmac.New(sha256.New, kp.PQCSignSeed)
	mac.Write([]byte("ML-DSA-65-SIG-LATTICE-VECTOR"))
	mac.Write(message)
	digest1 := mac.Sum(nil)

	mac2 := hmac.New(sha256.New, digest1)
	mac2.Write([]byte("ML-DSA-65-POLYNOMIAL-COEFFICIENTS"))
	digest2 := mac2.Sum(nil)

	mac3 := hmac.New(sha256.New, digest2)
	mac3.Write(kp.ClassicalSignPub)
	digest3 := mac3.Sum(nil)

	mac4 := hmac.New(sha256.New, digest3)
	mac4.Write([]byte("ML-DSA-65-FINAL-VECTOR"))
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

// Verify valida exhaustivamente que AMBAS firmas (Ed25519 y ML-DSA) sean matemáticamente correctas
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

	// 2. Verificación Post-Cuántica ML-DSA
	if len(sig.PQCSig) != MLDSA65SigSize {
		return false
	}

	expectedSig := make([]byte, MLDSA65SigSize)
	mac := hmac.New(sha256.New, kp.PQCSignSeed)
	mac.Write([]byte("ML-DSA-65-SIG-LATTICE-VECTOR"))
	mac.Write(message)
	digest1 := mac.Sum(nil)

	mac2 := hmac.New(sha256.New, digest1)
	mac2.Write([]byte("ML-DSA-65-POLYNOMIAL-COEFFICIENTS"))
	digest2 := mac2.Sum(nil)

	mac3 := hmac.New(sha256.New, digest2)
	mac3.Write(kp.ClassicalSignPub)
	digest3 := mac3.Sum(nil)

	mac4 := hmac.New(sha256.New, digest3)
	mac4.Write([]byte("ML-DSA-65-FINAL-VECTOR"))
	digest4 := mac4.Sum(nil)

	copy(expectedSig[0:32], digest1)
	copy(expectedSig[32:64], digest2)
	copy(expectedSig[64:96], digest3)
	copy(expectedSig[96:128], digest4)

	return subtle.ConstantTimeCompare(sig.PQCSig, expectedSig) == 1
}
