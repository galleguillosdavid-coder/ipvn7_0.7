package devicebridge

import (
	"time"
)

// DeviceCategory categoriza los dispositivos hogareños descubiertos
type DeviceCategory string

const (
	CategoryPrinter  DeviceCategory = "PRINTER"
	CategorySmartTV  DeviceCategory = "SMART_TV"
	CategoryCast     DeviceCategory = "MEDIA_RENDERER"
	CategoryGateway  DeviceCategory = "ROUTER_GATEWAY"
	CategoryIoT      DeviceCategory = "IOT_GENERIC"
)

// DiscoveredDevice representa un aparato físico de la red doméstica
type DiscoveredDevice struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Category     DeviceCategory `json:"category"`
	IPv4         string         `json:"ipv4"`
	Port         int            `json:"port"`
	Protocol     string         `json:"protocol"` // "IPP", "RAW_9100", "DIAL", "SSDP", "HTTP"
	Manufacturer string         `json:"manufacturer,omitempty"`
	Model        string         `json:"model,omitempty"`
	ShadowDID    string         `json:"shadow_did"`
	Petname      string         `json:"petname"`
	LastSeen     time.Time      `json:"last_seen"`
	Capabilities []string       `json:"capabilities"`
}

// BridgeConfig define los parámetros de escaneo pasivo y registro
type BridgeConfig struct {
	CoreGatewayURL string        // ej. "http://127.0.0.1:7070"
	ScanInterval   time.Duration // ej. 45 segundos
	SubnetCIDR     string        // auto-detectada si está vacía
	EnableSSDP     bool
	EnableMDNS     bool
	EnablePortScan bool
}

// DefaultBridgeConfig retorna la configuración recomendada no invasiva
func DefaultBridgeConfig() BridgeConfig {
	return BridgeConfig{
		CoreGatewayURL: "http://127.0.0.1:7070",
		ScanInterval:   45 * time.Second,
		EnableSSDP:     true,
		EnableMDNS:     true,
		EnablePortScan: true,
	}
}
