package l1

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestEgressForwarder_Authorization(t *testing.T) {
	f := NewEgressForwarder(false)
	didA := "did:ipvn7:nodo_amigo"
	didB := "did:ipvn7:desconocido"

	// 1. Deshabilitado por defecto
	if f.IsAuthorized(didA) {
		t.Fatal("No debió autorizar con salida deshabilitada")
	}

	// 2. Habilitar y autorizar didA
	f.SetAllowEgress(true)
	f.AuthorizeDID(didA)

	if !f.IsAuthorized(didA) {
		t.Fatal("Debió autorizar didA")
	}
	if f.IsAuthorized(didB) {
		t.Fatal("No debió autorizar didB (ZTNA default-deny)")
	}

	// 3. Revocar didA
	f.RevokeDID(didA)
	if f.IsAuthorized(didA) {
		t.Fatal("No debió autorizar didA tras revocación")
	}
}

func TestEgressForwarder_ForwardStream(t *testing.T) {
	// Servidor TCP local mock
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("falla al iniciar listener: %v", err)
	}
	defer ln.Close()

	serverEcho := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 128)
		n, _ := conn.Read(buf)
		msg := string(buf[:n])
		conn.Write([]byte("PONG:" + msg))
		serverEcho <- msg
	}()

	f := NewEgressForwarder(true)
	didA := "did:ipvn7:nodo_test"
	f.AuthorizeDID(didA)

	// Crear conexión virtual bidireccional usando net.Pipe
	clientSide, forwarderSide := net.Pipe()

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- f.ForwardStream(didA, ln.Addr().String(), forwarderSide)
	}()

	// Enviar desde el cliente
	_, err = clientSide.Write([]byte("PING_PAYLOAD"))
	if err != nil {
		t.Fatalf("falla al escribir cliente: %v", err)
	}

	// Leer respuesta en el cliente
	respBuf := make([]byte, 128)
	n, err := clientSide.Read(respBuf)
	if err != nil {
		t.Fatalf("falla al leer cliente: %v", err)
	}
	clientSide.Close()

	if !bytes.Contains(respBuf[:n], []byte("PONG:PING_PAYLOAD")) {
		t.Errorf("Cliente recibió respuesta incorrecta: %s", string(respBuf[:n]))
	}

	select {
	case msg := <-serverEcho:
		if msg != "PING_PAYLOAD" {
			t.Errorf("Servidor recibió mensaje incorrecto: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout esperando eco del servidor")
	}

	<-doneCh

	sessions, tx, rx, enabled := f.GetStats()
	if !enabled || sessions != 0 || tx == 0 || rx == 0 {
		t.Errorf("Métricas incoherentes: sessions=%d, tx=%d, rx=%d, enabled=%v", sessions, tx, rx, enabled)
	}
}
