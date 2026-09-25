package docker_proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrBackendNotFound = errors.New("docker_proxy: servicio o backend no encontrado para el DID solicitado")
	ErrProxyClosed     = errors.New("docker_proxy: proxy inverso cerrado")
)

// ContainerRoute representa el mapeo entre una identidad soberana y un socket local de contenedor
type ContainerRoute struct {
	DID         string    // Identificador soberano: ej. "did:ipvn7:web-service"
	BackendURL  *url.URL  // Ej: http://127.0.0.1:8080
	ContainerID string    // Hash Docker identificador
	CreatedAt   time.Time
}

// SovereignDockerProxy enruta peticiones hacia microservicios y contenedores locales basándose en DIDs
type SovereignDockerProxy struct {
	listenAddr   string
	server       *http.Server
	routes       map[string]*ContainerRoute
	mu           sync.RWMutex
	closed       atomic.Bool
	requestCount atomic.Uint64
	bytesServed  atomic.Uint64
}

// NewSovereignDockerProxy inicializa el proxy inverso soberano
func NewSovereignDockerProxy(listenAddr string) *SovereignDockerProxy {
	if listenAddr == "" {
		listenAddr = "127.0.0.1:0"
	}
	p := &SovereignDockerProxy{
		listenAddr: listenAddr,
		routes:     make(map[string]*ContainerRoute),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleProxyRequest)
	mux.HandleFunc("/healthz", p.handleHealthCheck)

	p.server = &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return p
}

// RegisterRoute asocia un DID de IPVN7 a una URL de servicio local
func (p *SovereignDockerProxy) RegisterRoute(did string, backendRawURL string, containerID string) error {
	u, err := url.Parse(backendRawURL)
	if err != nil {
		return fmt.Errorf("url de backend inválida: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.routes[strings.ToLower(did)] = &ContainerRoute{
		DID:         did,
		BackendURL:  u,
		ContainerID: containerID,
		CreatedAt:   time.Now(),
	}
	return nil
}

// UnregisterRoute elimina una ruta registrada
func (p *SovereignDockerProxy) UnregisterRoute(did string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.routes, strings.ToLower(did))
}

func (p *SovereignDockerProxy) resolveRoute(req *http.Request) *ContainerRoute {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 1. Búsqueda por Host header (ej. did-ipvn7-web.ipvn7.local o DID directo)
	host := strings.Split(req.Host, ":")[0]
	if route, exists := p.routes[strings.ToLower(host)]; exists {
		return route
	}

	// 2. Búsqueda por cabecera soberana X-IPVN7-Target-DID
	if targetDID := req.Header.Get("X-IPVN7-Target-DID"); targetDID != "" {
		if route, exists := p.routes[strings.ToLower(targetDID)]; exists {
			return route
		}
	}

	return nil
}

func (p *SovereignDockerProxy) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	if p.closed.Load() {
		http.Error(w, ErrProxyClosed.Error(), http.StatusServiceUnavailable)
		return
	}

	route := p.resolveRoute(r)
	if route == nil {
		http.Error(w, ErrBackendNotFound.Error(), http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(route.BackendURL)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		http.Error(rw, fmt.Sprintf("error de comunicación con backend soberano: %v", err), http.StatusBadGateway)
	}

	p.requestCount.Add(1)
	proxy.ServeHTTP(w, r)
}

func (p *SovereignDockerProxy) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "OK: IPVN7 Sovereign Docker Reverse Proxy\n")
}

// Start arranca el listener físico del servidor proxy
func (p *SovereignDockerProxy) Start() (net.Listener, error) {
	ln, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return nil, err
	}
	p.listenAddr = ln.Addr().String()

	go func() {
		_ = p.server.Serve(ln)
	}()

	return ln, nil
}

// Addr retorna la dirección en la que está escuchando el proxy
func (p *SovereignDockerProxy) Addr() string {
	return p.listenAddr
}

// Stats retorna el conteo de peticiones y rutas activas
func (p *SovereignDockerProxy) Stats() (requests uint64, routesCount int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.requestCount.Load(), len(p.routes)
}

// Close apaga el proxy limpiamente
func (p *SovereignDockerProxy) Close() error {
	p.closed.Store(true)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}
