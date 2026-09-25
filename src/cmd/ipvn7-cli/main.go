package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l3"
	"ipvn7/pkg/l4"
)

func printHelp() {
	fmt.Println(`ipvn7-cli: Consola de Gestión de Red Soberana ipvn7 (v0.6.0)

Uso:
  ipvn7-cli [comando] [argumentos...]

Comandos base del Núcleo Universal:
  status                     Muestra la identidad DID, IPs virtuales y estado del nodo
  peers                      Lista los pares registrados en los 12 anillos de Kleinberg
  components                 Inspecciona los componentes inteligentes acoplados al núcleo
  fsm                        Muestra la FSM de salud de los pares (Healthy, Degraded...)
  radar                      Visualización en consola de los anillos concéntricos
  pace                       Muestra la telemetría del marcapasos Packet Pacing
  services                   Lista los servicios bajo demanda (Zero Footprint)
  mcp                        Inicia el servidor MCP interactivo sobre stdio

Comandos de Innovación (Fases A, B, C, D):
  firewall [list|test <did>|allow <did>]  Cortafuegos ZTNA Default-Deny (Dimensión 1)
  dag [put <payload>|status]              Persistencia inmutable DAG y DTN (Dimensión 4)
  accounting                              Reciprocidad Tit-for-Tat (Dimensión 5)
  wot [vouch <did> <score> [razon]|score <did>]  Red social Web-of-Trust (Dimensión 9)
  alias [nombre] [did] [ctx]              Asigna petname dDNS con relatividad contextual
  ddns [resolve <name>|list|export-hosts] Resolución mnemotécnica local (Dimensión 2)
  multipath [list|strategy <mode>]        Planificador multi-camino (Dimensión 8)
  sas [peer_did]                          Código SAS y Emojis OOB (Dimensión 10)
  copilot [diagnose|stats]                Copiloto de IA y auto-curación (Dimensión 11)

Comandos de Gobernanza e Inmunología (Versión 0.5.0):
  constitution [view|verify <archivo>]    Constitución Digital Soberana (Artículo I y cláusulas)
  senate [list|propose|vote|veto|report]  Senado de Agentes y Democracia Líquida
  sentinel [status|audit <did>|alerts]    Inmunología Celular de Centinelas y Quórum Slashing
  bench <target_ip:port> [sec] [size]    Benchmark físico de saturación y jitter RFC 3550
  tunnel <puerto> [nombre_servicio]       Túnel soberano de ingreso P2P (Alternativa a Cloudflare)
  expose <puerto> [nombre_servicio]       Alias de 'tunnel' para exponer servicios locales a la malla
  diag [list|report <tipo> <grav> <det>]  Colector distribuido de diagnósticos y fallas de malla
  dfs [put <file>|get <cid>|manifest]     Almacén de Archivos Distribuido por contenido (CAS)
  version                                 Muestra la versión formal y capacidades WASM/Update
  help                                    Muestra este menú de ayuda`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]
	keystorePath := filepath.Join("keystore", "node_identity.key")

	var id *l0.Identity
	var err error
	if _, statErr := os.Stat(keystorePath); statErr == nil {
		id, err = l0.LoadFromFile(keystorePath)
		if err != nil {
			id, _ = l0.GenerateIdentity()
		}
	} else {
		id, _ = l0.GenerateIdentity()
	}

	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	pacer := l1.NewPacketPacer(l1.DefaultPacerConfig())
	firewall := l1.NewZTNAFirewall(true)
	dagStore := l1.NewDAGStore(id)
	accounting := l2.NewTransitAccounting()
	wot := l1.NewWebOfTrust()
	petnames := l4.NewPetnameResolver()
	multipath := l1.NewMultipathScheduler()
	multipath.RegisterInterface(&l1.PhysicalInterface{
		Name:      "primary-wlan",
		Type:      "WIFI",
		LatencyMs: 12.0,
		LossRate:  0.001,
		Weight:    1000,
		Active:    true,
	})
	multipath.RegisterInterface(&l1.PhysicalInterface{
		Name:      "secondary-eth",
		Type:      "ETHERNET",
		LatencyMs: 2.0,
		LossRate:  0.0001,
		Weight:    1000,
		Active:    true,
	})
	copilot := l3.NewAICopilotEngine(firewall, multipath, telemetry)

	// Gobernanza e Inmunología v0.5.0
	constVerifier := l3.NewConstitutionalVerifier(id)
	senate := l3.NewAgentSenateEngine(id, constVerifier, dagStore, wot, accounting)
	sentinel := l2.NewSentinelImmunologyEngine(id, firewall, accounting, wot)

	// Seed de propuesta demo si el senado está recién creado
	if len(senate.ListProposals()) == 0 {
		_, _ = senate.SubmitProposal(
			"Optimización Algorítmica Kleinberg L1",
			"Optimización heurística de saltos XOR para reducir latencia promedio a menos de 15ms sin alterar L0",
			l3.CategoryRoutingOptimization,
			"package main\n\nfunc OptimizeKleinbergRouting() bool {\n\treturn true\n}\n",
		)
	}

	// Inicializar gestor de servicios bajo demanda
	svcManager := l4.NewServiceLifecycleManager(5 * time.Minute)
	svcManager.RegisterDormantService("chat_e2ee")
	svcManager.RegisterDormantService("remote_desktop")
	svcManager.RegisterDormantService("dag_store")

	switch command {
	case "status":
		handleStatusCommand(id, router, telemetry)

	case "components":
		handleComponentsCommand()

	case "peers":
		handlePeersCommand(router)

	case "fsm":
		handleFSMCommand(router)

	case "pace":
		handlePaceCommand(pacer)

	case "services":
		handleServicesCommand(svcManager)

	case "radar":
		handleRadarCommand(id, router)

	case "firewall":
		handleFirewallCommand(firewall)

	case "dag":
		handleDAGCommand(dagStore)

	case "accounting":
		handleAccountingCommand(accounting)

	case "wot":
		handleWOTCommand(wot, id)

	case "alias":
		handleAliasCommand(petnames, id)

	case "ddns":
		handleDDNSCommand(petnames)

	case "multipath":
		handleMultipathCommand(multipath)

	case "sas":
		handleSASCommand(id)

	case "copilot":
		handleCopilotCommand(copilot)

	case "constitution":
		handleConstitutionCommand(constVerifier)

	case "senate":
		handleSenateCommand(senate, id)

	case "sentinel":
		handleSentinelCommand(sentinel, id)

	case "bench":
		handleBenchmarkCommand(os.Args[2:])

	case "tunnel", "expose":
		handleTunnelCommand(os.Args[2:], id)

	case "mcp":
		mcpServer := l3.NewMCPServer(id, router, telemetry)
		mcpServer.AttachSubsystems(
			firewall,
			dagStore,
			wot,
			multipath,
			copilot,
			func(name string) (string, string, string, bool) {
				rec, found := petnames.Resolve(name, "")
				if !found {
					return "", "", "", false
				}
				return rec.DID, rec.VirtualIPv6, rec.VirtualIPv4, true
			},
			func(peerDID string) (string, []string, error) {
				pubPeer, err := l0.PublicKeyFromDID(peerDID)
				if err != nil {
					return "", nil, err
				}
				sas := l4.DeriveSAS(id.PublicKey, pubPeer)
				return sas.Digits, sas.Emojis, nil
			},
		)
		mcpServer.AttachSenateAndSentinel(constVerifier, senate, sentinel)
		if err := mcpServer.ServeStdioDefault(); err != nil {
			fmt.Fprintf(os.Stderr, "Error en sesión MCP: %v\n", err)
		}

	case "diag":
		handleDiagCommand(os.Args[2:])

	case "dfs":
		handleDFSCommand(os.Args[2:])

	case "version", "--version", "-v":
		fmt.Printf("ipvn7-cli v0.7.0 (Sovereign Edition)\n")
		fmt.Printf("WebAssembly Engine: compatible (web/ipvn7.wasm / app_wasm.js)\n")
		fmt.Printf("Atomic Version Manager: activo (pkg/core/version_manager.go)\n")

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Printf("Comando desconocido: '%s'. Ejecute 'ipvn7-cli help' para opciones.\n", command)
	}
}

func init() {
	_ = json.Marshal
}
