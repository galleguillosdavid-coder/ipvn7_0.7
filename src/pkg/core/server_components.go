package core

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

func (s *CoreServer) handleComponents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	comps := s.gateway.ListComponents()
	_ = json.NewEncoder(w).Encode(comps)
}

func (s *CoreServer) handleRegisterComponent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var reg ComponentRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if reg.ID == "" || reg.Name == "" {
		http.Error(w, "ID y Name son requeridos", http.StatusBadRequest)
		return
	}

	if err := s.gateway.RegisterComponent(&reg); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "registered", "id": reg.ID})
}

func (s *CoreServer) handleUnregisterComponent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if err := s.gateway.UnregisterComponent(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "unregistered", "id": req.ID})
}

func (s *CoreServer) handleToggleComponent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID        string `json:"id"`
		Action    string `json:"action"`
		Installed *bool  `json:"installed,omitempty"`
		Enabled   *bool  `json:"enabled,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		http.Error(w, "ID requerido", http.StatusBadRequest)
		return
	}

	comps := s.gateway.ListComponents()
	var target *ComponentRegistration
	for _, c := range comps {
		if c.ID == req.ID {
			target = c
			break
		}
	}

	if target == nil {
		target = &ComponentRegistration{
			ID:           req.ID,
			Name:         req.ID,
			Version:      "0.7.0",
			Transport:    "inproc",
			Capabilities: []string{"plugin"},
			Installed:    true,
			Enabled:      true,
		}
		_ = s.gateway.RegisterComponent(target)
	}

	switch req.Action {
	case "install":
		_ = s.gateway.SetComponentInstalled(req.ID, true)
	case "uninstall":
		_ = s.gateway.SetComponentInstalled(req.ID, false)
	case "power_on":
		_ = s.gateway.SetComponentEnabled(req.ID, true)
	case "power_off":
		_ = s.gateway.SetComponentEnabled(req.ID, false)
	case "toggle_power":
		_ = s.gateway.SetComponentEnabled(req.ID, !target.Enabled)
	case "toggle_install":
		_ = s.gateway.SetComponentInstalled(req.ID, !target.Installed)
	default:
		if req.Installed != nil {
			_ = s.gateway.SetComponentInstalled(req.ID, *req.Installed)
		}
		if req.Enabled != nil {
			_ = s.gateway.SetComponentEnabled(req.ID, *req.Enabled)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "updated",
		"id":         req.ID,
		"components": s.gateway.ListComponents(),
	})
}

func (s *CoreServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if err := s.gateway.Heartbeat(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive", "id": req.ID})
}

func (s *CoreServer) handleSendDatagram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var raw struct {
		TargetDID string          `json:"target_did"`
		SourceDID string          `json:"source_did,omitempty"`
		Protocol  string          `json:"protocol"`
		Payload   json.RawMessage `json:"payload"`
		Priority  uint8           `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if raw.TargetDID == "" {
		http.Error(w, "TargetDID requerido", http.StatusBadRequest)
		return
	}

	var payloadBytes []byte
	if len(raw.Payload) > 0 {
		if raw.Payload[0] == '"' {
			var str string
			if json.Unmarshal(raw.Payload, &str) == nil {
				if dec, err := base64.StdEncoding.DecodeString(str); err == nil && len(dec) > 0 {
					payloadBytes = dec
				} else {
					payloadBytes = []byte(str)
				}
			}
		} else {
			_ = json.Unmarshal(raw.Payload, &payloadBytes)
		}
	}

	env := DatagramEnvelope{
		TargetDID: raw.TargetDID,
		SourceDID: raw.SourceDID,
		Protocol:  raw.Protocol,
		Payload:   payloadBytes,
		Priority:  raw.Priority,
	}

	if err := s.gateway.SendDatagram(&env); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "dispatched", "target": env.TargetDID})
}
