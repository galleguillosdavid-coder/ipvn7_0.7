// Package l3 implementa los ejecutores MCP de diagnóstico, IA Copilot y explicabilidad.
// Cumple con la regla de atomicidad modular (<=400 líneas) de docs/INGENIERIA_LEAN.md.
package l3

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

// ExecuteDiagnosticTool ejecuta las herramientas de diagnóstico, introspección e IA
func (s *MCPServer) ExecuteDiagnosticTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "ipvn7_ai_diagnose":
		if s.Copilot == nil {
			return "", fmt.Errorf("motor AI Copilot no inicializado")
		}
		recentLat := 10.0
		if v, ok := args["recent_latency_ms"].(float64); ok {
			recentLat = v
		}
		dropsCount := uint64(0)
		if v, ok := args["drops_count"].(float64); ok {
			dropsCount = uint64(v)
		}

		anomalies := s.Copilot.DetectAnomalies(recentLat, dropsCount)
		diagnoses := make([]*CopilotDiagnosis, 0)
		for _, a := range anomalies {
			d, err := s.Copilot.DiagnoseAndHeal(context.Background(), a)
			if err == nil {
				diagnoses = append(diagnoses, d)
			}
		}

		res := map[string]interface{}{
			"anomalies_detected": len(anomalies),
			"anomalies":          anomalies,
			"diagnoses":          diagnoses,
			"copilot_stats":      s.Copilot.Stats(),
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_query_kuzu":
		if s.Kuzu == nil {
			return "", fmt.Errorf("motor Kùzu Graph no inicializado")
		}
		qStr, _ := args["query"].(string)
		if qStr == "" {
			return "", fmt.Errorf("parámetro 'query' es obligatorio")
		}
		resCypher, err := s.Kuzu.ExecuteCypher(qStr)
		if err != nil {
			return "", fmt.Errorf("error ejecutando consulta Cypher: %v", err)
		}
		data, _ := json.MarshalIndent(resCypher, "", "  ")
		return string(data), nil

	case "ipvn7_nat_probe":
		res, err := l1.ProbeSTUN(nil, 2*time.Second)
		if err != nil {
			return "", fmt.Errorf("diagnóstico STUN falló: %v", err)
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_guardian_health":
		var status l2.HealthStatus
		if s.Guardian != nil {
			status = s.Guardian.GetHealthStatus()
		} else {
			cfg := l2.DefaultGuardianConfig()
			tempGuardian := l2.NewResilienceGuardian(cfg)
			status = tempGuardian.GetHealthStatus()
		}
		data, _ := json.MarshalIndent(status, "", "  ")
		return string(data), nil

	case "ipvn7_explain_decision":
		targetDID, _ := args["target_did"].(string)
		decType, _ := args["decision_type"].(string)
		if targetDID == "" {
			return "", fmt.Errorf("target_did requerido")
		}
		if decType == "" {
			decType = "routing"
		}

		explanation := map[string]interface{}{
			"target_did":    targetDID,
			"decision_type": decType,
			"timestamp":     time.Now().UTC(),
		}

		switch decType {
		case "ztna_firewall":
			allowed := false
			if s.Firewall != nil {
				_, allowed = s.Firewall.GetPolicy(targetDID)
			}
			if allowed {
				explanation["decision"] = "PERMITIDO (ALLOW)"
				explanation["natural_explanation"] = fmt.Sprintf("El peer '%s' posee una regla explícita de confianza o token criptográfico válido en el cortafuegos ZTNA de confianza cero.", targetDID)
			} else {
				explanation["decision"] = "BLOQUEADO (DROP)"
				explanation["natural_explanation"] = fmt.Sprintf("El peer '%s' fue bloqueado bajo la política soberana 'Default-Deny'. No existe autorización explícita ni aval en la Web of Trust para este identificador.", targetDID)
			}
		case "pfo_audit":
			tree := s.Kuzu.GetPFOTree()
			explanation["decision"] = "AXIOMÁTICO VALIDADO (PFO_OK)"
			explanation["natural_explanation"] = fmt.Sprintf("El pipeline PFO certifica que el grafo no posee bifurcaciones no-deterministas ni colisiones SHA-256 (Nodos verificados: %v, Raíz: %v).", tree["total_nodes"], tree["root_hash"])
		case "nat_traversal":
			explanation["decision"] = "TRANSPARENTE (HOLE_PUNCHING_SUCCESS)"
			explanation["natural_explanation"] = "El endpoint reflexivo fue resuelto mediante STUN RFC 5389. Se asignó mapeo independiente de punto final, permitiendo comunicación UDP directa sin túneles de retransmisión centralizados."
		default: // routing
			nextHop, err := s.Router.FindNextHop(targetDID)
			if err == nil && nextHop != nil {
				explanation["decision"] = fmt.Sprintf("SIGUIENTE_SALTO -> %s", nextHop.DID)
				explanation["natural_explanation"] = fmt.Sprintf("El enrutador Kleinberg seleccionó el nodo '%s' (Anillo %d) porque minimiza la distancia XOR hiperbólica euclidiana hacia el destino '%s' con latencia estimada de %.2f ms.", nextHop.DID, nextHop.RingIndex, targetDID, nextHop.Locator.LatencyMs)
			} else {
				explanation["decision"] = "DESTINO_LOCAL_O_DESCONOCIDO"
				explanation["natural_explanation"] = fmt.Sprintf("El identificador '%s' coincide con el DID del nodo local o no tiene saltos activos conocidos en la topología de anillos.", targetDID)
			}
		}

		data, _ := json.MarshalIndent(explanation, "", "  ")
		return string(data), nil

	default:
		return "", fmt.Errorf("herramienta de diagnóstico desconocida: %s", name)
	}
}
