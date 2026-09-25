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
)

func TestServer_IntentAPI(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(true)
	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	server := NewCoreServer(7070, id, router, telemetry, gw, fw, "")

	// 1. Enviar intención: "suspender equipo"
	body, _ := json.Marshal(map[string]string{
		"prompt": "Pon el PC en modo reposo",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intent", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleIntentProcess(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("código de estado esperado 200, obtenido %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error decodificando respuesta JSON: %v", err)
	}

	if resp["action"] != "sleep" {
		t.Fatalf("acción esperada sleep, obtenida: %v", resp["action"])
	}
}
