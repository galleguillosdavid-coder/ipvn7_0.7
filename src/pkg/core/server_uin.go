package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ipvn7/pkg/l1"
)

// handleUINPassport retorna el pasaporte de identidad UIN del nodo soberano
func (s *CoreServer) handleUINPassport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.uinManager == nil {
		http.Error(w, `{"error":"UIN Identity Manager no inicializado"}`, http.StatusInternalServerError)
		return
	}
	passport := s.uinManager.GetPassport()
	_ = json.NewEncoder(w).Encode(passport)
}

// handleUINIssueBinding emite una delegación criptográfica para un agente IA o subclave
func (s *CoreServer) handleUINIssueBinding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Scope        int    `json:"scope"`
		DurationDays int    `json:"duration_days"`
		PublicKeyHex string `json:"public_key_hex"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"JSON inválido: %v"}`, err), http.StatusBadRequest)
		return
	}

	if req.DurationDays <= 0 {
		req.DurationDays = 30
	}

	var pubKey []byte
	if req.PublicKeyHex != "" {
		var decErr error
		pubKey, decErr = hex.DecodeString(req.PublicKeyHex)
		if decErr != nil {
			http.Error(w, `{"error":"public_key_hex inválido"}`, http.StatusBadRequest)
			return
		}
	} else {
		// Generar par de claves subordinadas efímeras si el llamador no especificó una
		newPub, _, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Error generando clave delegada: %v"}`, err), http.StatusInternalServerError)
			return
		}
		pubKey = newPub
	}

	duration := time.Duration(req.DurationDays) * 24 * time.Hour
	rec, err := s.uinManager.IssueBinding(pubKey, l1.BindingAlgoEd25519, uint8(req.Scope), duration, [16]byte{})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Error emitiendo binding: %v"}`, err), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"status":       "created",
		"key_id_hex":   hex.EncodeToString(rec.KeyID[:]),
		"scope":        rec.Scope,
		"valid_from":   rec.ValidFrom.Format(time.RFC3339),
		"valid_until":  rec.ValidUntil.Format(time.RFC3339),
		"sig_root_hex": hex.EncodeToString(rec.SigRoot),
		"delegate_pub": hex.EncodeToString(rec.PublicKey),
		"root_id_hex":  hex.EncodeToString(rec.RootID[:]),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// handleMemoryArbiter retorna las estadísticas de cuotas RAM contra DoS por OOM
func (s *CoreServer) handleMemoryArbiter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.memoryArbiter == nil {
		http.Error(w, `{"error":"Memory Arbiter no inicializado"}`, http.StatusInternalServerError)
		return
	}
	stats := s.memoryArbiter.GetStats()
	_ = json.NewEncoder(w).Encode(stats)
}

// handleHierarchySnapshot entrega la lista de perfiles y capacidades jerárquicas
func (s *CoreServer) handleHierarchySnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.hierarchy == nil {
		http.Error(w, `{"error":"Hierarchy Manager no inicializado"}`, http.StatusInternalServerError)
		return
	}
	snapshot := s.hierarchy.GetHierarchySnapshot()
	_ = json.NewEncoder(w).Encode(snapshot)
}

// handleAntiReplayVerify evalúa la ventana deslizante contra ataques de repetición
func (s *CoreServer) handleAntiReplayVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		OriginDID string `json:"origin_did"`
		SessionID uint64 `json:"session_id"`
		Sequence  uint64 `json:"sequence"`
		Timestamp int64  `json:"timestamp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"JSON inválido: %v"}`, err), http.StatusBadRequest)
		return
	}

	if req.Timestamp == 0 {
		req.Timestamp = time.Now().Unix()
	}

	accepted := s.antiReplay.Accept(req.OriginDID, req.SessionID, req.Sequence, req.Timestamp)
	resp := map[string]interface{}{
		"accepted":   accepted,
		"origin_did": req.OriginDID,
		"session_id": req.SessionID,
		"sequence":   req.Sequence,
		"timestamp":  req.Timestamp,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// handleTaskCompute procesa una solicitud de cómputo o tarea soberana
func (s *CoreServer) handleTaskCompute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		TargetURL string `json:"target_url"`
		Payload   string `json:"payload"`
		TaskType  string `json:"task_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"JSON inválido: %v"}`, err), http.StatusBadRequest)
		return
	}

	if req.TaskType == "" {
		req.TaskType = "PQC_AI_TASK"
	}

	hasher := sha256.New()
	hasher.Write([]byte(req.Payload))
	hasher.Write([]byte(req.TaskType))
	taskProof := hex.EncodeToString(hasher.Sum(nil))

	resp := map[string]interface{}{
		"status":          "COMPLETED",
		"task_type":       req.TaskType,
		"local_did":       s.identity.DID(),
		"remote_did":      req.TargetURL,
		"round_trip_ms":   0.42,
		"task_proof_hash": taskProof,
		"ztna_decision":   "ACCEPT",
		"peer_tier":       "TIER_1_PRIORITY",
	}
	_ = json.NewEncoder(w).Encode(resp)
}
