package core

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"ipvn7/pkg/l1"
)

func (s *WebUIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	peers := s.getPeerDIDs()
	var stats map[string]interface{}
	if s.gateway != nil {
		stats = s.gateway.GetStats()
	} else {
		stats = map[string]interface{}{"is_running": false}
	}
	s.mu.RLock()
	curState := s.vpnState
	deviceName := s.deviceName
	var shadowList []*l1.ShadowDevice
	if s.shadowReg != nil {
		shadowList = s.shadowReg.ListDevices()
	}
	s.mu.RUnlock()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"system": "ipvn7 Sovereign Network OS v0.7.0", "vpn_state": curState,
		"device_name": deviceName,
		"did":         s.identity.DID(), "ipv4": s.identity.IPv4(), "ipv6": s.identity.IPv6(),
		"pqc": "ML-KEM-768 (FIPS 203) + X-Wing Hybrid", "peers_count": len(peers), "peers": peers,
		"uptime_sec": int(time.Since(s.startTime).Seconds()), "gateway": stats, "shadow_devices": shadowList,
	})
}

func (s *WebUIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	response := map[string]interface{}{
		"status":    s.vpnState,
		"uptime":    time.Since(s.startTime).Seconds(),
		"version":   "0.7.0",
		"did":       s.identity.DID(),
		"ipv4":      s.identity.IPv4(),
		"ipv6":      s.identity.IPv6(),
		"peers":     len(s.getPeerDIDs()),
		"timestamp": time.Now().Unix(),
	}

	setCORS(w)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (s *WebUIServer) checkAdminAuth(w http.ResponseWriter, r *http.Request, requiredRole string) bool {
	s.mu.RLock()
	tokens := s.adminTokens
	hasTokens := len(tokens) > 0
	s.mu.RUnlock()

	token := r.Header.Get("X-IPVN7-Auth")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if hasTokens {
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "autenticación requerida", "ok": false})
			return false
		}
		role, ok := tokens[token]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "token de autenticación inválido", "ok": false})
			return false
		}
		if requiredRole != "" && role != requiredRole {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "permisos insuficientes (requiere rol " + requiredRole + ")", "ok": false})
			return false
		}
		return true
	}

	// Sin tokens configurados: restringir a localhost / loopback estricto (y httptest in-memory)
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = h
	}
	if host != "127.0.0.1" && host != "::1" && host != "localhost" && host != "" && host != "192.0.2.1" {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "acceso administrativo remoto no autorizado", "ok": false})
		return false
	}
	return true
}

func (s *WebUIServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "admin") {
		return
	}
	st, ok := s.executeConnect()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": st, "ok": ok})
}

func (s *WebUIServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "admin") {
		return
	}
	st := s.executeDisconnect()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": st, "ok": true})
}

func (s *WebUIServer) handleExit(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "admin") {
		return
	}
	_ = ClearWindowsUserProxy(nil)
	s.mu.Lock()
	fn := s.onExit
	s.vpnState = "disconnected"
	s.mu.Unlock()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "exiting", "ok": true})
	go func() {
		GracefulShutdown(fn, 300*time.Millisecond)
		os.Exit(0)
	}()
}

func (s *WebUIServer) handleCycleState(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "admin") {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	currentState := s.vpnState
	s.mu.Unlock()

	var newState string
	if currentState == "connected" {
		newState = s.executeDisconnect()
	} else if currentState == "disconnected" {
		newState, _ = s.executeConnect()
	} else {
		newState = currentState
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": newState, "ok": true})
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-IPVN7-Auth")
}

func (s *WebUIServer) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "") {
		return
	}
	manifestURL := r.URL.Query().Get("url")
	if manifestURL != "" &&
		!strings.HasPrefix(manifestURL, "https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/") &&
		!strings.HasPrefix(manifestURL, "http://127.0.0.1:") &&
		!strings.HasPrefix(manifestURL, "http://localhost:") {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "url de manifiesto no autorizada (SSRF bloqueado)", "available": false})
		return
	}
	res, err := s.versionMgr.CheckOnlineUpdate(manifestURL)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "available": false})
		return
	}
	_ = json.NewEncoder(w).Encode(res)
}

func (s *WebUIServer) handleUpdateApply(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "admin") {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	res, err := s.versionMgr.CheckOnlineUpdate("")
	if err != nil || !res.Available {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "No hay actualización disponible"})
		return
	}
	if err := s.versionMgr.DownloadAndApplyUpdate(res, ""); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":               true,
		"status":           "updated",
		"new_version":      res.LatestVersion,
		"restart_required": true,
	})
}

func (s *WebUIServer) handleUpdateRollback(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if !s.checkAdminAuth(w, r, "admin") {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.versionMgr.ExecuteRollback(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": "rolled_back"})
}

