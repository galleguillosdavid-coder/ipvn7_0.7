package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"time"

	"ipvn7/pkg/dfs"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l4"
)

// CoreServer orquesta la API REST, streaming de eventos y panel de control para v0.7.0
type CoreServer struct {
	port              int
	identity          *l0.Identity
	router            *l1.KleinbergRouter
	telemetry         *l2.TelemetryRingBuffer
	gateway           *SmartComponentGateway
	firewall          *l1.ZTNAFirewall
	staticDir         string
	startTime         time.Time
	httpServer        *http.Server
	running           int32
	uinManager        *l1.UINIdentityManager
	memoryArbiter     *l1.GlobalMemoryArbiter
	hierarchy         *l1.NodeHierarchyManager
	antiReplay        *l1.AntiReplayFilter
	chatManager       *l4.ChatManager
	telemetryExporter *TelemetryExporter
	chunkStore        *dfs.ChunkStore
	collector         *DistributedCollector
}

// NewCoreServer inicializa el servidor API del Núcleo Universal
func NewCoreServer(
	port int,
	id *l0.Identity,
	router *l1.KleinbergRouter,
	telemetry *l2.TelemetryRingBuffer,
	gateway *SmartComponentGateway,
	firewall *l1.ZTNAFirewall,
	staticDir string,
) *CoreServer {
	uin, _ := l1.NewUINIdentityManager(l1.IdentityModeHybrid)
	arb := l1.NewGlobalMemoryArbiter(l1.DefaultTotalMemoryLimit)
	hier := l1.NewNodeHierarchyManager(id.DID(), l1.NodeClassAnchor)
	anti := l1.NewAntiReplayFilter(arb)
	chat := l4.NewChatManager(id, nil)
	chat.SetRouter(router)
	store, _ := dfs.NewChunkStore(filepath.Join("data", "storage"))
	col := NewDistributedCollector(store, id)

	return &CoreServer{
		port:              port,
		identity:          id,
		router:            router,
		telemetry:         telemetry,
		gateway:           gateway,
		firewall:          firewall,
		staticDir:         staticDir,
		startTime:         time.Now(),
		uinManager:        uin,
		memoryArbiter:     arb,
		hierarchy:         hier,
		antiReplay:        anti,
		chatManager:       chat,
		telemetryExporter: NewTelemetryExporter(),
		chunkStore:        store,
		collector:         col,
	}
}

// Start arranca el servidor HTTP en segundo plano
func (s *CoreServer) Start() error {
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return nil
	}

	mux := http.NewServeMux()

	// 1. Endpoints de Estado, Núcleo y Telemetría Abierta Prometheus
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/peers", s.handlePeers)
	mux.Handle("/metrics", s.telemetryExporter.Handler())
	mux.Handle("/api/v1/metrics", s.telemetryExporter.Handler())

	// 2. Endpoints del Smart Component Gateway
	mux.HandleFunc("/api/v1/components", s.handleComponents)
	mux.HandleFunc("/api/v1/components/register", s.handleRegisterComponent)
	mux.HandleFunc("/api/v1/components/unregister", s.handleUnregisterComponent)
	mux.HandleFunc("/api/v1/components/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("/api/v1/components/toggle", s.handleToggleComponent)
	mux.HandleFunc("/api/components/toggle", s.handleToggleComponent)
	mux.HandleFunc("/api/v1/send", s.handleSendDatagram)

	// 3. Streaming de Eventos en Vivo (SSE)
	mux.HandleFunc("/api/v1/events", s.handleEventsSSE)

	// 4. Endpoints del Ecosistema Doméstico (Smart Home Hub)
	mux.HandleFunc("/api/v1/home/cast", s.handleHomeCast)
	mux.HandleFunc("/api/v1/home/cast/project", s.handleHomeCastProject)
	mux.HandleFunc("/api/v1/home/cast/stop", s.handleHomeCastStop)
	mux.HandleFunc("/api/v1/home/cast/control", s.handleHomeCastControl)
	mux.HandleFunc("/api/v1/home/print", s.handleHomePrint)
	mux.HandleFunc("/api/v1/home/print/jobs", s.handleHomePrintJobs)
	mux.HandleFunc("/api/v1/home/wol", s.handleHomeWoL)
	mux.HandleFunc("/api/v1/home/iot-shield", s.handleHomeIoTShield)
	mux.HandleFunc("/api/v1/home/drop", s.handleHomeDrop)
	mux.HandleFunc("/api/v1/home/drop/open", s.handleHomeDropOpen)
	mux.HandleFunc("/tv", s.handleTVReceiver)

	// 5. Endpoints UIN, Agentes IA y Gobernanza de Recursos
	mux.HandleFunc("/api/v1/uin/passport", s.handleUINPassport)
	mux.HandleFunc("/api/v1/uin/binding/create", s.handleUINIssueBinding)
	mux.HandleFunc("/api/v1/memory/arbiter", s.handleMemoryArbiter)
	mux.HandleFunc("/api/v1/hierarchy/snapshot", s.handleHierarchySnapshot)
	mux.HandleFunc("/api/v1/antireplay/verify", s.handleAntiReplayVerify)
	mux.HandleFunc("/api/v1/task/compute", s.handleTaskCompute)

	// 7. Endpoints de Chat Soberano E2EE (L4)
	mux.HandleFunc("/api/v1/chat/contacts", s.handleChatContacts)
	mux.HandleFunc("/api/v1/chat/messages", s.handleChatMessages)
	mux.HandleFunc("/api/v1/chat/send", s.handleChatSend)
	mux.HandleFunc("/api/v1/chat/receive", s.handleChatReceive)

	// 7.1 Endpoints de Modo VPN (Botón Soberano)
	mux.HandleFunc("/api/v1/vpn/status", s.handleVPNStatus)
	mux.HandleFunc("/api/v1/vpn/toggle", s.handleVPNToggle)

	// 7.2 Diagnóstico de Malla y Almacén Distribuido (DFS)
	mux.HandleFunc("/api/v1/telemetry/report", s.handleTelemetryReport)
	mux.HandleFunc("/api/v1/telemetry/reports", s.handleTelemetryReports)
	mux.HandleFunc("/api/v1/dfs/upload", s.handleDFSUpload)
	mux.HandleFunc("/api/v1/dfs/file/", s.handleDFSFile)
	mux.HandleFunc("/api/v1/dfs/manifest/", s.handleDFSManifest)

	// 7.3 Procesador de Intenciones en Lenguaje Natural (Fase 36)
	mux.HandleFunc("/api/v1/intent", s.handleIntentProcess)

	// Endpoints de Emparejamiento P2P
	mux.HandleFunc("/api/v1/peers/connect", s.handlePeersConnect)
	mux.HandleFunc("/api/peers/connect", s.handlePeersConnect)

	// Alias de compatibilidad hacia atrás (/api/...)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/peers", s.handlePeers)
	mux.HandleFunc("/api/components", s.handleComponents)
	mux.HandleFunc("/api/events", s.handleEventsSSE)
	mux.HandleFunc("/api/chat/receive", s.handleChatReceive)
	mux.HandleFunc("/api/uin/passport", s.handleUINPassport)
	mux.HandleFunc("/api/uin/binding/create", s.handleUINIssueBinding)
	mux.HandleFunc("/api/memory/arbiter", s.handleMemoryArbiter)
	mux.HandleFunc("/api/hierarchy/snapshot", s.handleHierarchySnapshot)
	mux.HandleFunc("/api/antireplay/verify", s.handleAntiReplayVerify)
	mux.HandleFunc("/api/task/compute", s.handleTaskCompute)
	mux.HandleFunc("/api/chat/contacts", s.handleChatContacts)
	mux.HandleFunc("/api/chat/messages", s.handleChatMessages)
	mux.HandleFunc("/api/chat/send", s.handleChatSend)
	mux.HandleFunc("/api/home/print/jobs", s.handleHomePrintJobs)

	// 8. Servidor de Estáticos (Web Dashboard v0.6): dual disco + embebido
	mux.Handle("/", http.FileServer(s.getStaticFileSystem()))


	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // SSE requiere stream continuo
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[ADVERTENCIA CORE SERVER] Error escuchando en :%d: %v\n", s.port, err)
		}
	}()

	return nil
}

// Stop apaga el servidor HTTP ordenadamente
func (s *CoreServer) Stop() error {
	if !atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func (s *CoreServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	snap := s.telemetry.Snapshot()
	peers := s.router.GetAllPeers()
	components := s.gateway.ListComponents()

	status := CoreStatus{
		Version:            "0.7.0",
		DID:                s.identity.DID(),
		IPv6:               s.identity.IPv6().String(),
		IPv4:               s.identity.IPv4().String(),
		UptimeSeconds:      int64(time.Since(s.startTime).Seconds()),
		ActivePeersCount:   len(peers),
		AttachedComponents: components,
		TotalRxBytes:       snap.PacketsRx,
		TotalTxBytes:       snap.PacketsTx,
		TotalDrops:         snap.PacketsDropped,
	}

	_ = json.NewEncoder(w).Encode(status)
}

func (s *CoreServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	all := s.router.GetAllPeers()
	peers := make([]*l1.PeerNode, 0, len(all))
	seenIPs := make(map[string]bool)
	for _, p := range all {
		if p.DID == s.identity.DID() {
			continue // Un nodo jamás debe mostrarse a sí mismo como par externo
		}
		ip := ""
		if p.Locator.PhysicalAddr != nil {
			ip = p.Locator.PhysicalAddr.IP.String()
		}
		if ip != "" && seenIPs[ip] {
			continue // Evitar duplicar instancias previas del mismo host físico
		}
		if ip != "" {
			seenIPs[ip] = true
		}
		peers = append(peers, p)
	}
	_ = json.NewEncoder(w).Encode(peers)
}

// Manejadores de componentes desacoplados en server_components.go

// GetCollector expone el motor colector distribuido para diagnósticos y auditoría
func (s *CoreServer) GetCollector() *DistributedCollector {
	return s.collector
}

// GetChunkStore expone el motor de almacenamiento de bloques DFS
func (s *CoreServer) GetChunkStore() *dfs.ChunkStore {
	return s.chunkStore
}

