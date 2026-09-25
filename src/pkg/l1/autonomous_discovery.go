// Package l1 implementa el Motor de Descubrimiento Autónomo de Malla Soberana.
// Permite que nodos aislados, desconectados de la LAN o detrás de NAT/CGNAT,
// se descubran y emparejen automáticamente mediante balizas ciegas EBRA y STUN reflexivo.
package l1

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ipvn7/pkg/l0"
)

// KuzuPeerUpserter abstrae la sincronización topológica con KùzuDB
type KuzuPeerUpserter interface {
	UpsertPeer(did, alias, peerType string, isCurrent bool)
	UpsertXORLink(sourceDID, targetDID, relation string, latencyMs float64, hopDistance int8, iface string)
}

// AutonomousDiscoveryEngine orquesta el descubrimiento continuo y autónomo de pares
type AutonomousDiscoveryEngine struct {
	mu         sync.Mutex
	Identity   *l0.Identity
	Router     *KleinbergRouter
	Rendezvous *BlindRendezvousManager
	Firewall   *ZTNAFirewall
	Kuzu       KuzuPeerUpserter
	PacketConn net.PacketConn
	ListenPort int
	running           bool
	stopChan          chan struct{}
	discovered        map[string]time.Time
	reflexiveEndpoint string
}

// NewAutonomousDiscoveryEngine inicializa el motor de descubrimiento
func NewAutonomousDiscoveryEngine(
	id *l0.Identity,
	router *KleinbergRouter,
	rz *BlindRendezvousManager,
	fw *ZTNAFirewall,
	kuzu KuzuPeerUpserter,
	conn net.PacketConn,
	port int,
) *AutonomousDiscoveryEngine {
	return &AutonomousDiscoveryEngine{
		Identity:   id,
		Router:     router,
		Rendezvous: rz,
		Firewall:   fw,
		Kuzu:       kuzu,
		PacketConn: conn,
		ListenPort: port,
		discovered: make(map[string]time.Time),
		stopChan:   make(chan struct{}),
	}
}

// Start inicia el bucle de descubrimiento autónomo en segundo plano
func (e *AutonomousDiscoveryEngine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	go e.discoveryLoop()
}

// Stop detiene el bucle de descubrimiento
func (e *AutonomousDiscoveryEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return
	}
	e.running = false
	close(e.stopChan)
}

func (e *AutonomousDiscoveryEngine) discoveryLoop() {
	// Pulso inicial inmediato
	e.executeCycle()

	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.executeCycle()
		}
	}
}

func (e *AutonomousDiscoveryEngine) executeCycle() {
	// Contar sólo pares remotos con estado saludable
	remoteHealthyPeers := 0
	for _, p := range e.Router.GetAllPeers() {
		if p.DID != e.Identity.DID() && p.HealthState == HealthStateHealthy {
			remoteHealthyPeers++
		}
	}

	// Si no hay pares remotos saludables, reactivar señalización externa (Auto-curación post-desconexión)
	if remoteHealthyPeers < 1 && e.Rendezvous.IsCircuitBroken() {
		e.Rendezvous.ResetCircuitBreaker()
	}

	// Si ya alcanzamos al menos 2 pares remotos y el circuit breaker se disparó, reposar
	if remoteHealthyPeers >= MinPeerTripThreshold && e.Rendezvous.IsCircuitBroken() {
		return
	}

	// 0. Mantener pinhole NAT activo emitiendo STUN directamente por el socket físico
	e.SendSTUNKeepalive()

	// 1. Recolectar interfaces locales y sonda reflexiva STUN RFC 5389
	var candidateEndpoints []string

	e.mu.Lock()
	if e.reflexiveEndpoint != "" {
		candidateEndpoints = append(candidateEndpoints, e.reflexiveEndpoint)
	}
	e.mu.Unlock()

	localIPs := getLocalNonLoopbackIPs()
	for _, ip := range localIPs {
		candidateEndpoints = append(candidateEndpoints, fmt.Sprintf("%s:%d", ip, e.ListenPort))
	}

	if len(candidateEndpoints) == 0 {
		candidateEndpoints = append(candidateEndpoints, fmt.Sprintf("127.0.0.1:%d", e.ListenPort))
	}

	endpointPayload := strings.Join(candidateEndpoints, ",")

	// 2. Publicar Baliza Ciega Efímera EBRA
	if !e.Rendezvous.IsCircuitBroken() {
		env, err := e.Rendezvous.PublishBlindBeacon(endpointPayload, 0)
		if err != nil {
			fmt.Printf("[EBRA] Error publicando baliza: %v\n", err)
		} else if env != nil {
			fmt.Printf("[EBRA] Baliza publicada en tópico %s (Endpoints: %s)\n", env.TopicID[:16]+"...", endpointPayload)
		}
	}

	// 3. Enviar datagrama broadcast LAN de cortesía (por si ambos nodos comparten LAN)
	e.broadcastLocalBeacon()

	// 3.1 Probar pares de confianza persistidos (Auto-roaming WAN y NAT Hole Punching)
	e.probeTrustedPeers()

	// 4. Descubrir balizas de pares remotos en el tópico EBRA
	beacons, err := e.Rendezvous.DiscoverAllPeers(0)
	if err != nil || len(beacons) == 0 {
		return
	}

	for _, peerBeacon := range beacons {
		if peerBeacon.DID == e.Identity.DID() {
			continue
		}

		endpoints := strings.Split(peerBeacon.Endpoint, ",")
		for _, ep := range endpoints {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}

			udpAddr, err := net.ResolveUDPAddr("udp", ep)
			if err != nil {
				continue
			}

			// Autorizar en Cortafuegos ZTNA
			if e.Firewall != nil {
				e.Firewall.AuthorizeDID(&DIDPolicy{
					DID:           peerBeacon.DID,
					AllowInbound:  true,
					AllowOutbound: true,
					AllowRelay:    true,
				})
			}

			// Registrar en Kleinberg Router
			_ = e.Router.AddOrUpdatePeer(peerBeacon.DID, udpAddr, 2.5)

			// Sincronizar en Grafo KùzuDB
			if e.Kuzu != nil {
				e.Kuzu.UpsertPeer(peerBeacon.DID, "", "kleinberg_peer", false)
				e.Kuzu.UpsertXORLink(e.Identity.DID(), peerBeacon.DID, "xor_metric", 2.5, 0, "EBRA-WAN")
			}

			// Enviar datagrama UDP firmado de RoamingUpdate para perforar NAT (Hole Punching)
			e.sendHandshakePacket(peerBeacon.DID, udpAddr)

			e.mu.Lock()
			last, exists := e.discovered[peerBeacon.DID]
			if !exists || time.Since(last) > 30*time.Second {
				e.discovered[peerBeacon.DID] = time.Now()
				fmt.Printf("[DESCUBRIMIENTO AUTÓNOMO EBRA] ¡Par descubierto!: %s en %s\n", peerBeacon.DID[:24]+"...", ep)
			}
			e.mu.Unlock()

			e.Rendezvous.NotifyDirectPeerConnected()
		}
	}
}

func (e *AutonomousDiscoveryEngine) broadcastLocalBeacon() {
	if e.PacketConn == nil {
		return
	}

	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)

	pkt := l0.NewPacket(
		l0.MsgTypeRoamingUpdate,
		e.Identity.DID(),
		"",
		1,
		nonce,
		[]byte(fmt.Sprintf("%d", e.ListenPort)),
	)
	_ = pkt.SignPacket(e.Identity)
	encoded, err := pkt.Encode()
	if err != nil {
		return
	}

	targetPorts := []int{7777, 7778, 7001, 8080}
	for _, p := range targetPorts {
		dst := &net.UDPAddr{IP: net.IPv4bcast, Port: p}
		_, _ = e.PacketConn.WriteTo(encoded, dst)
	}
}

func (e *AutonomousDiscoveryEngine) sendHandshakePacket(targetDID string, targetAddr *net.UDPAddr) {
	if e.PacketConn == nil {
		return
	}

	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)

	pkt := l0.NewPacket(
		l0.MsgTypeRoamingUpdate,
		e.Identity.DID(),
		targetDID,
		uint64(time.Now().UnixNano()),
		nonce,
		[]byte(fmt.Sprintf("endpoint:%d", e.ListenPort)),
	)
	_ = pkt.SignPacket(e.Identity)
	encoded, err := pkt.Encode()
	if err != nil {
		return
	}

	_, _ = e.PacketConn.WriteTo(encoded, targetAddr)
}

// SetReflexiveEndpoint registra la IP:puerto público activo obtenido por STUN en el socket físico
func (e *AutonomousDiscoveryEngine) SetReflexiveEndpoint(ep string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.reflexiveEndpoint = ep
}

// SendSTUNKeepalive emite una trama STUN por el socket principal para abrir y mantener el cono NAT
func (e *AutonomousDiscoveryEngine) SendSTUNKeepalive() {
	if e.PacketConn == nil {
		return
	}
	req, _, err := BuildSTUNBindingRequest()
	if err != nil {
		return
	}
	for _, srv := range DefaultSTUNServers {
		dst, err := net.ResolveUDPAddr("udp4", srv)
		if err == nil {
			_, _ = e.PacketConn.WriteTo(req, dst)
			break
		}
	}
}

// TrustedPeerEntry representa la estructura de un par persistido en keystore
type TrustedPeerEntry struct {
	DID           string   `json:"did"`
	Name          string   `json:"name"`
	Endpoints     []string `json:"endpoints"`
	VirtualIPv4   string   `json:"virtual_ipv4"`
	VirtualIPv6   string   `json:"virtual_ipv6"`
	AutoReconnect bool     `json:"auto_reconnect"`
}

func (e *AutonomousDiscoveryEngine) probeTrustedPeers() {
	paths := []string{
		filepath.Join("keystore", "trusted_peers.json"),
		`C:\ipvn7\keystore\trusted_peers.json`,
		filepath.Join("..", "keystore", "trusted_peers.json"),
	}
	var data []byte
	for _, p := range paths {
		if b, err := os.ReadFile(p); err == nil && len(b) > 0 {
			data = b
			break
		}
	}
	if len(data) == 0 {
		return
	}

	var peers []TrustedPeerEntry
	if err := json.Unmarshal(data, &peers); err != nil {
		return
	}

	for _, peer := range peers {
		if peer.DID == e.Identity.DID() || peer.DID == "" {
			continue
		}
		if e.Firewall != nil {
			e.Firewall.AuthorizeDID(&DIDPolicy{
				DID:           peer.DID,
				AllowInbound:  true,
				AllowOutbound: true,
				AllowRelay:    true,
			})
		}
		for _, ep := range peer.Endpoints {
			if strings.HasPrefix(ep, "wan_dynamic") || ep == "" {
				continue
			}
			udpAddr, err := net.ResolveUDPAddr("udp", ep)
			if err != nil {
				continue
			}
			_ = e.Router.AddOrUpdatePeer(peer.DID, udpAddr, 1.2)
			e.sendHandshakePacket(peer.DID, udpAddr)
		}
	}
}

func getLocalNonLoopbackIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}
