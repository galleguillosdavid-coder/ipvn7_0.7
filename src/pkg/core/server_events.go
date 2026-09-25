package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handleEventsSSE gestiona la transmisión de eventos en vivo mediante Server-Sent Events
func (s *CoreServer) handleEventsSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	subscriberID := fmt.Sprintf("sub_%d", time.Now().UnixNano())
	events := s.gateway.SubscribeEvents(subscriberID)
	defer s.gateway.UnsubscribeEvents(subscriberID)

	// Enviar pulso de inicio
	_, _ = fmt.Fprintf(w, "event: init\ndata: {\"status\":\"connected\",\"subscriber\":\"%s\"}\n\n", subscriberID)
	flusher.Flush()

	ctx := r.Context()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case ev, ok := <-events:
			if !ok {
				return
			}
			payloadJSON, _ := json.Marshal(ev)
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(payloadJSON))
			flusher.Flush()
		}
	}
}
