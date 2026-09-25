package l3

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ipvn7/pkg/l1"
)

// ExecuteTool ejecuta la herramienta requerida por el agente de IA
func (s *MCPServer) ExecuteTool(name string, args map[string]interface{}) (string, error) {
	if strings.HasPrefix(name, "ipvn7_pqc") || strings.HasPrefix(name, "ipvn7_sphinx") || strings.HasPrefix(name, "ipvn7_memory_arbiter") {
		return s.ExecuteSecurityPqcTool(name, args)
	}
	if strings.HasPrefix(name, "ipvn7_senate") || strings.HasPrefix(name, "ipvn7_sentinel") || strings.HasPrefix(name, "ipvn7_constitution") || strings.HasPrefix(name, "ipvn7_law") {
		return s.ExecuteGovernanceTool(name, args)
	}
	if name == "ipvn7_ai_diagnose" || name == "ipvn7_query_kuzu" || name == "ipvn7_nat_probe" || name == "ipvn7_guardian_health" || name == "ipvn7_explain_decision" {
		return s.ExecuteDiagnosticTool(name, args)
	}
	if strings.HasPrefix(name, "ipvn7_system") || strings.HasPrefix(name, "ipvn7_hardware") {
		return s.ExecuteSystemControlTool(name, args)
	}
	if strings.HasPrefix(name, "ipvn7_robot") || strings.HasPrefix(name, "ipvn7_drone") || strings.HasPrefix(name, "ipvn7_vehicle") {
		return s.ExecuteRobotVehicleTool(name, args)
	}
	return s.ExecuteNetworkTool(name, args)
}

// ExecuteNetworkTool ejecuta las herramientas de red, identidad, criptografía y telemetría
func (s *MCPServer) ExecuteNetworkTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "get_node_status":
		snap := s.Telemetry.Snapshot()
		res := map[string]interface{}{
			"did":          s.Identity.DID(),
			"sovereign_v6": s.Identity.IPv6().String(),
			"virtual_v4":   s.Identity.IPv4().String(),
			"uptime_sec":   time.Since(s.StartTime).Seconds(),
			"peers_count":  len(s.Router.GetAllPeers()),
			"telemetry":    snap,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "list_peers":
		peers := s.Router.GetAllPeers()
		type peerSummary struct {
			DID       string  `json:"did"`
			Ring      int     `json:"ring"`
			LatencyMs float64 `json:"latency_ms"`
			Addr      string  `json:"physical_addr"`
		}
		list := make([]peerSummary, 0, len(peers))
		for _, p := range peers {
			addrStr := "none"
			if p.Locator.PhysicalAddr != nil {
				addrStr = p.Locator.PhysicalAddr.String()
			}
			list = append(list, peerSummary{
				DID:       p.DID,
				Ring:      p.RingIndex,
				LatencyMs: p.Locator.LatencyMs,
				Addr:      addrStr,
			})
		}
		data, _ := json.MarshalIndent(list, "", "  ")
		return string(data), nil

	case "route_packet":
		destDID, ok := args["dest_did"].(string)
		if !ok || destDID == "" {
			return "", fmt.Errorf("parámetro dest_did requerido")
		}
		nextHop, err := s.Router.FindNextHop(destDID)
		if err != nil {
			return "", fmt.Errorf("error calculando ruta: %w", err)
		}
		res := map[string]interface{}{
			"target_did":    destDID,
			"next_hop_did":  nextHop.DID,
			"next_hop_ring": nextHop.RingIndex,
			"next_hop_rtt":  nextHop.Locator.LatencyMs,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_firewall_rule":
		if s.Firewall == nil {
			return "", fmt.Errorf("subsistema de cortafuegos ZTNA no inicializado")
		}
		action, _ := args["action"].(string)
		targetDID, _ := args["did"].(string)
		if targetDID == "" {
			return "", fmt.Errorf("parámetro 'did' requerido")
		}

		if action == "set" {
			inbound, _ := args["allow_inbound"].(bool)
			outbound, _ := args["allow_outbound"].(bool)
			relay, _ := args["allow_relay"].(bool)
			policy := &l1.DIDPolicy{
				DID:           targetDID,
				AllowInbound:  inbound,
				AllowOutbound: outbound,
				AllowRelay:    relay,
				CreatedAt:     time.Now(),
			}
			s.Firewall.AuthorizeDID(policy)
			res := map[string]interface{}{
				"status": "authorized",
				"policy": policy,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			return string(data), nil
		} else {
			policy, exists := s.Firewall.GetPolicy(targetDID)
			res := map[string]interface{}{
				"did":          targetDID,
				"found":        exists,
				"policy":       policy,
				"default_deny": s.Firewall.IsDefaultDeny(),
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			return string(data), nil
		}

	case "ipvn7_dag_put":
		if s.DAGStore == nil {
			return "", fmt.Errorf("subsistema DAGStore no inicializado")
		}
		payload, ok := args["payload"].(string)
		if !ok || payload == "" {
			return "", fmt.Errorf("parámetro 'payload' requerido")
		}
		targetDID, _ := args["target_did"].(string)
		block, err := s.DAGStore.PutBlock([]byte(payload), nil, targetDID)
		if err != nil {
			return "", fmt.Errorf("error guardando bloque DAG: %w", err)
		}
		res := map[string]interface{}{
			"status":     "stored",
			"cid":        block.CID,
			"author_did": block.AuthorDID,
			"target_did": block.TargetDID,
			"timestamp":  block.Timestamp,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_wot_vouch":
		if s.WoT == nil {
			return "", fmt.Errorf("subsistema Web-of-Trust no inicializado")
		}
		subjectDID, _ := args["subject_did"].(string)
		if subjectDID == "" {
			return "", fmt.Errorf("parámetro 'subject_did' requerido")
		}
		trustLevel, ok := args["trust_level"].(float64)
		if !ok {
			trustLevel = 0.8
		}
		reason, _ := args["reason"].(string)
		if reason == "" {
			reason = "MCP Agent endorsement"
		}

		vouch, err := s.WoT.SignAndIssueVouch(s.Identity, subjectDID, trustLevel, reason, 30*24*time.Hour)
		if err != nil {
			return "", fmt.Errorf("error emitiendo aval WoT: %w", err)
		}
		score, hops := s.WoT.CalculateReputation(s.Identity.DID(), subjectDID)
		res := map[string]interface{}{
			"status":           "vouched",
			"vouch":            vouch,
			"reputation_score": score,
			"hops":             hops,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_ddns_resolve":
		name, _ := args["name"].(string)
		if name == "" {
			return "", fmt.Errorf("parámetro 'name' requerido")
		}
		if s.DDNSResolverFn == nil {
			return "", fmt.Errorf("resolutor dDNS no configurado")
		}
		did, v6, v4, found := s.DDNSResolverFn(name)
		res := map[string]interface{}{
			"name":         name,
			"found":        found,
			"did":          did,
			"virtual_ipv6": v6,
			"virtual_ipv4": v4,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_sas_derive":
		peerDID, _ := args["peer_did"].(string)
		if peerDID == "" {
			return "", fmt.Errorf("parámetro 'peer_did' requerido")
		}
		if s.SASDeriverFn != nil {
			digits, emojis, err := s.SASDeriverFn(peerDID)
			if err != nil {
				return "", fmt.Errorf("error derivando SAS: %w", err)
			}
			res := map[string]interface{}{
				"peer_did": peerDID,
				"digits":   digits,
				"emojis":   emojis,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			return string(data), nil
		}
		return "", fmt.Errorf("derivador SAS no configurado")

	case "ipvn7_socks5_status":
		var stats map[string]interface{}
		if s.SOCKS5 != nil {
			stats = s.SOCKS5.GetStats()
		} else {
			stats = map[string]interface{}{
				"listen_addr": "127.0.0.1:10807",
				"is_running":  false,
				"status":      "STANDBY",
				"rfc":         "RFC 1928 (Universal SOCKS5)",
			}
		}
		data, _ := json.MarshalIndent(stats, "", "  ")
		return string(data), nil

	case "ipvn7_uin_resolve_entity":
		var passport l1.UINPassport
		if s.UIN != nil {
			passport = s.UIN.GetPassport()
		} else {
			tempUIN, _ := l1.NewUINIdentityManager(l1.IdentityModeHybrid)
			passport = tempUIN.GetPassport()
		}
		data, _ := json.MarshalIndent(passport, "", "  ")
		return string(data), nil

	case "ipvn7_validate_intent":
		did, _ := args["did"].(string)
		intentStr, _ := args["intent"].(string)
		if did == "" {
			did = s.Identity.DID()
		}
		if intentStr == "" {
			intentStr = "general"
		}

		var allowed bool
		var reason string
		if s.Hierarchy != nil {
			allowed, reason = s.Hierarchy.ValidateIntent(did, l1.IntentScope(intentStr))
		} else {
			allowed = true
			reason = "Permitido por defecto (gestor en inicialización)"
		}

		res := map[string]interface{}{
			"did":         did,
			"intent":      intentStr,
			"is_allowed":  allowed,
			"evaluation":  reason,
			"timestamp":   time.Now().UTC(),
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	default:
		return "", fmt.Errorf("herramienta de red desconocida: %s", name)
	}
}
