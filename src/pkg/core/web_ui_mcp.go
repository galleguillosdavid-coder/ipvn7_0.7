package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ipvn7/pkg/l0"
)

func mcpText(t string) map[string]interface{} {
	return map[string]interface{}{"content": []map[string]interface{}{{"type": "text", "text": t}}}
}

func (s *WebUIServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodGet {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"mcp_version": "2024-11-05", "server": "ipvn7-sovereign-node", "status": "ready"})
		return
	}
	var req struct {
		ID     interface{}     `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": nil, "error": map[string]interface{}{"code": -32700, "message": "Parse error"}})
		return
	}
	resp := map[string]interface{}{"jsonrpc": "2.0", "id": req.ID}
	switch req.Method {
	case "initialize":
		resp["result"] = map[string]interface{}{"protocolVersion": "2024-11-05", "capabilities": map[string]interface{}{"tools": map[string]interface{}{}}, "serverInfo": map[string]interface{}{"name": "ipvn7-sovereign-node", "version": "0.7.0"}}
	case "tools/list":
		resp["result"] = map[string]interface{}{"tools": []map[string]interface{}{
			{"name": "get_network_status", "description": "Consulta el estado físico de la red IPvN7, túnel VPN y pares.", "inputSchema": map[string]interface{}{"type": "object"}},
			{"name": "toggle_vpn", "description": "Conecta o desconecta la VPN I7 Soberana.", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"action": map[string]interface{}{"type": "string", "enum": []string{"connect", "disconnect"}}}, "required": []string{"action"}}},
			{"name": "list_peers", "description": "Lista los pares físicos conectados en la malla soberana.", "inputSchema": map[string]interface{}{"type": "object"}},
			{"name": "route_inference", "description": "Enruta una solicitud al worker óptimo de inferencia Edge AI (BitNet/Ollama).", "inputSchema": map[string]interface{}{"type": "object"}},
			{"name": "get_egress_status", "description": "Consulta el estado y métricas de la pasarela de salida a Internet soberana.", "inputSchema": map[string]interface{}{"type": "object"}},
			{"name": "verify_delegation", "description": "Valida una cadena de delegacion APC (IETF adaptive authorization).", "inputSchema": map[string]interface{}{"type": "object"}},
			{"name": "verify_mandate", "description": "Valida un mandato de gasto agéntico AP2/x402.", "inputSchema": map[string]interface{}{"type": "object"}},
		}}
	case "tools/call":
		var p struct {
			Name string                 `json:"name"`
			Args map[string]interface{} `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &p)
		switch p.Name {
		case "get_network_status":
			s.mu.RLock()
			st := s.vpnState
			s.mu.RUnlock()
			resp["result"] = mcpText(fmt.Sprintf("VPN State: %s | Peers: %d | IPv4: %s | PQC: ML-KEM-768", st, len(s.getPeerDIDs()), s.identity.IPv4()))
		case "get_egress_status":
			st := "Modo Directo Local"
			if s.gateway != nil {
				g := s.gateway.GetStats()
				if did, ok := g["active_gateway_did"].(string); ok && did != "" {
					st = fmt.Sprintf("Salida activa por pasarela soberana: %s", did)
				}
			}
			resp["result"] = mcpText(st)
		case "route_inference":
			if w := s.getWorker(); w == nil {
				resp["result"] = mcpText("No Edge AI workers found. Using local CPU.")
			} else {
				resp["result"] = mcpText(fmt.Sprintf("Routed to worker %s (DID: %s, VIP: %s, Port: %v)", w.Name, w.DID, w.VirtualIPv4, w.AllowedPorts))
			}
		case "verify_delegation":
			raw, _ := json.Marshal(p.Args["chain"])
			var chain l0.AgenticPrincipalChain
			if json.Unmarshal(raw, &chain) != nil {
				resp["result"] = mcpText("Invalid chain JSON")
			} else if tier, err := chain.VerifyChain(); err != nil {
				resp["result"] = mcpText(fmt.Sprintf("Chain Denied: %v", err))
			} else {
				resp["result"] = mcpText(fmt.Sprintf("Chain Verified: Tier=%s (Root: %s, Hops: %d)", tier, chain.RootDID, len(chain.Links)))
			}
		case "verify_mandate":
			raw, _ := json.Marshal(p.Args["mandate"])
			var m l0.SpendingMandate
			if json.Unmarshal(raw, &m) != nil || !m.Verify() {
				resp["result"] = mcpText("Mandate Invalid or Expired")
			} else {
				resp["result"] = mcpText(fmt.Sprintf("Mandate Valid: Payer=%s, Resource=%s, Units=%d", m.PayerDID, m.Resource, m.MaxUnits))
			}
		case "toggle_vpn":
			if a, _ := p.Args["action"].(string); a == "connect" {
				s.executeConnect()
			} else {
				s.executeDisconnect()
			}
			s.mu.RLock()
			cur := s.vpnState
			s.mu.RUnlock()
			resp["result"] = mcpText(fmt.Sprintf("VPN state: %s", cur))
		case "list_peers":
			dids := s.getPeerDIDs()
			resp["result"] = mcpText(fmt.Sprintf("Active peers (%d): %v", len(dids), dids))
		default:
			resp["error"] = map[string]interface{}{"code": -32602, "message": "Unknown tool"}
		}
	default:
		resp["error"] = map[string]interface{}{"code": -32601, "message": "Method not found"}
	}
	_ = json.NewEncoder(w).Encode(resp)
}
