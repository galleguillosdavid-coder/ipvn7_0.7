// Package l1 implementa el resolver DNS soberano anti-fugas con caché zero-alloc en memoria.
package l1

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

// Constantes canónicas del resolver soberano
const (
	DefaultDNSCacheTTL   = 300 * time.Second
	DefaultDNSTimeout    = 3 * time.Second
	MaxDNSCacheEntries   = 2048
)

// DNSCacheEntry representa un registro resuelto almacenado en memoria
type DNSCacheEntry struct {
	IP        net.IP
	ExpiresAt time.Time
}

// SovereignDNSResolver resuelve nombres evitando fugas de privacidad hacia el ISP local
type SovereignDNSResolver struct {
	mu           sync.RWMutex
	cache        map[string]DNSCacheEntry
	upstreams    []string
	customDialer func(ctx context.Context, network, address string) (net.Conn, error)
	timeout      time.Duration
}

// NewSovereignDNSResolver inicializa el resolver soberano con servidores upstream de respaldo
func NewSovereignDNSResolver(upstreams []string) *SovereignDNSResolver {
	if len(upstreams) == 0 {
		upstreams = []string{"1.1.1.1:53", "9.9.9.9:53"}
	}
	return &SovereignDNSResolver{
		cache:     make(map[string]DNSCacheEntry),
		upstreams: upstreams,
		timeout:   DefaultDNSTimeout,
	}
}

// SetCustomDialer permite enrutar las consultas DNS a través del túnel soberano
func (r *SovereignDNSResolver) SetCustomDialer(dialer func(ctx context.Context, network, address string) (net.Conn, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.customDialer = dialer
}

// ResolveIP resuelve un nombre de host a una dirección IP (IPv4 preferida)
func (r *SovereignDNSResolver) ResolveIP(host string) (net.IP, error) {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return nil, errors.New("host vacío")
	}

	// 1. Si ya es una IP válida directa, retornar sin consulta de red
	if ip := net.ParseIP(host); ip != nil {
		return ip, nil
	}

	now := time.Now()

	// 2. Búsqueda en caché local (cero alocaciones de heap)
	r.mu.RLock()
	if entry, found := r.cache[host]; found && now.Before(entry.ExpiresAt) {
		r.mu.RUnlock()
		return entry.IP, nil
	}
	r.mu.RUnlock()

	// 3. Resolución upstream segura
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
	}

	r.mu.RLock()
	dialer := r.customDialer
	upstreams := r.upstreams
	r.mu.RUnlock()

	if dialer != nil {
		resolver.Dial = dialer
	} else if len(upstreams) > 0 {
		targetUpstream := upstreams[0]
		resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: r.timeout}
			return d.DialContext(ctx, "udp", targetUpstream)
		}
	}

	ips, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.New("no se encontraron IPs para el host")
	}

	// Preferir IPv4 para compatibilidad amplia
	selectedIP := ips[0]
	for _, ip := range ips {
		if ip.To4() != nil {
			selectedIP = ip
			break
		}
	}

	// 4. Actualizar caché con expiración TTL
	r.mu.Lock()
	if len(r.cache) >= MaxDNSCacheEntries {
		// Poda simple del mapa si se alcanza el límite
		r.cache = make(map[string]DNSCacheEntry)
	}
	r.cache[host] = DNSCacheEntry{
		IP:        selectedIP,
		ExpiresAt: now.Add(DefaultDNSCacheTTL),
	}
	r.mu.Unlock()

	return selectedIP, nil
}

// ClearCache purga los registros en memoria
func (r *SovereignDNSResolver) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]DNSCacheEntry)
}

// CacheSize retorna la cantidad de dominios en memoria
func (r *SovereignDNSResolver) CacheSize() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cache)
}
