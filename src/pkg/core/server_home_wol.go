package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ghthor/gowol"
	"ipvn7/pkg/l2"
)

func (s *CoreServer) handleHomeWoL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		entries := ReadLocalARPTable()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		MACAddress string `json:"mac_address"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	targetMAC := strings.TrimSpace(req.MACAddress)
	if targetMAC == "" {
		entries := ReadLocalARPTable()
		if len(entries) > 0 {
			targetMAC = entries[0].MAC
		}
	}
	if targetMAC == "" {
		http.Error(w, "Se requiere una dirección MAC de destino válida", http.StatusBadRequest)
		return
	}

	// Transmisión dinámica a todas las subredes IPv4 activas mediante gowol
	broadcasts := GetLocalBroadcasts()
	sentCount := 0
	for _, bcast := range broadcasts {
		cleanBcast := strings.Split(bcast, ":")[0]
		if err := gowol.MagicWake(targetMAC, cleanBcast); err == nil {
			sentCount++
		}
	}

	if sentCount == 0 {
		http.Error(w, fmt.Sprintf("No se pudo transmitir Magic Packet a %s en ninguna interfaz de red", targetMAC), http.StatusInternalServerError)
		return
	}

	s.telemetry.RecordEvent(l2.EventTxPacket, uint32(sentCount*102), 100, 0)
	s.gateway.PublishEvent(MeshEvent{
		Type:      "HOME_WOL_SENT",
		Timestamp: time.Now().UnixNano(),
		Source:    s.identity.DID(),
		Payload: map[string]interface{}{
			"mac":        targetMAC,
			"library":    "github.com/ghthor/gowol",
			"broadcasts": broadcasts,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "sent",
		"message":    fmt.Sprintf("Magic Packet WoL transmitido vía gowol a %d interfaces para encender %s", sentCount, targetMAC),
		"mac":        targetMAC,
		"library":    "github.com/ghthor/gowol",
		"bytes_sent": sentCount * 102,
	})
}
