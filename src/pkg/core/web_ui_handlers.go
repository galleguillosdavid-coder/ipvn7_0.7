package core

import (
	"encoding/json"
	"net/http"
	"os"
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

func (s *WebUIServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	st, ok := s.executeConnect()
	setCORS(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": st, "ok": ok})
}

func (s *WebUIServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	st := s.executeDisconnect()
	setCORS(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": st, "ok": true})
}

func (s *WebUIServer) handleExit(w http.ResponseWriter, r *http.Request) {
	_ = ClearWindowsUserProxy(nil)
	s.mu.Lock()
	fn := s.onExit
	s.vpnState = "disconnected"
	s.mu.Unlock()
	setCORS(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "exiting", "ok": true})
	go func() {
		GracefulShutdown(fn, 300*time.Millisecond)
		os.Exit(0)
	}()
}

func (s *WebUIServer) handleCycleState(w http.ResponseWriter, r *http.Request) {
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

	setCORS(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": newState, "ok": true})
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *WebUIServer) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	manifestURL := r.URL.Query().Get("url")
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

