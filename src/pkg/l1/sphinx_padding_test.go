package l1_test

import (
	"sync/atomic"
	"testing"
	"time"

	"ipvn7/pkg/l1"
)

func TestStochasticPaddingEngine_LifecycleAndFrameSize(t *testing.T) {
	keys, err := l1.GenerateHybridKeyPair("did:ipvn7:sphinx-padding-node")
	if err != nil {
		t.Fatalf("Error generando claves: %v", err)
	}

	router := l1.NewSphinxRouter(keys)
	circuit, err := router.BuildDynamicCircuit(nil, nil)
	if err != nil {
		t.Fatalf("Error creando circuito dinámico: %v", err)
	}

	engine := l1.NewStochasticPaddingEngine(router, 10*time.Millisecond, 30*time.Millisecond)

	// 1. Generar paquete individual y verificar tamaño exacto wire 1280B
	dummyPkt, err := engine.GenerateDummyPacket(circuit)
	if err != nil {
		t.Fatalf("Error generando dummy packet: %v", err)
	}

	wire := dummyPkt.Serialize()
	if len(wire) != l1.SphinxPacketSize {
		t.Fatalf("Tamaño wire de dummy (%d) no coincide con 1280 bytes deterministas", len(wire))
	}

	// 2. Probar ejecución del motor estocástico
	var sentCount int32
	engine.Start(circuit, func(w []byte) error {
		if len(w) == l1.SphinxPacketSize {
			atomic.AddInt32(&sentCount, 1)
		}
		return nil
	})

	if !engine.IsRunning() {
		t.Errorf("Engine debió reportar estado activo")
	}

	time.Sleep(100 * time.Millisecond)
	engine.Stop()

	if engine.IsRunning() {
		t.Errorf("Engine debió reportar estado inactivo tras Stop")
	}

	sent := atomic.LoadInt32(&sentCount)
	if sent < 2 {
		t.Errorf("Se esperaban al menos 2 paquetes dummy enviados en 100ms (enviados: %d)", sent)
	}
}

func TestStochasticPaddingEngine_EntropyFloorUnderCongestion(t *testing.T) {
	keys, _ := l1.GenerateHybridKeyPair("did:ipvn7:sphinx-floor-test")
	router := l1.NewSphinxRouter(keys)
	engine := l1.NewStochasticPaddingEngine(router, 10*time.Millisecond, 50*time.Millisecond)

	// Bajo congestión severa WAN (BBR throttling activo), el intervalo preserva proceso de renovación estocástico
	floor := engine.EnforceEntropyFloor(true)
	if floor < l1.MinEntropyFloorBase || floor > l1.MinEntropyFloorBase+l1.MinEntropyFloorJitter {
		t.Fatalf("Piso de entropía estocástico fuera de rango [%v, %v]: %v",
			l1.MinEntropyFloorBase, l1.MinEntropyFloorBase+l1.MinEntropyFloorJitter, floor)
	}

	// Sin congestión, retorna intervalo estocástico dinámico
	normal := engine.EnforceEntropyFloor(false)
	if normal < 10*time.Millisecond || normal > 50*time.Millisecond {
		t.Fatalf("Intervalo normal fuera de límites: %v", normal)
	}

	// Cuantización fija inmutable de 1280B
	if l1.QuantizedCanonicalMTU != 1280 {
		t.Fatalf("MTU canónico cuantizado debe ser 1280B")
	}
}

func BenchmarkStochasticPaddingEngine_Generate(b *testing.B) {
	keys, _ := l1.GenerateHybridKeyPair("did:ipvn7:bench-padding")
	router := l1.NewSphinxRouter(keys)
	circuit, _ := router.BuildDynamicCircuit(nil, nil)
	engine := l1.NewStochasticPaddingEngine(router, 50*time.Millisecond, 200*time.Millisecond)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GenerateDummyPacket(circuit)
	}
}
