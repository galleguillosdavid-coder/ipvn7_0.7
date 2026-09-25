package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"ipvn7/pkg/dfs"
	"ipvn7/pkg/l0"
)

func TestCoreServer_DiagnosticsAndDFSEndpoints(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ipvn7_server_dfs_test_*")
	if err != nil {
		t.Fatalf("Error creando directorio temporal: %v", err)
	}
	defer os.RemoveAll(tempDir)

	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error creando identidad: %v", err)
	}

	cs, err := dfs.NewChunkStore(filepath.Join(tempDir, "storage"))
	if err != nil {
		t.Fatalf("Error inicializando ChunkStore: %v", err)
	}

	collector := NewDistributedCollector(cs, id)
	server := &CoreServer{
		identity:   id,
		chunkStore: cs,
		collector:  collector,
	}

	// 1. Probar POST /api/v1/telemetry/report
	reportPayload := map[string]interface{}{
		"reporter_did": id.DID(),
		"timestamp":    1700000000,
		"anomaly_type": "NAT_FAILURE",
		"severity":     "CRITICAL",
		"details":      "STUN symmetric timeout en puerto UDP",
		"version":      "0.7.0",
		"os":           "windows",
		"arch":         "amd64",
		"metrics": map[string]float64{
			"timeout_ms": 3000.0,
			"retries":    5.0,
		},
	}
	body, _ := json.Marshal(reportPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/report", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.handleTelemetryReport(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handleTelemetryReport código esperado 200, obtenido %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Probar GET /api/v1/telemetry/reports
	req = httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/reports", nil)
	rec = httptest.NewRecorder()
	server.handleTelemetryReports(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handleTelemetryReports código esperado 200, obtenido %d", rec.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["total_reports"].(float64) < 1 {
		t.Fatalf("Se esperaba al menos 1 reporte registrado")
	}

	// 3. Probar POST /api/v1/dfs/upload
	uploadPayload := map[string]interface{}{
		"file_name": "patch_v0.7.1.bin",
		"content":   "contenido_binario_soberano_distribuido_en_malla",
	}
	upBody, _ := json.Marshal(uploadPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/dfs/upload", bytes.NewReader(upBody))
	rec = httptest.NewRecorder()
	server.handleDFSUpload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handleDFSUpload código esperado 200, obtenido %d: %s", rec.Code, rec.Body.String())
	}
	var upResp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &upResp)
	cid := upResp["cid"].(string)
	if cid == "" {
		t.Fatalf("handleDFSUpload no retornó CID válido")
	}

	// 4. Probar GET /api/v1/dfs/file/{cid}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/dfs/file/"+cid, nil)
	rec = httptest.NewRecorder()
	server.handleDFSFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handleDFSFile código esperado 200, obtenido %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "contenido_binario_soberano_distribuido_en_malla" {
		t.Fatalf("Contenido recuperado no coincide con el original subido")
	}
}
