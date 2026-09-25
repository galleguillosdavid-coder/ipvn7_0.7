package l1

import (
	"testing"
)

func TestAdversarialRouteGuard_PoisoningAndDampening(t *testing.T) {
	guard := NewAdversarialRouteGuard()

	srcDID := "did:ipvn7:nodeA"
	targetDID := "did:ipvn7:nodeC"
	legitNextHop := "did:ipvn7:nodeB"

	// 1. Probar intento de bucle o auto-enrutamiento
	valid, err := guard.ValidateRouteProposal(srcDID, targetDID, srcDID, 10.0, 10.0)
	if valid || err != ErrAdversarialLoopDetected {
		t.Fatalf("se esperaba ErrAdversarialLoopDetected, recibido: %v", err)
	}

	// 2. Probar intento de envenenamiento por discrepancia física:
	// La IA alucina 5ms pero la física UDP real mide 150ms
	valid, err = guard.ValidateRouteProposal(srcDID, targetDID, legitNextHop, 5.0, 150.0)
	if valid || err != ErrPhysicalDiscrepancyTooHigh {
		t.Fatalf("se esperaba ErrPhysicalDiscrepancyTooHigh ante discrepancia física, recibido: %v", err)
	}

	// 3. Probar propuesta legítima con RTT físico coherente y progreso métrico
	valid, err = guard.ValidateRouteProposal(srcDID, targetDID, targetDID, 25.0, 24.0)
	if !valid || err != nil {
		t.Fatalf("propuesta legítima debió ser aprobada, err: %v", err)
	}

	// 4. Probar amortiguador de tormentas de señalización (Flap Dampener)
	for i := 0; i < 5; i++ {
		valid, err = guard.ValidateRouteProposal(srcDID, targetDID, targetDID, 20.0, 20.0)
	}
	if valid || err != ErrSignalingStormDampened {
		t.Fatalf("se esperaba ErrSignalingStormDampened ante oscilaciones repetitivas, recibido: %v", err)
	}
}

func TestAdversarialRouteGuard_OneMillionEvaluationsStress(t *testing.T) {
	guard := NewAdversarialRouteGuard()
	srcDID := "did:ipvn7:nodeA"
	targetDID := "did:ipvn7:nodeC"

	rejectedCount := 0
	// 1,000,000 de evaluaciones de ataques de bucle / envenenamiento
	for i := 0; i < 1000000; i++ {
		// Simular intento de auto-enrutamiento cíclico
		valid, _ := guard.ValidateRouteProposal(srcDID, targetDID, srcDID, 10.0, 10.0)
		if !valid {
			rejectedCount++
		}
	}

	if rejectedCount != 1000000 {
		t.Fatalf("el 100%% de los 1,000,000 de intentos de bucle debieron ser rechazados en O(1), rechazados: %d", rejectedCount)
	}
}
