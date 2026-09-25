package l1

import (
	"bytes"
	"testing"
	"time"
)

func TestKeyRotationSentinel_LifecycleAndThresholds(t *testing.T) {
	sessionID := "sess-ipvn7-test-node-b"
	sentinel, err := NewKeyRotationSentinel(sessionID)
	if err != nil {
		t.Fatalf("NewKeyRotationSentinel failed: %v", err)
	}

	keys := sentinel.GetKeys()
	if keys.Epoch != 1 {
		t.Fatalf("expected initial epoch 1, got: %d", keys.Epoch)
	}
	if keys.HasPrevious {
		t.Fatal("expected no previous key in initial epoch")
	}

	initialKey := keys.ActiveKey

	// 1. Configurar límites reducidos para pruebas: 1000 bytes, 5 paquetes, 1 hora
	sentinel.SetLimits(1000, 5, 1*time.Hour)

	// Enviar 4 paquetes de 100 bytes (total 400B) -> no debe requerir rotación aún
	for i := 0; i < 4; i++ {
		if sentinel.TrackTraffic(100) {
			t.Fatal("should not trigger rotation at 400 bytes / 4 packets")
		}
	}

	// 5to paquete -> debe activar rotación por límite de paquetes (5 paquetes)
	if !sentinel.TrackTraffic(100) {
		t.Fatal("expected rotation required after 5 packets")
	}

	// 2. Ejecutar rotación
	newEpoch, err := sentinel.Rotate()
	if err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}
	if newEpoch != 2 {
		t.Fatalf("expected epoch 2 after rotation, got: %d", newEpoch)
	}

	keysAfter := sentinel.GetKeys()
	if keysAfter.Epoch != 2 {
		t.Fatalf("expected epoch 2, got: %d", keysAfter.Epoch)
	}
	if !keysAfter.HasPrevious {
		t.Fatal("expected HasPrevious=true after rotation")
	}
	if !bytes.Equal(keysAfter.PreviousKey[:], initialKey[:]) {
		t.Fatal("previous key should match initial active key")
	}
	if bytes.Equal(keysAfter.ActiveKey[:], initialKey[:]) {
		t.Fatal("new active key must be distinct from initial key")
	}

	// 3. Probar rotación por volumen de bytes (límite 500B, 100 paquetes)
	sentinel.SetLimits(500, 100, 1*time.Hour)
	if sentinel.TrackTraffic(499) {
		t.Fatal("should not trigger at 499 bytes")
	}
	if !sentinel.TrackTraffic(1) {
		t.Fatal("expected rotation required at 500 bytes")
	}

	newEpoch, _ = sentinel.Rotate()
	if newEpoch != 3 {
		t.Fatalf("expected epoch 3, got: %d", newEpoch)
	}
}
