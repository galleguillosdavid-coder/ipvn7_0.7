package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l3"
)

func handleConstitutionCommand(constVerifier *l3.ConstitutionalVerifier) {
	fmt.Println("=== CONSTITUCIÓN DIGITAL SOBERANA (Dimensión 12 / v0.5.0) ===")
	subCmd := "view"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "verify":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli constitution verify <archivo_fuente.go>")
			return
		}
		filePath := os.Args[3]
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("[-] Error leyendo archivo: %v\n", err)
			return
		}
		report, err := constVerifier.VerifyCode(string(content), filePath)
		if err != nil {
			fmt.Printf("[-] Error durante análisis constitucional: %v\n", err)
			return
		}
		fmt.Printf("Objetivo auditado : %s\n", report.ProposalOrCodeID)
		fmt.Printf("Certificado       : %v\n", report.Certified)
		if len(report.SignatureHex) >= 32 {
			fmt.Printf("Firma Autoridad   : %s...\n", report.SignatureHex[:32])
		}
		if len(report.Violations) > 0 {
			fmt.Printf("\n[!] Violaciones Detectadas (%d):\n", len(report.Violations))
			for _, v := range report.Violations {
				fmt.Printf("  - Cláusula: %d | Severidad: %s\n    Detalle: %s\n",
					v.ClauseNumber, v.Severity, v.Description)
			}
		} else {
			fmt.Println("\n[+] CÓDIGO 100% CONFORME CON EL ARTÍCULO I (Privacidad Absoluta, Zero-Trust, Anti-Plutocracia)")
		}
	case "view":
		fallthrough
	default:
		fmt.Println(constVerifier.GetConstitutionText())
	}
}

func handleSenateCommand(senate *l3.AgentSenateEngine, id *l0.Identity) {
	fmt.Println("=== SENADO DE AGENTES Y DEMOCRACIA LÍQUIDA (Dimensión 13 / v0.5.0) ===")
	subCmd := "list"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "propose":
		if len(os.Args) < 5 {
			fmt.Println("Uso: ipvn7-cli senate propose <titulo> <descripcion> [categoria] [archivo_codigo]")
			return
		}
		title := os.Args[3]
		desc := os.Args[4]
		cat := l3.CategoryRoutingOptimization
		if len(os.Args) >= 6 {
			cat = l3.ProposalCategory(os.Args[5])
		}
		code := "package main\n// Código verificado conforme a la Constitución\n"
		if len(os.Args) >= 7 {
			c, err := os.ReadFile(os.Args[6])
			if err == nil {
				code = string(c)
			}
		}
		prop, err := senate.SubmitProposal(title, desc, cat, code)
		if err != nil {
			fmt.Printf("[-] Error enviando propuesta: %v\n", err)
			return
		}
		fmt.Printf("[+] Propuesta radicada con éxito:\n")
		fmt.Printf("    ID: %s\n", prop.ID)
		fmt.Printf("    Título: %s\n", prop.Title)
		fmt.Printf("    Estado: %s\n", prop.Status)
		fmt.Printf("    Certificación Constitucional: %v\n", prop.ConstitutionalReport.Certified)
	case "vote":
		if len(os.Args) < 5 {
			fmt.Println("Uso: ipvn7-cli senate vote <prop_id> <yes|no> [justificacion]")
			return
		}
		propID := os.Args[3]
		stanceStr := strings.ToLower(os.Args[4])
		stance := l3.StanceSupport
		if stanceStr == "no" || stanceStr == "oppose" {
			stance = l3.StanceOppose
		}
		just := "Voto de agente ponderado por Proof-of-Contribution"
		if len(os.Args) >= 6 {
			just = os.Args[5]
		}
		vote, err := senate.CastAgentVote(propID, stance, just)
		if err != nil {
			fmt.Printf("[-] Error al votar: %v\n", err)
			return
		}
		fmt.Printf("[+] Voto computado y firmado por agente:\n")
		fmt.Printf("    Propuesta: %s\n", vote.ProposalID)
		fmt.Printf("    Postura  : %s\n", vote.Stance)
		fmt.Printf("    Peso PoC : %.2f\n", vote.Weight)
		if len(vote.SignatureHex) >= 24 {
			fmt.Printf("    Firma    : %s...\n", vote.SignatureHex[:24])
		}
	case "veto":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli senate veto <prop_id> [razon]")
			return
		}
		propID := os.Args[3]
		reason := "Veto soberano humano inalienable ejercido vía CLI"
		if len(os.Args) >= 5 {
			reason = os.Args[4]
		}
		err := senate.SovereignHumanVeto(propID, reason)
		if err != nil {
			fmt.Printf("[-] Error al ejercer veto: %v\n", err)
			return
		}
		fmt.Printf("[!] VETO SOBERANO HUMANO EJERCIDO EXITOSAMENTE\n")
		fmt.Printf("    Propuesta: %s\n", propID)
		fmt.Printf("    Razón    : %s\n", reason)
	case "report":
		rep := senate.GenerateMorningReport()
		fmt.Printf("=== INFORME MATUTINO DE SUPERVISIÓN HUMANA (%s) ===\n", rep.Date)
		fmt.Printf("Debates activos      : %d\n", rep.TotalActiveDebates)
		fmt.Printf("Votos emitidos por IA: %d\n", rep.VotesCastByAgent)
		if len(rep.Decisions) == 0 {
			fmt.Println("(Sin decisiones pendientes ni registradas en este ciclo)")
		} else {
			for _, d := range rep.Decisions {
				vStr := "Activo"
				if d.Vetoed {
					vStr = "VETADO POR HUMANO"
				}
				fmt.Printf("  - [%s] %s | Postura: %s (Peso PoC: %.1f) | Estado: %s\n    Justificación: %s\n",
					d.ProposalID, d.Title, d.AgentStance, d.WeightUsed, vStr, d.Justification)
			}
		}
	case "list":
		fallthrough
	default:
		props := senate.ListProposals()
		fmt.Printf("Propuestas en el Senado: %d\n", len(props))
		for _, p := range props {
			fmt.Printf("\n* [%s] %s (Categoría: %s)\n", p.ID, p.Title, p.Category)
			fmt.Printf("  Estado: %s | Soporte PoC: %.1f | Oposición PoC: %.1f\n", p.Status, p.SupportWeight, p.OpposeWeight)
			fmt.Printf("  Votos de agentes: %d | Argumentos técnicos: %d\n", len(p.Votes), len(p.Arguments))
			if p.VetoReason != "" {
				fmt.Printf("  [!] Veto Soberano: %s\n", p.VetoReason)
			}
		}
	}
}

func handleSentinelCommand(sentinel *l2.SentinelImmunologyEngine, id *l0.Identity) {
	fmt.Println("=== SISTEMA INMUNOLÓGICO CELULAR DE CENTINELAS (Dimensión 14 / v0.5.0) ===")
	subCmd := "status"
	if len(os.Args) >= 3 {
		subCmd = os.Args[2]
	}
	switch subCmd {
	case "audit":
		if len(os.Args) < 4 {
			fmt.Println("Uso: ipvn7-cli sentinel audit <peer_did>")
			return
		}
		targetDID := os.Args[3]
		chal, err := sentinel.IssueAuditChallenge(targetDID)
		if err != nil {
			fmt.Printf("[-] Error al despachar desafío: %v\n", err)
			return
		}
		fmt.Printf("[+] Desafío de auditoría cruzada emitido:\n")
		fmt.Printf("    ID Desafío    : %s\n", chal.ChallengeID)
		fmt.Printf("    Nodo Auditado : %s\n", chal.TargetDID)
		fmt.Printf("    Nonce         : %s\n", chal.NonceHex)
		fmt.Printf("    Hash Esperado : %s\n", chal.ExpectedSandboxHash)
		fmt.Printf("    Timestamp     : %s\n", chal.Timestamp.Format(time.RFC3339))
	case "alerts":
		alerts := sentinel.GetActiveAlerts()
		fmt.Printf("Alertas inmunológicas activas: %d\n", len(alerts))
		for _, a := range alerts {
			fmt.Printf("\n[!] Alerta: %s\n", a.AlertID)
			fmt.Printf("    Incidente: %s | Infractor: %s\n", a.Incident, a.OffenderDID)
			fmt.Printf("    Evidencia: %s\n", a.EvidencePayload)
			fmt.Printf("    Quórum: %d firmas (Alcanzado: %v, Slashed: %v)\n", len(a.Signatures), a.QuorumReached, a.Slashed)
		}
	case "status":
		fallthrough
	default:
		st := sentinel.GetStats()
		fmt.Printf("Centinela Local       : %s\n", id.DID())
		fmt.Printf("Umbral de Quórum      : %d firmas criptográficas\n", st.QuorumThreshold)
		fmt.Printf("Auditorías Ejecutadas : %d\n", st.TotalAuditsExecuted)
		fmt.Printf("Alertas Registradas   : %d\n", st.TotalAlertsEmitted)
		fmt.Printf("Nodos Neutralizados   : %d (Slashed de por vida)\n", len(st.NeutralizedDIDs))
		if len(st.NeutralizedDIDs) > 0 {
			fmt.Println("\nLista Negra Inmunológica (Drop Default-Deny ZTNA):")
			for _, did := range st.NeutralizedDIDs {
				fmt.Printf("  - [EXPULSADO] %s\n", did)
			}
		}
	}
}
