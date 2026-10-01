package core

import (
	"net"
	"os"
	"sync"
)

// NetworkConfig centraliza la configuración de red dinámica
type NetworkConfig struct {
	mu sync.RWMutex

	// IPs dinámicas del sistema
	LocalIP        string
	LocalHostname  string

	// Puertos configurables
	DefaultWebPort     int
	DefaultUDPPort     int
	AlternativeWebPort int
	AlternativeUDPPort int

	// Peers conocidos (puede cargarse desde archivo)
	KnownPeers map[string]string // nombre -> ip:puerto
}

// DefaultNetworkConfig crea configuración con valores por defecto
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		DefaultWebPort:     7070,
		DefaultUDPPort:     7777,
		AlternativeWebPort: 8080,
		AlternativeUDPPort: 7001,
		KnownPeers:        make(map[string]string),
	}
}

// DetectLocalIP detecta automáticamente la IP local
func (nc *NetworkConfig) DetectLocalIP() error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	hostname, _ := os.Hostname()
	nc.LocalHostname = hostname

	// Intentar detectar IP local conectando a dirección externa
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: usar primera interfaz disponible
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					nc.LocalIP = ipnet.IP.String()
					return nil
				}
			}
		}
		return err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	nc.LocalIP = localAddr.IP.String()
	return nil
}

// GetLocalIP retorna la IP local detectada
func (nc *NetworkConfig) GetLocalIP() string {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.LocalIP
}

// GetWebPorts retorna puertos web configurados
func (nc *NetworkConfig) GetWebPorts() []int {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return []int{nc.DefaultWebPort, nc.AlternativeWebPort}
}

// GetUDPPorts retorna puertos UDP configurados
func (nc *NetworkConfig) GetUDPPorts() []int {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return []int{nc.DefaultUDPPort, nc.AlternativeUDPPort}
}

// AddKnownPeer agrega un peer conocido
func (nc *NetworkConfig) AddKnownPeer(name, address string) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.KnownPeers[name] = address
}

// GetKnownPeer retorna dirección de peer conocido
func (nc *NetworkConfig) GetKnownPeer(name string) (string, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	addr, ok := nc.KnownPeers[name]
	return addr, ok
}

// ResolvePeerName intenta resolver nombre descriptivo para una IP
func (nc *NetworkConfig) ResolvePeerName(ip string) string {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	// Primero buscar en peers conocidos
	for name, addr := range nc.KnownPeers {
		if addr == ip || contains(addr, ip) {
			return name
		}
	}

	// Fallback: usar hostname si es IP local
	if ip == nc.LocalIP {
		if nc.LocalHostname != "" {
			return nc.LocalHostname + " (Núcleo Local)"
		}
		return "Núcleo Local"
	}

	// Fallback: nombre genérico
	return "Par (" + ip + ")"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
