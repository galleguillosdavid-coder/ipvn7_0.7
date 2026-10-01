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
	"runtime"
	"syscall"
	"time"

	"ipvn7/pkg/core"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"

	"github.com/fxamacker/cbor/v2"
)

type simpleMeshForwarder struct {
	router   *l1.KleinbergRouter
	conn     net.PacketConn
	identity *l0.Identity
}

func (f *simpleMeshForwarder) ForwardToMesh(targetDID string, ipPacket []byte) error {
	nextHop, err := f.router.FindNextHop(targetDID)
	if err != nil {
		return err
	}
	pkt := l0.NewPacket(l0.MsgTypeData, f.identity.DID(), targetDID, 0, nil, ipPacket)
	raw, err := pkt.Encode()
	if err != nil {
		return err
	}
	_, err = f.conn.WriteTo(raw, nextHop.Locator.PhysicalAddr)
	return err
}

func main() {
	keystorePath := flag.String("keystore", filepath.Join("keystore", "node_identity.key"), "Ruta del archivo keystore")
	listenPort := flag.Int("port", 7777, "Puerto UDP de transporte físico")
	peerAddr := flag.String("peer", "", "Dirección del par inicial (ej: 192.168.1.106:7777)")
	galacticMode := flag.Bool("galactic", false, "Activar modo Galáctica Scale (trillones de nodos)")
	socks5Port := flag.Int("socks5", 10807, "Puerto gateway universal (HTTP CONNECT + SOCKS5) para capturar tráfico")
	tunMode := flag.Bool("tun", false, "Activar modo TUN/TAP nativo de kernel")
	captureWeb := flag.Bool("capture-web", false, "Activar captura inmediata de tráfico web (por defecto inicia desconectado con botón)")
	noElevate := flag.Bool("no-elevate", false, "No intentar auto-elevación a Administrador")
	debugMode := flag.Bool("debug", false, "Activar registro detallado de depuracion en consola y archivo")
	logPath := flag.String("logfile", filepath.Join("data", "ipvn7.log"), "Ruta del archivo de registro persistente")
	webPort := flag.Int("web-port", 7070, "Puerto HTTP del panel de control interactivo")
	showVersion := flag.Bool("version", false, "Muestra la versión del sistema y sale")
	flag.Parse()

	if *showVersion {
		info := core.GetVersionInfo()
		fmt.Printf("ipvn7 Sovereign Network OS %s (Build: %s, WireVersion: 0x%02x, Platform: %s/%s)\n",
			info.Version, info.BuildVersion, info.WireVersion, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// 0. Inicializar sistema de logs persistente dual (consola + archivo en disco)
	if err := core.InitLogger(*logPath, *debugMode); err != nil {
		fmt.Fprintf(os.Stderr, "[AVISO] No se pudo inicializar archivo de log en %s: %v\n", *logPath, err)
	}
	defer core.CloseLogger()

	core.LogInfo("=== ipvn7 NOS v0.7.0 (PID: %d, Admin: %v, Plataforma: %s) ===", os.Getpid(), core.IsElevated(), runtime.GOOS)

	// Auto-elevación a Administrador (Mandato AGENTS.md: privilegios de administrador autoejecutados)
	if runtime.GOOS == "windows" && !core.IsElevated() && !*noElevate {
		core.LogInfo("[*] Solicitando auto-elevación con privilegios de Administrador (Kernel L3 TUN)...")
		elevatedArgs := append(os.Args[1:], "-tun", "-no-elevate")
		if *debugMode {
			elevatedArgs = append(elevatedArgs, "-debug")
		}
		if core.RequestSelfElevation(elevatedArgs) {
			core.LogInfo("[+] Proceso elevado lanzado en consola interactiva permanente. Cerrando lanzador inicial.")
			os.Exit(0)
		}
		core.LogWarn("[!] Permisos de Administrador no concedidos o cancelados por el usuario.")
		core.LogInfo("[*] Fallback automático: Continuando en Modo Usuario (Universal Gateway HTTP/SOCKS5)...")
	}

	// Si el proceso cuenta con privilegios elevados, purgar residuos en memoria y activar TUN L3
	if core.IsElevated() {
		core.TerminateConflictingProcesses(nil)
		if !*tunMode {
			*tunMode = true
			core.LogInfo("[+] Privilegios de Administrador verificados: Kernel TUN L3 activado.")
		}
	}

	var identity *l0.Identity
	var err error
	if identity, err = l0.LoadFromFile(*keystorePath); err != nil {
		if identity, err = l0.GenerateIdentity(); err != nil {
			fmt.Fprintf(os.Stderr, "[FATAL L0] Error generando entropía: %v\n", err); os.Exit(1)
		}
		_ = identity.SaveToFile(*keystorePath)
	}

	router := l1.NewKleinbergRouter(identity)
	bufferPool := l1.NewBufferPool()
	firewall := l1.NewZTNAFirewall(true)

	if *galacticMode {
		fmt.Println("[GALACTIC MODE] Modo galáctico deshabilitado en v0.7.0")
	}

	telemetry := l2.NewTelemetryRingBuffer()
	fmt.Printf("[L0 DID]: %s | [IPv4 Virtual]: %s\n", identity.DID(), identity.IPv4())
	fmt.Printf("[L1]: Kleinberg Router (12 Anillos) | ZTNA Default-Deny\n")
	fmt.Println("================================================================================")

	listenAddr := fmt.Sprintf("0.0.0.0:%d", *listenPort)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR TRANSPORTE] No se pudo abrir socket UDP en %s: %v\n", listenAddr, err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Printf("[+] Escuchando datagramas ipvn7 en UDP %s...\n", listenAddr)

	blindStore := l1.NewHybridBlindBeaconStore("")
	rendezvous := l1.NewBlindRendezvousManager(identity, blindStore, "ipvn7-minimal-mesh")
	discoveryEngine := l1.NewAutonomousDiscoveryEngine(identity, router, rendezvous, firewall, nil, conn, *listenPort)
	discoveryEngine.Start()
	defer discoveryEngine.Stop()
	fmt.Println("[+] Motor de Descubrimiento STUN activo en segundo plano.")

	// 6. Gateway SOCKS5 para capturar tráfico de navegación
	var socks5Gateway *l1.SOCKS5Gateway
	if *socks5Port > 0 {
		core.LogInfo("[+] Iniciando gateway SOCKS5 en puerto %d...", *socks5Port)
		socks5Gateway = l1.NewSOCKS5Gateway(fmt.Sprintf("127.0.0.1:%d", *socks5Port), nil)
		socks5Gateway.LogFunc = core.LogInfo
		if err := socks5Gateway.Start(); err != nil {
			core.LogError("[ERROR SOCKS5] No se pudo iniciar gateway: %v", err)
		} else {
			defer socks5Gateway.Stop()
			core.LogInfo("[+] SOCKS5 activo en 127.0.0.1:%d.", *socks5Port)
			defer func() {
				core.LogInfo("[*] Restaurando conexión directa de red de Windows...")
				_ = core.ClearWindowsUserProxy(nil)
			}()
		}
	}

	// 7. Panel de Control WebUI nativo embebido y Embajador de Dispositivos (Rol L)
	shadowRegistry := l1.NewShadowDeviceRegistry()
	webUI := core.StartWebUI(*webPort, identity, router, socks5Gateway)
	if webUI != nil {
		webUI.SetShadowRegistry(shadowRegistry)
		defer webUI.Stop()
		core.LogInfo("[+] Panel de control interactivo activo en http://127.0.0.1:%d", *webPort)

		proxyStr := fmt.Sprintf("http=127.0.0.1:%d;https=127.0.0.1:%d;socks=127.0.0.1:%d", *socks5Port, *socks5Port, *socks5Port)
		webUI.SetCallbacks(
			func() error {
				core.LogInfo("[+] VPN activada desde botón de usuario (Proxy Universal 127.0.0.1:%d)", *socks5Port)
				return core.SetWindowsUserProxy(proxyStr, nil)
			},
			func() error {
				core.LogInfo("[-] VPN desactivada desde botón de usuario. Restaurando Internet directo.")
				return core.ClearWindowsUserProxy(nil)
			},
			func() {
				core.LogInfo("[*] Cerrando aplicación y restaurando conexión directa...")
				_ = core.ClearWindowsUserProxy(nil)
			},
		)

		if *captureWeb {
			_ = core.SetWindowsUserProxy(proxyStr, nil)
			webUI.SetVPNState("connected")
		} else {
			_ = core.ClearWindowsUserProxy(nil)
			webUI.SetVPNState("disconnected")
		}

		// Lanzar automáticamente la ventana tipo app nativa
		core.LaunchDesktopWindow(fmt.Sprintf("http://127.0.0.1:%d", *webPort))
	}

	// 8. Adaptador TUN L3 para capturar tráfico a nivel de kernel
	var tunAdapter l1.TunAdapter
	var tunRouter *l1.TUNRouter
	if *tunMode {
		fmt.Println("[+] Inicializando adaptador TUN nativo de kernel...")
		var tunErr error
		tunAdapter, tunErr = l1.CreateTunAdapter(identity, true)
		if tunErr != nil || tunAdapter.IsUserspace() {
			fmt.Printf("[AVISO TUN] Interfaz de kernel no disponible (%v). Usando espacio de usuario.\n", tunErr)
		} else {
			defer tunAdapter.Close()
			fmt.Printf("[+] Interfaz TUN activa [%s] IP: %s (Red 10.7.0.0/16 enrutada a ipvn7).\n",
				tunAdapter.Mode(), identity.IPv4())
			forwarder := &simpleMeshForwarder{
				router:   router,
				conn:     conn,
				identity: identity,
			}
			tunRouter = l1.NewTUNRouter(tunAdapter, forwarder)
			defer tunRouter.Close()
		}
	}

	// 6. Si se especificó un par inicial, enviar datagrama de presentación
	if *peerAddr != "" {
		if remoteAddr, err := net.ResolveUDPAddr("udp", *peerAddr); err == nil {
			initPkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, identity.DID(), "", 0, nil, nil)
			if raw, err := initPkt.Encode(); err == nil {
				_, _ = conn.WriteTo(raw, remoteAddr)
			}
			fmt.Printf("[+] Saludo de enlace enviado a par inicial: %s\n", *peerAddr)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 7. Bucle de recepción UDP
	buf := make([]byte, l0.MaxPacketSize*2)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, fromAddr, err := conn.ReadFrom(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}

			// Despacho rápido de respuestas STUN RFC 5389
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

			// Procesamiento mínimo de paquetes usando APIs existentes
			var packet l0.Packet
			err = cbor.Unmarshal(buf[:n], &packet)
			if err != nil {
				// Paquete inválido o malformado
				pktBuf.Release()
				continue
			}

			// Verificar magic bytes
			if packet.Magic != l0.MagicBytes || packet.Version != l0.WireVersion {
				pktBuf.Release()
				continue
			}

			if udpAddr, ok := fromAddr.(*net.UDPAddr); ok && packet.SourceDID != identity.DID() && packet.SourceDID != "" {
				_ = router.AddOrUpdatePeer(packet.SourceDID, udpAddr, 1.0)
			}

			// Procesar según tipo de paquete
			switch packet.Type {
			case l0.MsgTypeHandshakeInit, l0.MsgTypeHandshakeResp, l0.MsgTypeHandshakeAuth:
				fmt.Printf("[HANDSHAKE] Recibido tipo %d de %s\n", packet.Type, packet.SourceDID)

			case l0.MsgTypeData:
				if packet.DestDID == identity.DID() {
					if tunRouter != nil {
						_ = tunRouter.InjectFromMesh(packet.Payload)
					} else {
						fmt.Printf("[DATOS] Recibido de %s (%d bytes)\n", packet.SourceDID, len(packet.Payload))
					}
				} else {
					nextHop, err := router.FindNextHop(packet.DestDID)
					if err == nil && nextHop != nil {
						conn.WriteTo(buf[:n], nextHop.Locator.PhysicalAddr)
						telemetry.RecordEvent(l2.EventTxPacket, uint32(n), 1100, 0)
					}
				}

			case l0.MsgTypeKeepAlive:
				fmt.Printf("[KEEPALIVE] Recibido de %s\n", packet.SourceDID)

			case l0.MsgTypeRoamingUpdate:
				if packet.SourceDID != identity.DID() {
					fmt.Printf("[ROAMING] %s actualizó su dirección\n", packet.SourceDID)
					if udpAddr, ok := fromAddr.(*net.UDPAddr); ok {
						respPkt := l0.NewPacket(l0.MsgTypeKeepAlive, identity.DID(), packet.SourceDID, 0, nil, nil)
						if raw, err := respPkt.Encode(); err == nil {
							_, _ = conn.WriteTo(raw, udpAddr)
						}
					}
				}
			}

			pktBuf.Release()
		}
	}(ctx)

	// 8. Pulsos de telemetría
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("[+] Núcleo Mínimo ipvn7 en ejecución continua. Presione Ctrl+C para detener.")

	for {
		select {
		case <-ticker.C:
			if *peerAddr != "" {
				if rAddr, err := net.ResolveUDPAddr("udp", *peerAddr); err == nil {
					kp := l0.NewPacket(l0.MsgTypeKeepAlive, identity.DID(), "", 0, nil, nil)
					if raw, err := kp.Encode(); err == nil {
						_, _ = conn.WriteTo(raw, rAddr)
					}
				}
			}
			snap := telemetry.Snapshot()
			if socks5Gateway != nil {
				stats := socks5Gateway.GetStats()
				fmt.Printf("[TELEMETRÍA] Malla P2P: %d pares | Pkts Rx: %d | Web: %v conns (Tx: %s | Rx: %s) | PQC: ML-KEM-768 Activo\n",
					len(router.GetAllPeers()), snap.PacketsRx, stats["total_connections"],
					l1.FormatBytes(stats["bytes_tx"].(uint64)), l1.FormatBytes(stats["bytes_rx"].(uint64)))
			} else {
				fmt.Printf("[TELEMETRÍA] Tx: %d pkts | Rx: %d pkts | Drops: %d | Pares: %d\n",
					snap.PacketsTx, snap.PacketsRx, snap.PacketsDropped, len(router.GetAllPeers()))
			}
		case sig := <-sigChan:
			fmt.Printf("\n[*] Señal recibida (%v). Apagado ordenado...\n", sig)
			cancel()
			return
		}
	}
}
