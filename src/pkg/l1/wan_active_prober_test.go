package l1

import (
	"context"
	"net"
	"testing"
	"time"

	"ipvn7/pkg/l0"
)

func TestWANActiveProber_LifecycleAndEcho(t *testing.T) {
	idA, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("failed to generate identity A: %v", err)
	}

	router := NewKleinbergRouter(idA)
	healing := NewLinkHealingEngine(router)

	// Crear dos sockets UDP reales en loopback
	connA, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen UDP A: %v", err)
	}
	defer connA.Close()

	connB, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen UDP B: %v", err)
	}
	defer connB.Close()

	prober := NewWANActiveProber(router, healing, connA)
	prober.SetFrequency(50 * time.Millisecond)

	idB, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("failed to generate identity B: %v", err)
	}
	peerDID := idB.DID()
	endpointB := connB.LocalAddr().String()

	// Receptor B devuelve eco de cualquier sondeo entrante
	stopEcho := make(chan struct{})
	go func() {
		buf := make([]byte, 128)
		for {
			select {
			case <-stopEcho:
				return
			default:
			}
			_ = connB.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, rAddr, err := connB.ReadFrom(buf)
			if err != nil {
				continue
			}
			if n >= ProbeWireSize {
				_, _ = connB.WriteTo(buf[:n], rAddr)
			}
		}
	}()
	defer close(stopEcho)

	// Registrar par en el router
	udpAddrB, err := net.ResolveUDPAddr("udp", endpointB)
	if err != nil {
		t.Fatalf("failed to resolve endpoint B: %v", err)
	}
	_ = router.AddOrUpdatePeer(peerDID, udpAddrB, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prober.Start(ctx)
	defer prober.Stop()

	// Emitir sondeo directo
	prober.ProbeSinglePeer(peerDID, endpointB)

	// Bucle de lectura en connA para procesar la respuesta
	bufA := make([]byte, 128)
	_ = connA.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	n, rAddr, err := connA.ReadFrom(bufA)
	if err != nil {
		t.Fatalf("did not receive echo response: %v", err)
	}

	handled := prober.HandleIncomingProbe(bufA[:n], rAddr, true)
	if !handled {
		t.Errorf("HandleIncomingProbe returned false for valid echo")
	}

	stats, found := prober.GetPeerStats(peerDID)
	if !found {
		t.Fatalf("peer stats not found for %s", peerDID)
	}

	if stats.ProbesSent == 0 {
		t.Errorf("expected ProbesSent > 0, got %d", stats.ProbesSent)
	}
	if stats.ProbesRecv == 0 {
		t.Errorf("expected ProbesRecv > 0, got %d", stats.ProbesRecv)
	}
	if stats.LastRTTMs < 0 {
		t.Errorf("invalid RTT: %f", stats.LastRTTMs)
	}
}

func TestWANActiveProber_LossAndHealing(t *testing.T) {
	idA, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("failed to generate identity A: %v", err)
	}

	router := NewKleinbergRouter(idA)
	healing := NewLinkHealingEngine(router)

	connA, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen UDP: %v", err)
	}
	defer connA.Close()

	prober := NewWANActiveProber(router, healing, connA)

	targetDID := "did:ipvn7:dead_peer"
	invalidEndpoint := "127.0.0.1:9999"

	// Registrar 3 pérdidas consecutivas
	for i := 0; i < 3; i++ {
		prober.RecordLoss(targetDID, invalidEndpoint)
	}

	stats, found := prober.GetPeerStats(targetDID)
	if !found {
		t.Fatalf("peer stats not found for %s", targetDID)
	}

	if stats.ConsecLosses < 3 {
		t.Errorf("expected at least 3 consecutive losses, got %d", stats.ConsecLosses)
	}
}
