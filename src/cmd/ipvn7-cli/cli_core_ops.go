package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ipvn7/pkg/core"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l4"
)

func handleStatusCommand(id *l0.Identity, router *l1.KleinbergRouter, telemetry *l2.TelemetryRingBuffer) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:7070/api/v1/status")
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var st core.CoreStatus
		if err := json.NewDecoder(resp.Body).Decode(&st); err == nil {
			fmt.Println("=== ESTADO DEL NODO ipvn7 (Demonio Activo en :7070) ===")
			fmt.Printf("Versión Núcleo : %s\n", st.Version)
			fmt.Printf("DID Soberano  : %s\n", st.DID)
			fmt.Printf("IPv6 Soberana : %s/64\n", st.IPv6)
			fmt.Printf("IPv4 Virtual  : %s/16\n", st.IPv4)
			fmt.Printf("Tiempo Activo : %ds\n", st.UptimeSeconds)
			fmt.Println("\n--- Telemetría en Vivo ---")
			fmt.Printf("Paquetes Tx   : %d\n", st.TotalTxBytes)
			fmt.Printf("Paquetes Rx   : %d\n", st.TotalRxBytes)
			fmt.Printf("Descartes ZTNA: %d\n", st.TotalDrops)
			fmt.Printf("Pares Malla   : %d\n", st.ActivePeersCount)
			fmt.Printf("Componentes   : %d acoplados al gateway\n", len(st.AttachedComponents))
			return
		}
	}
	fmt.Println("=== ESTADO DEL NODO ipvn7 [MODO LOCAL / OFFLINE] ===")
	fmt.Printf("DID Soberano  : %s\n", id.DID())
	fmt.Printf("IPv6 Soberana : %s/64\n", id.IPv6())
	fmt.Printf("IPv4 Virtual  : %s/16\n", id.IPv4())
	fmt.Printf("MTU Canónico  : %d bytes\n", l0.MaxPacketSize)
	snap := telemetry.Snapshot()
	fmt.Println("\n--- Telemetría Local ---")
	fmt.Printf("Paquetes Tx   : %d\n", snap.PacketsTx)
	fmt.Printf("Paquetes Rx   : %d\n", snap.PacketsRx)
	fmt.Printf("Descartes     : %d\n", snap.PacketsDropped)
	fmt.Printf("Ratio Tit-Tat : %.2f\n", snap.TitForTatRatio)
	rCfg := router.GetConfig()
	fmt.Println("\n--- Topología Elástica de Red (N-Anillos) ---")
	fmt.Printf("Perfil Topológico : %s\n", rCfg.Profile)
	fmt.Printf("Resolución Anillos: %d anillos logarítmicos\n", rCfg.NumRings)
	fmt.Printf("Capacidad Pares   : %d máx (%d slots/anillo)\n", rCfg.MaxTotalPeers, rCfg.PeersPerRing)
	fmt.Printf("Métrica Voraz 2D  : Latencia %.0f%% / XOR %.0f%%\n", rCfg.AlphaLatencyWeight*100, (1.0-rCfg.AlphaLatencyWeight)*100)
}

func handleComponentsCommand() {
	resp, err := http.Get("http://127.0.0.1:7070/api/v1/components")
	if err != nil {
		fmt.Println("=== SMART COMPONENT GATEWAY (v0.6.0) ===")
		fmt.Printf("[!] No se pudo conectar con el demonio local en :7070: %v\n", err)
		fmt.Println("[*] Asegúrate de que el nodo ipvn7 esté ejecutándose (`go run cmd/ipvn7/main.go`)")
		return
	}
	defer resp.Body.Close()
	var comps []*core.ComponentRegistration
	if err := json.NewDecoder(resp.Body).Decode(&comps); err != nil {
		fmt.Printf("[!] Error decodificando respuesta: %v\n", err)
		return
	}
	fmt.Printf("=== SMART COMPONENT GATEWAY (v0.6.0) — %d Componentes Acoplados ===\n", len(comps))
	for i, c := range comps {
		fmt.Printf("%d. [%s] %s (v%s) | Transporte: %s | Estado: %s\n", i+1, c.ID, c.Name, c.Version, c.Transport, c.State)
		fmt.Printf("   Capacidades: %s\n", strings.Join(c.Capabilities, ", "))
		if c.Endpoint != "" {
			fmt.Printf("   Endpoint:    %s\n", c.Endpoint)
		}
	}
}

func handlePeersCommand(router *l1.KleinbergRouter) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:7070/api/v1/peers")
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var peers []*l1.PeerNode
		if err := json.NewDecoder(resp.Body).Decode(&peers); err == nil {
			fmt.Printf("=== TABLA DE ENRUTAMIENTO KLEINBERG (Demonio en Vivo — %d Pares) ===\n", len(peers))
			if len(peers) == 0 {
				fmt.Println("(No hay pares remotos conectados en este momento)")
				return
			}
			for _, p := range peers {
				addr := "desconocido"
				if p.Locator.PhysicalAddr != nil {
					addr = p.Locator.PhysicalAddr.String()
				}
				fmt.Printf("Anillo [%02d] -> %s @ %s (Latencia: %.2f ms, Score: %.1f)\n",
					p.RingIndex, p.DID, addr, p.Locator.LatencyMs, p.HealthScore)
			}
			return
		}
	}
	peers := router.GetAllPeers()
	rCfg := router.GetConfig()
	fmt.Printf("=== TABLA DE ENRUTAMIENTO KLEINBERG [OFFLINE] (%d / %d slots · Perfil: %s) ===\n", len(peers), rCfg.MaxTotalPeers, rCfg.Profile)
	if len(peers) == 0 {
		fmt.Println("(No hay pares remotos conectados en este momento)")
		return
	}
	for _, p := range peers {
		fmt.Printf("Anillo [%02d] -> %s (Latencia: %.2f ms, Score: %.1f)\n",
			p.RingIndex, p.DID, p.Locator.LatencyMs, p.HealthScore)
	}
}

func handleFSMCommand(router *l1.KleinbergRouter) {
	peers := router.GetAllPeers()
	fmt.Println("=== FSM DE SALUD DE PARES (Make-Before-Break) ===")
	if len(peers) == 0 {
		fmt.Println("(No hay pares registrados para evaluar la FSM)")
		return
	}
	stateNames := []string{"HEALTHY", "DEGRADED", "UNSTABLE", "UNREACHABLE", "QUARANTINED"}
	for _, p := range peers {
		stName := "UNKNOWN"
		if p.HealthState >= 0 && p.HealthState < len(stateNames) {
			stName = stateNames[p.HealthState]
		}
		fmt.Printf("DID: %s\n", p.DID[:30]+"...")
		fmt.Printf("  Estado FSM   : %s\n", stName)
		fmt.Printf("  Health Score : %.1f / 100\n", p.HealthScore)
		fmt.Printf("  Jitter / Loss: %.2f ms | %.1f%%\n\n", p.JitterMs, p.LossRate*100)
	}
}

func handlePaceCommand(pacer *l1.PacketPacer) {
	stats := pacer.Stats()
	fmt.Println("=== MARCAPASOS DE RED (PACKET PACING) ===")
	fmt.Printf("Tasa Efectiva Sostenible : %d bytes/seg (%.2f Mbps)\n",
		stats.EffectiveRateBytesPerSec, float64(stats.EffectiveRateBytesPerSec*8)/1000000.0)
	fmt.Printf("Intervalo entre Paquetes : %d µs (reloj estricto anti-bufferbloat)\n", stats.PacketIntervalUs)
	fmt.Printf("Paquetes Procesados      : %d (Regulados: %d)\n", stats.PacketsSent, stats.ThrottledCount)
}

func handleServicesCommand(svcManager *l4.ServiceLifecycleManager) {
	fmt.Println("=== GESTOR DE SERVICIOS BAJO DEMANDA (Zero Footprint) ===")
	active := svcManager.GetActiveServices()
	if len(active) == 0 {
		fmt.Println("Todos los servicios están DORMIDOS en disco (0 bytes RAM consumidos).")
		fmt.Println("Disponibles para invocación perezosa: chat_e2ee, remote_desktop, dag_store")
		return
	}
	for _, s := range active {
		fmt.Printf("  - %s [ACTIVO | Alcance: %d | Sesiones: %d]\n", s.Name, s.Scope, s.ActiveSessions)
	}
}

func handleRadarCommand(id *l0.Identity, router *l1.KleinbergRouter) {
	rCfg := router.GetConfig()
	dist := router.GetRingDistribution()
	fmt.Printf("=== RADAR DE ANILLOS CONCÉNTRICOS KLEINBERG (%d Anillos · Perfil: %s) ===\n", rCfg.NumRings, rCfg.Profile)
	fmt.Printf("Identidad Centro: %s\n", id.DID())
	fmt.Printf("Capacidad Máxima : %d pares (%d slots por anillo)\n\n", rCfg.MaxTotalPeers, rCfg.PeersPerRing)

	fmt.Println("Distribución Topológica por Anillos [0 = Vecindad Local -> N-1 = Confines XOR]:")
	for i := 0; i < len(dist); i++ {
		barLen := dist[i] * 3
		bar := strings.Repeat("█", barLen)
		if bar == "" {
			bar = "·"
		}
		label := ""
		if i == 0 {
			label = " [Inmediato / LAN]"
		} else if i == len(dist)-1 {
			label = " [Antípoda XOR]"
		}
		fmt.Printf("  Anillo [%02d] (%2d/%2d) |%-24s|%s\n", i, dist[i], rCfg.PeersPerRing, bar, label)
	}
	starved := router.RebalanceRings()
	if len(starved) > 0 {
		fmt.Printf("\n[i] Anillos despoblados detectados: %d de %d anillos (auto-prospección lista)\n", len(starved), rCfg.NumRings)
	} else {
		fmt.Println("\n[+] Topología 100% poblada para saltos logarítmicos O(log N).")
	}
}
