package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)


// handlePeersConnect empareja físicamente este nodo con una dirección IP:puerto remota
func (s *CoreServer) handlePeersConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PeerAddress string `json:"peer_address"`
		Reciprocal  bool   `json:"reciprocal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.PeerAddress == "" {
		http.Error(w, "peer_address requerido (ej: 192.168.1.106:7001)", http.StatusBadRequest)
		return
	}

	host, portStr, err := net.SplitHostPort(req.PeerAddress)
	if err != nil {
		http.Error(w, "formato ip:puerto inválido", http.StatusBadRequest)
		return
	}
	port, _ := strconv.Atoi(portStr)

	// Consultar el status del peer para obtener su DID real
	targetWebPort := 8080
	if port == 7777 {
		targetWebPort = 7070
	} else if port == 7001 {
		targetWebPort = 8080
	}

	client := &http.Client{Timeout: 3 * time.Second}
	statusURL := fmt.Sprintf("http://%s:%d/api/v1/status", host, targetWebPort)
	resp, err := client.Get(statusURL)
	if err != nil {
		// Intentar puerto web alternativo
		if targetWebPort == 8080 {
			targetWebPort = 7070
		} else {
			targetWebPort = 8080
		}
		statusURL = fmt.Sprintf("http://%s:%d/api/v1/status", host, targetWebPort)
		resp, err = client.Get(statusURL)
	}

	remoteDID := ""
	if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
		var st CoreStatus
		if json.NewDecoder(resp.Body).Decode(&st) == nil {
			remoteDID = st.DID
		}
		resp.Body.Close()
	}

	if remoteDID == "" {
		remoteDID = findDIDInTrustedPeers(host)
	}

	if remoteDID == "" || remoteDID == s.identity.DID() {
		http.Error(w, "No se pudo resolver el DID del par remoto", http.StatusBadGateway)
		return
	}

	// Limpiar pares anteriores con la misma IP para evitar duplicados o fantasmas
	for _, p := range s.router.GetAllPeers() {
		if p.Locator.PhysicalAddr != nil && p.Locator.PhysicalAddr.IP.String() == host && p.DID != remoteDID {
			s.router.RemovePeer(p.DID)
		}
	}

	// Agregar o actualizar el par en el enrutador Kleinberg
	udpAddr := &net.UDPAddr{IP: net.ParseIP(host), Port: port}
	_ = s.router.AddOrUpdatePeer(remoteDID, udpAddr, 1.2)

	// Emparejamiento recíproco para que el otro nodo también registre a este nodo
	if req.Reciprocal {
		go func() {
			myHost := "192.168.1.198"
			if host == "192.168.1.198" {
				myHost = "192.168.1.106"
			}
			myUDPPort := 7777
			if s.port == 8080 {
				myUDPPort = 7001
			}
			pairReq := map[string]interface{}{
				"peer_address": fmt.Sprintf("%s:%d", myHost, myUDPPort),
				"reciprocal":   false,
			}
			data, _ := json.Marshal(pairReq)
			recipURL := fmt.Sprintf("http://%s:%d/api/v1/peers/connect", host, targetWebPort)
			recipResp, rErr := client.Post(recipURL, "application/json", bytes.NewReader(data))
			if rErr == nil && recipResp != nil {
				recipResp.Body.Close()
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "connected",
		"peer":   remoteDID,
		"addr":   req.PeerAddress,
	})
}

func findDIDInTrustedPeers(host string) string {
	paths := []string{
		filepath.Join("keystore", "trusted_peers.json"),
		`C:\ipvn7\keystore\trusted_peers.json`,
		filepath.Join("..", "keystore", "trusted_peers.json"),
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil || len(b) == 0 {
			continue
		}
		var peers []struct {
			DID       string   `json:"did"`
			Endpoints []string `json:"endpoints"`
		}
		if json.Unmarshal(b, &peers) == nil {
			for _, peer := range peers {
				for _, ep := range peer.Endpoints {
					if strings.Contains(ep, host) {
						return peer.DID
					}
				}
				if len(peers) == 1 {
					return peer.DID
				}
			}
		}
	}
	return ""
}
