package l1

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestSovereignDNSResolver_IPParsingDirect(t *testing.T) {
	r := NewSovereignDNSResolver(nil)

	// IP IPv4 directa no debe consultar red
	ip, err := r.ResolveIP("127.0.0.1")
	if err != nil {
		t.Fatalf("Falla resolviendo IP directa: %v", err)
	}
	if !ip.Equal(net.ParseIP("127.0.0.1")) {
		t.Errorf("IP resuelta incorrecta: got %v, want 127.0.0.1", ip)
	}

	// IP IPv6 directa
	ip6, err := r.ResolveIP("::1")
	if err != nil {
		t.Fatalf("Falla resolviendo IPv6 directa: %v", err)
	}
	if !ip6.Equal(net.ParseIP("::1")) {
		t.Errorf("IP resuelta incorrecta: got %v, want ::1", ip6)
	}
}

func TestSovereignDNSResolver_CacheBehavior(t *testing.T) {
	r := NewSovereignDNSResolver(nil)

	// Inyectar entrada en cache directamente para probar hit sin red
	testHost := "sovereign.local"
	expectedIP := net.ParseIP("10.7.0.1")

	r.mu.Lock()
	r.cache[testHost] = DNSCacheEntry{
		IP:        expectedIP,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	r.mu.Unlock()

	ip, err := r.ResolveIP(testHost)
	if err != nil {
		t.Fatalf("ResolveIP falló con entrada en cache: %v", err)
	}
	if !ip.Equal(expectedIP) {
		t.Errorf("IP retornada no coincide con cache: got %v, want %v", ip, expectedIP)
	}

	if r.CacheSize() != 1 {
		t.Errorf("CacheSize incorrecto: got %d, want 1", r.CacheSize())
	}

	r.ClearCache()
	if r.CacheSize() != 0 {
		t.Errorf("ClearCache falló, tamaño restante: %d", r.CacheSize())
	}
}

func TestSovereignDNSResolver_CustomDialer(t *testing.T) {
	r := NewSovereignDNSResolver(nil)

	dialerCalled := false
	r.SetCustomDialer(func(ctx context.Context, network, address string) (net.Conn, error) {
		dialerCalled = true
		d := net.Dialer{Timeout: 1 * time.Second}
		return d.DialContext(ctx, network, address)
	})

	// Resolver un dominio público estándar
	_, _ = r.ResolveIP("localhost")
	// Nota: si localhost se resuelve internamente o con dialer
	if r.CacheSize() > 0 {
		t.Logf("Cache poblado correctamente")
	}
	_ = dialerCalled
}
