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

func TestCoreServer_Endpoints(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false)
	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	server := NewCoreServer(0, id, router, telemetry, gw, fw, "")

	// 1. Probar GET /api/v1/status
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	w := httptest.NewRecorder()
	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200 en status, obtenido %d", w.Code)
	}

	var status CoreStatus
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("error decodificando status: %v", err)
	}
	if status.DID != id.DID() {
		t.Errorf("DID no coincide: %s vs %s", status.DID, id.DID())
	}
	if status.Version != "0.7.0" {
		t.Errorf("versión esperada 0.7.0, obtenida: %s", status.Version)
	}

	// 2. Probar registro de componente vía POST /api/v1/components/register
	compPayload := ComponentRegistration{
		ID:           "test_ext_agent",
		Name:         "External AI Sentinel",
		Version:      "1.0.0",
		Capabilities: []string{"audit", "healing"},
		Transport:    "http",
	}
	data, _ := json.Marshal(compPayload)
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/components/register", bytes.NewReader(data))
	wReg := httptest.NewRecorder()
	server.handleRegisterComponent(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("esperado 201 en registro, obtenido %d: %s", wReg.Code, wReg.Body.String())
	}

	// 3. Probar GET /api/v1/components
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/components", nil)
	wList := httptest.NewRecorder()
	server.handleComponents(wList, reqList)

	var list []*ComponentRegistration
	if err := json.NewDecoder(wList.Body).Decode(&list); err != nil {
		t.Fatalf("error decodificando lista: %v", err)
	}
	if len(list) != 1 || list[0].ID != "test_ext_agent" {
		t.Errorf("componente registrado no encontrado en lista: %+v", list)
	}

	// 4. Probar heartbeat
	heartbeatPayload := map[string]string{"id": "test_ext_agent"}
	hbData, _ := json.Marshal(heartbeatPayload)
	reqHb := httptest.NewRequest(http.MethodPost, "/api/v1/components/heartbeat", bytes.NewReader(hbData))
	wHb := httptest.NewRecorder()
	server.handleHeartbeat(wHb, reqHb)

	if wHb.Code != http.StatusOK {
		t.Fatalf("esperado 200 en heartbeat, obtenido %d", wHb.Code)
	}

	// 5. Probar desregistro
	reqUnreg := httptest.NewRequest(http.MethodPost, "/api/v1/components/unregister", bytes.NewReader(hbData))
	wUnreg := httptest.NewRecorder()
	server.handleUnregisterComponent(wUnreg, reqUnreg)

	if wUnreg.Code != http.StatusOK {
		t.Fatalf("esperado 200 en unregister, obtenido %d", wUnreg.Code)
	}
}

func TestCoreServer_UINEndpoints(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	fw := l1.NewZTNAFirewall(false)
	gw := NewSmartComponentGateway(id, router, telemetry, fw)
	defer gw.Stop()

	server := NewCoreServer(0, id, router, telemetry, gw, fw, "")

	// 1. GET /api/v1/uin/passport
	reqPass := httptest.NewRequest(http.MethodGet, "/api/v1/uin/passport", nil)
	wPass := httptest.NewRecorder()
	server.handleUINPassport(wPass, reqPass)
	if wPass.Code != http.StatusOK {
		t.Fatalf("passport status %d", wPass.Code)
	}
	var passport map[string]interface{}
	_ = json.NewDecoder(wPass.Body).Decode(&passport)
	if passport["root_id_hex"] == "" || passport["entity_did"] == "" {
		t.Errorf("passport inválido: %+v", passport)
	}

	// 2. POST /api/v1/uin/binding/create
	bindPayload := map[string]interface{}{
		"scope":         3, // AI_AGENT
		"duration_days": 15,
	}
	bData, _ := json.Marshal(bindPayload)
	reqBind := httptest.NewRequest(http.MethodPost, "/api/v1/uin/binding/create", bytes.NewReader(bData))
	wBind := httptest.NewRecorder()
	server.handleUINIssueBinding(wBind, reqBind)
	if wBind.Code != http.StatusOK {
		t.Fatalf("issue binding status %d: %s", wBind.Code, wBind.Body.String())
	}
	var binding map[string]interface{}
	_ = json.NewDecoder(wBind.Body).Decode(&binding)
	if binding["key_id_hex"] == "" || binding["sig_root_hex"] == "" {
		t.Errorf("binding inválido: %+v", binding)
	}

	// 3. GET /api/v1/memory/arbiter
	reqArb := httptest.NewRequest(http.MethodGet, "/api/v1/memory/arbiter", nil)
	wArb := httptest.NewRecorder()
	server.handleMemoryArbiter(wArb, reqArb)
	if wArb.Code != http.StatusOK {
		t.Fatalf("memory arbiter status %d", wArb.Code)
	}
	var arbStats map[string]interface{}
	_ = json.NewDecoder(wArb.Body).Decode(&arbStats)
	if arbStats["total_limit_bytes"] == nil {
		t.Errorf("memory stats sin total_limit_bytes")
	}

	// 4. GET /api/v1/hierarchy/snapshot
	reqHier := httptest.NewRequest(http.MethodGet, "/api/v1/hierarchy/snapshot", nil)
	wHier := httptest.NewRecorder()
	server.handleHierarchySnapshot(wHier, reqHier)
	if wHier.Code != http.StatusOK {
		t.Fatalf("hierarchy status %d", wHier.Code)
	}
	var hierList []map[string]interface{}
	_ = json.NewDecoder(wHier.Body).Decode(&hierList)
	if len(hierList) == 0 {
		t.Errorf("hierarchy vacía")
	}

	// 5. POST /api/v1/antireplay/verify
	replayPayload := map[string]interface{}{
		"origin_did": "did:ipvn7:test:sender",
		"session_id": 42,
		"sequence":   1,
	}
	rData, _ := json.Marshal(replayPayload)
	reqRep1 := httptest.NewRequest(http.MethodPost, "/api/v1/antireplay/verify", bytes.NewReader(rData))
	wRep1 := httptest.NewRecorder()
	server.handleAntiReplayVerify(wRep1, reqRep1)
	var repRes1 map[string]interface{}
	_ = json.NewDecoder(wRep1.Body).Decode(&repRes1)
	if repRes1["accepted"] != true {
		t.Errorf("paquete 1 debió ser aceptado: %+v", repRes1)
	}

	// Replay duplicado
	reqRep2 := httptest.NewRequest(http.MethodPost, "/api/v1/antireplay/verify", bytes.NewReader(rData))
	wRep2 := httptest.NewRecorder()
	server.handleAntiReplayVerify(wRep2, reqRep2)
	var repRes2 map[string]interface{}
	_ = json.NewDecoder(wRep2.Body).Decode(&repRes2)
	if repRes2["accepted"] != false {
		t.Errorf("paquete 2 debió ser rechazado por replay: %+v", repRes2)
	}

	// 6. POST /api/v1/task/compute
	taskPayload := map[string]interface{}{
		"target_url": "did:ipvn7:remote:node",
		"payload":    "echo sovereign message",
		"task_type":  "PQC_TEST",
	}
	tData, _ := json.Marshal(taskPayload)
	reqTask := httptest.NewRequest(http.MethodPost, "/api/v1/task/compute", bytes.NewReader(tData))
	wTask := httptest.NewRecorder()
	server.handleTaskCompute(wTask, reqTask)
	if wTask.Code != http.StatusOK {
		t.Fatalf("task compute status %d", wTask.Code)
	}
	var taskRes map[string]interface{}
	_ = json.NewDecoder(wTask.Body).Decode(&taskRes)
	if taskRes["status"] != "COMPLETED" || taskRes["task_proof_hash"] == "" {
		t.Errorf("task compute inválido: %+v", taskRes)
	}

	// 7. GET /metrics (Prometheus exporter)
	reqMet := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	wMet := httptest.NewRecorder()
	server.telemetryExporter.Handler().ServeHTTP(wMet, reqMet)
	if wMet.Code != http.StatusOK {
		t.Fatalf("esperado 200 en /metrics, obtenido %d", wMet.Code)
	}
}
