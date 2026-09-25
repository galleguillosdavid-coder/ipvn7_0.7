package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l3"
	"ipvn7/pkg/l4"
)

func handleFirewallCommand(firewall *l1.ZTNAFirewall) {
	fmt.Println("=== CORTAFUEGOS DE MICRO-SEGMENTACIÓN ZTNA (Dimensión 1) ===")
	subCmd := "list"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "test":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli firewall test <did>")
			return
		}
		targetDID := os.Args[3]
		dec, reason := firewall.EvaluateInbound(targetDID, 7001)
		fmt.Printf("Evaluando DID : %s\n", targetDID)
		fmt.Printf("Decisión      : %s (%s)\n", dec, reason)
	case "allow":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli firewall allow <did>")
			return
		}
		targetDID := os.Args[3]
		firewall.AuthorizeDID(&l1.DIDPolicy{
			DID:           targetDID,
			AllowInbound:  true,
			AllowOutbound: true,
			AllowRelay:    true,
			CreatedAt:     time.Now(),
		})
		fmt.Printf("[+] DID Autorizado en libreta ZTNA: %s\n", targetDID)
	case "list":
		fallthrough
	default:
		fmt.Printf("Política Global : Default-Deny (%v)\n", firewall.IsDefaultDeny())
		policies := firewall.GetAllPolicies()
		fmt.Printf("Reglas activas  : %d\n", len(policies))
		for _, p := range policies {
			fmt.Printf("  - DID: %s [Inbound:%v | Outbound:%v | Relay:%v]\n",
				p.DID, p.AllowInbound, p.AllowOutbound, p.AllowRelay)
		}
	}
}

func handleDAGCommand(dagStore *l1.DAGStore) {
	fmt.Println("=== ALMACÉN INMUTABLE DAG & COLA DTN (Dimensión 4) ===")
	subCmd := "status"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "put":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli dag put <payload> [target_did]")
			return
		}
		payload := os.Args[3]
		targetDID := ""
		if len(os.Args) >= 5 {
			targetDID = os.Args[4]
		}
		blk, err := dagStore.PutBlock([]byte(payload), nil, targetDID)
		if err != nil {
			fmt.Printf("[-] Error al crear bloque: %v\n", err)
			return
		}
		fmt.Printf("[+] Bloque creado exitosamente en DAG:\n")
		fmt.Printf("    CID        : %s\n", blk.CID)
		fmt.Printf("    Autor      : %s\n", blk.AuthorDID)
		fmt.Printf("    Destino    : %s\n", blk.TargetDID)
		fmt.Printf("    Timestamp  : %s\n", blk.Timestamp.Format(time.RFC3339))
	case "status":
		fallthrough
	default:
		fmt.Printf("Bloques almacenados : %d\n", dagStore.BlocksStored)
		fmt.Printf("Paquetes encolados  : %d\n", dagStore.BundlesQueued)
		fmt.Printf("Bloques totales     : %d\n", len(dagStore.ListBlocks(0)))
	}
}

func handleAccountingCommand(accounting *l2.TransitAccounting) {
	fmt.Println("=== ECONOMÍA DE RECIPROCIDAD TIT-FOR-TAT (Dimensión 5) ===")
	stats := accounting.Stats()
	fmt.Printf("Pares auditados    : %d\n", stats.TotalPeersTracked)
	fmt.Printf("Nivel PRIORITY     : %d\n", stats.PriorityPeers)
	fmt.Printf("Nivel NORMAL       : %d\n", stats.NormalPeers)
	fmt.Printf("Nivel BEST_EFFORT  : %d\n", stats.BestEffortPeers)
	fmt.Printf("Nivel THROTTLED    : %d\n", stats.ThrottledPeers)
	fmt.Printf("Total Estrangulado : %d sanciones\n", stats.TotalThrottled)
}

func handleWOTCommand(wot *l1.WebOfTrust, id *l0.Identity) {
	fmt.Println("=== RED SOCIAL CRIPTOGRÁFICA WEB-OF-TRUST (Dimensión 9) ===")
	subCmd := "status"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "vouch":
		if len(os.Args) < 5 {
			fmt.Println("Uso: ipvn7-cli wot vouch <subject_did> <trust_level 0.0-1.0> [razon]")
			return
		}
		targetDID := os.Args[3]
		level, _ := strconv.ParseFloat(os.Args[4], 64)
		reason := "Atestación directa por CLI"
		if len(os.Args) >= 6 {
			reason = os.Args[5]
		}
		vouch, err := wot.SignAndIssueVouch(id, targetDID, level, reason, 30*24*time.Hour)
		if err != nil {
			fmt.Printf("[-] Error emitiendo aval: %v\n", err)
			return
		}
		score, hops := wot.CalculateReputation(id.DID(), targetDID)
		fmt.Printf("[+] Aval emitido con firma Ed25519:\n")
		fmt.Printf("    Sujeto     : %s\n", vouch.SubjectDID)
		fmt.Printf("    Confianza  : %.2f\n", vouch.TrustLevel)
		fmt.Printf("    Reputación : %.2f / 100 (%d saltos)\n", score, hops)
	case "score":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli wot score <target_did>")
			return
		}
		targetDID := os.Args[3]
		score, hops := wot.CalculateReputation(id.DID(), targetDID)
		fmt.Printf("Reputación calculada para %s:\n", targetDID)
		fmt.Printf("  Puntaje Atenuado: %.2f / 100\n", score)
		fmt.Printf("  Distancia social: %d saltos\n", hops)
	default:
		fmt.Println("Subcomandos: wot vouch <did> <score> [razon] | wot score <did>")
	}
}

func petnamesFilePath() string {
	return filepath.Join("data", "petnames.json")
}

func loadPetnames(pr *l4.PetnameResolver) {
	data, err := os.ReadFile(petnamesFilePath())
	if err != nil {
		return
	}
	var recs []*l4.PetnameRecord
	if err := json.Unmarshal(data, &recs); err == nil {
		for _, r := range recs {
			_, _ = pr.RegisterPetname(r.Name, r.DID, r.VirtualIPv6, r.VirtualIPv4, r.ContextRoot, r.Comment)
		}
	}
}

func savePetnames(pr *l4.PetnameResolver) {
	_ = os.MkdirAll("data", 0755)
	recs := pr.ListRecords()
	data, err := json.MarshalIndent(recs, "", "  ")
	if err == nil {
		_ = os.WriteFile(petnamesFilePath(), data, 0644)
	}
}

func handleAliasCommand(petnames *l4.PetnameResolver, id *l0.Identity) {
	loadPetnames(petnames)
	if len(os.Args) < 4 {
		fmt.Println("Uso: ipvn7-cli alias <nombre.ipv7> <did:ipvn7:...> [contexto_raiz]")
		return
	}
	aliasName := os.Args[2]
	targetDID := os.Args[3]

	// Si el usuario ingresó primero el DID y luego el nombre, invertirlos asertivamente
	if strings.HasPrefix(aliasName, "did:") && !strings.HasPrefix(targetDID, "did:") {
		aliasName, targetDID = targetDID, aliasName
	}

	var ctxRoot string
	if len(os.Args) >= 5 {
		ctxRoot = os.Args[4]
	}

	// Derivar IPs soberanas para el registro si es DID válido
	pub, err := l0.PublicKeyFromDID(targetDID)
	v6 := "fd07::"
	v4 := "10.7.0.2"
	if err == nil {
		targetID := &l0.Identity{PublicKey: pub}
		v6 = targetID.IPv6().String()
		v4 = targetID.IPv4().String()
	}

	rec, err := petnames.RegisterPetname(aliasName, targetDID, v6, v4, ctxRoot, "Registrado vía CLI")
	if err != nil {
		fmt.Printf("[-] Error al registrar petname: %v\n", err)
		return
	}

	savePetnames(petnames)

	if ctxRoot != "" {
		fmt.Printf("[+] Petname Contextual Asignado: '%s' (Ámbito raíz: %s) -> %s\n",
			rec.Name, ctxRoot, targetDID)
	} else {
		fmt.Printf("[+] Petname Asignado: '%s' -> %s (Almacén local dDNS persistido)\n", rec.Name, targetDID)
	}
}

func handleDDNSCommand(petnames *l4.PetnameResolver) {
	loadPetnames(petnames)
	fmt.Println("=== RESOLUCIÓN DESCENTRALIZADA PETNAMES dDNS (Dimensión 2) ===")
	subCmd := "list"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "resolve":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli ddns resolve <nombre.ipv7>")
			return
		}
		name := os.Args[3]
		rec, found := petnames.Resolve(name, "")
		if !found {
			fmt.Printf("[-] Nombre '%s' no encontrado en el almacén local\n", name)
			return
		}
		fmt.Printf("[+] Nombre Resuelto: %s\n", rec.Name)
		fmt.Printf("    DID          : %s\n", rec.DID)
		fmt.Printf("    IPv6 Soberana: %s\n", rec.VirtualIPv6)
		fmt.Printf("    IPv4 Virtual : %s\n", rec.VirtualIPv4)
	case "export-hosts":
		hosts := petnames.ExportHostsFormat()
		fmt.Println(hosts)
	default:
		recs := petnames.ListRecords()
		fmt.Printf("Nombres registrados: %d\n", len(recs))
		for _, r := range recs {
			didShort := r.DID
			if len(didShort) > 24 {
				didShort = didShort[:24] + "..."
			}
			fmt.Printf("  - %s -> %s (%s)\n", r.Name, didShort, r.VirtualIPv6)
		}
	}
}

func handleMultipathCommand(multipath *l1.MultipathScheduler) {
	fmt.Println("=== PLANIFICADOR DE CAMINOS MÚLTIPLES (Dimensión 8) ===")
	stats := multipath.Stats()
	fmt.Printf("Enlaces registrados  : %d (Activos: %d)\n", stats.TotalInterfaces, stats.ActiveInterfaces)
	fmt.Printf("Paquetes enrutados   : %d\n", stats.PacketsRouted)
	fmt.Printf("Paquetes duplicados  : %d\n", stats.PacketsDuplicated)
	fmt.Printf("Eventos de failover  : %d\n", stats.FailoverEvents)
}

func handleSASCommand(id *l0.Identity) {
	fmt.Println("=== AUTENTICACIÓN FUERA-DE-BANDA SAS (Dimensión 10) ===")
	if len(os.Args) < 3 {
		fmt.Println("Uso: ipvn7-cli sas <peer_did>")
		return
	}
	peerDID := os.Args[2]
	pubPeer, err := l0.PublicKeyFromDID(peerDID)
	if err != nil {
		fmt.Printf("[-] Error extrayendo clave pública del DID: %v\n", err)
		return
	}
	sas := l4.DeriveSAS(id.PublicKey, pubPeer)
	fmt.Printf("Identidad Local  : %s\n", id.DID())
	fmt.Printf("Identidad Par    : %s\n", peerDID)
	fmt.Printf("Código Numérico  : [%s]\n", sas.Digits)
	fmt.Printf("Secuencia Visual : %s %s %s %s\n",
		sas.Emojis[0], sas.Emojis[1], sas.Emojis[2], sas.Emojis[3])
}

func handleCopilotCommand(copilot *l3.AICopilotEngine) {
	fmt.Println("=== COPILOTO DE IA Y AUTO-CURACIÓN (Dimensión 11) ===")
	subCmd := "diagnose"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "stats":
		st := copilot.Stats()
		fmt.Printf("Auto-curación activa: %v\n", st.AutoHealingActive)
		fmt.Printf("Diagnósticos totales: %d\n", st.TotalDiagnoses)
		fmt.Printf("Acciones aplicadas  : %d\n", st.ActionsApplied)
	case "diagnose":
		fallthrough
	default:
		anomalies := copilot.DetectAnomalies(55.0, 8)
		fmt.Printf("Escaneo de telemetría: %d anomalías detectadas\n", len(anomalies))
		for _, an := range anomalies {
			diag, _ := copilot.DiagnoseAndHeal(nil, an)
			fmt.Printf("\n[!] Anomalía: %s (%s)\n", an.Type, an.Severity)
			fmt.Printf("    Causa Raíz      : %s\n", diag.RootCause)
			fmt.Printf("    Solución        : %s\n", diag.RecommendedFix)
			fmt.Printf("    Acción Aplicada : %s (Auto-reparación: %v)\n", diag.ActionType, diag.ExecutedAction)
			fmt.Printf("    Modelo Usado    : %s (Confianza: %.2f)\n", diag.ModelUsed, diag.ConfidenceScore)
		}
	}
}
