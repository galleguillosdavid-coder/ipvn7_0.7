package l3

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// PeeringClaimSubject modela las afirmaciones de capacidad y SLA de un nodo par en la malla
type PeeringClaimSubject struct {
	ID               string  `json:"id"`
	NodeCapability   string  `json:"node_capability"` // "RELAY_NODE", "EDGE_ROUTER", "STORAGE_CACHE"
	MaxBandwidthMbps int     `json:"max_bandwidth_mbps"`
	MinUptimePct     float64 `json:"min_uptime_pct"`
	TrustTier        string  `json:"trust_tier"`
}

// VCProof representa la firma criptográfica W3C del emisor soberano
type VCProof struct {
	Type               string `json:"type"`                // "Ed25519Signature2020"
	Created            string `json:"created"`
	VerificationMethod string `json:"verificationMethod"` // DID del emisor + clave
	ProofPurpose       string `json:"proofPurpose"`       // "assertionMethod"
	ProofValue         string `json:"proofValue"`         // Hex de la firma Ed25519
}

// VerifiableCredential estructura de credencial verificable W3C para Peering Autónomo
type VerifiableCredential struct {
	Context           []string            `json:"@context"`
	ID                string              `json:"id"`
	Type              []string            `json:"type"`
	Issuer            string              `json:"issuer"`
	IssuanceDate      string              `json:"issuanceDate"`
	ExpirationDate    string              `json:"expirationDate,omitempty"`
	CredentialSubject PeeringClaimSubject `json:"credentialSubject"`
	Proof             VCProof             `json:"proof"`
}

// CanonicalClaimPayload genera el payload binario canónico para firma/verificación
func (vc *VerifiableCredential) CanonicalClaimPayload() []byte {
	return []byte(fmt.Sprintf("%s|%s|%s|%s|%s|%d|%.2f|%s",
		vc.ID, vc.Issuer, vc.IssuanceDate,
		vc.CredentialSubject.ID, vc.CredentialSubject.NodeCapability,
		vc.CredentialSubject.MaxBandwidthMbps, vc.CredentialSubject.MinUptimePct,
		vc.CredentialSubject.TrustTier))
}

// VCRegistryValidator valida y almacena credenciales verificables de la red
type VCRegistryValidator struct {
	mu           sync.RWMutex
	trustedRoots map[string]ed25519.PublicKey // Issuer DID -> PubKey
}

// NewVCRegistryValidator inicializa el validador W3C
func NewVCRegistryValidator() *VCRegistryValidator {
	return &VCRegistryValidator{
		trustedRoots: make(map[string]ed25519.PublicKey),
	}
}

// AddTrustedIssuer añade una clave pública de emisor autorizada
func (v *VCRegistryValidator) AddTrustedIssuer(issuerDID string, pubKey ed25519.PublicKey) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.trustedRoots[issuerDID] = pubKey
}

// SignCredential calcula y adosa la firma Ed25519 a una credencial
func SignCredential(vc *VerifiableCredential, privKey ed25519.PrivateKey) error {
	if len(privKey) != ed25519.PrivateKeySize {
		return errors.New("invalid private key size for Ed25519")
	}
	payload := vc.CanonicalClaimPayload()
	sig := ed25519.Sign(privKey, payload)
	vc.Proof = VCProof{
		Type:               "Ed25519Signature2020",
		Created:            time.Now().UTC().Format(time.RFC3339),
		VerificationMethod: fmt.Sprintf("%s#keys-1", vc.Issuer),
		ProofPurpose:       "assertionMethod",
		ProofValue:         hex.EncodeToString(sig),
	}
	return nil
}

// ValidateCredential valida integridad criptográfica, temporal y emisor de una VC
func (v *VCRegistryValidator) ValidateCredential(vc *VerifiableCredential) (bool, error) {
	if vc == nil {
		return false, errors.New("credential is nil")
	}

	// 1. Verificación temporal
	if vc.ExpirationDate != "" {
		exp, err := time.Parse(time.RFC3339, vc.ExpirationDate)
		if err == nil && time.Now().After(exp) {
			return false, errors.New("credential has expired")
		}
	}

	// 2. Comprobar emisor confiable
	v.mu.RLock()
	pubKey, ok := v.trustedRoots[vc.Issuer]
	v.mu.RUnlock()

	if !ok {
		return false, fmt.Errorf("issuer %s is not recognized in trusted roots", vc.Issuer)
	}

	// 3. Verificar firma criptográfica
	sigBytes, err := hex.DecodeString(vc.Proof.ProofValue)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false, errors.New("invalid signature proof format")
	}

	payload := vc.CanonicalClaimPayload()
	if !ed25519.Verify(pubKey, payload, sigBytes) {
		return false, errors.New("cryptographic signature is invalid")
	}

	return true, nil
}
