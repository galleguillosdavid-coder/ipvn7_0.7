package core

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"
)

// Constantes canónicas del Núcleo Mínimo I7 (Axiomas de Arquitectura)
const (
	I7StandardMTU      = 1280 // Límite determinista de MTU de alambre (RFC 8200 / IPv6 base)
	I7HeaderSize       = 64   // Cabecera compacta fija de contenedor
	I7MaxPayloadSize   = I7StandardMTU - I7HeaderSize
	I7Version          = 0x07
)

// Errores canónicos del Núcleo Mínimo
var (
	ErrPayloadTooLarge    = errors.New("i7: carga útil excede el tamaño máximo permitido por el contenedor (1280B)")
	ErrIntegrityViolation = errors.New("i7: violación de integridad criptográfica en el contenedor")
	ErrExpiredCapability  = errors.New("i7: capacidad (token ZTNA) expirada")
	ErrInvalidHopCount    = errors.New("i7: conteo de saltos agotado o inválido")
)

// 1. Identity: Identidad soberana criptográfica basada en DID
type I7Identity struct {
	DID       string `json:"did"`
	PublicKey []byte `json:"public_key"` // Ed25519 o PQC Key Bytes
}

// 2. Object: Carga útil tipada universal
type I7Object struct {
	Type     string `json:"type"`      // ej. "application/octet-stream", "text/plain", "i7/control"
	Payload  []byte `json:"payload"`
	Metadata []byte `json:"metadata,omitempty"`
}

// 3. Container: Trama fija de alambre delimitada por la MTU
type I7Container struct {
	Version   uint8  `json:"version"`
	ChannelID uint32 `json:"channel_id"`
	Payload   []byte `json:"payload"`
	Tag       []byte `json:"tag"` // 16 bytes AEAD Tag / Integrity
}

// NewContainer empaqueta un objeto asegurando el invariante de MTU <= 1280 bytes
func NewContainer(channelID uint32, payload, tag []byte) (*I7Container, error) {
	if len(payload)+I7HeaderSize > I7StandardMTU {
		return nil, fmt.Errorf("%w: tamaño total %d B", ErrPayloadTooLarge, len(payload)+I7HeaderSize)
	}
	return &I7Container{
		Version:   I7Version,
		ChannelID: channelID,
		Payload:   payload,
		Tag:       tag,
	}, nil
}

// 4. Session: Canal criptográfico autenticado establecido entre dos identidades
type I7Session struct {
	SessionID string     `json:"session_id"`
	LocalDID  string     `json:"local_did"`
	RemoteDID string     `json:"remote_did"`
	SharedKey []byte     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// IsExpired verifica la validez temporal de la sesión
func (s *I7Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// 5. Channel: Flujo lógico multiplexado dentro de una sesión
type I7Channel struct {
	ID        uint32 `json:"id"`
	Priority  uint8  `json:"priority"` // 0 = Normal, 1 = Alta, 2 = Control
	Active    bool   `json:"active"`
}

// 6. Path: Vector de enrutamiento determinista entre identidades
type I7Path struct {
	SourceDID string `json:"source_did"`
	TargetDID string `json:"target_did"`
	NextHop   string `json:"next_hop"`
	HopCount  uint8  `json:"hop_count"`
	MaxHops   uint8  `json:"max_hops"`
}

// IncrementHop valida y avanza el camino enrutado
func (p *I7Path) IncrementHop() error {
	p.HopCount++
	if p.HopCount > p.MaxHops {
		return ErrInvalidHopCount
	}
	return nil
}

// 7. MTU: Verificador del invariante de longitud
func ValidateMTU(packetSize int) bool {
	return packetSize <= I7StandardMTU
}

// 8. Integrity: Verificación de autenticidad en tiempo constante
func VerifyIntegrity(expectedTag, actualTag []byte) bool {
	if len(expectedTag) == 0 || len(actualTag) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(expectedTag, actualTag) == 1
}

// 9. Routing: Primitiva de decisión de reenvío por distancia XOR
type I7RoutingPrimitive struct {
	LocalDID string
}

// XORMetric calcula la distancia XOR métrica entre dos identificadores
func XORMetric(a, b []byte) []byte {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	metric := make([]byte, minLen)
	for i := 0; i < minLen; i++ {
		metric[i] = a[i] ^ b[i]
	}
	return metric
}

// 10. Capability: Token de autorización soberana ZTNA
type I7Capability struct {
	IssuerDID string    `json:"issuer_did"`
	TargetDID string    `json:"target_did"`
	Resource  string    `json:"resource"`
	Actions   []string  `json:"actions"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Validate verifica que la capacidad no haya expirado y aplique a la acción
func (c *I7Capability) Validate(requiredAction string) error {
	if time.Now().After(c.ExpiresAt) {
		return ErrExpiredCapability
	}
	for _, action := range c.Actions {
		if action == requiredAction || action == "*" {
			return nil
		}
	}
	return fmt.Errorf("i7: acción '%s' no autorizada por capability", requiredAction)
}
