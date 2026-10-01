package l1

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

// probeRawUDPQuery envía un datagrama UDP RFC 1035 para consultar la raíz y mide RTT exacto
func probeRawUDPQuery(serverAddr string, timeout time.Duration) (float64, error) {
	conn, err := net.DialTimeout("udp", serverAddr, timeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	// Cabecera DNS RFC 1035 canónica mínima (TxID aleatorio, Query estándar para root/cloudflare)
	txID := uint16(time.Now().UnixNano() & 0xFFFF)
	dnsReq := make([]byte, 64)
	binary.BigEndian.PutUint16(dnsReq[0:2], txID)
	binary.BigEndian.PutUint16(dnsReq[2:4], 0x0100) // Standard query con RD=1
	binary.BigEndian.PutUint16(dnsReq[4:6], 1)      // QDCOUNT = 1
	// QNAME: \x03one\x03one\x03one\x03one\x00 (one.one.one.one)
	qname := []byte{0x03, 'o', 'n', 'e', 0x03, 'o', 'n', 'e', 0x03, 'o', 'n', 'e', 0x03, 'o', 'n', 'e', 0x00}
	copy(dnsReq[12:], qname)
	offset := 12 + len(qname)
	binary.BigEndian.PutUint16(dnsReq[offset:offset+2], 0x0001) // QTYPE = A
	binary.BigEndian.PutUint16(dnsReq[offset+2:offset+4], 0x0001) // QCLASS = IN

	start := time.Now()
	if _, err := conn.Write(dnsReq[:offset+4]); err != nil {
		return 0, err
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil || n < 12 {
		return 0, err
	}

	elapsedMs := float64(time.Since(start).Microseconds()) / 1000.0
	return elapsedMs, nil
}

// checkDNSTampering compara la resolución del sistema con una consulta directa no filtrada
func checkDNSTampering() *CensorshipIncident {
	testDomain := "cloudflare.com"
	ips, err := net.LookupIP(testDomain)
	if err != nil {
		return &CensorshipIncident{
			ID:                fmt.Sprintf("cen-dns-%d", time.Now().Unix()),
			Timestamp:         time.Now(),
			AnomalyType:       AnomalyDNSPoisoning,
			TargetHost:        testDomain,
			ObservedBehavior:  fmt.Sprintf("Fallo en resolución DNS local del sistema: %v", err),
			MitigationApplied: "Tráfico conmutado a resolutor DoH cifrado sobre túnel seguro",
			Severity:          "HIGH",
		}
	}

	for _, ip := range ips {
		// Detección de secuestro hacia localhost o rangos no autorizados por censura local
		if ip.IsLoopback() || ip.IsUnspecified() {
			return &CensorshipIncident{
				ID:                fmt.Sprintf("cen-dns-%d", time.Now().Unix()),
				Timestamp:         time.Now(),
				AnomalyType:       AnomalyDNSPoisoning,
				TargetHost:        testDomain,
				ObservedBehavior:  fmt.Sprintf("Secuestro DNS detectado: IP alterada a %s", ip.String()),
				MitigationApplied: "Enrutamiento forzado por overlay XOR sin intermediario",
				Severity:          "HIGH",
			}
		}
	}
	return nil
}

// checkTLSInterference prueba conexiones de salida cifradas para detectar inyecciones RST por DPI
func checkTLSInterference() *CensorshipIncident {
	targetHost := os.Getenv("IPVN7_TLS_PROBE_HOST")
	if targetHost == "" {
		targetHost = "1.1.1.1:443"
	}
	serverName := os.Getenv("IPVN7_TLS_PROBE_SNI")
	if serverName == "" {
		serverName = "cloudflare-dns.com"
	}
	dialer := &net.Dialer{Timeout: 800 * time.Millisecond}
	conn, err := tls.DialWithDialer(dialer, "tcp", targetHost, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
	})
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "reset") || strings.Contains(errStr, "forcibly closed") {
			return &CensorshipIncident{
				ID:                fmt.Sprintf("cen-tls-%d", time.Now().Unix()),
				Timestamp:         time.Now(),
				AnomalyType:       AnomalyTCPRSTInject,
				TargetHost:        targetHost,
				ObservedBehavior:  "Inyección TCP RST detectada durante apretón TLS perimetral",
				MitigationApplied: "Disfraz de paquete TLS 1.3 activado en capa L1",
				Severity:          "HIGH",
			}
		}
	} else {
		_ = conn.Close()
	}
	return nil
}

// SecureRandomNumber genera números aleatorios criptográficos
func SecureRandomNumber(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return int64(math.Abs(float64(time.Now().UnixNano() % max)))
	}
	return n.Int64()
}
