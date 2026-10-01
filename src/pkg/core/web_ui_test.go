package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestWebUI_Endpoints(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatalf("StartWebUI fallo")
	}
	defer ui.Stop()

	// Test status endpoint con ShadowDeviceRegistry
	sreg := l1.NewShadowDeviceRegistry()
	_, _ = sreg.RegisterDevice("Test Printer", "192.168.1.50", "00:11:22:33:44:55", l1.DeviceTypePrinter, []int{9100})
	ui.SetShadowRegistry(sreg)

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	ui.handleStatus(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("esperado 200 en status, obtenido: %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("error decodificando json: %v", err)
	}
	if data["did"] != id.DID() {
		t.Errorf("DID no coincide: %v vs %v", data["did"], id.DID())
	}
	shadows, ok := data["shadow_devices"].([]interface{})
	if !ok || len(shadows) != 1 {
		t.Errorf("Esperado 1 shadow device en respuesta status, obtenido: %v", data["shadow_devices"])
	}

	// Test index endpoint
	reqIndex := httptest.NewRequest("GET", "/", nil)
	wIndex := httptest.NewRecorder()
	ui.handleIndex(wIndex, reqIndex)
	if wIndex.Code != http.StatusOK {
		t.Errorf("esperado 200 en index, obtenido: %d", wIndex.Code)
	}

	// Test VPN connect / disconnect endpoints
	connCalled, disconnCalled := false, false
	ui.SetCallbacks(
		func() error { connCalled = true; return nil },
		func() error { disconnCalled = true; return nil },
		nil,
	)

	reqConn := httptest.NewRequest("POST", "/api/v1/vpn/connect", nil)
	wConn := httptest.NewRecorder()
	ui.handleConnect(wConn, reqConn)
	if wConn.Code != http.StatusOK || !connCalled {
		t.Errorf("falla en handleConnect: code=%d, called=%v", wConn.Code, connCalled)
	}

	reqDisconn := httptest.NewRequest("POST", "/api/v1/vpn/disconnect", nil)
	wDisconn := httptest.NewRecorder()
	ui.handleDisconnect(wDisconn, reqDisconn)
	if wDisconn.Code != http.StatusOK || !disconnCalled {
		t.Errorf("falla en handleDisconnect: code=%d, called=%v", wDisconn.Code, disconnCalled)
	}

	// Test guide endpoint redirect
	reqGuide := httptest.NewRequest("GET", "/guide", nil)
	wGuide := httptest.NewRecorder()
	ui.server.Handler.ServeHTTP(wGuide, reqGuide)
	if wGuide.Code != http.StatusMovedPermanently {
		t.Errorf("esperado 301 en /guide, obtenido: %d", wGuide.Code)
	}
}

func TestWebUI_MCPEndpoint(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatalf("StartWebUI fallo")
	}
	defer ui.Stop()

	// 1. GET /mcp
	reqGet := httptest.NewRequest("GET", "/mcp", nil)
	wGet := httptest.NewRecorder()
	ui.handleMCP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("GET /mcp esperado 200, obtenido %d", wGet.Code)
	}

	// 2. OPTIONS /mcp
	reqOpt := httptest.NewRequest("OPTIONS", "/mcp", nil)
	wOpt := httptest.NewRecorder()
	ui.handleMCP(wOpt, reqOpt)
	if wOpt.Code != http.StatusOK {
		t.Errorf("OPTIONS /mcp esperado 200, obtenido %d", wOpt.Code)
	}

	// 3. POST /mcp initialize
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	reqInit := httptest.NewRequest("POST", "/mcp", strings.NewReader(initBody))
	wInit := httptest.NewRecorder()
	ui.handleMCP(wInit, reqInit)
	if wInit.Code != http.StatusOK {
		t.Errorf("POST /mcp initialize esperado 200, obtenido %d", wInit.Code)
	}

	// 4. POST /mcp tools/list
	listBody := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	reqList := httptest.NewRequest("POST", "/mcp", strings.NewReader(listBody))
	wList := httptest.NewRecorder()
	ui.handleMCP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/list esperado 200, obtenido %d", wList.Code)
	}

	// 5. POST /mcp tools/call get_network_status
	callBody := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_network_status","arguments":{}}}`
	reqCall := httptest.NewRequest("POST", "/mcp", strings.NewReader(callBody))
	wCall := httptest.NewRecorder()
	ui.handleMCP(wCall, reqCall)
	if wCall.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/call status esperado 200, obtenido %d", wCall.Code)
	}

	// 6. POST /mcp tools/call toggle_vpn
	toggleBody := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"toggle_vpn","arguments":{"action":"connect"}}}`
	reqToggle := httptest.NewRequest("POST", "/mcp", strings.NewReader(toggleBody))
	wToggle := httptest.NewRecorder()
	ui.handleMCP(wToggle, reqToggle)
	if wToggle.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/call toggle_vpn esperado 200, obtenido %d", wToggle.Code)
	}

	// 7. POST /mcp tools/call list_peers
	peersBody := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"list_peers","arguments":{}}}`
	reqPeers := httptest.NewRequest("POST", "/mcp", strings.NewReader(peersBody))
	wPeers := httptest.NewRecorder()
	ui.handleMCP(wPeers, reqPeers)
	if wPeers.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/call list_peers esperado 200, obtenido %d", wPeers.Code)
	}

	// 8. POST /mcp tools/call route_inference
	infBody := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"route_inference","arguments":{}}}`
	reqInf := httptest.NewRequest("POST", "/mcp", strings.NewReader(infBody))
	wInf := httptest.NewRecorder()
	ui.handleMCP(wInf, reqInf)
	if wInf.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/call route_inference esperado 200, obtenido %d", wInf.Code)
	}

	// 9. POST /mcp tools/call verify_delegation
	delBody := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"verify_delegation","arguments":{"chain":{"root_did":"did:ipvn7:test","links":[]}}}}`
	reqDel := httptest.NewRequest("POST", "/mcp", strings.NewReader(delBody))
	wDel := httptest.NewRecorder()
	ui.handleMCP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/call verify_delegation esperado 200, obtenido %d", wDel.Code)
	}

	// 10. POST /mcp tools/call verify_mandate
	mandBody := `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"verify_mandate","arguments":{"mandate":{"payer_did":"did:ipvn7:test"}}}}`
	reqMand := httptest.NewRequest("POST", "/mcp", strings.NewReader(mandBody))
	wMand := httptest.NewRecorder()
	ui.handleMCP(wMand, reqMand)
	if wMand.Code != http.StatusOK {
		t.Errorf("POST /mcp tools/call verify_mandate esperado 200, obtenido %d", wMand.Code)
	}
}

func TestWebUI_A2AAndAgentCard(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}
	ui := StartWebUI(0, id, nil, nil)
	if ui == nil {
		t.Fatalf("StartWebUI fallo")
	}
	defer ui.Stop()

	// 1. GET /.well-known/agent-card.json
	reqCard := httptest.NewRequest("GET", "/.well-known/agent-card.json", nil)
	wCard := httptest.NewRecorder()
	ui.handleAgentCard(wCard, reqCard)
	if wCard.Code != http.StatusOK {
		t.Errorf("GET /.well-known/agent-card.json esperado 200, obtenido %d", wCard.Code)
	}
	var card map[string]interface{}
	if err := json.NewDecoder(wCard.Body).Decode(&card); err != nil {
		t.Fatalf("error decodificando agent card: %v", err)
	}
	if card["did"] != id.DID() {
		t.Errorf("DID en agent card no coincide: %v vs %v", card["did"], id.DID())
	}

	// 2. OPTIONS /a2a
	reqOpt := httptest.NewRequest("OPTIONS", "/a2a", nil)
	wOpt := httptest.NewRecorder()
	ui.handleA2A(wOpt, reqOpt)
	if wOpt.Code != http.StatusOK {
		t.Errorf("OPTIONS /a2a esperado 200, obtenido %d", wOpt.Code)
	}

	// 3. GET /a2a
	reqGet := httptest.NewRequest("GET", "/a2a", nil)
	wGet := httptest.NewRecorder()
	ui.handleA2A(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("GET /a2a esperado 200, obtenido %d", wGet.Code)
	}

	// 4. POST /a2a handshake
	handshakeBody := `{"jsonrpc":"2.0","id":1,"method":"handshake","params":{}}`
	reqHandshake := httptest.NewRequest("POST", "/a2a", strings.NewReader(handshakeBody))
	wHandshake := httptest.NewRecorder()
	ui.handleA2A(wHandshake, reqHandshake)
	if wHandshake.Code != http.StatusOK {
		t.Errorf("POST /a2a handshake esperado 200, obtenido %d", wHandshake.Code)
	}

	// 5. POST /a2a peers/discover
	peersBody := `{"jsonrpc":"2.0","id":2,"method":"peers/discover","params":{}}`
	reqPeers := httptest.NewRequest("POST", "/a2a", strings.NewReader(peersBody))
	wPeers := httptest.NewRecorder()
	ui.handleA2A(wPeers, reqPeers)
	if wPeers.Code != http.StatusOK {
		t.Errorf("POST /a2a peers esperado 200, obtenido %d", wPeers.Code)
	}

	// 6. POST /a2a tasks/send
	sendBody := `{"jsonrpc":"2.0","id":3,"method":"tasks/send","params":{"task":"sync"}}`
	reqSend := httptest.NewRequest("POST", "/a2a", strings.NewReader(sendBody))
	wSend := httptest.NewRecorder()
	ui.handleA2A(wSend, reqSend)
	if wSend.Code != http.StatusOK {
		t.Errorf("POST /a2a tasks/send esperado 200, obtenido %d", wSend.Code)
	}

	// 7. POST /a2a a2a/inference
	a2aInfBody := `{"jsonrpc":"2.0","id":4,"method":"a2a/inference","params":{}}`
	reqA2AInf := httptest.NewRequest("POST", "/a2a", strings.NewReader(a2aInfBody))
	wA2AInf := httptest.NewRecorder()
	ui.handleA2A(wA2AInf, reqA2AInf)
	if wA2AInf.Code != http.StatusOK {
		t.Errorf("POST /a2a a2a/inference esperado 200, obtenido %d", wA2AInf.Code)
	}

	// 8. POST /a2a tasks/send con cadena de delegación válida
	subAgent, _ := l0.GenerateIdentity()
	link, _ := l0.CreateDelegationLink(id, subAgent.DID(), "task:delegate", l0.TierOrdinary, 5*time.Minute)
	chainJson, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0", "id": 5, "method": "tasks/send",
		"params": map[string]interface{}{
			"task": "distributed_inference",
			"chain": l0.AgenticPrincipalChain{RootDID: id.DID(), Links: []l0.DelegationLink{*link}},
		},
	})
	reqAPC := httptest.NewRequest("POST", "/a2a", strings.NewReader(string(chainJson)))
	wAPC := httptest.NewRecorder()
	ui.handleA2A(wAPC, reqAPC)
	if wAPC.Code != http.StatusOK {
		t.Errorf("POST /a2a tasks/send con APC esperado 200, obtenido %d", wAPC.Code)
	}

	// 9. POST /a2a mandate/redeem (AP2)
	mandate, _ := l0.CreateSpendingMandate(id, subAgent.DID(), "inference:gpu_sec", 60, 5*time.Minute)
	mandateJson, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0", "id": 6, "method": "mandate/redeem",
		"params": map[string]interface{}{"mandate": mandate},
	})
	reqRedeem := httptest.NewRequest("POST", "/a2a", strings.NewReader(string(mandateJson)))
	wRedeem := httptest.NewRecorder()
	ui.handleA2A(wRedeem, reqRedeem)
	if wRedeem.Code != http.StatusOK {
		t.Errorf("POST /a2a mandate/redeem esperado 200, obtenido %d", wRedeem.Code)
	}
}
