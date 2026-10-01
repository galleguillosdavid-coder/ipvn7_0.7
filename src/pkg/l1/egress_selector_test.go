package l1

import (
	"testing"
	"time"
)

func TestEgressSelector_AutoSelectionAndHysteresis(t *testing.T) {
	reg := NewEgressRegistry()
	didA := "did:ipvn7:nodo_a"
	didB := "did:ipvn7:nodo_b"

	reg.Set(didA, &EgressCapability{CanExit: true, ActiveSessions: 2})
	reg.Set(didB, &EgressCapability{CanExit: true, ActiveSessions: 10})

	selector := NewEgressSelector(reg)
	now := time.Now()

	peers := []*PeerNode{
		{
			DID:         didA,
			HealthState: HealthStateHealthy,
			Locator:     PeerLocator{LatencyMs: 10.0},
			LossRate:    0.0,
		},
		{
			DID:         didB,
			HealthState: HealthStateHealthy,
			Locator:     PeerLocator{LatencyMs: 30.0},
			LossRate:    0.02,
		},
	}

	// 1. Debe seleccionar Nodo A (mejor score)
	selected := selector.EvaluatePeers(peers, now)
	if selected != didA {
		t.Fatalf("Esperaba selección de %s, obtuvo: %s", didA, selected)
	}

	// 2. Simular que Nodo B mejora solo un poco (latencia 9ms vs 10ms de A).
	// No debe conmutar de inmediato debido a la histéresis del 25%.
	peers[1].Locator.LatencyMs = 9.0
	peers[1].LossRate = 0.0
	peers[1].HealthState = HealthStateHealthy
	now = now.Add(1 * time.Second)

	selected = selector.EvaluatePeers(peers, now)
	if selected != didA {
		t.Fatalf("Histéresis rota: no debió conmutar a B por una mejora menor al 25%%. Seleccionó: %s", selected)
	}

	// 3. Simular que Nodo A sufre degradación crítica (Unreachable)
	peers[0].HealthState = HealthStateUnreachable
	now = now.Add(2 * time.Second)

	selected = selector.EvaluatePeers(peers, now)
	if selected != didB {
		t.Fatalf("Esperaba conmutación a %s ante fallo de A, obtuvo: %s", didB, selected)
	}

	// 4. Modo directo: debe retornar vacío (enlace local)
	selector.SetMode(EgressModeDirect, "")
	selected = selector.EvaluatePeers(peers, now)
	if selected != "" {
		t.Fatalf("En EgressModeDirect debía retornar cadena vacía, obtuvo: %s", selected)
	}
}

func TestEgressSelector_SpecificMode(t *testing.T) {
	reg := NewEgressRegistry()
	didTarget := "did:ipvn7:nodo_especifico"

	selector := NewEgressSelector(reg)
	selector.SetMode(EgressModeSpecific, didTarget)
	now := time.Now()

	peers := []*PeerNode{
		{
			DID:         didTarget,
			HealthState: HealthStateHealthy,
			Locator:     PeerLocator{LatencyMs: 15.0},
		},
	}

	// Nodo específico saludable
	selected := selector.EvaluatePeers(peers, now)
	if selected != didTarget {
		t.Fatalf("Esperaba %s en modo específico, obtuvo: %s", didTarget, selected)
	}

	// Si nodo específico cae, fallback a directo (cadena vacía)
	peers[0].HealthState = HealthStateUnreachable
	selected = selector.EvaluatePeers(peers, now)
	if selected != "" {
		t.Fatalf("Esperaba fallback vacío ante caída de nodo específico, obtuvo: %s", selected)
	}
}
