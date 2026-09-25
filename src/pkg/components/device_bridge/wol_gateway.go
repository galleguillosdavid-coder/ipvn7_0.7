// Package devicebridge implementa el gateway físico Wake-on-LAN (WoL) y la gestión de
// Shadow DIDs para periféricos de red local integrados a la malla ipvn7 v0.7.
package devicebridge

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// ShadowDevice representa un periférico de LAN mapeado a la identidad soberana
type ShadowDevice struct {
	ShadowDID string    `json:"shadow_did"` // did:ipvn7:shadow:<mac_hex>
	MAC       net.HardwareAddr `json:"mac"`
	IP        net.IP    `json:"ip"`
	Name      string    `json:"name"`
	LastWake  time.Time `json:"last_wake"`
}

// WoLGateway gestiona la emisión de Magic Packets físicos y el catálogo de Shadow DIDs
type WoLGateway struct {
	mu            sync.RWMutex
	devices       map[string]*ShadowDevice
	broadcastAddr string
}

// NewWoLGateway inicializa el gateway de Wake-on-LAN
func NewWoLGateway(broadcastAddr string) *WoLGateway {
	if broadcastAddr == "" {
		broadcastAddr = "255.255.255.255:9"
	}
	return &WoLGateway{
		devices:       make(map[string]*ShadowDevice),
		broadcastAddr: broadcastAddr,
	}
}

// RegisterShadowDevice registra un periférico asignándole un Shadow DID canónico
func (g *WoLGateway) RegisterShadowDevice(macStr, name string, ip net.IP) (*ShadowDevice, error) {
	mac, err := net.ParseMAC(macStr)
	if err != nil {
		return nil, fmt.Errorf("wol: direccion MAC invalida: %w", err)
	}

	cleanHex := strings.ToLower(strings.ReplaceAll(mac.String(), ":", ""))
	shadowDID := fmt.Sprintf("did:ipvn7:shadow:%s", cleanHex)

	g.mu.Lock()
	defer g.mu.Unlock()

	dev := &ShadowDevice{
		ShadowDID: shadowDID,
		MAC:       mac,
		IP:        ip,
		Name:      name,
	}
	g.devices[shadowDID] = dev
	return dev, nil
}

// BuildMagicPacket construye el paquete físico Wake-on-LAN (6x 0xFF + 16x MAC = 102 bytes)
func BuildMagicPacket(mac net.HardwareAddr) ([]byte, error) {
	if len(mac) != 6 {
		return nil, errors.New("wol: la MAC debe tener exactamente 6 octetos")
	}

	packet := make([]byte, 102)
	// 6 bytes de sincronización 0xFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	// 16 repeticiones de la dirección MAC
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:6+(i+1)*6], mac)
	}
	return packet, nil
}

// Wake emite un Magic Packet físico real hacia la dirección broadcast de la subred
func (g *WoLGateway) Wake(shadowDID string) error {
	g.mu.Lock()
	dev, exists := g.devices[shadowDID]
	if !exists || dev == nil {
		g.mu.Unlock()
		return fmt.Errorf("wol: periferico con DID %s no encontrado", shadowDID)
	}
	mac := dev.MAC
	dev.LastWake = time.Now()
	broadcast := g.broadcastAddr
	g.mu.Unlock()

	packet, err := BuildMagicPacket(mac)
	if err != nil {
		return err
	}

	uAddr, err := net.ResolveUDPAddr("udp", broadcast)
	if err != nil {
		return fmt.Errorf("wol: fallo al resolver broadcast: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, uAddr)
	if err != nil {
		return fmt.Errorf("wol: fallo al abrir socket UDP: %w", err)
	}
	defer conn.Close()

	n, err := conn.Write(packet)
	if err != nil {
		return fmt.Errorf("wol: fallo al emitir paquete: %w", err)
	}
	if n != len(packet) {
		return fmt.Errorf("wol: envio incompleto (%d de %d bytes)", n, len(packet))
	}

	return nil
}

// GetDevice retorna la ficha de un periférico registrado por su Shadow DID
func (g *WoLGateway) GetDevice(shadowDID string) (*ShadowDevice, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	dev, exists := g.devices[shadowDID]
	return dev, exists
}
