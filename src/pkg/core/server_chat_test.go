package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l4"
)

func TestCoreServer_ChatAndSpoolEndpoints(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad: %v", err)
	}

	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false)
	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	server := NewCoreServer(7070, id, router, telemetry, gw, fw, "")

	// 1. Test GET /api/v1/chat/contacts
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/chat/contacts", nil)
	rr1 := httptest.NewRecorder()
	server.handleChatContacts(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("Esperado 200 en contacts, obtenido: %d", rr1.Code)
	}
	var contacts []*l4.ChatContact
	if err := json.Unmarshal(rr1.Body.Bytes(), &contacts); err != nil || len(contacts) == 0 {
		t.Fatalf("Error deserializando contactos o lista vacía: %v", err)
	}

	// 2. Test POST /api/v1/chat/send
	payload := map[string]string{
		"target_did": id.DID(),
		"text":       "Mensaje de prueba de integración soberana",
	}
	body, _ := json.Marshal(payload)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/chat/send", bytes.NewReader(body))
	rr2 := httptest.NewRecorder()
	server.handleChatSend(rr2, req2)
	if rr2.Code != http.StatusCreated {
		t.Fatalf("Esperado 201 en chat send, obtenido: %d, body: %s", rr2.Code, rr2.Body.String())
	}

	// 3. Test GET /api/v1/chat/messages
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/chat/messages?peer="+id.DID(), nil)
	rr3 := httptest.NewRecorder()
	server.handleChatMessages(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("Esperado 200 en chat messages, obtenido: %d", rr3.Code)
	}

	// 4. Test GET /api/v1/home/print/jobs
	req4 := httptest.NewRequest(http.MethodGet, "/api/v1/home/print/jobs", nil)
	rr4 := httptest.NewRecorder()
	server.handleHomePrintJobs(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Fatalf("Esperado 200 en print jobs, obtenido: %d", rr4.Code)
	}
}
