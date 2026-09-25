package pqc_ssh

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPQCSSHGateway_LifecycleAndTunnel(t *testing.T) {
	// 1. Iniciar un servidor echo TCP de destino emulando el demonio SSH
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error creando servidor destino: %v", err)
	}
	defer targetLn.Close()

	go func() {
		for {
			conn, aErr := targetLn.Accept()
			if aErr != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c) // Echo real
			}(conn)
		}
	}()

	// 2. Inicializar la puerta de enlace SSH PQC
	cfg := PQCSSHConfig{
		LocalListenAddr: "127.0.0.1:0",
		TargetSSHAddr:   targetLn.Addr().String(),
		TargetPeerDID:   "did:ipvn7:test-node-b",
		ConnectTimeout:  2 * time.Second,
	}

	gw, err := NewPQCSSHGateway(cfg)
	if err != nil {
		t.Fatalf("error inicializando gateway: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = gw.Start(ctx)
	}()

	// 3. Conectar cliente SSH a la puerta de enlace local
	clientConn, err := net.DialTimeout("tcp", gw.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatalf("error conectando al gateway ssh: %v", err)
	}
	defer clientConn.Close()

	testPayload := []byte("SSH-2.0-OpenSSH_9.2p1 IPVN7-PQC-Protected\r\n")
	if _, err := clientConn.Write(testPayload); err != nil {
		t.Fatalf("error escribiendo en el tunel: %v", err)
	}

	buf := make([]byte, len(testPayload))
	_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := io.ReadFull(clientConn, buf)
	if err != nil {
		t.Fatalf("error leyendo respuesta por el tunel: %v", err)
	}

	if string(buf[:n]) != string(testPayload) {
		t.Errorf("datos recibidos no coinciden: got %q, want %q", string(buf[:n]), string(testPayload))
	}

	// 4. Comprobar telemetría acumulada
	_, rx, tx := gw.Stats()
	if tx == 0 || rx == 0 {
		t.Errorf("telemetría esperada no nula: tx=%d, rx=%d", tx, rx)
	}

	// 5. Cierre ordenado
	if err := gw.Close(); err != nil {
		t.Errorf("error cerrando gateway: %v", err)
	}
}
