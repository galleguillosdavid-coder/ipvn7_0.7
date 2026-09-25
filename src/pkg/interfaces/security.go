// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import "time"

// SecurityDecision representa el resultado formal de la evaluación de acceso de un paquete.
type SecurityDecision uint8

const (
	// DecisionDeny indica que el paquete debe ser descartado silenciosamente (Drop).
	DecisionDeny SecurityDecision = iota

	// DecisionAllow indica que el paquete cumple con todas las políticas de micro-segmentación.
	DecisionAllow

	// DecisionRateLimit indica que el remitente superó la cuota de ancho de banda y debe amortiguarse.
	DecisionRateLimit

	// DecisionQuarantine indica que el remitente ha sido confinado por la capa de inmunología celular.
	DecisionQuarantine

	// DecisionDropACL indica descarte por política explícita de lista de control de acceso.
	DecisionDropACL
)

func (d SecurityDecision) String() string {
	switch d {
	case DecisionAllow:
		return "ACCEPT"
	case DecisionDeny:
		return "DROP_DEFAULT_DENY"
	case DecisionDropACL:
		return "DROP_ACL"
	case DecisionRateLimit:
		return "DROP_RATE_LIMIT"
	case DecisionQuarantine:
		return "QUARANTINE"
	default:
		return "UNKNOWN"
	}
}

// ZTNAFirewall abstrae el cortafuegos de confianza cero (Zero Trust Network Access).
// Principio de Diseño: Default-Deny por defecto. Ningún tráfico pasa sin regla explícita.
type ZTNAFirewall interface {
	// EvaluateInbound dictamina si un paquete proveniente de 'srcDID' hacia el puerto 'destPort' es admitido.
	EvaluateInbound(srcDID string, destPort uint16) (SecurityDecision, string)

	// AllowPeer autoriza explícitamente el tráfico proveniente de un par de confianza para puertos específicos.
	AllowPeer(peerDID string, ports []uint16)

	// RevokePeer revoca cualquier permiso concedido a un par, restableciendo el bloqueo estricto.
	RevokePeer(peerDID string)

	// IsAllowed verifica si un par cuenta con autorización activa en las tablas ZTNA.
	IsAllowed(peerDID string) bool
}

// ThreatSeverity categoriza la gravedad de una anomalía detectada por los centinelas.
type ThreatSeverity string

const (
	SeverityLow      ThreatSeverity = "LOW"
	SeverityMedium   ThreatSeverity = "MEDIUM"
	SeverityHigh     ThreatSeverity = "HIGH"
	SeverityCritical ThreatSeverity = "CRITICAL"
)

// ThreatIncident describe un evento de agresión o violación de protocolo registrado por centinelas.
type ThreatIncident struct {
	ID          string         `json:"id"`
	SourceDID   string         `json:"source_did"`
	Severity    ThreatSeverity `json:"severity"`
	Description string         `json:"description"`
	Timestamp   time.Time      `json:"timestamp"`
}

// SentinelImmunology gestiona la detección activa de ataques (Sybil, Replay, Amplification) y aislamiento celular.
type SentinelImmunology interface {
	// InspectPacket examina el patrón de tráfico de un paquete para detectar firmas de ataque o flood.
	InspectPacket(srcDID string, packetSize int, packetType uint8) *ThreatIncident

	// QuarantineNode aisla inmediatamente un DID malicioso impidiendo toda comunicación entrante y saliente.
	QuarantineNode(did string, reason string, duration time.Duration)

	// IsQuarantined retorna true si el nodo se encuentra actualmente en estado de aislamiento celular.
	IsQuarantined(did string) bool
}
