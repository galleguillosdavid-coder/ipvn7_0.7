package core

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"ipvn7/pkg/core/templates"
	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

// WebUIServer encapsula el servidor HTTP del panel de control
type WebUIServer struct {
	port         int
	identity     *l0.Identity
	router       *l1.KleinbergRouter
	gateway      *l1.SOCKS5Gateway
	shadowReg    *l1.ShadowDeviceRegistry
	versionMgr   *VersionManager
	server       *http.Server
	mu           sync.RWMutex
	startTime    time.Time
	deviceName   string
	vpnState     string // "disconnected", "connecting", "connected"
	onConnect    func() error
	onDisconnect func() error
	onExit       func()
}

func (s *WebUIServer) SetShadowRegistry(reg *l1.ShadowDeviceRegistry) {
	s.mu.Lock()
	s.shadowReg = reg
	s.mu.Unlock()
}
func (s *WebUIServer) SetCallbacks(onConn, onDisconn func() error, onExit func()) {
	s.mu.Lock()
	s.onConnect, s.onDisconnect, s.onExit = onConn, onDisconn, onExit
	s.mu.Unlock()
}
func (s *WebUIServer) SetVPNState(state string) { s.mu.Lock(); s.vpnState = state; s.mu.Unlock() }

// StartWebUI lanza el servidor del panel interactivo en segundo plano
func StartWebUI(port int, id *l0.Identity, router *l1.KleinbergRouter, gw *l1.SOCKS5Gateway) *WebUIServer {
	if port < 0 {
		return nil
	}

	// Obtener nombre del dispositivo del sistema
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Dispositivo Desconocido"
	}

	s := &WebUIServer{
		port:       port,
		identity:   id,
		router:     router,
		gateway:    gw,
		versionMgr: NewVersionManager(""),
		startTime:  time.Now(),
		deviceName: hostname,
		vpnState:   "connected", // Iniciar en verde (conectado)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/vpn/connect", s.handleConnect)
	mux.HandleFunc("/api/v1/vpn/disconnect", s.handleDisconnect)
	mux.HandleFunc("/api/v1/vpn/exit", s.handleExit)
	mux.HandleFunc("/api/v1/vpn/cycle", s.handleCycleState)
	mux.HandleFunc("/api/v1/update/check", s.handleUpdateCheck)
	mux.HandleFunc("/api/v1/update/apply", s.handleUpdateApply)
	mux.HandleFunc("/api/v1/update/rollback", s.handleUpdateRollback)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/mcp", s.handleMCP)
	mux.HandleFunc("/a2a", s.handleA2A)
	mux.HandleFunc("/.well-known/agent-card.json", s.handleAgentCard)
	mux.Handle("/guide/", http.StripPrefix("/guide/", http.FileServer(http.Dir("guide"))))
	mux.HandleFunc("/guide", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/guide/", http.StatusMovedPermanently)
	})
	s.server = &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port), Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "[ERROR WEB] No se pudo abrir servidor HTTP en puerto %d: %v\n", port, err)
		}
	}()
	return s
}

func (s *WebUIServer) Stop() {
	if s != nil && s.server != nil {
		_ = s.server.Close()
	}
}

func LaunchDesktopWindow(url string) {
	go func() {
		time.Sleep(600 * time.Millisecond)
		switch runtime.GOOS {
		case "windows":
			for _, b := range []string{"msedge.exe", "chrome.exe"} {
				if exec.Command("cmd.exe", "/c", "start", b, fmt.Sprintf("--app=%s", url), "--window-size=480,600").Run() == nil {
					return
				}
			}
			_ = exec.Command("cmd.exe", "/c", "start", url).Run()
		case "linux":
			if os.Getenv("SSH_CLIENT") != "" || (os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "") {
				return
			}
			for _, b := range []string{"google-chrome", "chromium"} {
				if exec.Command(b, fmt.Sprintf("--app=%s", url)).Start() == nil {
					return
				}
			}
			_ = exec.Command("xdg-open", url).Start()
		case "darwin":
			_ = exec.Command("open", url).Start()
		}
	}()
}

func (s *WebUIServer) executeConnect() (string, bool) {
	s.mu.Lock()
	s.vpnState = "connecting"
	fn := s.onConnect
	s.mu.Unlock()
	var err error
	if fn != nil {
		err = fn()
	}
	s.mu.Lock()
	if err == nil {
		s.vpnState = "connected"
	} else {
		s.vpnState = "disconnected"
	}
	st := s.vpnState
	s.mu.Unlock()
	return st, err == nil
}

func (s *WebUIServer) executeDisconnect() string {
	s.mu.Lock()
	s.vpnState = "disconnecting"
	fn := s.onDisconnect
	s.mu.Unlock()
	if fn != nil {
		_ = fn()
	}
	s.mu.Lock()
	s.vpnState = "disconnected"
	s.mu.Unlock()
	_ = ClearWindowsUserProxy(nil)
	return "disconnected"
}

func (s *WebUIServer) getPeerDIDs() []string {
	if s.router == nil {
		return nil
	}
	var res []string
	for _, p := range s.router.GetAllPeers() {
		res = append(res, p.DID)
	}
	return res
}

func (s *WebUIServer) getWorker() *l1.ShadowDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.shadowReg == nil {
		return nil
	}
	return s.shadowReg.FindOptimalInferenceWorker()
}

func (s *WebUIServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(templates.WebUITemplate, s.deviceName, s.identity.DID())
	_, _ = w.Write([]byte(html))
}
