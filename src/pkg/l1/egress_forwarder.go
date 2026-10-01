// Package l1 implementa el reenviador de salida (Egress Forwarder) en el nodo pasarela.
package l1

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Errores canónicos de pasarela de salida
var (
	ErrEgressDisabled  = errors.New("salida a internet deshabilitada en este nodo")
	ErrEgressForbidden = errors.New("did no autorizado para usar esta pasarela (ZTNA default-deny)")
)

// EgressForwarder gestiona las conexiones salientes hacia el Internet público
type EgressForwarder struct {
	mu             sync.RWMutex
	allowEgress    bool
	allowedDIDs    map[string]bool
	allowAnyPeer   bool
	activeSessions int64
	bytesTx        uint64
	bytesRx        uint64
	dialTimeout    time.Duration
}

// NewEgressForwarder inicializa el reenviador con política ZTNA Default-Deny
func NewEgressForwarder(allowEgress bool) *EgressForwarder {
	return &EgressForwarder{
		allowEgress:  allowEgress,
		allowedDIDs:  make(map[string]bool),
		dialTimeout:  8 * time.Second,
	}
}

// SetAllowEgress activa o desactiva la función de pasarela en este nodo
func (f *EgressForwarder) SetAllowEgress(allowed bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allowEgress = allowed
}

// SetAllowAnyPeer define si cualquier par de la malla puede usar el gateway
func (f *EgressForwarder) SetAllowAnyPeer(allow bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allowAnyPeer = allow
}

// AuthorizeDID añade un DID a la lista blanca de clientes autorizados
func (f *EgressForwarder) AuthorizeDID(did string) {
	if did == "" {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allowedDIDs[did] = true
}

// RevokeDID revoca el acceso de un DID a esta pasarela
func (f *EgressForwarder) RevokeDID(did string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.allowedDIDs, did)
}

// IsAuthorized comprueba si un DID emisor tiene permiso de salida según política ZTNA
func (f *EgressForwarder) IsAuthorized(did string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if !f.allowEgress {
		return false
	}
	if f.allowAnyPeer {
		return true
	}
	return f.allowedDIDs[did]
}

// GetStats retorna las métricas operativas del reenviador
func (f *EgressForwarder) GetStats() (sessions int64, tx uint64, rx uint64, enabled bool) {
	f.mu.RLock()
	enabled = f.allowEgress
	f.mu.RUnlock()
	return atomic.LoadInt64(&f.activeSessions), atomic.LoadUint64(&f.bytesTx), atomic.LoadUint64(&f.bytesRx), enabled
}

// ForwardStream reenvía un flujo bidireccional desde un cliente hacia el destino en Internet
func (f *EgressForwarder) ForwardStream(clientDID string, targetAddr string, clientStream io.ReadWriter) error {
	if !f.IsAuthorized(clientDID) {
		return fmt.Errorf("%w: %s", ErrEgressForbidden, clientDID)
	}

	outConn, err := net.DialTimeout("tcp", targetAddr, f.dialTimeout)
	if err != nil {
		return fmt.Errorf("falla al conectar con destino externo %s: %w", targetAddr, err)
	}
	defer outConn.Close()

	atomic.AddInt64(&f.activeSessions, 1)
	defer atomic.AddInt64(&f.activeSessions, -1)

	var wg sync.WaitGroup
	wg.Add(2)

	// Flujo de subida: Cliente -> Internet
	go func() {
		defer wg.Done()
		n, _ := io.Copy(outConn, clientStream)
		atomic.AddUint64(&f.bytesTx, uint64(n))
		if tcpConn, ok := outConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	// Flujo de bajada: Internet -> Cliente
	go func() {
		defer wg.Done()
		n, _ := io.Copy(clientStream, outConn)
		atomic.AddUint64(&f.bytesRx, uint64(n))
	}()

	wg.Wait()
	return nil
}
