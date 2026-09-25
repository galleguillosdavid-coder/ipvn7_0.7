package devicebridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"ipvn7/pkg/core"
)

// SovereignDeviceBridge orquesta el descubrimiento y registro de periféricos en el Gateway
type SovereignDeviceBridge struct {
	mu         sync.RWMutex
	cfg        BridgeConfig
	devices    map[string]*DiscoveredDevice
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	httpClient *http.Client
}

// NewSovereignDeviceBridge crea un nuevo puente de periféricos hogareños
func NewSovereignDeviceBridge(cfg BridgeConfig) *SovereignDeviceBridge {
	ctx, cancel := context.WithCancel(context.Background())
	return &SovereignDeviceBridge{
		cfg:        cfg,
		devices:    make(map[string]*DiscoveredDevice),
		ctx:        ctx,
		cancel:     cancel,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Start inicia el ciclo de vida continuo del puente
func (b *SovereignDeviceBridge) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.mu.Unlock()

	// Registrar el propio puente como un componente en el Smart Gateway
	b.registerSelf()

	// Ejecutar primer escaneo inmediatamente en segundo plano
	go b.runLoop()
	return nil
}

// Stop apaga ordenadamente el puente
func (b *SovereignDeviceBridge) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}
	b.running = false
	b.cancel()
}

// GetDevices devuelve copia de los dispositivos descubiertos
func (b *SovereignDeviceBridge) GetDevices() []*DiscoveredDevice {
	b.mu.RLock()
	defer b.mu.RUnlock()

	res := make([]*DiscoveredDevice, 0, len(b.devices))
	for _, d := range b.devices {
		res = append(res, d)
	}
	return res
}

func (b *SovereignDeviceBridge) registerSelf() {
	reg := core.ComponentRegistration{
		ID:           "component:sovereign_device_bridge",
		Name:         "Puente Soberano de Dispositivos Hogar (TV/Impresoras)",
		Version:      "1.0.0",
		Description:  "Descubre periféricos LAN (SSDP/mDNS) y los convierte en DIDs sombra soberanos",
		Capabilities: []string{"bridge:lan_devices", "proxy:printer_ipp", "proxy:smart_tv_cast"},
		Transport:    "http",
		Endpoint:     b.cfg.CoreGatewayURL,
	}

	data, _ := json.Marshal(reg)
	url := fmt.Sprintf("%s/api/v1/components/register", b.cfg.CoreGatewayURL)
	_, _ = b.httpClient.Post(url, "application/json", bytes.NewReader(data))
}

func (b *SovereignDeviceBridge) runLoop() {
	ticker := time.NewTicker(b.cfg.ScanInterval)
	defer ticker.Stop()

	// Primer escaneo al iniciar
	b.scanAndSync()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.scanAndSync()
		}
	}
}

func (b *SovereignDeviceBridge) scanAndSync() {
	scanCtx, cancel := context.WithTimeout(b.ctx, 8*time.Second)
	defer cancel()

	devices, err := DiscoverHomeDevices(scanCtx, b.cfg)
	if err != nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, dev := range devices {
		b.devices[dev.ID] = dev

		// Registrar cada dispositivo descubierto como componente sombra en el Gateway
		compReg := core.ComponentRegistration{
			ID:          dev.ID,
			Name:        dev.Name,
			Version:     "1.0.0",
			Description: fmt.Sprintf("Dispositivo Hogar Sombra (%s) accesible en %s:%d", dev.Petname, dev.IPv4, dev.Port),
			Capabilities: append(dev.Capabilities,
				fmt.Sprintf("shadow_did:%s", dev.ShadowDID),
				fmt.Sprintf("petname:%s", dev.Petname),
			),
			Transport: dev.Protocol,
			Endpoint:  fmt.Sprintf("http://%s:%d", dev.IPv4, dev.Port),
		}

		data, _ := json.Marshal(compReg)
		url := fmt.Sprintf("%s/api/v1/components/register", b.cfg.CoreGatewayURL)
		resp, err := b.httpClient.Post(url, "application/json", bytes.NewReader(data))
		if err == nil {
			if resp.StatusCode == http.StatusConflict {
				hbPayload, _ := json.Marshal(map[string]string{"id": dev.ID})
				hbURL := fmt.Sprintf("%s/api/v1/components/heartbeat", b.cfg.CoreGatewayURL)
				if hbResp, hbErr := b.httpClient.Post(hbURL, "application/json", bytes.NewReader(hbPayload)); hbErr == nil {
					_ = hbResp.Body.Close()
				}
			}
			_ = resp.Body.Close()
		}
	}
}
