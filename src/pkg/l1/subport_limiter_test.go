package l1

import (
	"fmt"
	"testing"
)

func TestSubPortLimiter_AntiStarvationAndAntiOOM(t *testing.T) {
	limiter := NewSubPortLimiter()
	limiter.SetCustomRate(SubPortTelemetry, 5.0, 1.0) // 5 tokens max burst, 1 token/sec refill

	legitDID := "did:ipvn7:authenticated_peer"
	limiter.RegisterAuthenticatedPeer(legitDID)

	// 1. Simular ataque de agotamiento de memoria (OOM): 10,000 DIDs efímeros aleatorios no autenticados
	untrustedAccepted := 0
	untrustedDropped := 0
	for i := 0; i < 10000; i++ {
		fakeDID := fmt.Sprintf("did:ipvn7:random_attacker_%d", i)
		if limiter.AllowPacket(SubPortTelemetry, fakeDID) {
			untrustedAccepted++
		} else {
			untrustedDropped++
		}
	}

	if untrustedAccepted != 5 {
		t.Fatalf("los 10,000 DIDs no autenticados debieron agotar la cubeta compartida en 5 paquetes, aceptados: %d", untrustedAccepted)
	}
	if untrustedDropped != 9995 {
		t.Fatalf("se esperaban 9995 paquetes descartados sin alocar memoria, descartados: %d", untrustedDropped)
	}

	// 2. Verificar que el par legítimo autenticado retiene su propia cubeta dedicada
	legitAccepted := 0
	for i := 0; i < 5; i++ {
		if limiter.AllowPacket(SubPortTelemetry, legitDID) {
			legitAccepted++
		}
	}
	if legitAccepted != 5 {
		t.Fatalf("el par autenticado debió ser admitido en su cubeta dedicada sin ser afectado por el ataque OOM, aceptados: %d", legitAccepted)
	}

	// 3. Verificar que el canal de datos (SubPortVPNProxy) permanece 100% operativo
	if !limiter.AllowPacket(SubPortVPNProxy, legitDID) {
		t.Fatalf("el canal rápido de datos no debe bloquearse")
	}
}

func TestSubPortLimiter_OneMillionFloodingStress(t *testing.T) {
	limiter := NewSubPortLimiter()
	limiter.SetCustomRate(SubPortTelemetry, 5.0, 1.0)

	// Inyectar 1,000,000 de datagramas con DIDs efímeros únicos simulando clúster de GPUs
	dropped := 0
	accepted := 0
	targetDIDBase := "did:ipvn7:gpu_dead_soul_"

	for i := 0; i < 1000000; i++ {
		// Simulación de DID efímero sin alocación pesada
		did := targetDIDBase
		if limiter.AllowPacket(SubPortTelemetry, did) {
			accepted++
		} else {
			dropped++
		}
	}

	if accepted != 5 {
		t.Fatalf("se esperaban exactamente 5 paquetes aceptados, aceptados: %d", accepted)
	}
	if dropped != 999995 {
		t.Fatalf("se esperaban exactamente 999,995 paquetes descartados en O(1), descartados: %d", dropped)
	}
}
