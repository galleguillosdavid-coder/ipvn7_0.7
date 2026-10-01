package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ipvn7/pkg/l0"
)

func (s *WebUIServer) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"name": "ipvn7-network-os-agent", "version": "0.7.0", "did": s.identity.DID(),
		"protocols":    []string{"a2a/1.0", "mcp/2024-11-05", "ap2/2025-09"},
		"endpoints":    map[string]string{"a2a": "/a2a", "mcp": "/mcp", "card": "/.well-known/agent-card.json"},
		"capabilities": []string{"p2p_routing", "nat_traversal", "pqc_encryption", "ztna_shield", "physical_ai_vla", "ap2_mandates"},
		"security":     map[string]string{"pqc": "ML-KEM-768", "signature": "Ed25519", "ztna": "default-deny"},
	})
}

func (s *WebUIServer) handleA2A(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodGet {
		s.handleAgentCard(w, r)
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
	case "handshake", "a2a/handshake":
		resp["result"] = map[string]interface{}{"status": "accepted", "agent": "ipvn7-network-os-agent", "did": s.identity.DID(), "pqc": "ML-KEM-768"}
	case "peers", "peers/discover":
		peers := s.getPeerDIDs()
		resp["result"] = map[string]interface{}{"peers": peers, "count": len(peers)}
	case "a2a/inference", "inference/route":
		if worker := s.getWorker(); worker == nil {
			resp["result"] = map[string]interface{}{"status": "not_available", "message": "no edge ai worker"}
		} else {
			resp["result"] = map[string]interface{}{"status": "ready", "worker_did": worker.DID, "virtual_ip": worker.VirtualIPv4, "ports": worker.AllowedPorts}
		}
	case "delegate", "tasks/send":
		var p struct {
			Chain *l0.AgenticPrincipalChain `json:"chain"`
		}
		_ = json.Unmarshal(req.Params, &p)
		tier := l0.TierOrdinary
		if p.Chain != nil {
			var err error
			if tier, err = p.Chain.VerifyChain(); err != nil {
				resp["error"] = map[string]interface{}{"code": -32003, "message": fmt.Sprintf("Delegation denied: %v", err)}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
		}
		resp["result"] = map[string]interface{}{"status": "routed", "tier": tier, "transport": "p2p_sovereign_datagram"}
	case "mandate/redeem", "ap2/redeem":
		var p struct {
			Mandate *l0.SpendingMandate `json:"mandate"`
		}
		_ = json.Unmarshal(req.Params, &p)
		if p.Mandate == nil || !p.Mandate.Verify() {
			resp["error"] = map[string]interface{}{"code": -32004, "message": "Invalid or expired AP2 spending mandate"}
		} else {
			resp["result"] = map[string]interface{}{"status": "accepted", "resource": p.Mandate.Resource, "units": p.Mandate.MaxUnits}
		}
	default:
		resp["error"] = map[string]interface{}{"code": -32601, "message": "A2A method not found"}
	}
	_ = json.NewEncoder(w).Encode(resp)
}
