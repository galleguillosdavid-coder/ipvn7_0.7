package devicebridge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDeviceBridge_BuildDeviceModel(t *testing.T) {
	dev := buildDeviceModel("192.168.1.50", 631, "Impresora HP LaserJet", CategoryPrinter, "IPP", []string{"ipp:print"})
	if dev.Category != CategoryPrinter {
		t.Errorf("categoría inesperada: %s", dev.Category)
	}
	if dev.Petname != "printer.192-168-1-50.ipv7" {
		t.Errorf("petname inesperado: %s", dev.Petname)
	}
	if dev.ShadowDID == "" {
		t.Errorf("shadow DID no debe ser vacío")
	}

	tv := buildDeviceModel("192.168.1.100", 8008, "Smart TV Samsung", CategorySmartTV, "DIAL", []string{"tv:cast"})
	if tv.Category != CategorySmartTV {
		t.Errorf("categoría inesperada: %s", tv.Category)
	}
	if tv.Petname != "tv.192-168-1-100.ipv7" {
		t.Errorf("petname inesperado: %s", tv.Petname)
	}
}

func TestDeviceBridge_IntegrationWithGateway(t *testing.T) {
	var registrationsCount int32

	// Servidor falso que simula el Core Gateway
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/components/register" {
			atomic.AddInt32(&registrationsCount, 1)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
			return
		}
		if r.URL.Path == "/api/v1/components/heartbeat" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cfg := BridgeConfig{
		CoreGatewayURL: server.URL,
		ScanInterval:   100 * time.Millisecond,
		EnableSSDP:     false, // evitar sockets reales en tests unitarios
		EnablePortScan: false,
	}

	bridge := NewSovereignDeviceBridge(cfg)
	if err := bridge.Start(); err != nil {
		t.Fatalf("error iniciando puente: %v", err)
	}
	defer bridge.Stop()

	// Esperar que registre a sí mismo
	time.Sleep(150 * time.Millisecond)

	count := atomic.LoadInt32(&registrationsCount)
	if count < 1 {
		t.Errorf("se esperaba al menos 1 registro en el gateway, se obtuvieron %d", count)
	}
}

func TestDiscoverHomeDevices_TimeoutHygiene(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cfg := BridgeConfig{
		EnableSSDP:     true,
		EnablePortScan: true,
	}

	devices, err := DiscoverHomeDevices(ctx, cfg)
	if err != nil {
		t.Fatalf("error en escaneo: %v", err)
	}
	// El test debe finalizar limpiamente antes del timeout sin colgarse
	_ = devices
}
