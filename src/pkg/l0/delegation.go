package l0

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ExecutionTier define el nivel de control graduado segun IETF draft-das-agentic-adaptive-authorization-00
type ExecutionTier string

const (
	TierOrdinary   ExecutionTier = "ordinary"   // Lectura, telemetria, estado de red
	TierEscalated  ExecutionTier = "escalated"  // Control de VPN, enrutamiento, puentes LAN
	TierQuarantine ExecutionTier = "quarantine" // No verificado / aislado
	TierDeny       ExecutionTier = "deny"       // Acceso bloqueado
)

// DelegationLink representa un eslabon criptografico en la Agentic Principal Chain (APC)
type DelegationLink struct {
	IssuerDID  string        `json:"issuer_did"`
	SubjectDID string        `json:"subject_did"`
	Scope      string        `json:"scope"`
	Tier       ExecutionTier `json:"tier"`
	ExpiresAt  int64         `json:"expires_at"`
	Signature  string        `json:"signature"`
}

// CanonicalBytes produce la representacion determinista para la firma Ed25519
func (d *DelegationLink) CanonicalBytes() []byte {
	raw := fmt.Sprintf("%s|%s|%s|%s|%d", d.IssuerDID, d.SubjectDID, d.Scope, d.Tier, d.ExpiresAt)
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}

// Sign firma el eslabon con la identidad soberana del emisor
func (d *DelegationLink) Sign(issuer *Identity) error {
	if issuer.DID() != d.IssuerDID {
		return errors.New("identidad emisora no coincide con IssuerDID")
	}
	sig := issuer.Sign(d.CanonicalBytes())
	d.Signature = hex.EncodeToString(sig)
	return nil
}

// Verify comprueba la firma digital Ed25519 del emisor
func (d *DelegationLink) Verify() bool {
	pub, err := PublicKeyFromDID(d.IssuerDID)
	if err != nil {
		return false
	}
	sigBytes, err := hex.DecodeString(d.Signature)
	if err != nil {
		return false
	}
	return VerifySignature(pub, d.CanonicalBytes(), sigBytes)
}

// AgenticPrincipalChain agrupa una sucesion verificable de delegaciones
type AgenticPrincipalChain struct {
	RootDID string           `json:"root_did"`
	Links   []DelegationLink `json:"links"`
}

// CreateDelegationLink construye y firma un nuevo eslabon de delegacion
func CreateDelegationLink(issuer *Identity, subjectDID, scope string, tier ExecutionTier, ttl time.Duration) (*DelegationLink, error) {
	link := &DelegationLink{
		IssuerDID:  issuer.DID(),
		SubjectDID: subjectDID,
		Scope:      scope,
		Tier:       tier,
		ExpiresAt:  time.Now().Add(ttl).Unix(),
	}
	if err := link.Sign(issuer); err != nil {
		return nil, err
	}
	return link, nil
}

// VerifyChain valida la cadena de custodia completa y atenuacion de privilegios
func (c *AgenticPrincipalChain) VerifyChain() (ExecutionTier, error) {
	if len(c.Links) == 0 {
		return TierQuarantine, errors.New("cadena de delegacion vacia")
	}
	now := time.Now().Unix()
	currentTier := TierEscalated

	for i, link := range c.Links {
		if i == 0 {
			if link.IssuerDID != c.RootDID {
				return TierDeny, fmt.Errorf("primer eslabon no coincide con RootDID %s", c.RootDID)
			}
		} else {
			if link.IssuerDID != c.Links[i-1].SubjectDID {
				return TierDeny, fmt.Errorf("ruptura de custodia en eslabon %d: %s != %s", i, link.IssuerDID, c.Links[i-1].SubjectDID)
			}
		}
		if !link.Verify() {
			return TierDeny, fmt.Errorf("firma invalida en eslabon %d (%s)", i, link.IssuerDID)
		}
		if now > link.ExpiresAt {
			return TierDeny, fmt.Errorf("token expirado en eslabon %d", i)
		}
		// Atenuacion criptografica: un eslabon no puede elevar privilegios
		if currentTier == TierOrdinary && link.Tier == TierEscalated {
			return TierDeny, fmt.Errorf("violacion de atenuacion en eslabon %d: intento de escalamiento", i)
		}
		if link.Tier == TierOrdinary {
			currentTier = TierOrdinary
		}
	}
	return currentTier, nil
}

// SpendingMandate representa un mandato de gasto agéntico pre-autorizado (AP2 / x402)
type SpendingMandate struct {
	PayerDID     string `json:"payer_did"`
	RecipientDID string `json:"recipient_did"`
	MaxUnits     uint64 `json:"max_units"`     // Unidades soberanas (ms GPU, bytes ancho de banda)
	Resource     string `json:"resource"`      // "bandwidth:prio", "inference:gpu_sec", "print:job"
	ExpiresAt    int64  `json:"expires_at"`    // Unix timestamp
	Signature    string `json:"signature"`     // Ed25519 hex
}

// CanonicalBytes produce el hash determinista del mandato AP2
func (m *SpendingMandate) CanonicalBytes() []byte {
	raw := fmt.Sprintf("%s|%s|%d|%s|%d", m.PayerDID, m.RecipientDID, m.MaxUnits, m.Resource, m.ExpiresAt)
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}

// Sign firma el mandato con la clave privada del pagador
func (m *SpendingMandate) Sign(payer *Identity) error {
	if payer.DID() != m.PayerDID {
		return errors.New("identidad pagadora no coincide con PayerDID")
	}
	sig := payer.Sign(m.CanonicalBytes())
	m.Signature = hex.EncodeToString(sig)
	return nil
}

// Verify comprueba la validez matemática y de vigencia del mandato AP2
func (m *SpendingMandate) Verify() bool {
	if time.Now().Unix() > m.ExpiresAt {
		return false
	}
	pub, err := PublicKeyFromDID(m.PayerDID)
	if err != nil {
		return false
	}
	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return false
	}
	return VerifySignature(pub, m.CanonicalBytes(), sigBytes)
}

// CreateSpendingMandate emite un nuevo mandato de gasto agéntico
func CreateSpendingMandate(payer *Identity, recipientDID, resource string, maxUnits uint64, ttl time.Duration) (*SpendingMandate, error) {
	mandate := &SpendingMandate{
		PayerDID:     payer.DID(),
		RecipientDID: recipientDID,
		MaxUnits:     maxUnits,
		Resource:     resource,
		ExpiresAt:    time.Now().Add(ttl).Unix(),
	}
	if err := mandate.Sign(payer); err != nil {
		return nil, err
	}
	return mandate, nil
}

