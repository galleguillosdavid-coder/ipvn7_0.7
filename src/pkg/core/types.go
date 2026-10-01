package core

import (
	"time"
)

// ComponentState representa el estado del ciclo de vida de un componente externo
type ComponentState string

const (
	StateRegistered ComponentState = "REGISTERED"
	StateActive     ComponentState = "ACTIVE"
	StateIdle       ComponentState = "IDLE"
	StateSuspended  ComponentState = "SUSPENDED"
	StateFailed     ComponentState = "FAILED"
)

// ComponentRegistration define la metadata de un componente que se acopla al núcleo
type ComponentRegistration struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Description  string         `json:"description,omitempty"`
	Capabilities []string       `json:"capabilities"`
	Endpoint     string         `json:"endpoint,omitempty"`
	Transport    string         `json:"transport"` // "websocket", "ipc", "http", "inproc"
	RegisteredAt time.Time      `json:"registered_at"`
	LastSeenAt   time.Time      `json:"last_seen_at"`
	State        ComponentState `json:"state"`
	Installed    bool           `json:"installed"`
	Enabled      bool           `json:"enabled"`
}

// EventType define los tipos de eventos emitidos por el bus del núcleo
type EventType string

const (
	EventPeerJoined     EventType = "PEER_JOINED"
	EventPeerLeft       EventType = "PEER_LEFT"
	EventPacketRx       EventType = "PACKET_RX"
	EventPacketTx       EventType = "PACKET_TX"
	EventPacketDrop     EventType = "PACKET_DROP"
	EventZTNAAlert      EventType = "ZTNA_ALERT"
	EventComponentBound EventType = "COMPONENT_BOUND"
	EventComponentUnbound EventType = "COMPONENT_UNBOUND"
	EventTopologyChange EventType = "TOPOLOGY_CHANGE"
)

// MeshEvent es un evento que se difunde a los componentes suscritos
type MeshEvent struct {
	Type      EventType              `json:"type"`
	Timestamp int64                  `json:"timestamp_ns"`
	Source    string                 `json:"source"`
	Payload   map[string]interface{} `json:"payload"`
}

// DatagramEnvelope encapsula un paquete de datos transmitido a través del bus
type DatagramEnvelope struct {
	TargetDID string `json:"target_did"`
	SourceDID string `json:"source_did,omitempty"`
	Protocol  string `json:"protocol"`
	Payload   []byte `json:"payload"`
	Priority  uint8  `json:"priority"` // 0: Normal, 1: Interactive, 2: Control
}

// CoreStatus describe el estado global del Núcleo Universal ipvn7
type CoreStatus struct {
	Version            string                  `json:"version"`
	DID                string                  `json:"did"`
	IPv6               string                  `json:"ipv6"`
	IPv4               string                  `json:"ipv4"`
	UptimeSeconds      int64                   `json:"uptime_seconds"`
	ActivePeersCount   int                     `json:"active_peers_count"`
	AttachedComponents []*ComponentRegistration `json:"attached_components"`
	TotalRxBytes       uint64                  `json:"total_rx_bytes"`
	TotalTxBytes       uint64                  `json:"total_tx_bytes"`
	TotalDrops         uint64                  `json:"total_drops"`
}
