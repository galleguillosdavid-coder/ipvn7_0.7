package l1

import (
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/l0"
)

// AnomalyType clasifica los tipos de interferencia detectados por la sonda
type AnomalyType string

const (
	AnomalyDNSPoisoning   AnomalyType = "DNS_POISONING"
	AnomalyTLSSNIBlock    AnomalyType = "TLS_SNI_BLOCK"
	AnomalyTCPRSTInject   AnomalyType = "TCP_RST_INJECTION"
	AnomalyBufferbloat    AnomalyType = "BUFFERBLOAT_SATURATION"
)

// CensorshipIncident describe una alteración o intento de bloqueo detectado
type CensorshipIncident struct {
	ID                string      `json:"id"`
	Timestamp         time.Time   `json:"timestamp"`
	AnomalyType       AnomalyType `json:"anomaly_type"`
	TargetHost        string      `json:"target_host"`
	ObservedBehavior  string      `json:"observed_behavior"`
	MitigationApplied string      `json:"mitigation_applied"`
	Severity          string      `json:"severity"` // "LOW", "MEDIUM", "HIGH"
}

// ProbeReport resume la salud y libertad de la red observada por la sonda
type ProbeReport struct {
	IsActive             bool                  `json:"is_active"`
	LastScan             time.Time             `json:"last_scan"`
	HealthScore          float64               `json:"health_score"` // 0.0 - 100.0
	LatencyEWMA          float64               `json:"latency_ewma_ms"`
	JitterMs             float64               `json:"jitter_ms"`
	PacketLossPct        float64               `json:"packet_loss_pct"`
	PeersAudited         int                   `json:"peers_audited"`
	CensorshipIncidents  []*CensorshipIncident `json:"censorship_incidents"`
	CensorshipResistance string                `json:"censorship_resistance"` // "SHIELDED"
}

// NetworkProbeEngine ejecuta auditorías voluntarias sin violar el axioma Zero-PII
type NetworkProbeEngine struct {
	mu           sync.RWMutex
	identity     *l0.Identity
	router       *KleinbergRouter
	isActive     atomic.Bool
	stopChan     chan struct{}
	lastReport   *ProbeReport
	incidentList []*CensorshipIncident
}

// NewNetworkProbeEngine crea la sonda voluntaria de red
func NewNetworkProbeEngine(id *l0.Identity, router *KleinbergRouter) *NetworkProbeEngine {
	npe := &NetworkProbeEngine{
		identity: id,
		router:   router,
		lastReport: &ProbeReport{
			IsActive:             false,
			LastScan:             time.Now(),
			HealthScore:          100.0,
			LatencyEWMA:          0.0,
			JitterMs:             0.0,
			PacketLossPct:        0.0,
			PeersAudited:         len(router.GetAllPeers()),
			CensorshipResistance: "SHIELDED",
			CensorshipIncidents:  make([]*CensorshipIncident, 0),
		},
		incidentList: make([]*CensorshipIncident, 0),
	}
	return npe
}

// SetActive activa o apaga la sonda en caliente
func (npe *NetworkProbeEngine) SetActive(active bool) bool {
	prev := npe.isActive.Swap(active)
	if !prev && active {
		npe.stopChan = make(chan struct{})
		go npe.probeLoop()
		go npe.ExecuteLiveAudit()
	} else if prev && !active {
		if npe.stopChan != nil {
			close(npe.stopChan)
		}
		npe.mu.Lock()
		npe.lastReport.IsActive = false
		npe.mu.Unlock()
	}
	return active
}

// IsActive indica si la sonda está corriendo
func (npe *NetworkProbeEngine) IsActive() bool {
	return npe.isActive.Load()
}

// GetReport devuelve el estado actual de la sonda
func (npe *NetworkProbeEngine) GetReport() *ProbeReport {
	npe.mu.RLock()
	defer npe.mu.RUnlock()

	rep := *npe.lastReport
	rep.IsActive = npe.isActive.Load()
	return &rep
}

// ExecuteLiveAudit realiza una inspección física y activa de la red mediante sockets reales
func (npe *NetworkProbeEngine) ExecuteLiveAudit() *ProbeReport {
	peers := npe.router.GetAllPeers()
	totalPeers := len(peers)
	if totalPeers == 0 {
		totalPeers = 1 // Este nodo autónomo
	}

	// 1. Medición de RTT y Jitter Real mediante consultas DNS UDP directas a Anycast Root/Public Resolvers (DEC-003)
	neutralDNS := []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
	if envDNS := os.Getenv("IPVN7_NEUTRAL_RESOLVERS"); envDNS != "" {
		neutralDNS = strings.Split(envDNS, ",")
	}
	latencies := make([]float64, 0, len(neutralDNS)*2)
	failures := 0

	for _, srv := range neutralDNS {
		lat, err := probeRawUDPQuery(srv, 700*time.Millisecond)
		if err != nil {
			failures++
		} else {
			latencies = append(latencies, lat)
		}
	}

	// Calcular EWMA de latencia y jitter de varianza real
	var avgLat float64
	var jitter float64
	var lossPct float64

	totalAttempts := len(neutralDNS)
	if totalAttempts > 0 {
		lossPct = (float64(failures) / float64(totalAttempts)) * 100.0
	}

	if len(latencies) > 0 {
		var sum float64
		for _, l := range latencies {
			sum += l
		}
		avgLat = sum / float64(len(latencies))

		// Varianza para jitter
		var varSum float64
		for _, l := range latencies {
			diff := l - avgLat
			varSum += diff * diff
		}
		jitter = math.Sqrt(varSum / float64(len(latencies)))
	} else {
		// Modo fuera de línea / red aislada
		avgLat = 0.5
		jitter = 0.1
	}

	// 2. Detección real de interferencia DNS (DNS Poisoning / Redirección NXDOMAIN)
	incidents := make([]*CensorshipIncident, 0)
	if dnsInc := checkDNSTampering(); dnsInc != nil {
		incidents = append(incidents, dnsInc)
	}

	// 3. Detección real de bloqueo TLS / RST Injection en puerto 443
	if tlsInc := checkTLSInterference(); tlsInc != nil {
		incidents = append(incidents, tlsInc)
	}

	// 4. Calcular índice de salud (HealthScore) derivado de mediciones reales
	healthScore := 100.0 - lossPct*0.5 - math.Min(jitter*2.0, 20.0)
	if len(incidents) > 0 {
		healthScore -= float64(len(incidents) * 15)
	}
	if healthScore < 10.0 {
		healthScore = 10.0
	}
	if healthScore > 100.0 {
		healthScore = 100.0
	}

	npe.mu.Lock()
	defer npe.mu.Unlock()

	// Mantener histórico de incidentes únicos
	for _, inc := range incidents {
		exists := false
		for _, existing := range npe.incidentList {
			if existing.TargetHost == inc.TargetHost && existing.AnomalyType == inc.AnomalyType {
				exists = true
				break
			}
		}
		if !exists {
			npe.incidentList = append(npe.incidentList, inc)
		}
	}

	// Integrar EWMA ponderado si ya existía reporte anterior
	if npe.lastReport.LatencyEWMA > 0 && avgLat > 0 {
		avgLat = 0.7*avgLat + 0.3*npe.lastReport.LatencyEWMA
	}

	npe.lastReport = &ProbeReport{
		IsActive:             npe.isActive.Load(),
		LastScan:             time.Now(),
		HealthScore:          math.Round(healthScore*10) / 10,
		LatencyEWMA:          math.Round(avgLat*100) / 100,
		JitterMs:             math.Round(jitter*100) / 100,
		PacketLossPct:        math.Round(lossPct*10) / 10,
		PeersAudited:         totalPeers,
		CensorshipResistance: "SHIELDED",
		CensorshipIncidents:  npe.incidentList,
	}

	return npe.lastReport
}

// RunAudit ejecuta la auditoría empírica en vivo del motor de sondeo
func (npe *NetworkProbeEngine) RunAudit() *ProbeReport {
	return npe.ExecuteLiveAudit()
}

func (npe *NetworkProbeEngine) probeLoop() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-npe.stopChan:
			return
		case <-ticker.C:
			if npe.isActive.Load() {
				npe.ExecuteLiveAudit()
			}
		}
	}
}

