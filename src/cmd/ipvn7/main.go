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

)

func main() {
	keystorePath := flag.String("keystore", filepath.Join("keystore", "node_identity.key"), "Ruta keystore")
	listenPort := flag.Int("port", 7777, "Puerto UDP")
	peerAddr := flag.String("peer", "", "Par inicial")
	galacticMode := flag.Bool("galactic", false, "Modo Galáctica Scale")
	socks5Port := flag.Int("socks5", 10807, "Puerto gateway universal")
	tunMode := flag.Bool("tun", false, "Modo TUN/TAP")
	captureWeb := flag.Bool("capture-web", false, "Captura web inmediata")
	noElevate := flag.Bool("no-elevate", false, "No auto-elevar")
	debugMode := flag.Bool("debug", false, "Depuración")
	logPath := flag.String("logfile", filepath.Join("data", "ipvn7.log"), "Ruta log")
	webPort := flag.Int("web-port", 7070, "Puerto WebUI")
	showVersion := flag.Bool("version", false, "Muestra versión")
	flag.Parse()

	if *showVersion {
		info := core.GetVersionInfo()
		fmt.Printf("ipvn7 %s (Build: %s, WireVersion: 0x%02x, %s/%s)\n", info.Version, info.BuildVersion, info.WireVersion, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if err := core.InitLogger(*logPath, *debugMode); err != nil {
		fmt.Fprintf(os.Stderr, "[AVISO] Log en %s: %v\n", *logPath, err)
	}
	defer core.CloseLogger()
	core.LogInfo("=== ipvn7 NOS v0.7.0 (PID: %d, Admin: %v, Plataforma: %s) ===", os.Getpid(), core.IsElevated(), runtime.GOOS)

	if runtime.GOOS == "windows" && !core.IsElevated() && !*noElevate {
		elevatedArgs := append(os.Args[1:], "-tun", "-no-elevate")
		if *debugMode { elevatedArgs = append(elevatedArgs, "-debug") }
		if core.RequestSelfElevation(elevatedArgs) {
			os.Exit(0)
		}
	}
	if core.IsElevated() {
		core.TerminateConflictingProcesses(nil)
		if !*tunMode { *tunMode = true }
	}

	var identity *l0.Identity
	var err error
	if identity, err = l0.LoadFromFile(*keystorePath); err != nil {
		if identity, err = l0.GenerateIdentity(); err != nil {
			fmt.Fprintf(os.Stderr, "[FATAL] Entropía: %v\n", err); os.Exit(1)
		}
		_ = identity.SaveToFile(*keystorePath)
	}

	router := l1.NewKleinbergRouter(identity)
	firewall := l1.NewZTNAFirewall(true)
	hybridKeys, err := l1.GenerateHybridKeyPair(identity.DID())
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL PQC] No se pudieron generar claves híbridas ML-KEM-768: %v\n", err)
		os.Exit(1)
	}
	sessionMgr := l1.NewPQCSessionManager(identity, hybridKeys, firewall)
	antiReplay := l1.NewAntiReplayFilter(nil)

	if *galacticMode {
		fmt.Println("[GALACTIC MODE] Modo galáctico deshabilitado en v0.7.0")
	}

	telemetry := l2.NewTelemetryRingBuffer()
	fmt.Printf("[L0 DID]: %s | [IPv4 Virtual]: %s\n", identity.DID(), identity.IPv4())
	fmt.Printf("[L1]: Kleinberg Router (%d Anillos) | ZTNA Default-Deny\n", router.GetConfig().NumRings)
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

	// 6. Gateway SOCKS5 para captura de tráfico
	var socks5Gateway *l1.SOCKS5Gateway
	if *socks5Port > 0 {
		socks5Gateway = l1.NewSOCKS5Gateway(fmt.Sprintf("127.0.0.1:%d", *socks5Port), nil)
		socks5Gateway.LogFunc = core.LogInfo
		if err := socks5Gateway.Start(); err != nil {
			core.LogError("[ERROR SOCKS5] %v", err)
		} else {
			defer socks5Gateway.Stop()
			defer func() { _ = core.ClearWindowsUserProxy(nil) }()
			core.LogInfo("[+] SOCKS5 activo en 127.0.0.1:%d", *socks5Port)
		}
	}

	// 7. Panel de Control WebUI nativo
	shadowRegistry := l1.NewShadowDeviceRegistry()
	webUI := core.StartWebUI(*webPort, identity, router, socks5Gateway)
	if webUI != nil {
		webUI.SetShadowRegistry(shadowRegistry)
		defer webUI.Stop()
		proxyStr := fmt.Sprintf("http=127.0.0.1:%d;https=127.0.0.1:%d;socks=127.0.0.1:%d", *socks5Port, *socks5Port, *socks5Port)
		enableVPN := func() error { return core.SetWindowsUserProxy(proxyStr, nil) }
		disableVPN := func() error { return core.ClearWindowsUserProxy(nil) }
		webUI.SetCallbacks(enableVPN, disableVPN, func() { _ = disableVPN() })
		if *captureWeb {
			_ = enableVPN()
			webUI.SetVPNState("connected")
		} else {
			_ = disableVPN()
			webUI.SetVPNState("disconnected")
		}
		core.LaunchDesktopWindow(fmt.Sprintf("http://127.0.0.1:%d", *webPort))
	}

	// 8. Adaptador TUN L3 de kernel
	var tunAdapter l1.TunAdapter
	var tunRouter *l1.TUNRouter
	if *tunMode {
		var tunErr error
		tunAdapter, tunErr = l1.CreateTunAdapter(identity, true)
		if tunErr != nil || tunAdapter.IsUserspace() {
			fmt.Printf("[AVISO TUN] Interfaz kernel no disponible (%v). Espacio usuario activo.\n", tunErr)
		} else {
			defer tunAdapter.Close()
			forwarder := &simpleMeshForwarder{
				router: router, conn: conn, identity: identity,
				sessionMgr: sessionMgr, firewall: firewall, listenPort: uint16(*listenPort),
			}
			tunRouter = l1.NewTUNRouter(tunAdapter, forwarder)
			defer tunRouter.Close()
			fmt.Printf("[+] Interfaz TUN activa [%s] IP: %s\n", tunAdapter.Mode(), identity.IPv4())
		}
	}

	if *peerAddr != "" {
		if remoteAddr, err := net.ResolveUDPAddr("udp", *peerAddr); err == nil {
			initPkt := l0.NewPacket(l0.MsgTypeRoamingUpdate, identity.DID(), "", 0, nil, nil)
			_ = initPkt.SignPacket(identity)
			if raw, err := initPkt.Encode(); err == nil { _, _ = conn.WriteTo(raw, remoteAddr) }
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() { continue }
				return
			}

			if n >= 20 && binary.BigEndian.Uint32(buf[4:8]) == l1.STUNMagicCookie {
				if ip, port, err := l1.ParseSTUNBindingResponse(buf[:n], [12]byte{}); err == nil {
					discoveryEngine.SetReflexiveEndpoint(fmt.Sprintf("%s:%d", ip.String(), port))
				}
				continue
			}

			packet, err := l0.DecodePacket(buf[:n])
			if err != nil { continue }
			telemetry.RecordEvent(l2.EventRxPacket, uint32(n), 1100, 0)

			// Anti-Replay L1: aislamiento por DID y SessionID (RFC 4303)
			var sessionID uint64
			if sKeys, hasS := sessionMgr.GetSession(packet.SourceDID); hasS {
				sessionID = binary.BigEndian.Uint64(sKeys.SessionID[:8])
			}
			if packet.Sequence > 0 && !antiReplay.Accept(packet.SourceDID, sessionID, packet.Sequence, packet.Timestamp) {
				telemetry.RecordEvent(l2.EventDrop, uint32(n), 0, 0)
				continue
			}

			udpAddr, isUDP := fromAddr.(*net.UDPAddr)

			switch packet.Type {
			case l0.MsgTypeHandshakeInit:
				if respPkt, err := sessionMgr.HandleHandshakeInitPacket(packet); err == nil {
					canonicalPeer := l0.CanonicalDID(packet.SourceDID)
					if isUDP { _ = router.AddOrUpdatePeer(canonicalPeer, udpAddr, 1.0) }
					if raw, err := respPkt.Encode(); err == nil { _, _ = conn.WriteTo(raw, fromAddr) }
					fmt.Printf("[PQC] Handshake completado con %s (ML-KEM-768 FIPS 203)\n", canonicalPeer)
				}
			case l0.MsgTypeHandshakeResp:
				if err := sessionMgr.HandleHandshakeRespPacket(packet); err == nil {
					canonicalPeer := l0.CanonicalDID(packet.SourceDID)
					if isUDP { _ = router.AddOrUpdatePeer(canonicalPeer, udpAddr, 1.0) }
					fmt.Printf("[PQC] Sesión 1-RTT confirmada con %s\n", canonicalPeer)
				}
			case l0.MsgTypeRoamingUpdate:
				if packet.SourceDID != identity.DID() {
					if valid, err := packet.VerifyPacketSignature(); err == nil && valid {
						if isUDP { _ = router.AddOrUpdatePeer(packet.SourceDID, udpAddr, 1.1) }
						firewall.AuthorizeDID(&l1.DIDPolicy{DID: packet.SourceDID, AllowInbound: true, AllowOutbound: true, AllowRelay: true})
						respPkt := l0.NewPacket(l0.MsgTypeKeepAlive, identity.DID(), packet.SourceDID, 0, nil, nil)
						if raw, err := respPkt.Encode(); err == nil { _, _ = conn.WriteTo(raw, fromAddr) }
					}
				}
			case l0.MsgTypeKeepAlive:
			case l0.MsgTypeData:
				// ZTNA Default-Deny: evaluación estricta en el camino de datos principal
				decision, reason := firewall.EvaluateInbound(packet.SourceDID, uint16(*listenPort))
				if decision != l1.DecisionAccept {
					telemetry.RecordEvent(l2.EventDrop, uint32(n), 0, 0)
					core.LogWarn("[ZTNA DENEGADO] Paquete de %s descartado: %s", packet.SourceDID, reason)
					continue
				}

				if isUDP && packet.SourceDID != identity.DID() && packet.SourceDID != "" {
					_ = router.AddOrUpdatePeer(packet.SourceDID, udpAddr, 1.0)
				}

				if packet.DestDID == identity.DID() {
					payload := packet.Payload
					// Si existe sesión PQC activa, descifrar con ChaCha20-Poly1305
					if sessionMgr.HasSession(packet.SourceDID) {
						decrypted, err := sessionMgr.DecryptDataPacket(packet)
						if err != nil {
							telemetry.RecordEvent(l2.EventDrop, uint32(n), 0, 0)
							continue
						}
						payload = decrypted
					}

					if tunRouter != nil {
						_ = tunRouter.InjectFromMesh(payload)
					} else {
						fmt.Printf("[DATOS PQC] Recibido de %s (%d bytes)\n", packet.SourceDID, len(payload))
					}
				} else {
					if outDec, _ := firewall.EvaluateOutbound(packet.DestDID, uint16(*listenPort)); outDec == l1.DecisionAccept {
						if nextHop, err := router.FindNextHop(packet.DestDID); err == nil && nextHop != nil {
							conn.WriteTo(buf[:n], nextHop.Locator.PhysicalAddr)
							telemetry.RecordEvent(l2.EventTxPacket, uint32(n), 1100, 0)
						}
					}
				}
			}
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
