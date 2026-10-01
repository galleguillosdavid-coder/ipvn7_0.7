// Package l1 implementa el Registro de Dispositivos Sombra (Shadow DIDs) y Pasarela L4.
// Permite que un nodo ipvn7 actúe como Nodo Guardián (Edge Ambassador) para impresoras,
// Smart TVs y dispositivos IoT locales sin modificar su firmware original.
package l1

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DeviceType categoriza el hardware local puenteado
type DeviceType string

const (
	DeviceTypePrinter DeviceType = "printer"
	DeviceTypeDisplay DeviceType = "display"
	DeviceTypeCamera  DeviceType = "camera"
	DeviceTypeIoT     DeviceType = "iot"
	DeviceTypeEdgeAI  DeviceType = "edge_ai"
	DeviceTypeGeneric DeviceType = "generic"
)

// ShadowDevice representa un dispositivo físico local con identidad soberana virtual
type ShadowDevice struct {
	DID          string     `json:"did"`
	Name         string     `json:"name"`
	DeviceType   DeviceType `json:"device_type"`
	PhysicalIP   string     `json:"physical_ip"`
	MAC          string     `json:"mac"`
	VirtualIPv4  string     `json:"virtual_ipv4"`
	AllowedPorts []int      `json:"allowed_ports"`
	CreatedAt    time.Time  `json:"created_at"`
	Enabled      bool       `json:"enabled"`
}

// ShadowDeviceRegistry gestiona el catálogo de dispositivos LAN representados
type ShadowDeviceRegistry struct {
	mu           sync.RWMutex
	devicesByDID map[string]*ShadowDevice
	devicesByVIP map[string]*ShadowDevice
	nextIPIndex  int
}

// NewShadowDeviceRegistry inicializa el registro
func NewShadowDeviceRegistry() *ShadowDeviceRegistry {
	return &ShadowDeviceRegistry{
		devicesByDID: make(map[string]*ShadowDevice),
		devicesByVIP: make(map[string]*ShadowDevice),
		nextIPIndex:  10,
	}
}

// DeriveShadowDID calcula un DID determinista basado en el hash de la MAC física
func DeriveShadowDID(mac string) string {
	cleanMAC := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))
	h := sha256.Sum256([]byte("ipvn7:shadow:" + cleanMAC))
	return "did:ipvn7:shadow:" + hex.EncodeToString(h[:16])
}

// RegisterDevice registra un nuevo dispositivo físico en la malla soberana
func (r *ShadowDeviceRegistry) RegisterDevice(name, physicalIP, mac string, devType DeviceType, ports []int) (*ShadowDevice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if physicalIP == "" || mac == "" {
		return nil, errors.New("IP física y MAC son requeridas")
	}

	did := DeriveShadowDID(mac)
	if existing, ok := r.devicesByDID[did]; ok {
		existing.Name = name
		existing.PhysicalIP = physicalIP
		existing.AllowedPorts = ports
		existing.Enabled = true
		return existing, nil
	}

	if r.nextIPIndex > 254 {
		return nil, errors.New("límite de IPs virtuales sombra alcanzado en el bloque local")
	}

	vip := fmt.Sprintf("10.7.100.%d", r.nextIPIndex)
	r.nextIPIndex++

	dev := &ShadowDevice{
		DID:          did,
		Name:         name,
		DeviceType:   devType,
		PhysicalIP:   physicalIP,
		MAC:          mac,
		VirtualIPv4:  vip,
		AllowedPorts: ports,
		CreatedAt:    time.Now(),
		Enabled:      true,
	}

	r.devicesByDID[did] = dev
	r.devicesByVIP[vip] = dev
	return dev, nil
}

// LookupByDID busca un dispositivo sombra por su DID
func (r *ShadowDeviceRegistry) LookupByDID(did string) *ShadowDevice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.devicesByDID[did]
}

// LookupByVirtualIP busca un dispositivo sombra por su IP virtual 10.7.X.Y
func (r *ShadowDeviceRegistry) LookupByVirtualIP(vip string) *ShadowDevice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.devicesByVIP[vip]
}

// ListDevices retorna todos los dispositivos registrados
func (r *ShadowDeviceRegistry) ListDevices() []*ShadowDevice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]*ShadowDevice, 0, len(r.devicesByDID))
	for _, d := range r.devicesByDID {
		res = append(res, d)
	}
	return res
}

// RegisterEdgeAIWorker registra un nodo de cómputo/inferencia local (BitNet, Ollama, Exo)
func (r *ShadowDeviceRegistry) RegisterEdgeAIWorker(name, physicalIP, mac string, port int) (*ShadowDevice, error) {
	if port <= 0 {
		port = 11434
	}
	return r.RegisterDevice(name, physicalIP, mac, DeviceTypeEdgeAI, []int{port})
}

// FindOptimalInferenceWorker busca el primer worker de inferencia Edge AI disponible
func (r *ShadowDeviceRegistry) FindOptimalInferenceWorker() *ShadowDevice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.devicesByDID {
		if d.DeviceType == DeviceTypeEdgeAI && d.Enabled {
			return d
		}
	}
	return nil
}

// ForwardStream reenvía un flujo TCP desde el túnel de malla hacia el socket físico local
func (r *ShadowDeviceRegistry) ForwardStream(clientConn net.Conn, targetDID string, targetPort int) error {
	dev := r.LookupByDID(targetDID)
	if dev == nil || !dev.Enabled {
		return errors.New("dispositivo sombra no encontrado o deshabilitado")
	}

	portAllowed := false
	for _, p := range dev.AllowedPorts {
		if p == targetPort {
			portAllowed = true
			break
		}
	}
	if !portAllowed && len(dev.AllowedPorts) > 0 {
		return fmt.Errorf("puerto %d no permitido para el dispositivo %s", targetPort, dev.Name)
	}

	targetAddr := net.JoinHostPort(dev.PhysicalIP, strconv.Itoa(targetPort))
	targetConn, err := net.DialTimeout("tcp", targetAddr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("error conectando con dispositivo físico %s: %w", targetAddr, err)
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(targetConn, clientConn)
		_ = targetConn.Close()
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(clientConn, targetConn)
		_ = clientConn.Close()
	}()

	wg.Wait()
	return nil
}

// WakeDevice inyecta un Magic Packet Wake-on-LAN físico hacia el broadcast de la LAN
func (r *ShadowDeviceRegistry) WakeDevice(mac string) error {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("MAC inválida: %w", err)
	}

	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 6; i < 102; i += 6 {
		copy(packet[i:i+6], hw)
	}

	bcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9")
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, bcastAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}
