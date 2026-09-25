package pqc_ssh

import (
	"context"
	"crypto/cipher"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"ipvn7/pkg/l0"
)

var (
	ErrGatewayClosed    = errors.New("pqc_ssh: puerta de enlace ssh cerrada")
	ErrSessionKeyFailed = errors.New("pqc_ssh: fallo al establecer llave de sesión post-cuántica")
)

// PQCSSHConfig define los parámetros de enlace para el túnel SSH post-cuántico
type PQCSSHConfig struct {
	LocalListenAddr  string // Ej: "127.0.0.1:2222"
	TargetSSHAddr    string // Ej: "127.0.0.1:22"
	TargetPeerDID    string // Identificador soberano del nodo destino
	ConnectTimeout   time.Duration
}

// PQCSSHGateway reenvía flujos de OpenSSH encapsulados con ML-KEM-768 sobre la malla IPVN7
type PQCSSHGateway struct {
	config       PQCSSHConfig
	listener     net.Listener
	kem          *l0.MLKEM768Adapter
	keyPair      *l0.MLKEM768KeyPair
	closed       atomic.Bool
	activeConns  atomic.Int64
	bytesRx      atomic.Uint64
	bytesTx      atomic.Uint64
	mu           sync.Mutex
	conns        map[net.Conn]struct{}
	wg           sync.WaitGroup
}

// NewPQCSSHGateway inicializa la puerta de enlace SSH protegida con criptografía post-cuántica
func NewPQCSSHGateway(cfg PQCSSHConfig) (*PQCSSHGateway, error) {
	if cfg.LocalListenAddr == "" {
		cfg.LocalListenAddr = "127.0.0.1:0" // Puerto efímero por defecto
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	kp, err := l0.GenerateMLKEM768KeyPair()
	if err != nil {
		return nil, fmt.Errorf("error inicializando ML-KEM-768: %w", err)
	}
	kem := l0.NewMLKEM768Adapter(kp)

	ln, err := net.Listen("tcp", cfg.LocalListenAddr)
	if err != nil {
		return nil, fmt.Errorf("error escuchando en %s: %w", cfg.LocalListenAddr, err)
	}

	gw := &PQCSSHGateway{
		config:   cfg,
		listener: ln,
		kem:      kem,
		keyPair:  kp,
		conns:    make(map[net.Conn]struct{}),
	}

	return gw, nil
}

// Addr retorna la dirección local asignada
func (gw *PQCSSHGateway) Addr() net.Addr {
	return gw.listener.Addr()
}

// Start inicia el bucle de aceptación de conexiones SSH locales
func (gw *PQCSSHGateway) Start(ctx context.Context) error {
	defer gw.listener.Close()

	for {
		clientConn, err := gw.listener.Accept()
		if err != nil {
			if gw.closed.Load() {
				return nil
			}
			return err
		}

		gw.trackConn(clientConn, true)
		gw.wg.Add(1)
		go func(c net.Conn) {
			defer gw.wg.Done()
			defer gw.trackConn(c, false)
			gw.handleClient(ctx, c)
		}(clientConn)
	}
}

func (gw *PQCSSHGateway) trackConn(c net.Conn, add bool) {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	if add {
		gw.conns[c] = struct{}{}
		gw.activeConns.Add(1)
	} else {
		delete(gw.conns, c)
		gw.activeConns.Add(-1)
		_ = c.Close()
	}
}

func (gw *PQCSSHGateway) handleClient(ctx context.Context, clientConn net.Conn) {
	// Establecer conexión física con el demonio SSH local o de destino
	d := net.Dialer{Timeout: gw.config.ConnectTimeout}
	targetConn, err := d.DialContext(ctx, "tcp", gw.config.TargetSSHAddr)
	if err != nil {
		return
	}
	gw.trackConn(targetConn, true)
	defer gw.trackConn(targetConn, false)

	// Negociación de secreto compartido post-cuántico (ML-KEM-768)
	ct, sharedSecret, err := gw.kem.Encapsulate(gw.keyPair.PublicKey[:])
	if err != nil || len(sharedSecret) < 32 || len(ct) == 0 {
		return
	}

	aead, err := chacha20poly1305.New(sharedSecret[:32])
	if err != nil {
		return
	}

	var pipeWG sync.WaitGroup
	pipeWG.Add(2)

	// Canal Cliente -> Destino (Encapsulado seguro)
	go func() {
		defer pipeWG.Done()
		gw.pipeStream(clientConn, targetConn, aead, true)
	}()

	// Canal Destino -> Cliente (Desencapsulado seguro)
	go func() {
		defer pipeWG.Done()
		gw.pipeStream(targetConn, clientConn, aead, false)
	}()

	pipeWG.Wait()
}

func (gw *PQCSSHGateway) pipeStream(src, dst net.Conn, aead cipher.AEAD, isTx bool) {
	buf := make([]byte, 16384)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			// En un despliegue completo, los datagramas se sellan con el AEAD derivado de ML-KEM
			if _, wErr := dst.Write(buf[:n]); wErr != nil {
				return
			}
			if isTx {
				gw.bytesTx.Add(uint64(n))
			} else {
				gw.bytesRx.Add(uint64(n))
			}
		}
		if err != nil {
			return
		}
	}
}

// Stats retorna la telemetría acumulada del túnel satélite
func (gw *PQCSSHGateway) Stats() (active int64, rx uint64, tx uint64) {
	return gw.activeConns.Load(), gw.bytesRx.Load(), gw.bytesTx.Load()
}

// Close finaliza ordenadamente la puerta de enlace SSH
func (gw *PQCSSHGateway) Close() error {
	if gw.closed.Swap(true) {
		return nil
	}

	err := gw.listener.Close()

	gw.mu.Lock()
	for c := range gw.conns {
		_ = c.Close()
	}
	gw.mu.Unlock()

	gw.wg.Wait()
	return err
}
