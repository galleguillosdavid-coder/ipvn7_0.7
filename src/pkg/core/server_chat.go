package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ipvn7/pkg/l4"
)


// handleChatContacts retorna la lista de contactos conocidos en la malla y el agente local
func (s *CoreServer) handleChatContacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.chatManager == nil {
		_ = json.NewEncoder(w).Encode([]*l4.ChatContact{})
		return
	}

	contacts := []*l4.ChatContact{}

	// 1. Agregar pares físicos reales descubiertos por el enrutador
	seenDIDs := make(map[string]bool)
	for _, p := range s.router.GetAllPeers() {
		if p.DID == s.identity.DID() || seenDIDs[p.DID] {
			continue
		}
		seenDIDs[p.DID] = true

		ipStr := ""
		port := 0
		if p.Locator.PhysicalAddr != nil {
			ipStr = p.Locator.PhysicalAddr.IP.String()
			port = p.Locator.PhysicalAddr.Port
		}
		contactName := fmt.Sprintf("Par (%s)", ipStr)
		if ipStr == "192.168.1.106" {
			contactName = "Notebook Dvd (192.168.1.106)"
		} else if ipStr == "192.168.1.198" {
			contactName = "PC Principal (192.168.1.198)"
		}

		contacts = append(contacts, &l4.ChatContact{
			DID:      p.DID,
			Name:     contactName,
			Avatar:   "💻",
			Status:   "online",
			Endpoint: fmt.Sprintf("%s:%d", ipStr, port),
			LastSeen: p.Locator.LastSeen,
		})
	}

	// 2. Contacto soberano del núcleo local
	hostName, _ := os.Hostname()
	if hostName == "" {
		hostName = "Nodo Soberano"
	}
	daemonContact := &l4.ChatContact{
		DID:      s.identity.DID(),
		Name:     fmt.Sprintf("%s (Núcleo Local)", hostName),
		Avatar:   "🤖",
		Status:   "online",
		Endpoint: fmt.Sprintf("inproc://127.0.0.1:%d", s.port),
		LastSeen: time.Now(),
	}
	contacts = append(contacts, daemonContact)

	_ = json.NewEncoder(w).Encode(contacts)
}

// handleChatMessages retorna el historial cronológico de mensajes con un par o canal
func (s *CoreServer) handleChatMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.chatManager == nil {
		_ = json.NewEncoder(w).Encode([]*l4.ChatMessage{})
		return
	}

	peerDID := r.URL.Query().Get("peer")
	if peerDID == "" {
		peerDID = s.identity.DID()
	}

	history := s.chatManager.GetHistory(peerDID)
	_ = json.NewEncoder(w).Encode(history)
}

// handleChatSend despacha un mensaje E2EE firmado criptográficamente y lo transmite por red si es remoto
func (s *CoreServer) handleChatSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TargetDID string `json:"target_did"`
		Text      string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		http.Error(w, "El mensaje no puede estar vacío", http.StatusBadRequest)
		return
	}
	if req.TargetDID == "" {
		req.TargetDID = s.identity.DID()
	}

	msg, err := s.chatManager.SendMessage(req.TargetDID, req.Text, "")
	if err != nil {
		http.Error(w, "Error al enviar mensaje: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Notificar en el bus SSE local
	s.gateway.PublishEvent(MeshEvent{
		Type:      "CHAT_MESSAGE_SENT",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload: map[string]interface{}{
			"id":         msg.ID,
			"target_did": msg.TargetDID,
			"author_did": msg.AuthorDID,
			"text":       msg.Text,
			"signature":  msg.Signature,
			"timestamp":  msg.Timestamp.Format(time.RFC3339),
			"delivered":  msg.Delivered,
			"latency_ms": msg.DeliveryLatencyMs,
		},
	})

	// Si el destinatario es un par remoto, transmitirlo por la red a su API
	if req.TargetDID != s.identity.DID() && req.TargetDID != "did:ipvn7:local:daemon" {
		go func(remoteDID string, m *l4.ChatMessage) {
			var targetIP string
			for _, p := range s.router.GetAllPeers() {
				if p.DID == remoteDID && p.Locator.PhysicalAddr != nil {
					targetIP = p.Locator.PhysicalAddr.IP.String()
					break
				}
			}
			if targetIP == "" {
				if strings.Contains(remoteDID, "bda3fed") || strings.Contains(remoteDID, "d45e") {
					targetIP = "192.168.1.106"
				} else if strings.Contains(remoteDID, "e9352") {
					targetIP = "192.168.1.198"
				}
			}

			if targetIP != "" {
				targetPort := 7070
				if targetIP == "192.168.1.106" {
					targetPort = 8080
				}
				for _, peer := range s.router.GetAllPeers() {
					if peer.DID == remoteDID && peer.Locator.PhysicalAddr != nil {
						if peer.Locator.PhysicalAddr.Port == 7001 {
							targetPort = 8080
						} else if peer.Locator.PhysicalAddr.Port == 7777 {
							targetPort = 7070
						}
						break
					}
				}

				ports := []int{targetPort}
				if targetPort == 7070 {
					ports = append(ports, 8080)
				} else {
					ports = append(ports, 7070)
				}

				data, _ := json.Marshal(m)
				client := &http.Client{Timeout: 1500 * time.Millisecond}
				for _, p := range ports {
					url := fmt.Sprintf("http://%s:%d/api/v1/chat/receive", targetIP, p)
					resp, postErr := client.Post(url, "application/json", bytes.NewReader(data))
					if postErr == nil && resp != nil {
						resp.Body.Close()
						break
					}
				}
			}

		}(req.TargetDID, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// handleChatReceive procesa un mensaje entrante transmitido por un par remoto a través de la red
func (s *CoreServer) handleChatReceive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var msg l4.ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if msg.Text == "" || msg.AuthorDID == "" {
		http.Error(w, "Mensaje incompleto", http.StatusBadRequest)
		return
	}

	// Almacenar en el gestor de chat local
	s.chatManager.AddInboundMessage(&msg)

	// Emitir evento SSE localmente para que el navegador del receptor lo dibuje de inmediato
	s.gateway.PublishEvent(MeshEvent{
		Type:      "CHAT_MESSAGE_RECEIVED",
		Timestamp: time.Now().UnixNano(),
		Source:    msg.AuthorDID,
		Payload: map[string]interface{}{
			"id":         msg.ID,
			"target_did": msg.TargetDID,
			"author_did": msg.AuthorDID,
			"text":       msg.Text,
			"signature":  msg.Signature,
			"timestamp":  msg.Timestamp.Format(time.RFC3339),
			"delivered":  true,
			"latency_ms": 0.5,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "delivered"})
}

