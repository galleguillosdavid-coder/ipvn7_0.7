package devicebridge

import (
	"testing"
)

func TestBlastRadiusGuard_IoTQuarantineIsolation(t *testing.T) {
	rootDID := "did:ipvn7:e93524a43b6487f4d6b0b0881251b5088d06f1ad3c400edd1d26d2eb777e838a"
	guard := NewBlastRadiusGuard(rootDID)

	lightbulbDID := "did:ipvn7:shadow:a4c138010203"
	thermostatDID := "did:ipvn7:shadow:b4c239040506"

	// 1. Simular anomalía extrema en la bombilla IoT (telemetría corrupta)
	quarantined, rootUnaffected := guard.EvaluateTelemetryAnomaly(lightbulbDID, 0.99)
	if !quarantined {
		t.Fatalf("se esperaba que la bombilla fuera puesta en cuarentena")
	}
	if !rootUnaffected {
		t.Fatalf("el nodo residencial nunca debe ser afectado por fallas IoT")
	}

	// 2. Verificar que la bombilla está aislada pero el termostato y el root siguen activos
	if !guard.IsQuarantined(lightbulbDID) {
		t.Fatalf("la bombilla deberia figurar en cuarentena")
	}
	if guard.IsQuarantined(thermostatDID) {
		t.Fatalf("el termostato no debe ser penalizado por la bombilla")
	}
	if guard.IsQuarantined(rootDID) {
		t.Fatalf("el nodo residencial jamas puede estar en cuarentena")
	}

	// 3. Intentar aislar al nodo raíz directamente: debe ser rechazado
	_, err := guard.QuarantineDevice(rootDID, "Ataque o anomalia simulada")
	if err != ErrRootImmuneToIoTQuarantine {
		t.Fatalf("se esperaba ErrRootImmuneToIoTQuarantine, recibido: %v", err)
	}

	// 4. Liberar la bombilla tras corrección
	guard.ReleaseDevice(lightbulbDID)
	if guard.IsQuarantined(lightbulbDID) {
		t.Fatalf("la bombilla deberia estar liberada de la cuarentena")
	}
}
