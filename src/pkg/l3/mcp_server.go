package l3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

// MCP Request/Response según especificación Model Context Protocol (JSON-RPC 2.0)
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool Definition para MCP
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// PetnameResolverFunc resuelve nombres mnemotécnicos dDNS
type PetnameResolverFunc func(name string) (did string, ipv6 string, ipv4 string, found bool)

// SASDeriverFunc deriva el código SAS para emparejamiento OOB
type SASDeriverFunc func(peerDID string) (digits string, emojis []string, err error)

// MCPServer expone el plano de control para asistentes de inteligencia artificial (Dimensión 6)
type MCPServer struct {
	Identity       *l0.Identity
	Router         *l1.KleinbergRouter
	Telemetry      *l2.TelemetryRingBuffer
	Firewall       *l1.ZTNAFirewall
	DAGStore       *l1.DAGStore
	WoT            *l1.WebOfTrust
	Multipath      *l1.MultipathScheduler
	Copilot        *AICopilotEngine
	DDNSResolverFn PetnameResolverFunc
	SASDeriverFn   SASDeriverFunc
	Kuzu           *KuzuGraphEngine
	HybridKeys     *l1.HybridKeyPair
	Sphinx         *l1.SphinxRouter
	SOCKS5         *l1.SOCKS5Gateway
	Guardian       *l2.ResilienceGuardian
	UIN            *l1.UINIdentityManager
	MemoryArbiter  *l1.GlobalMemoryArbiter
	Hierarchy      *l1.NodeHierarchyManager
	Constitution   *ConstitutionalVerifier
	Senate         *AgentSenateEngine
	Sentinel       *l2.SentinelImmunologyEngine
	LawEngine      *ComputationalLawEngine
	StartTime      time.Time
	mu             sync.Mutex
}

// NewMCPServer inicializa el servidor MCP con acceso a las capas inferiores
func NewMCPServer(id *l0.Identity, router *l1.KleinbergRouter, telemetry *l2.TelemetryRingBuffer) *MCPServer {
	return &MCPServer{
		Identity:  id,
		Router:    router,
		Telemetry: telemetry,
		StartTime: time.Now(),
	}
}

// AttachSenateAndSentinel enlaza los subsistemas del Senado de Agentes y Centinelas (docs/GOBERNANZA.md)
func (s *MCPServer) AttachSenateAndSentinel(constVerifier *ConstitutionalVerifier, senate *AgentSenateEngine, sentinel *l2.SentinelImmunologyEngine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Constitution = constVerifier
	s.Senate = senate
	s.Sentinel = sentinel
}

// AttachLawEngine enlaza el motor de Derecho Computable y resolución de antinomias (docs/GOBERNANZA.md)
func (s *MCPServer) AttachLawEngine(law *ComputationalLawEngine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LawEngine = law
}

// AttachLegacyRescues enlaza los componentes rescatados de D:\David (SOCKS5, Guardian)
func (s *MCPServer) AttachLegacyRescues(socks5 *l1.SOCKS5Gateway, guardian *l2.ResilienceGuardian) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SOCKS5 = socks5
	s.Guardian = guardian
}

// AttachUINStack enlaza los componentes rescatados de G:\Mi unidad\03_Programacion (UIN, Arbiter, Hierarchy)
func (s *MCPServer) AttachUINStack(uin *l1.UINIdentityManager, arbiter *l1.GlobalMemoryArbiter, hier *l1.NodeHierarchyManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UIN = uin
	s.MemoryArbiter = arbiter
	s.Hierarchy = hier
}

// AttachSubsystems enlaza los subsistemas avanzados de las Fases A, B, C y D
func (s *MCPServer) AttachSubsystems(
	fw *l1.ZTNAFirewall,
	dag *l1.DAGStore,
	wot *l1.WebOfTrust,
	mp *l1.MultipathScheduler,
	copilot *AICopilotEngine,
	ddnsFn PetnameResolverFunc,
	sasFn SASDeriverFunc,
) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Firewall = fw
	s.DAGStore = dag
	s.WoT = wot
	s.Multipath = mp
	s.Copilot = copilot
	s.DDNSResolverFn = ddnsFn
	s.SASDeriverFn = sasFn
}

// AttachKuzu enlaza el motor de grafos relacional KùzuDB
func (s *MCPServer) AttachKuzu(kuzu *KuzuGraphEngine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Kuzu = kuzu
}

// AttachHybridCrypto enlaza los módulos cuántico-resistentes y enrutamiento cebolla Sphinx
func (s *MCPServer) AttachHybridCrypto(keys *l1.HybridKeyPair, sphinx *l1.SphinxRouter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HybridKeys = keys
	s.Sphinx = sphinx
}

// GetSupportedTools retorna el catálogo de herramientas para el modelo de IA

func (s *MCPServer) HandleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "ipvn7-mcp-server",
					"version": "0.3.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]bool{"listChanged": true},
				},
			},
		}

	case "tools/list":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": s.GetSupportedTools(),
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32602, Message: "Argumentos inválidos"},
			}
		}

		result, err := s.ExecuteTool(callParams.Name, callParams.Arguments)
		if err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: -32000, Message: err.Error()},
			}
		}

		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]string{
					{
						"type": "text",
						"text": result,
					},
				},
			},
		}

	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: fmt.Sprintf("Método no soportado: %s", req.Method)},
		}
	}
}

// ExecuteTool ejecuta la herramienta requerida por el agente de IA

func (s *MCPServer) ServeStdio(reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)
	encoder := json.NewEncoder(writer)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := &JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   &RPCError{Code: -32700, Message: "Parse error JSON"},
			}
			_ = encoder.Encode(resp)
			continue
		}

		resp := s.HandleRequest(&req)
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// ServeStdioDefault ejecuta el servidor MCP sobre os.Stdin y os.Stdout
func (s *MCPServer) ServeStdioDefault() error {
	return s.ServeStdio(os.Stdin, os.Stdout)
}
