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
	remoteHealthyPeers := 0
	for _, p := range e.Router.GetAllPeers() {
		if p.DID != e.Identity.DID() && p.HealthState == HealthStateHealthy {
			remoteHealthyPeers++
		}
	}
	if remoteHealthyPeers < 1 && e.Rendezvous.IsCircuitBroken() {
		e.Rendezvous.ResetCircuitBreaker()
	}
	if remoteHealthyPeers >= MinPeerTripThreshold && e.Rendezvous.IsCircuitBroken() {
		return
	}

	e.SendSTUNKeepalive()

	var endpoints []string
	e.mu.Lock()
	if e.reflexiveEndpoint != "" {
		endpoints = append(endpoints, e.reflexiveEndpoint)
	}
	e.mu.Unlock()
	for _, ip := range getLocalNonLoopbackIPs() {
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", ip, e.ListenPort))
	}
	if len(endpoints) == 0 {
		endpoints = append(endpoints, fmt.Sprintf("127.0.0.1:%d", e.ListenPort))
	}

	endpointPayload := strings.Join(endpoints, ",")
	if !e.Rendezvous.IsCircuitBroken() {
		_, _ = e.Rendezvous.PublishBlindBeacon(endpointPayload, 0)
	}

	e.broadcastLocalBeacon()
	e.probeTrustedPeers()

	beacons, err := e.Rendezvous.DiscoverAllPeers(0)
	if err != nil || len(beacons) == 0 {
		return
	}

	for _, b := range beacons {
		if b.DID == e.Identity.DID() {
			continue
		}
		for _, ep := range strings.Split(b.Endpoint, ",") {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			udpAddr, err := net.ResolveUDPAddr("udp", ep)
			if err != nil {
				continue
			}
			if e.Firewall != nil {
				e.Firewall.AuthorizeDID(&DIDPolicy{DID: b.DID, AllowInbound: true, AllowOutbound: true, AllowRelay: true})
			}
			_ = e.Router.AddOrUpdatePeer(b.DID, udpAddr, 2.5)
			if e.Kuzu != nil {
				e.Kuzu.UpsertPeer(b.DID, "", "kleinberg_peer", false)
				e.Kuzu.UpsertXORLink(e.Identity.DID(), b.DID, "xor_metric", 2.5, 0, "EBRA-WAN")
			}
			e.sendHandshakePacket(b.DID, udpAddr)
			e.mu.Lock()
			if last, exists := e.discovered[b.DID]; !exists || time.Since(last) > 30*time.Second {
				e.discovered[b.DID] = time.Now()
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
	pkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, e.Identity.DID(), "", 1, nonce, []byte(fmt.Sprintf("%d", e.ListenPort)))
	_ = pkt.SignPacket(e.Identity)
	if enc, err := pkt.Encode(); err == nil {
		for _, p := range []int{7777, 7778, 7001, 8080} {
			_, _ = e.PacketConn.WriteTo(enc, &net.UDPAddr{IP: net.IPv4bcast, Port: p})
		}
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
	var data []byte
	for _, p := range []string{filepath.Join("keystore", "trusted_peers.json"), `C:\ipvn7\keystore\trusted_peers.json`, filepath.Join("..", "keystore", "trusted_peers.json")} {
		if b, err := os.ReadFile(p); err == nil && len(b) > 0 {
			data = b
			break
		}
	}
	var peers []TrustedPeerEntry
	if len(data) == 0 || json.Unmarshal(data, &peers) != nil {
		return
	}
	for _, peer := range peers {
		if peer.DID == e.Identity.DID() || peer.DID == "" {
			continue
		}
		if e.Firewall != nil {
			e.Firewall.AuthorizeDID(&DIDPolicy{DID: peer.DID, AllowInbound: true, AllowOutbound: true, AllowRelay: true})
		}
		for _, ep := range peer.Endpoints {
			if strings.HasPrefix(ep, "wan_dynamic") || ep == "" {
				continue
			}
			if udpAddr, err := net.ResolveUDPAddr("udp", ep); err == nil {
				_ = e.Router.AddOrUpdatePeer(peer.DID, udpAddr, 1.2)
				e.sendHandshakePacket(peer.DID, udpAddr)
			}
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
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					ips = append(ips, ipNet.IP.String())
				}
			}
		}
	}
	return ips
}
