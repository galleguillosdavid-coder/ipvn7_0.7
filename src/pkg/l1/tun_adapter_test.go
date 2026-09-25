package l1

import (
	"bytes"
	"testing"

	"ipvn7/pkg/l0"
)

func TestUserspaceVirtualAdapter_Operations(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando id: %v", err)
	}

	adapter := NewUserspaceVirtualAdapter(id)
	defer adapter.Close()

	if adapter.MTU() != DefaultMTU {
		t.Errorf("MTU inesperado: %d != %d", adapter.MTU(), DefaultMTU)
	}

	if adapter.IPv4() == nil || adapter.IPv6() == nil {
		t.Fatal("las IPs virtuales asignadas al adaptador no deben ser nulas")
	}

	if !adapter.IsUserspace() {
		t.Errorf("IsUserspace debe ser verdadero")
	}

	if adapter.Mode() != ModeUserspaceVirtual {
		t.Errorf("modo inesperado: %s", adapter.Mode())
	}

	testData := []byte("ipvn7_virtual_packet_payload")
	if err := adapter.WritePacket(testData); err != nil {
		t.Fatalf("WritePacket falló: %v", err)
	}

	// Probar inyección
	if err := adapter.InjectPacket(testData); err != nil {
		t.Fatalf("InjectPacket falló: %v", err)
	}

	readPkt, err := adapter.ReadPacket()
	if err != nil {
		t.Fatalf("ReadPacket falló: %v", err)
	}

	if !bytes.Equal(readPkt, testData) {
		t.Errorf("el paquete leído no coincide con el inyectado")
	}
}

func TestCreateTunAdapter_Fallback(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	adapter, err := CreateTunAdapter(id, false)
	if err != nil {
		t.Fatalf("CreateTunAdapter falló: %v", err)
	}
	defer adapter.Close()

	if !adapter.IsUserspace() {
		t.Errorf("se esperaba adaptador de userspace en fallback")
	}
}
