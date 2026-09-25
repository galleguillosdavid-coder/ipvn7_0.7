// Package l3 implementa el plano de control Zero-Trust Network Access (ZTNA)
// basado en Intenciones y Contexto (UIN Capas 2 y 3), rescatado de Ipv7IEU
// (core/bridge/group_control.go) y formalizado para ipvn7 v0.7.
package l3

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// IntentAction define la acción declarada en la intención del nodo o agente
type IntentAction string

const (
	IntentActionTelemetryPush IntentAction = "telemetry:push"
	IntentActionChatSend      IntentAction = "chat:send"
	IntentActionDAGSync       IntentAction = "dag:sync"
	IntentActionBenchmarkRun  IntentAction = "bench:run"
	IntentActionTunnelProxy   IntentAction = "tunnel:proxy"
	IntentActionControlAdmin  IntentAction = "control:admin"
)

// ContextFlag define banderas de contexto operacional del nodo emisor
type ContextFlag string

const (
	ContextFlagMissionCritical ContextFlag = "mission_critical"
	ContextFlagFieldOps        ContextFlag = "field_ops"
	ContextFlagDegradedMesh    ContextFlag = "degraded_mesh"
	ContextFlagLocalLoop       ContextFlag = "local_loop"
)

// IntentPolicy modela la política de acceso ZTNA basada en identidad y capacidades
type IntentPolicy struct {
	PolicyID             string        `json:"policy_id"`
	SubjectDID           string        `json:"subject_did"` // Wildcard "*" permitido
	AllowedActions       []IntentAction `json:"allowed_actions"`
	RequiredContextFlags []ContextFlag  `json:"required_context_flags"`
	MaxRateLimitPerSec   int           `json:"max_rate_limit_per_sec"`
	ExpiresAt            time.Time     `json:"expires_at"`
}

// IntentRequest representa la solicitud declarada de un nodo para realizar una acción
type IntentRequest struct {
	SubjectDID   string       `json:"subject_did"`
	Action       IntentAction `json:"action"`
	ContextFlags []ContextFlag `json:"context_flags"`
	Timestamp    time.Time    `json:"timestamp"`
}

// IntentDecision resultado de la evaluación ZTNA
type IntentDecision struct {
	Allowed   bool   `json:"allowed"`
	Reason    string `json:"reason"`
	PolicyID  string `json:"policy_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// IntentEngine orquesta la evaluación ZTNA declarativa en L3
type IntentEngine struct {
	mu           sync.RWMutex
	policies     map[string]IntentPolicy
	rateCounters map[string]int
	lastReset    time.Time
}

// NewIntentEngine inicializa el motor de políticas de intención
func NewIntentEngine() *IntentEngine {
	return &IntentEngine{
		policies:     make(map[string]IntentPolicy),
		rateCounters: make(map[string]int),
		lastReset:    time.Now(),
	}
}

// AddPolicy añade o actualiza una política ZTNA
func (ie *IntentEngine) AddPolicy(policy IntentPolicy) error {
	if policy.PolicyID == "" {
		return errors.New("intent: PolicyID no puede ser vacío")
	}
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.policies[policy.PolicyID] = policy
	return nil
}

// RemovePolicy elimina una política por su ID
func (ie *IntentEngine) RemovePolicy(policyID string) {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	delete(ie.policies, policyID)
}

// EvaluateIntent evalúa la solicitud frente a las políticas activas
func (ie *IntentEngine) EvaluateIntent(req IntentRequest) IntentDecision {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	now := time.Now()
	// Reiniciar limitadores de tasa cada segundo
	if now.Sub(ie.lastReset) >= time.Second {
		ie.rateCounters = make(map[string]int)
		ie.lastReset = now
	}

	for _, pol := range ie.policies {
		// Validar expiración si está definida
		if !pol.ExpiresAt.IsZero() && now.After(pol.ExpiresAt) {
			continue
		}

		// Validar sujeto DID (coincidencia exacta o wildcard)
		if pol.SubjectDID != "*" && pol.SubjectDID != req.SubjectDID {
			continue
		}

		// Validar si la acción solicitada está autorizada
		actionAllowed := false
		for _, a := range pol.AllowedActions {
			if a == req.Action {
				actionAllowed = true
				break
			}
		}
		if !actionAllowed {
			continue
		}

		// Validar contexto requerido (UIN Capa 3)
		if !hasRequiredContext(req.ContextFlags, pol.RequiredContextFlags) {
			return IntentDecision{
				Allowed:   false,
				Reason:    fmt.Sprintf("contexto requerido insatisfecho para politica %s", pol.PolicyID),
				PolicyID:  pol.PolicyID,
				Timestamp: now,
			}
		}

		// Validar límite de tasa
		if pol.MaxRateLimitPerSec > 0 {
			key := fmt.Sprintf("%s:%s", req.SubjectDID, req.Action)
			if ie.rateCounters[key] >= pol.MaxRateLimitPerSec {
				return IntentDecision{
					Allowed:   false,
					Reason:    fmt.Sprintf("tasa excedida (limite: %d req/s)", pol.MaxRateLimitPerSec),
					PolicyID:  pol.PolicyID,
					Timestamp: now,
				}
			}
			ie.rateCounters[key]++
		}

		return IntentDecision{
			Allowed:   true,
			Reason:    "intencion autorizada por politica ZTNA",
			PolicyID:  pol.PolicyID,
			Timestamp: now,
		}
	}

	return IntentDecision{
		Allowed:   false,
		Reason:    "ninguna politica ZTNA autoriza esta intencion (default deny)",
		Timestamp: now,
	}
}

// hasRequiredContext verifica si todas las banderas requeridas están presentes en la solicitud
func hasRequiredContext(reqFlags, requiredFlags []ContextFlag) bool {
	if len(requiredFlags) == 0 {
		return true
	}
	flagMap := make(map[ContextFlag]bool, len(reqFlags))
	for _, f := range reqFlags {
		flagMap[f] = true
	}
	for _, reqF := range requiredFlags {
		if !flagMap[reqF] {
			return false
		}
	}
	return true
}
