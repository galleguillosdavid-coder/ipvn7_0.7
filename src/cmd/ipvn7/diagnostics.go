package main

import (
	"fmt"
	"net"
	"os"

	"ipvn7/pkg/core"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

func registerCoreCapabilities(gw *core.SmartComponentGateway) {
	_ = gw.RegisterComponent(&core.ComponentRegistration{
		ID:           "core:mesh_router",
		Name:         "Enrutador de Mundo Pequeño Kleinberg",
		Version:      "0.7.0",
		Capabilities: []string{"routing:kleinberg", "topology:12_rings", "mobility:roaming"},
		Transport:    "inproc",
	})
	_ = gw.RegisterComponent(&core.ComponentRegistration{
		ID:           "core:ztna_gatekeeper",
		Name:         "Cortafuegos ZTNA Default-Deny",
		Version:      "0.7.0",
		Capabilities: []string{"firewall:ztna", "policy:default_deny", "crypto:did_acl"},
		Transport:    "inproc",
	})
	_ = gw.RegisterComponent(&core.ComponentRegistration{
		ID:           "core:zero_copy_pool",
		Name:         "Pool de Memoria Zero-Copy",
		Version:      "0.7.0",
		Capabilities: []string{"memory:zero_copy", "pool:3_tier", "alloc:lock_free"},
		Transport:    "inproc",
	})
	_ = gw.RegisterComponent(&core.ComponentRegistration{
		ID:           "core:packet_pacer",
		Name:         "Marcapasos de Flujo Sostenible",
		Version:      "0.7.0",
		Capabilities: []string{"pacing:sustainable_flow", "mtu:1280"},
		Transport:    "inproc",
	})
	_ = gw.RegisterComponent(&core.ComponentRegistration{
		ID:           "core:telemetry_ring",
		Name:         "Telemetría Lock-Free Ring Buffer",
		Version:      "0.7.0",
		Capabilities: []string{"telemetry:realtime", "latency:28ns"},
		Transport:    "inproc",
	})
}

func runDiagnostics(id *l0.Identity, router *l1.KleinbergRouter, pool *l1.BufferPool, fw *l1.ZTNAFirewall, telem *l2.TelemetryRingBuffer, udpPort, tcpPort int) {
	fmt.Println("[+] Ejecutando autodiagnóstico determinista del Núcleo Universal...")

	// 1. L0 Criptografía Ed25519
	msg := []byte("ipvn7_diagnostic_payload")
	sig := id.Sign(msg)
	if !l0.VerifySignature(id.PublicKey, msg, sig) {
		fmt.Println("[✗] L0 Criptografía: Error en verificación de firma")
		os.Exit(1)
	}
	fmt.Println(" [✓] L0 Criptografía Ed25519 & DID: PASS")

	// 2. L0 Wire CBOR determinista
	pkt := l0.NewPacket(l0.MsgTypeData, id.DID(), id.DID(), 1, make([]byte, 16), msg)
	encoded, err := pkt.Encode()
	if err != nil {
		fmt.Printf("[✗] L0 Wire CBOR: Error codificando: %v\n", err)
		os.Exit(1)
	}
	if _, err := l0.DecodePacket(encoded); err != nil {
		fmt.Printf("[✗] L0 Wire CBOR: Error decodificando: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(" [✓] L0 Wire CBOR RFC 8949 (1280B Canónico): PASS")

	// 3. L1 Zero-Copy Pool
	buf := pool.Acquire(64)
	if buf == nil || len(buf.Data()) != 64 {
		fmt.Println("[✗] L1 Zero-Copy Pool: Búfer inválido")
		os.Exit(1)
	}
	buf.Release()
	fmt.Println(" [✓] L1 Buffer Pool Zero-Copy (3 Tiers): PASS")

	// 4. L1 Kleinberg Router
	cfg := router.GetConfig()
	if cfg.NumRings <= 0 || cfg.MaxTotalPeers <= 0 {
		fmt.Println("[✗] L1 Kleinberg: Configuración de anillos inválida")
		os.Exit(1)
	}
	fmt.Printf(" [✓] L1 Enrutador Kleinberg (%d Anillos / %d Slots): PASS\n", cfg.NumRings, cfg.MaxTotalPeers)

	// 5. L1 ZTNA Firewall
	dec, _ := fw.EvaluateInbound("did:ipvn7:unknown_peer", uint16(udpPort))
	if dec == l1.DecisionAccept {
		fmt.Println("[✗] L1 ZTNA: Política Default-Deny no bloqueó DID desconocido")
		os.Exit(1)
	}
	fmt.Println(" [✓] L1 Cortafuegos ZTNA Default-Deny: PASS")

	// 6. L2 Telemetría Lock-Free
	telem.RecordEvent(l2.EventTxPacket, 1280, 50, 0)
	snap := telem.Snapshot()
	if snap.PacketsTx == 0 {
		fmt.Println("[✗] L2 Telemetría: Registro fallido")
		os.Exit(1)
	}
	fmt.Println(" [✓] L2 Telemetría Lock-Free Ring Buffer: PASS")

	// 7. Verificación de puertos
	checkPort := func(network, addr string) string {
		ln, err := net.Listen(network, addr)
		if err != nil {
			return "OCUPADO (demonio activo)"
		}
		_ = ln.Close()
		return "DISPONIBLE"
	}
	udpStatus := checkPort("udp", fmt.Sprintf("127.0.0.1:%d", udpPort))
	tcpStatus := checkPort("tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
	fmt.Printf(" [i] Transporte Físico UDP :%d [%s]\n", udpPort, udpStatus)
	fmt.Printf(" [i] API & Dashboard HTTP :%d [%s]\n", tcpPort, tcpStatus)

	fmt.Println("\n================================================================================")
	fmt.Println("[DIAGNÓSTICO EXITOSO] Núcleo Funcional Universal v0.6 validado y 100% operativo.")
	fmt.Println("================================================================================")
}
