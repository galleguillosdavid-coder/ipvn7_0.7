package l1

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/interfaces"
)

// Decision representa el veredicto del motor ZTNA (interfaces.SecurityDecision)
type Decision = interfaces.SecurityDecision

const (
	DecisionAccept          = interfaces.DecisionAllow
	DecisionDropDefaultDeny = interfaces.DecisionDeny
	DecisionDropACL         = interfaces.DecisionDropACL
	DecisionDropRateLimit   = interfaces.DecisionRateLimit
)

// DIDPolicy define los permisos granulares concedidos a una identidad soberana
type DIDPolicy struct {
	DID           string    `json:"did"`
	AllowInbound  bool      `json:"allow_inbound"`
	AllowOutbound bool      `json:"allow_outbound"`
	AllowRelay    bool      `json:"allow_relay"`
	AllowedPorts  []uint16  `json:"allowed_ports"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type cachedDecision struct {
	decision Decision
	reason   string
	expires  time.Time
}

// Estado Taxonómico: IMPLEMENTADO (Control de Acceso Local Basado en DIDs).
// NOTA TÉCNICA: Implementa filtrado de paquetes y micro-segmentación por identidad soberana (DID ACLs)
// con política Default-Deny y caché. No pretende sustituir una infraestructura ZTNA corporativa
// completa con evaluación continua de postura de seguridad del dispositivo (device posture/EDR).

// ZTNAFirewall implementa el cortafuegos de identidad por DID (L1)
type ZTNAFirewall struct {
	mu           sync.RWMutex
	defaultDeny  bool
	policies     map[string]*DIDPolicy
	policyCache  map[string]cachedDecision
	cacheTTL     time.Duration
	cacheCleanup time.Time

	PacketsEvaluated  uint64
	PacketsAccepted   uint64
	PacketsDroppedDD  uint64
	PacketsDroppedACL uint64
}

// NewZTNAFirewall inicializa el firewall con la política Default-Deny configurada
func NewZTNAFirewall(defaultDeny bool) *ZTNAFirewall {
	return &ZTNAFirewall{
		defaultDeny:  defaultDeny,
		policies:     make(map[string]*DIDPolicy),
		policyCache:  make(map[string]cachedDecision),
		cacheTTL:     30 * time.Second,
		cacheCleanup: time.Now().Add(30 * time.Second),
	}
}

// SetDefaultDeny conmuta la política global de denegación por defecto
func (fw *ZTNAFirewall) SetDefaultDeny(enabled bool) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.defaultDeny = enabled
	fw.policyCache = make(map[string]cachedDecision)
}

// IsDefaultDeny consulta si la denegación por defecto está activa
func (fw *ZTNAFirewall) IsDefaultDeny() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.defaultDeny
}

// AuthorizeDID otorga permisos a un DID en la libreta de políticas
func (fw *ZTNAFirewall) AuthorizeDID(p *DIDPolicy) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	fw.policies[p.DID] = p
	fw.policyCache = make(map[string]cachedDecision)
}

// RevokeDID revoca el acceso concedido a un DID
func (fw *ZTNAFirewall) RevokeDID(did string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	delete(fw.policies, did)
	fw.policyCache = make(map[string]cachedDecision)
}

// AuthorizeDIDsBatch autoriza múltiples DIDs atómicamente
func (fw *ZTNAFirewall) AuthorizeDIDsBatch(policies []*DIDPolicy) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	now := time.Now()
	for _, p := range policies {
		if p.CreatedAt.IsZero() {
			p.CreatedAt = now
		}
		fw.policies[p.DID] = p
	}
	fw.policyCache = make(map[string]cachedDecision)
}

// RevokeDIDsBatch revoca múltiples DIDs atómicamente
func (fw *ZTNAFirewall) RevokeDIDsBatch(dids []string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	for _, did := range dids {
		delete(fw.policies, did)
	}
	fw.policyCache = make(map[string]cachedDecision)
}

// GetPolicy consulta la política asignada a un DID
func (fw *ZTNAFirewall) GetPolicy(did string) (*DIDPolicy, bool) {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	p, ok := fw.policies[did]
	return p, ok
}

// GetAllPolicies retorna una copia de todas las políticas activas
func (fw *ZTNAFirewall) GetAllPolicies() []*DIDPolicy {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	res := make([]*DIDPolicy, 0, len(fw.policies))
	for _, p := range fw.policies {
		res = append(res, p)
	}
	return res
}

// AllowPeer autoriza el tráfico proveniente de un par para puertos específicos
func (fw *ZTNAFirewall) AllowPeer(peerDID string, ports []uint16) {
	fw.AuthorizeDID(&DIDPolicy{
		DID:           peerDID,
		AllowInbound:  true,
		AllowOutbound: true,
		AllowedPorts:  ports,
	})
}

// RevokePeer revoca cualquier permiso concedido a un par
func (fw *ZTNAFirewall) RevokePeer(peerDID string) {
	fw.RevokeDID(peerDID)
}

// IsAllowed verifica si un par cuenta con autorización activa
func (fw *ZTNAFirewall) IsAllowed(peerDID string) bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	p, exists := fw.policies[peerDID]
	if !exists {
		return !fw.defaultDeny
	}
	if !p.ExpiresAt.IsZero() && time.Now().After(p.ExpiresAt) {
		return false
	}
	return p.AllowInbound
}

func (fw *ZTNAFirewall) cleanupExpiredCache() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	now := time.Now()
	if now.Before(fw.cacheCleanup) {
		return
	}
	fw.cacheCleanup = now.Add(30 * time.Second)
	for k, c := range fw.policyCache {
		if now.After(c.expires) {
			delete(fw.policyCache, k)
		}
	}
}

func (fw *ZTNAFirewall) eval(did string, port uint16, isInbound bool) (Decision, string) {
	atomic.AddUint64(&fw.PacketsEvaluated, 1)
	cacheKey := fmt.Sprintf("%s:%d:%t", did, port, isInbound)

	fw.mu.RLock()
	if c, ok := fw.policyCache[cacheKey]; ok && time.Now().Before(c.expires) {
		fw.mu.RUnlock()
		return c.decision, c.reason
	}
	fw.mu.RUnlock()

	fw.cleanupExpiredCache()

	fw.mu.RLock()
	policy, exists := fw.policies[did]
	fw.mu.RUnlock()

	var decision Decision
	var reason string

	if !exists {
		if fw.defaultDeny {
			atomic.AddUint64(&fw.PacketsDroppedDD, 1)
			decision = DecisionDropDefaultDeny
			reason = fmt.Sprintf("DID %s no figura en identidades autorizadas (Default-Deny)", did)
		} else {
			atomic.AddUint64(&fw.PacketsAccepted, 1)
			decision = DecisionAccept
			reason = "Tráfico aceptado por modo permisivo"
		}
	} else if !policy.ExpiresAt.IsZero() && time.Now().After(policy.ExpiresAt) {
		atomic.AddUint64(&fw.PacketsDroppedACL, 1)
		decision = DecisionDropACL
		reason = fmt.Sprintf("La política del DID %s ha expirado", did)
	} else if isInbound && !policy.AllowInbound {
		atomic.AddUint64(&fw.PacketsDroppedACL, 1)
		decision = DecisionDropACL
		reason = fmt.Sprintf("Tráfico entrante denegado explícitamente para DID %s", did)
	} else if !isInbound && !policy.AllowOutbound {
		atomic.AddUint64(&fw.PacketsDroppedACL, 1)
		decision = DecisionDropACL
		reason = fmt.Sprintf("Tráfico saliente denegado explícitamente para DID %s", did)
	} else if isInbound && len(policy.AllowedPorts) > 0 {
		portAllowed := false
		for _, p := range policy.AllowedPorts {
			if p == port {
				portAllowed = true
				break
			}
		}
		if !portAllowed {
			atomic.AddUint64(&fw.PacketsDroppedACL, 1)
			decision = DecisionDropACL
			reason = fmt.Sprintf("Puerto virtual %d no autorizado para DID %s", port, did)
		} else {
			atomic.AddUint64(&fw.PacketsAccepted, 1)
			decision = DecisionAccept
			reason = "Datagrama validado por política ZTNA"
		}
	} else {
		atomic.AddUint64(&fw.PacketsAccepted, 1)
		decision = DecisionAccept
		reason = "Datagrama validado por política ZTNA"
	}

	fw.mu.Lock()
	fw.policyCache[cacheKey] = cachedDecision{decision: decision, reason: reason, expires: time.Now().Add(fw.cacheTTL)}
	fw.mu.Unlock()
	return decision, reason
}

// EvaluateInbound dictamina si un datagrama entrante de srcDID hacia dstPort debe aceptarse
func (fw *ZTNAFirewall) EvaluateInbound(srcDID string, dstPort uint16) (Decision, string) {
	return fw.eval(srcDID, dstPort, true)
}

// EvaluateOutbound dictamina si un datagrama saliente hacia dstDID y dstPort debe transmitirse
func (fw *ZTNAFirewall) EvaluateOutbound(dstDID string, dstPort uint16) (Decision, string) {
	return fw.eval(dstDID, dstPort, false)
}

// FirewallStats estructura las métricas de rendimiento y auditoría del firewall
type FirewallStats struct {
	DefaultDeny       bool   `json:"default_deny"`
	ActivePolicies    int    `json:"active_policies"`
	PacketsEvaluated  uint64 `json:"packets_evaluated"`
	PacketsAccepted   uint64 `json:"packets_accepted"`
	PacketsDroppedDD  uint64 `json:"packets_dropped_default_deny"`
	PacketsDroppedACL uint64 `json:"packets_dropped_acl"`
}

// Stats genera una instantánea atómica para telemetría L2
func (fw *ZTNAFirewall) Stats() FirewallStats {
	fw.mu.RLock()
	dd := fw.defaultDeny
	count := len(fw.policies)
	fw.mu.RUnlock()

	return FirewallStats{
		DefaultDeny:       dd,
		ActivePolicies:    count,
		PacketsEvaluated:  atomic.LoadUint64(&fw.PacketsEvaluated),
		PacketsAccepted:   atomic.LoadUint64(&fw.PacketsAccepted),
		PacketsDroppedDD:  atomic.LoadUint64(&fw.PacketsDroppedDD),
		PacketsDroppedACL: atomic.LoadUint64(&fw.PacketsDroppedACL),
	}
}

// ToOpenMetrics genera el reporte en formato OpenMetrics/Prometheus
func (fw *ZTNAFirewall) ToOpenMetrics() string {
	s := fw.Stats()
	dd := 0
	if s.DefaultDeny {
		dd = 1
	}
	metrics := []struct {
		name, help, typ string
		val             any
	}{
		{"packets_total", "Total evaluados", "counter", s.PacketsEvaluated},
		{"accepted_total", "Aceptados", "counter", s.PacketsAccepted},
		{"dropped_default_deny_total", "Descartados DD", "counter", s.PacketsDroppedDD},
		{"dropped_acl_total", "Descartados ACL", "counter", s.PacketsDroppedACL},
		{"active_policies", "Políticas activas", "gauge", s.ActivePolicies},
		{"default_deny", "Modo Default-Deny", "gauge", dd},
	}
	var sb strings.Builder
	for _, m := range metrics {
		fmt.Fprintf(&sb, "# HELP ipvn7_firewall_%s %s\n# TYPE ipvn7_firewall_%s %s\nipvn7_firewall_%s %v\n", m.name, m.help, m.name, m.typ, m.name, m.val)
	}
	return sb.String()
}
