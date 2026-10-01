package l1_test

import (
	"net"
	"sync"
	"testing"
	"time"

	"ipvn7/pkg/l1"
)

func TestNATHolePunchMessageSerialization(t *testing.T) {
	msg := &l1.HolePunchMessage{
		Magic:     l1.HolePunchMagic,
		OpCode:    l1.HolePunchProbeOp,
		SessionID: [16]byte{0x01, 0x02, 0x03, 0x04},
		Sequence:  42,
	}

	wire := msg.Encode()
	if len(wire) != 23 {
		t.Fatalf("Tamaño wire incorrecto: %d bytes (esperado 23)", len(wire))
	}

	decoded, err := l1.DecodeHolePunchMessage(wire)
	if err != nil {
		t.Fatalf("DecodeHolePunchMessage falló: %v", err)
	}

	if decoded.Magic != msg.Magic || decoded.OpCode != msg.OpCode || decoded.Sequence != msg.Sequence {
		t.Errorf("Campos decodificados no coinciden")
	}

	if decoded.SessionID != msg.SessionID {
		t.Errorf("SessionID decodificado no coincide")
	}
}

func TestNATHolePunchEngine_SimultaneousPunch(t *testing.T) {
	// Crear dos sockets UDP locales simulando Nodo A y Nodo B detrás de NAT
	addrA, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	connA, err := net.ListenUDP("udp", addrA)
	if err != nil {
		t.Fatalf("Error abriendo connA: %v", err)
	}
	defer connA.Close()

	addrB, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	connB, err := net.ListenUDP("udp", addrB)
	if err != nil {
		t.Fatalf("Error abriendo connB: %v", err)
	}
	defer connB.Close()

	engineA := l1.NewNATHolePunchEngine(connA)
	engineB := l1.NewNATHolePunchEngine(connB)

	// Bucle de lectura de connA
	stopLoop := make(chan struct{})
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-stopLoop:
				return
			default:
				_ = connA.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
				n, from, err := connA.ReadFromUDP(buf)
				if err == nil {
					engineA.HandleIncomingPacket(buf[:n], from)
				}
			}
		}
	}()

	// Bucle de lectura de connB
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-stopLoop:
				return
			default:
				_ = connB.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
				n, from, err := connB.ReadFromUDP(buf)
				if err == nil {
					engineB.HandleIncomingPacket(buf[:n], from)
				}
			}
		}
	}()

	// Perforación simultánea (ICE-Lite)
	var wg sync.WaitGroup
	wg.Add(2)

	var resA, resB *net.UDPAddr
	var errA, errB error

	go func() {
		defer wg.Done()
		resA, errA = engineA.ExecutePunch(connB.LocalAddr().(*net.UDPAddr), 1*time.Second)
	}()

	go func() {
		defer wg.Done()
		resB, errB = engineB.ExecutePunch(connA.LocalAddr().(*net.UDPAddr), 1*time.Second)
	}()

	wg.Wait()
	close(stopLoop)

	if errA != nil {
		t.Fatalf("Perforación en Nodo A falló: %v", errA)
	}
	if errB != nil {
		t.Fatalf("Perforación en Nodo B falló: %v", errB)
	}

	if resA.Port != connB.LocalAddr().(*net.UDPAddr).Port {
		t.Errorf("Puerto perforado en A no coincide con B")
	}
	if resB.Port != connA.LocalAddr().(*net.UDPAddr).Port {
		t.Errorf("Puerto perforado en B no coincide con A")
	}
}

func BenchmarkHolePunchMessage_EncodeDecode(b *testing.B) {
	msg := &l1.HolePunchMessage{
		Magic:     l1.HolePunchMagic,
		OpCode:    l1.HolePunchProbeOp,
		SessionID: [16]byte{1, 2, 3, 4},
		Sequence:  100,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		wire := msg.Encode()
		_, _ = l1.DecodeHolePunchMessage(wire)
	}
}
