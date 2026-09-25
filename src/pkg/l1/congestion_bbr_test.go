package l1

import (
	"testing"
	"time"
)

func TestBBRController_StartupAndConvergence(t *testing.T) {
	// Inicializar con 10 MB/s y 10ms RTT -> BDP = 100,000 bytes
	bbr := NewBBRController(10000000, 10*time.Millisecond)

	state, btlBw, rtProp, pacingRate, cwnd := bbr.Metrics()
	if state != BBRStateStartup {
		t.Fatalf("expected initial state BBRStateStartup, got: %s", state)
	}
	if btlBw != 10000000 {
		t.Fatalf("expected btlBw 10MB/s, got: %.0f", btlBw)
	}
	if rtProp != 10*time.Millisecond {
		t.Fatalf("expected rtProp 10ms, got: %v", rtProp)
	}
	if cwnd < 5120 {
		t.Fatalf("expected minimum cwnd >= 5120 bytes, got: %d", cwnd)
	}
	if pacingRate <= btlBw {
		t.Fatalf("pacingRate in startup should be accelerated (2.885x), got: %.0f", pacingRate)
	}

	// Verificar permiso de envío CanSend
	if !bbr.CanSend(cwnd - 1) {
		t.Fatal("expected CanSend=true when inflight < cwnd")
	}
	if bbr.CanSend(cwnd) {
		t.Fatal("expected CanSend=false when inflight == cwnd")
	}

	// Simular llegada de Acks con mayor ancho de banda y menor RTT
	bbr.OnAck(20000000, 8*time.Millisecond)
	_, newBtlBw, newRtProp, _, newCwnd := bbr.Metrics()
	if newBtlBw != 20000000 {
		t.Fatalf("expected btlBw updated to 20MB/s, got: %.0f", newBtlBw)
	}
	if newRtProp != 8*time.Millisecond {
		t.Fatalf("expected rtProp updated to 8ms, got: %v", newRtProp)
	}
	if newCwnd <= cwnd {
		t.Fatalf("expected cwnd to expand with higher bandwidth, got: %d vs %d", newCwnd, cwnd)
	}

	// Forzar avance temporal para transitar de STARTUP a DRAIN
	bbr.lastCycle = time.Now().Add(-600 * time.Millisecond)
	bbr.OnAck(20000000, 8*time.Millisecond)
	state, _, _, pacingDrain, _ := bbr.Metrics()
	if state != BBRStateDrain {
		t.Fatalf("expected transition to BBRStateDrain, got: %s", state)
	}
	if pacingDrain >= newBtlBw {
		t.Fatalf("pacing in DRAIN should be slowed down (< 1.0x), got: %.0f vs %.0f", pacingDrain, newBtlBw)
	}

	// Forzar avance temporal para transitar de DRAIN a PROBE_BW
	bbr.lastCycle = time.Now().Add(-400 * time.Millisecond)
	bbr.OnAck(20000000, 8*time.Millisecond)
	state, _, _, _, _ = bbr.Metrics()
	if state != BBRStateProbeBW {
		t.Fatalf("expected transition to BBRStateProbeBW, got: %s", state)
	}
}
