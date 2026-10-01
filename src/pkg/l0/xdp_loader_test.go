package l0

import (
	"runtime"
	"testing"
)

func TestXDPAttachMode_String(t *testing.T) {
	tests := []struct {
		mode     XDPAttachMode
		expected string
	}{
		{XDPModeNative, "xdp-native (NIC driver)"},
		{XDPModeGeneric, "xdp-generic (SKB fallback)"},
		{XDPModeOffload, "xdp-offload (Hardware SmartNIC)"},
		{XDPAttachMode(99), "xdp-unknown (99)"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.mode.String())
		}
	}
}

func TestXDPManager_Lifecycle(t *testing.T) {
	mgr := DefaultXDPManager()
	if mgr == nil {
		t.Fatal("DefaultXDPManager no debe retornar nil")
	}

	cfg := XDPConfig{
		InterfaceName: "eth0",
		ObjectPath:    "xdp_ipvn7.o",
		Mode:          XDPModeNative,
	}

	if runtime.GOOS != "linux" {
		err := mgr.Attach(cfg)
		if err != ErrXDPUnsupportedPlatform {
			t.Errorf("en %s se esperaba ErrXDPUnsupportedPlatform, pero se obtuvo: %v", runtime.GOOS, err)
		}
		if mgr.IsAttached("eth0") {
			t.Error("en plataforma no-Linux IsAttached no debe ser verdadero")
		}
		if err := mgr.Detach("eth0"); err != nil {
			t.Errorf("Detach inesperado fallo: %v", err)
		}
	}

	stats, err := mgr.GetStats()
	if err != nil {
		t.Fatalf("GetStats fallo: %v", err)
	}
	if stats == nil {
		t.Fatal("stats no debe ser nil")
	}
}
