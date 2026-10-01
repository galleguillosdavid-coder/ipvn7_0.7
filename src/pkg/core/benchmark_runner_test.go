package core

import (
	"net"
	"testing"
	"time"
)

func TestCalculateRFC3550Jitter(t *testing.T) {
	jitter := 0.0
	// Secuencia de variaciones de tránsito
	deltas := []float64{2.5, 3.0, 1.2, 0.8, 4.1}
	for _, d := range deltas {
		jitter = CalculateRFC3550Jitter(jitter, d)
	}

	if jitter <= 0 {
		t.Fatalf("jitter RFC 3550 esperado > 0, obtenido %f", jitter)
	}
}

func TestRunBenchmarkClient_Loopback(t *testing.T) {
	// Iniciar receptor dummy local
	serverConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("falla al iniciar listener de prueba: %v", err)
	}
	defer serverConn.Close()

	target := serverConn.LocalAddr().String()

	cfg := BenchmarkConfig{
		Target:     target,
		Duration:   200 * time.Millisecond,
		PacketSize: 512,
	}

	res, err := RunBenchmarkClient(cfg)
	if err != nil {
		t.Fatalf("RunBenchmarkClient fallo: %v", err)
	}

	if res.PacketsSent == 0 {
		t.Fatal("se esperaba haber transmitido al menos 1 paquete")
	}
	if res.ThroughputMbps <= 0 {
		t.Fatal("throughput esperado > 0")
	}
	if res.PPS <= 0 {
		t.Fatal("PPS esperado > 0")
	}
}
