package l3_test

import (
	"encoding/json"
	"strings"
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l3"
)

func TestMCPSecurityAndPqcTools(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	server := l3.NewMCPServer(id, router, telemetry)

	testTools := []struct {
		name         string
		expectedSubstr string
	}{
		{
			name:         "ipvn7_pqc_status",
			expectedSubstr: "ML-KEM-768",
		},
		{
			name:         "ipvn7_sphinx_status",
			expectedSubstr: "3-Hop",
		},
		{
			name:         "ipvn7_memory_arbiter_status",
			expectedSubstr: "OOM-Immune",
		},
	}

	for _, tt := range testTools {
		t.Run(tt.name, func(t *testing.T) {
			callReq := l3.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      10,
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name": "` + tt.name + `", "arguments": {}}`),
			}
			resp := server.HandleRequest(&callReq)
			if resp.Error != nil {
				t.Fatalf("herramienta %s falló: %s", tt.name, resp.Error.Message)
			}
			outMap, ok := resp.Result.(map[string]interface{})
			if !ok {
				t.Fatalf("resultado de %s no es un mapa: %v", tt.name, resp.Result)
			}
			content, ok := outMap["content"].([]map[string]string)
			if !ok || len(content) == 0 {
				t.Fatalf("contenido inválido para %s: %v", tt.name, outMap)
			}
			text := content[0]["text"]
			if !strings.Contains(text, tt.expectedSubstr) {
				t.Fatalf("salida de %s no contiene '%s': %s", tt.name, tt.expectedSubstr, text)
			}
		})
	}
}
