package core

import (
	"testing"
)

func TestOperatingModeController_Lifecycle(t *testing.T) {
	ctrl := NewOperatingModeController(ModeOpen)

	if ctrl.CurrentMode() != ModeOpen {
		t.Fatalf("Esperado ModeOpen inicial, obtenido %s", ctrl.CurrentMode())
	}
	if !ctrl.AllowsDiscovery() {
		t.Fatal("ModeOpen debe permitir discovery")
	}
	if !ctrl.AllowsUnpinnedPeers() {
		t.Fatal("ModeOpen debe permitir unpinned peers")
	}

	// Transición a ModeSafe
	err := ctrl.SetMode(ModeSafe, "aislamiento manual dark node")
	if err != nil {
		t.Fatalf("SetMode falló: %v", err)
	}
	if ctrl.CurrentMode() != ModeSafe {
		t.Fatalf("Esperado ModeSafe, obtenido %s", ctrl.CurrentMode())
	}
	if ctrl.AllowsDiscovery() {
		t.Fatal("ModeSafe debe prohibir discovery")
	}
	if ctrl.AllowsUnpinnedPeers() {
		t.Fatal("ModeSafe debe prohibir unpinned peers")
	}

	// Transición a ModeDegraded
	err = ctrl.SetMode(ModeDegraded, "congestión severa >30% pérdida")
	if err != nil {
		t.Fatalf("SetMode falló: %v", err)
	}
	if !ctrl.IsDegraded() {
		t.Fatal("Esperado IsDegraded == true")
	}

	// Verificar histórico
	history := ctrl.History()
	if len(history) != 2 {
		t.Fatalf("Esperadas 2 transiciones en historial, obtenidas %d", len(history))
	}
	if history[0].FromMode != ModeOpen || history[0].ToMode != ModeSafe {
		t.Errorf("Transición 0 inesperada: %+v", history[0])
	}
	if history[1].FromMode != ModeSafe || history[1].ToMode != ModeDegraded {
		t.Errorf("Transición 1 inesperada: %+v", history[1])
	}

	// Validación de modo inválido
	err = ctrl.SetMode("MODO_INVENTADO", "prueba error")
	if err == nil {
		t.Fatal("Se esperaba error al configurar modo inválido")
	}
}

func TestValidateModeString(t *testing.T) {
	cases := []struct {
		input    string
		expected NodeOperatingMode
		wantErr  bool
	}{
		{"MODE_OPEN", ModeOpen, false},
		{"open", ModeOpen, false},
		{"SAFE", ModeSafe, false},
		{"degraded", ModeDegraded, false},
		{"invalido", "", true},
	}

	for _, tc := range cases {
		m, err := ValidateModeString(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateModeString(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
		}
		if !tc.wantErr && m != tc.expected {
			t.Errorf("ValidateModeString(%q) = %v, expected %v", tc.input, m, tc.expected)
		}
	}
}
