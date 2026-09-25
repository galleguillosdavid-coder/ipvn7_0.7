package l3

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestVerifiableCredentials(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Key generation error: %v", err)
	}

	issuerDID := "did:ipvn7:authority:core-mesh"
	validator := NewVCRegistryValidator()
	validator.AddTrustedIssuer(issuerDID, pub)

	vc := &VerifiableCredential{
		Context:        []string{"https://www.w3.org/2018/credentials/v1"},
		ID:             "urn:uuid:peering-claim-12345",
		Type:           []string{"VerifiableCredential", "IPVN7PeeringAgreement"},
		Issuer:         issuerDID,
		IssuanceDate:   time.Now().UTC().Format(time.RFC3339),
		ExpirationDate: time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339),
		CredentialSubject: PeeringClaimSubject{
			ID:               "did:ipvn7:node:super-relay-a",
			NodeCapability:   "RELAY_NODE",
			MaxBandwidthMbps: 1000,
			MinUptimePct:     99.99,
			TrustTier:        "PLATINUM",
		},
	}

	// Sign credential
	if err := SignCredential(vc, priv); err != nil {
		t.Fatalf("Failed to sign credential: %v", err)
	}

	// Validate valid credential
	valid, err := validator.ValidateCredential(vc)
	if err != nil || !valid {
		t.Fatalf("Credential validation failed: %v", err)
	}

	// Tamper with subject
	vc.CredentialSubject.MaxBandwidthMbps = 10000
	validTampered, errTampered := validator.ValidateCredential(vc)
	if validTampered || errTampered == nil {
		t.Fatalf("Expected validation error on tampered credential, but passed")
	}

	// Test expired credential
	vc.CredentialSubject.MaxBandwidthMbps = 1000
	vc.ExpirationDate = time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	SignCredential(vc, priv)
	validExpired, errExpired := validator.ValidateCredential(vc)
	if validExpired || errExpired == nil {
		t.Fatalf("Expected expiration error, but passed")
	}
}
