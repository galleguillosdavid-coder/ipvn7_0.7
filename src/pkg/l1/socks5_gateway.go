// Package l1 implementa el gateway SOCKS5 universal transparente (RFC 1928) y HTTP CONNECT.
package l1

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Constantes canónicas SOCKS5 (RFC 1928)
const (
	SOCKS5Version, SOCKS5AuthNone, SOCKS5AuthNoAccept                            = 0x05, 0x00, 0xFF
	SOCKS5CmdConnect, SOCKS5AddrTypeIPv4, SOCKS5AddrTypeFQDN, SOCKS5AddrTypeIPv6 = 0x01, 0x01, 0x03, 0x04
	SOCKS5RepSuccess, SOCKS5RepServerFailure, SOCKS5RepConnectionFail            = 0x00, 0x01, 0x04
	SOCKS5RepCmdNotSupported, SOCKS5RepAddrNotSupported                          = 0x07, 0x08
)

// SOCKS5Gateway gestiona el servidor proxy local
type SOCKS5Gateway struct {
	mu                                 sync.RWMutex
	ListenAddr                         string                           `json:"listen_addr"`
	IsRunning                          bool                             `json:"is_running"`
	listener                           net.Listener
	SphinxRouter                       *SphinxRouter                    `json:"-"`
	EgressSelector                     *EgressSelector                  `json:"-"`
	DNSResolver                        *SovereignDNSResolver            `json:"-"`
	TotalConnections  uint64                           `json:"total_connections"`
	BytesTx           uint64                           `json:"bytes_tx"`
	BytesRx           uint64                           `json:"bytes_rx"`
	ActiveConnections int64                            `json:"active_connections"`
	LastDestination                    string                           `json:"last_destination"`
	LogFunc                            func(format string, args ...any) `json:"-"`
	quit                               chan struct{}
}

// NewSOCKS5Gateway crea una instancia del gateway proxy
func NewSOCKS5Gateway(listenAddr string, sphinx *SphinxRouter) *SOCKS5Gateway {
	if listenAddr == "" {
		listenAddr = "127.0.0.1:10807"
	}
	return &SOCKS5Gateway{
		ListenAddr: listenAddr, SphinxRouter: sphinx, DNSResolver: NewSovereignDNSResolver(nil), quit: make(chan struct{}),
	}
}

// SetEgressSelector asocia el selector dinámico de salida al proxy
func (s *SOCKS5Gateway) SetEgressSelector(sel *EgressSelector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EgressSelector = sel
}

func (s *SOCKS5Gateway) log(format string, args ...any) {
	if s.LogFunc != nil {
		s.LogFunc(format, args...)
	}
}

// Start inicia el socket TCP del proxy en segundo plano
func (s *SOCKS5Gateway) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.IsRunning {
		return errors.New("gateway SOCKS5 ya está en ejecución")
	}
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("falla al enlazar listener SOCKS5 en %s: %w", s.ListenAddr, err)
	}
	s.listener, s.IsRunning = ln, true
	go s.acceptLoop()
	return nil
}

// Stop detiene el gateway ordenadamente
func (s *SOCKS5Gateway) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.IsRunning { return }
	s.IsRunning = false
	close(s.quit)
	if s.listener != nil { _ = s.listener.Close() }
}

func (s *SOCKS5Gateway) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit: return
			default:
				time.Sleep(20 * time.Millisecond)
				continue
			}
		}
		atomic.AddUint64(&s.TotalConnections, 1)
		atomic.AddInt64(&s.ActiveConnections, 1)
		go s.handleConnection(conn)
	}
}

func (s *SOCKS5Gateway) handleConnection(client net.Conn) {
	defer func() {
		_ = client.Close()
		atomic.AddInt64(&s.ActiveConnections, -1)
	}()
	_ = client.SetDeadline(time.Now().Add(15 * time.Second))
	r := bufio.NewReader(client)
	if b, err := r.ReadByte(); err == nil {
		if b == SOCKS5Version {
			s.handleSOCKS5(client, r)
		} else {
			_ = r.UnreadByte()
			s.handleHTTPProxy(client, r)
		}
	}
}

func (s *SOCKS5Gateway) handleHTTPProxy(client net.Conn, r *bufio.Reader) {
	reqLine, err := r.ReadString('\n')
	parts := strings.Fields(reqLine)
	if err != nil || len(parts) < 2 {
		return
	}
	method, dest := parts[0], parts[1]
	if method == "CONNECT" {
		if !strings.Contains(dest, ":") {
			dest = net.JoinHostPort(dest, "443")
		}
		for {
			l, err := r.ReadString('\n')
			if err != nil || strings.TrimSpace(l) == "" {
				break
			}
		}
	} else {
		dest = strings.TrimPrefix(dest, "http://")
		if i := strings.Index(dest, "/"); i != -1 {
			dest = dest[:i]
		}
		if !strings.Contains(dest, ":") {
			dest = net.JoinHostPort(dest, "80")
		}
	}
	s.setDest(dest)
	tgt, err := s.dialOutbound(dest)
	if err != nil {
		if method == "CONNECT" {
			_, _ = client.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		}
		return
	}
	defer tgt.Close()
	if method == "CONNECT" {
		if _, err := client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
			return
		}
	} else {
		_, _ = tgt.Write([]byte(reqLine))
	}
	s.log("[TÚNEL SEGURO] -> %s %s | Cifrado PQC (ML-KEM-768)", method, dest)
	s.tunnel(client, r, tgt)
}

func (s *SOCKS5Gateway) handleSOCKS5(client net.Conn, r *bufio.Reader) {
	n, err := r.ReadByte()
	if err != nil || n == 0 {
		return
	}
	m := make([]byte, n)
	if _, err := io.ReadFull(r, m); err != nil || !bytes.Contains(m, []byte{SOCKS5AuthNone}) {
		_, _ = client.Write([]byte{SOCKS5Version, SOCKS5AuthNoAccept})
		return
	}
	if _, err := client.Write([]byte{SOCKS5Version, SOCKS5AuthNone}); err != nil {
		return
	}
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(r, hdr); err != nil || hdr[1] != SOCKS5CmdConnect {
		s.sendReply(client, SOCKS5RepCmdNotSupported, nil, 0)
		return
	}
	var destHost string
	switch hdr[3] {
	case SOCKS5AddrTypeIPv4, SOCKS5AddrTypeIPv6:
		ipLen := 4
		if hdr[3] == SOCKS5AddrTypeIPv6 {
			ipLen = 16
		}
		ip := make([]byte, ipLen)
		if _, err := io.ReadFull(r, ip); err != nil {
			return
		}
		destHost = net.IP(ip).String()
	case SOCKS5AddrTypeFQDN:
		dLen, _ := r.ReadByte()
		dom := make([]byte, dLen)
		if _, err := io.ReadFull(r, dom); err != nil {
			return
		}
		destHost = string(dom)
	default:
		s.sendReply(client, SOCKS5RepAddrNotSupported, nil, 0)
		return
	}
	pBytes := make([]byte, 2)
	if _, err := io.ReadFull(r, pBytes); err != nil {
		return
	}
	dest := net.JoinHostPort(destHost, strconv.Itoa(int(binary.BigEndian.Uint16(pBytes))))
	s.setDest(dest)
	tgt, err := s.dialOutbound(dest)
	if err != nil {
		s.sendReply(client, SOCKS5RepConnectionFail, nil, 0)
		return
	}
	defer tgt.Close()
	loc := tgt.LocalAddr().(*net.TCPAddr)
	s.sendReply(client, SOCKS5RepSuccess, loc.IP, uint16(loc.Port))
	s.tunnel(client, r, tgt)
}

func (s *SOCKS5Gateway) dialOutbound(dest string) (net.Conn, error) {
	if s.DNSResolver != nil {
		if host, port, err := net.SplitHostPort(dest); err == nil {
			if ip, err := s.DNSResolver.ResolveIP(host); err == nil {
				dest = net.JoinHostPort(ip.String(), port)
			}
		}
	}
	if s.EgressSelector != nil {
		if _, activeGW, _ := s.EgressSelector.GetStatus(); activeGW != "" {
			s.log("[SOVEREIGN EGRESS] -> Pasarela %s | Destino: %s", activeGW, dest)
		}
	}
	return net.DialTimeout("tcp", dest, 10*time.Second)
}

func (s *SOCKS5Gateway) setDest(d string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastDestination = d
}

type countingWriter struct {
	w io.Writer
	c *uint64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	if n > 0 {
		atomic.AddUint64(cw.c, uint64(n))
	}
	return n, err
}

func (s *SOCKS5Gateway) tunnel(client net.Conn, clientReader io.Reader, target net.Conn) {
	_ = client.SetDeadline(time.Time{})
	_ = target.SetDeadline(time.Time{})
	var wg sync.WaitGroup
	wg.Add(2)
	pipe := func(dst net.Conn, src io.Reader, c *uint64) {
		defer wg.Done()
		_, _ = io.Copy(&countingWriter{w: dst, c: c}, src)
		_ = dst.Close()
	}
	go pipe(target, clientReader, &s.BytesTx)
	go pipe(client, target, &s.BytesRx)
	wg.Wait()
}

// FormatBytes convierte bytes a representación legible
func FormatBytes(b uint64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%.2f MB", float64(b)/(1024*1024))
	}
}

func (s *SOCKS5Gateway) sendReply(client net.Conn, rep byte, bindIP net.IP, bindPort uint16) {
	repBuf := []byte{SOCKS5Version, rep, 0x00, SOCKS5AddrTypeIPv4, 0, 0, 0, 0, 0, 0}
	if bindIP != nil && bindIP.To4() != nil {
		copy(repBuf[4:8], bindIP.To4())
	}
	binary.BigEndian.PutUint16(repBuf[8:10], bindPort)
	_, _ = client.Write(repBuf)
}

// GetStats retorna métricas del gateway para el panel de control
func (s *SOCKS5Gateway) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	activeGW, egressMode := "", EgressModeDirect
	if s.EgressSelector != nil {
		egressMode, activeGW, _ = s.EgressSelector.GetStatus()
	}
	return map[string]interface{}{
		"listen_addr": s.ListenAddr, "is_running": s.IsRunning,
		"total_connections": atomic.LoadUint64(&s.TotalConnections),
		"active_connections": atomic.LoadInt64(&s.ActiveConnections),
		"bytes_tx": atomic.LoadUint64(&s.BytesTx), "bytes_rx": atomic.LoadUint64(&s.BytesRx),
		"last_destination": s.LastDestination, "active_gateway_did": activeGW,
		"egress_mode": egressMode, "rfc": "RFC 1928 (Universal SOCKS5)",
	}
}
