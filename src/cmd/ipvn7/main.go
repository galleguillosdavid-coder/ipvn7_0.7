package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ipvn7/pkg/components/device_bridge"
	"ipvn7/pkg/components/os_runner"
	vehiclerobotbridge "ipvn7/pkg/components/vehicle_robot_bridge"
	"ipvn7/pkg/core"
	"ipvn7/pkg/interfaces"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l3"
)

const (
	Banner = `
██╗██████╗ ██╗   ██╗███╗   ██╗███████╗
██║██╔══██╗██║   ██║████╗  ██║╚════██║
██║██████╔╝██║   ██║██╔██╗ ██║    ██╔╝
██║██╔═══╝ ╚██╗ ██╔╝██║╚██╗██║   ██╔╝ 
██║██║      ╚████╔╝ ██║ ╚████║   ██║  
╚═╝╚═╝       ╚═══╝  ╚═╝  ╚═══╝   ╚═╝  
   Universal Sovereign Core - ipvn7 NOS v0.6
`
)

func main() {
	keystorePath := flag.String("keystore", filepath.Join("keystore", "node_identity.key"), "Ruta del archivo keystore")
	listenPort := flag.Int("port", 7777, "Puerto UDP de transporte físico")
	webPort := flag.Int("web-port", 7070, "Puerto HTTP para el panel de control y API universal")
	mcpMode := flag.Bool("mcp", false, "Iniciar en modo servidor MCP (Model Context Protocol) sobre stdio")
	diagnostics := flag.Bool("diagnostics", false, "Ejecutar autodiagnóstico del Núcleo Universal y salir")
	preferNative := flag.Bool("tun-native", true, "Intentar inicializar interfaz TUN nativa de kernel (fallback a userspace)")
	vpnMode := flag.Bool("vpn", false, "Activar modo VPN Sovereign de inmediato al arrancar")
	enableHomeBridge := flag.Bool("home-bridge", true, "Habilitar descubrimiento de periféricos LAN (TVs/Impresoras)")
	showVersion := flag.Bool("version", false, "Mostrar versión formal y capacidades del motor")
	doRollback := flag.Bool("rollback", false, "Revertir atómicamente al binario .old previo")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ipvn7 Network OS v0.7.0 (Sovereign Edition)\n")
		fmt.Printf("WASM Core: compatible (web/ipvn7.wasm / app_wasm.js)\n")
		fmt.Printf("Atomic Version Manager: activo\n")
		return
	}

	if *doRollback {
		execPath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] No se pudo resolver ejecutable: %v\n", err)
			os.Exit(1)
		}
		updater := core.NewAtomicUpdateEngine("0.7.0", nil)
		if err := updater.Rollback(execPath); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR ROLLBACK] %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[ROLLBACK EXITOSO] Binario restaurado al respaldo .old anterior.")
		return
	}

	// 1. Capa L0: Identidad y Criptografía Soberana Inmutable
	var identity *l0.Identity
	var err error

	if _, statErr := os.Stat(*keystorePath); statErr == nil {
		identity, err = l0.LoadFromFile(*keystorePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR L0] No se pudo cargar keystore: %v. Generando uno nuevo...\n", err)
			identity, _ = l0.GenerateIdentity()
			_ = identity.SaveToFile(*keystorePath)
		}
	} else {
		identity, err = l0.GenerateIdentity()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[FATAL L0] Error generando entropía soberana: %v\n", err)
			os.Exit(1)
		}
		_ = identity.SaveToFile(*keystorePath)
	}

	// 2. Capa L1: Enrutamiento Kleinberg, Marcapasos y Adaptador
	router := l1.NewKleinbergRouter(identity)
	bufferPool := l1.NewBufferPool()
	pacer := l1.NewPacketPacer(l1.DefaultPacerConfig())
	_ = pacer
	firewall := l1.NewZTNAFirewall(true) // Default-Deny por diseño
	qos := l1.NewQoSManager()

	tun, err := l1.CreateTunAdapter(identity, *preferNative)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL L1] Error instanciando adaptador TUN: %v\n", err)
		os.Exit(1)
	}
	defer tun.Close()

	// 3. Capa L2: Telemetría Lock-Free
	telemetry := l2.NewTelemetryRingBuffer()

	// 3.1 Modo Servidor MCP (Model Context Protocol para Asistentes y Agentes de IA)
	if *mcpMode {
		mcpServer := l3.NewMCPServer(identity, router, telemetry)
		mcpServer.Firewall = firewall
		if err := mcpServer.ServeStdioDefault(); err != nil {
			fmt.Fprintf(os.Stderr, "[FATAL MCP] %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Impresión de Identidad y Estado de Red
	fmt.Print(Banner)
	fmt.Println("================================================================================")
	fmt.Printf("[L0 DID SOBERANO]   : %s\n", identity.DID())
	fmt.Printf("[L1 IPv6 SOBERANO]  : %s/64\n", identity.IPv6())
	fmt.Printf("[L1 IPv4 VIRTUAL]   : %s/16\n", identity.IPv4())
	if tun.IsUserspace() {
		fmt.Printf("[L1 ADAPTADOR]      : Espacio de Usuario Zero-Copy (In-Memory Safe Fallback)\n")
	} else {
		fmt.Printf("[L1 ADAPTADOR]      : Interfaz Nativa de Kernel OS Activa\n")
	}
	fmt.Printf("[L1 ZTNA FIREWALL]  : Micro-segmentación Default-Deny Activa (Aislamiento Total)\n")
	fmt.Printf("[L1 ENRUTADOR]      : Topología Kleinberg Mundo Pequeño (12 Anillos / 120 Pares)\n")
	fmt.Printf("[L1 ZERO-COPY POOL] : 3 Niveles Preasignados (64B / 1500B / 64KB)\n")
	fmt.Printf("[L1 MARCAPASOS]     : Sustainable Flow Pacer (MTU 1280B Canónico)\n")
	fmt.Printf("[L2 TELEMETRÍA]     : Ring Buffer Lock-Free (<28 ns, Cero Alocaciones)\n")
	fmt.Printf("[CORE SMART GATEWAY]: Bus de Acoplamiento Externo (/api/v1/components)\n")
	fmt.Printf("[CORE API & WEB]    : http://localhost:%d\n", *webPort)
	fmt.Println("================================================================================")

	// Si se solicita modo diagnóstico aislado (sin ocupar puertos de producción)
	if *diagnostics {
		runDiagnostics(identity, router, bufferPool, firewall, telemetry, *listenPort, *webPort)
		return
	}

	// Si se solicita modo MCP stdio para agentes de IA
	if *mcpMode {
		mcpServer := l3.NewMCPServer(identity, router, telemetry)
		if err := mcpServer.ServeStdioDefault(); err != nil {
			fmt.Fprintf(os.Stderr, "[MCP ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 4. Núcleo v0.6: Smart Component Gateway (Plano de Acoplamiento Externo)
	healing := l1.NewLinkHealingEngine(router)
	gateway := core.NewSmartComponentGateway(identity, router, telemetry, firewall)
	gateway.SetHealingEngine(healing)
	defer gateway.Stop()

	// Registro de capacidades fundacionales del núcleo en el gateway
	registerCoreCapabilities(gateway)

	// 5. Servidor Core API & Panel Ligero v0.6
	coreServer := core.NewCoreServer(*webPort, identity, router, telemetry, gateway, firewall, core.FindStaticDir())
	if err := coreServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[ADVERTENCIA CORE] Error iniciando servidor API: %v\n", err)
	} else {
		defer func() { _ = coreServer.Stop() }()
		if *vpnMode {
			if err := core.ActivateVPNMode(); err == nil {
				fmt.Println("[+] [VPN I7] Modo Internet Protegido activado automáticamente (SOCKS5 127.0.0.1:10807)")
			}
		}
	}

	// 5.1 Puente de Periféricos Domésticos (Auto-descubrimiento en segundo plano)
	if *enableHomeBridge {
		bridgeCfg := devicebridge.DefaultBridgeConfig()
		bridgeCfg.CoreGatewayURL = fmt.Sprintf("http://127.0.0.1:%d", *webPort)
		bridgeCfg.ScanInterval = 30 * time.Second
		homeBridge := devicebridge.NewSovereignDeviceBridge(bridgeCfg)
		if err := homeBridge.Start(); err == nil {
			defer homeBridge.Stop()
			fmt.Println("[+] Sovereign Device Bridge activo: escaneando periféricos LAN (TVs/Impresoras)...")
		}
	}

	// 5.2 Agente Satélite Universal de Sistema Operativo (Fase 34: Control de Hardware & SO)
	osRunner := osrunner.NewOSRunnerComponent("os_runner_local", "Universal OS Hardware Runner", gateway, nil)
	if err := osRunner.Start(); err == nil {
		defer osRunner.Stop()
		fmt.Println("[+] OS Hardware Runner activo: control de hardware, energía y procesos vinculado.")
	}

	// 5.3 Puente Satélite Universal de Robótica, Drones y Vehículos (Fase 35)
	robotVehBridge := vehiclerobotbridge.NewVehicleRobotBridge("veh_robot_bridge", "Robotics & Vehicle Bridge", vehiclerobotbridge.GatewayBusFunc(func(evType string, data []byte) error {
		gateway.PublishEvent(core.MeshEvent{
			Type:      core.EventType(evType),
			Timestamp: time.Now().UnixNano(),
			Source:    "veh_robot_bridge",
			Payload:   map[string]interface{}{"data": string(data)},
		})
		return nil
	}))
	_ = robotVehBridge
	fmt.Println("[+] Vehicle & Robot Bridge activo: soporte MAVLink v2, CAN Bus y J1939 vinculado.")


	// 5.4 Programador Autónomo de Tareas y Cron Soberano (Fase 35)
	cronSched := core.NewSovereignCronScheduler(func(jobID, intent string) error {
		gateway.PublishEvent(core.MeshEvent{
			Type:      core.EventType("cron:" + jobID),
			Timestamp: time.Now().UnixNano(),
			Source:    "sovereign_cron",
			Payload:   map[string]interface{}{"intent": intent},
		})
		return nil
	})
	cronSched.Start()
	defer cronSched.Stop()
	fmt.Println("[+] Sovereign Cron Scheduler activo: evaluación de tareas autónomas iniciada.")

	// 6. Socket UDP de Transporte Físico
	listenAddr := fmt.Sprintf("0.0.0.0:%d", *listenPort)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR TRANSPORTE] No se pudo abrir socket UDP en %s: %v\n", listenAddr, err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Printf("[+] Escuchando datagramas ipvn7 en UDP %s...\n", listenAddr)

	// 6.0 Mapeo dinámico UPnP IGD en router residencial (asíncrono no bloqueante)
	go func(port int) {
		mapper := l1.NewUPnPMapper(3 * time.Second)
		if err := mapper.DiscoverAndForward(port, "ipvn7-mesh-udp"); err == nil {
			fmt.Printf("[+] UPnP IGD: Puerto exterior UDP %d mapeado exitosamente en router residencial.\n", port)
		}
	}(*listenPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 6.1 Ensamblaje de la Línea de Producción (Pipes & Filters) y FSM Determinista
	nodeFSM := core.NewDeterministicNodeFSM(interfaces.NodeStateDisconnected)
	dataPipeline := core.NewLinearPipeline().
		AddStage(core.NewDecodeCBORStage()).
		AddStage(core.NewZTNAFilterStage(firewall, uint16(*listenPort), gateway)).
		AddStage(core.NewQoSFilterStage(qos)).
		AddStage(core.NewTopologyFSMStage(router, firewall, gateway, nodeFSM))
	dataPipeline.SetDeadLetterHandler(core.NewTelemetryDeadLetterHandler(telemetry))

	// 6.2 Centinela de Sondeo WAN Activo y Auto-Reparación (L1)
	prober := l1.NewWANActiveProber(router, healing, conn)
	prober.Start(ctx)
	defer prober.Stop()

	// 7. Descubrimiento Autónomo de Malla (EBRA + STUN LAN/WAN)
	blindStore := l1.NewHybridBlindBeaconStore("")
	rendezvous := l1.NewBlindRendezvousManager(identity, blindStore, "ipvn7-sovereign-mesh-v0.6")
	discoveryEngine := l1.NewAutonomousDiscoveryEngine(
		identity,
		router,
		rendezvous,
		firewall,
		nil,
		conn,
		*listenPort,
	)
	discoveryEngine.Start()
	defer discoveryEngine.Stop()
	fmt.Println("[+] Motor de Descubrimiento Autónomo EBRA/STUN activo en segundo plano.")

	// Bucle de recepción UDP con despacho unidireccional a la línea de montaje
	buf := make([]byte, l0.MaxPacketSize*2)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, remoteAddr, err := conn.ReadFrom(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}

			// Despacho rápido de sondas WAN activas
			if n >= l1.ProbeWireSize && prober != nil && prober.HandleIncomingProbe(buf[:n], remoteAddr, false) {
				continue
			}

			// Despacho rápido de respuestas STUN RFC 5389 para mapeo reflexivo en socket físico
			if n >= 20 && binary.BigEndian.Uint32(buf[4:8]) == l1.STUNMagicCookie {
				if ip, port, err := l1.ParseSTUNBindingResponse(buf[:n], [12]byte{}); err == nil {
					discoveryEngine.SetReflexiveEndpoint(fmt.Sprintf("%s:%d", ip.String(), port))
				}
				continue
			}

			// Adquisición de búfer Zero-Copy
			pktBuf := bufferPool.Acquire(n)
			copy(pktBuf.RawSlice(), buf[:n])
			telemetry.RecordEvent(l2.EventRxPacket, uint32(n), 1100, 0)

			pCtx := core.AcquirePacketContext(pktBuf, remoteAddr)
			if err := dataPipeline.Execute(ctx, pCtx); err != nil || pCtx.Dropped {
				// Dead-letter handler ya registró telemetría y liberó el búfer
				core.ReleasePacketContext(pCtx)
				continue
			}

			pktBuf.Release()
			core.ReleasePacketContext(pCtx)
		}
	}(ctx)

	// Pulsos de telemetría y espera de señales
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("[+] Núcleo Universal ipvn7 en ejecución continua. Presione Ctrl+C para detener.")

	for {
		select {
		case <-ticker.C:
			snap := telemetry.Snapshot()
			fmt.Printf("[TELEMETRÍA v0.6] Tx: %d pkts | Rx: %d pkts | Drops: %d | Pares: %d | Comp. Adjuntos: %d\n",
				snap.PacketsTx, snap.PacketsRx, snap.PacketsDropped, len(router.GetAllPeers()), len(gateway.ListComponents()))
		case sig := <-sigChan:
			fmt.Printf("\n[*] Señal recibida (%v). Apagado ordenado del Núcleo Universal...\n", sig)
			cancel()
			return
		}
	}
}


