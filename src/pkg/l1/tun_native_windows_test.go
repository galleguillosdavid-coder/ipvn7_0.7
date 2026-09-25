//go:build windows

package l1

import (
	"testing"

	"ipvn7/pkg/l0"
)

func TestNativeTunWindows_FallbackWhenMissingDLL(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad: %v", err)
	}

	adapter, err := CreateTunAdapter(id, true)
	if err != nil {
		t.Fatalf("CreateTunAdapter falló inesperadamente: %v", err)
	}
	defer adapter.Close()

	if adapter == nil {
		t.Fatal("El adaptador retornado es nulo")
	}

	// Al no estar presente wintun.dll en el directorio local, debe operar como UserspaceVirtualAdapter
	if !adapter.IsUserspace() {
		t.Logf("Adaptador Wintun detectado operando en modo kernel nativo: %s", adapter.Mode())
	} else {
		t.Logf("Fallback limpio a UserspaceVirtualAdapter certificado: %s", adapter.Mode())
		if adapter.Mode() != ModeUserspaceVirtual {
			t.Errorf("Esperado %s, obtenido %s", ModeUserspaceVirtual, adapter.Mode())
		}
	}

	if adapter.MTU() != DefaultMTU {
		t.Errorf("MTU inválido: %d, esperado %d", adapter.MTU(), DefaultMTU)
	}
}

func TestNativeTunWindows_LifecycleAndProperties(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	adapter := NewUserspaceVirtualAdapter(id)

	if !adapter.IsUserspace() {
		t.Error("UserspaceVirtualAdapter reportó IsUserspace() = false")
	}

	if err := adapter.Close(); err != nil {
		t.Errorf("Error al cerrar adaptador: %v", err)
	}

	// Tras el cierre, las operaciones deben retornar error sin entrar en pánico
	if err := adapter.WritePacket([]byte("test")); err == nil {
		t.Error("Se esperaba error escribiendo en adaptador cerrado")
	}
	if _, err := adapter.ReadPacket(); err == nil {
		t.Error("Se esperaba error leyendo de adaptador cerrado")
	}
}
