package l1

import (
	"testing"
)

func TestAdaptiveMTUManager_DegradationAndProbing(t *testing.T) {
	mgr := NewAdaptiveMTUManager()
	peer := "did:ipvn7:field-sensor-node"

	// 1. Por defecto, debe retornar el MTU canónico 1280B
	mtu := mgr.GetEffectiveMTU(peer)
	if mtu != MTUTierCanonical {
		t.Fatalf("expected default MTU %d, got: %d", MTUTierCanonical, mtu)
	}

	// 2. Pérdidas aisladas (1 y 2) no deben degradar de inmediato
	mgr.OnPacketLoss(peer)
	mgr.OnPacketLoss(peer)
	if mgr.GetEffectiveMTU(peer) != MTUTierCanonical {
		t.Fatal("MTU should not degrade before 3 consecutive losses")
	}

	// 3. La 3ra pérdida consecutiva debe degradar a MTUTierConstrain (512B)
	newMTU := mgr.OnPacketLoss(peer)
	if newMTU != MTUTierConstrain {
		t.Fatalf("expected degraded MTU %d, got: %d", MTUTierConstrain, newMTU)
	}

	// 4. Otras 3 pérdidas deben degradar al nivel mínimo LoRa (256B)
	mgr.OnPacketLoss(peer)
	mgr.OnPacketLoss(peer)
	newMTU = mgr.OnPacketLoss(peer)
	if newMTU != MTUTierLoRa {
		t.Fatalf("expected LoRa MTU %d, got: %d", MTUTierLoRa, newMTU)
	}

	// 5. Pérdidas adicionales en el mínimo no deben causar panic ni desborde de índice
	newMTU = mgr.OnPacketLoss(peer)
	if newMTU != MTUTierLoRa {
		t.Fatalf("MTU should remain at minimum %d, got: %d", MTUTierLoRa, newMTU)
	}

	// 6. Sonda exitosa de 1280B restaura el MTU a MTUTierCanonical
	mgr.OnSuccessProbe(peer, 1280)
	if mgr.GetEffectiveMTU(peer) != MTUTierCanonical {
		t.Fatalf("expected restored MTU %d, got: %d", MTUTierCanonical, mgr.GetEffectiveMTU(peer))
	}

	// 7. Sonda exitosa en LAN directa (1420B)
	mgr.OnSuccessProbe(peer, 1420)
	if mgr.GetEffectiveMTU(peer) != MTUTierLAN {
		t.Fatalf("expected LAN MTU %d, got: %d", MTUTierLAN, mgr.GetEffectiveMTU(peer))
	}

	// 8. Reset restaura a 1280B
	mgr.Reset(peer)
	if mgr.GetEffectiveMTU(peer) != MTUTierCanonical {
		t.Fatalf("expected MTU 1280 after reset, got: %d", mgr.GetEffectiveMTU(peer))
	}
}
