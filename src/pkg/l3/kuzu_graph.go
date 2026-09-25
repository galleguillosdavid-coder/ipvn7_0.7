package l3

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Modelos de Nodos del Grafo Kùzu (2.md Sección 2)

// PeerNode representa una identidad soberana en el grafo topológico
type PeerNode struct {
	DID           string `json:"did"`
	MLDSA         string `json:"ml_dsa"`
	Anycast       bool   `json:"anycast"`
	DeviceProfile string `json:"device_profile"`
}

// AxiomPrinciple representa una ley inquebrantable de la arquitectura (genesis.md)
type AxiomPrinciple struct {
	PrincipleID string `json:"principle_id"` // ej. AXIOM_ZERO_PII, AXIOM_CORE_FREEZE
	Statement   string `json:"statement"`    // Descripción axiomática
	Immutable   bool   `json:"immutable"`
}

// OperationalFunction representa un módulo o adaptador en código
type OperationalFunction struct {
	FunctionID string `json:"function_id"` // ej. L1_ZTNA_FIREWALL, L1_WDRR_SCHEDULER
	ModulePath string `json:"module_path"` // ej. pkg/l1/firewall.go
	Status     string `json:"status"`      // ACTIVE, VERIFIED
}

// EmpiricalObservation representa una métrica empírica validada en campo
type EmpiricalObservation struct {
	ObservationID      string    `json:"observation_id"`
	MetricsFingerprint string    `json:"metrics_fingerprint"`
	Value              float64   `json:"value"`
	Timestamp          time.Time `json:"timestamp"`
}

// EngineeringDecision representa un registro inmutable de decisión técnica
type EngineeringDecision struct {
	DecisionID string    `json:"decision_id"`
	Rationale  string    `json:"rationale"`
	Timestamp  time.Time `json:"timestamp"`
}

// Modelos de Relaciones (Aristas) en Kùzu

// XORLinkRel representa la distancia matemática y enlace entre pares en Kleinberg
type XORLinkRel struct {
	FromDID     string    `json:"from_did"`
	ToDID       string    `json:"to_did"`
	XORDistance string    `json:"xor_distance"`
	LatencyMs   float64   `json:"latency_ms"`
	Degree      int8      `json:"degree"` // Anillo [0..11]
	RFBand      string    `json:"rf_band"`
	LastSeen    time.Time `json:"last_seen"`
}

// GovernsRel representa una acción de gobernanza FSM firmada criptográficamente
type GovernsRel struct {
	FromDID   string    `json:"from_did"`
	ToDID     string    `json:"to_did"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"`
}


// KuzuGraphEngine implementa la fuente de verdad viva en memoria/grafo (genesis.md L178 y 2.md Sec 2)
type KuzuGraphEngine struct {
	mu           sync.RWMutex
	peers        map[string]*PeerNode
	axioms       map[string]*AxiomPrinciple
	functions    map[string]*OperationalFunction
	observations map[string]*EmpiricalObservation
	decisions    map[string]*EngineeringDecision

	// Relaciones
	xorLinks    []*XORLinkRel
	governs     []*GovernsRel
	derivesFrom map[string]string   // function_id -> principle_id
	validatedBy map[string][]string // function_id -> []observation_id
}

// NewKuzuGraphEngine inicializa el motor de grafos y siembra el esquema PFO canónico
func NewKuzuGraphEngine() *KuzuGraphEngine {
	kg := &KuzuGraphEngine{
		peers:        make(map[string]*PeerNode),
		axioms:       make(map[string]*AxiomPrinciple),
		functions:    make(map[string]*OperationalFunction),
		observations: make(map[string]*EmpiricalObservation),
		decisions:    make(map[string]*EngineeringDecision),
		xorLinks:     make([]*XORLinkRel, 0),
		governs:      make([]*GovernsRel, 0),
		derivesFrom:  make(map[string]string),
		validatedBy:  make(map[string][]string),
	}

	kg.seedCanonicalAxioms()
	kg.seedCanonicalFunctions()
	kg.seedCanonicalDecisions()
	return kg
}



// UpsertPeer registra o actualiza un nodo par en el grafo
func (kg *KuzuGraphEngine) UpsertPeer(did, mlDSA, deviceProfile string, anycast bool) {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	kg.peers[did] = &PeerNode{
		DID:           did,
		MLDSA:         mlDSA,
		Anycast:       anycast,
		DeviceProfile: deviceProfile,
	}
}

// UpsertXORLink registra o actualiza una arista de enrutamiento Kleinberg
func (kg *KuzuGraphEngine) UpsertXORLink(from, to, xorDist string, latency float64, degree int8, rfBand string) {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	// Actualizar si ya existe la arista
	for _, l := range kg.xorLinks {
		if l.FromDID == from && l.ToDID == to {
			l.LatencyMs = latency
			l.Degree = degree
			l.RFBand = rfBand
			l.LastSeen = time.Now()
			return
		}
	}

	kg.xorLinks = append(kg.xorLinks, &XORLinkRel{
		FromDID:     from,
		ToDID:       to,
		XORDistance: xorDist,
		LatencyMs:   latency,
		Degree:      degree,
		RFBand:      rfBand,
		LastSeen:    time.Now(),
	})
}

// AddObservation registra una observación empírica validando una función operativa
func (kg *KuzuGraphEngine) AddObservation(functionID, fingerprint string, value float64) string {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%.4f:%d", functionID, fingerprint, value, time.Now().UnixNano())))
	obsID := "obs:" + hex.EncodeToString(h[:8])

	obs := &EmpiricalObservation{
		ObservationID:      obsID,
		MetricsFingerprint: fingerprint,
		Value:              value,
		Timestamp:          time.Now(),
	}

	kg.observations[obsID] = obs
	kg.validatedBy[functionID] = append(kg.validatedBy[functionID], obsID)
	return obsID
}

// RecordDecision registra una decisión técnica inmutable
func (kg *KuzuGraphEngine) RecordDecision(decisionID, rationale string) {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	kg.decisions[decisionID] = &EngineeringDecision{
		DecisionID: decisionID,
		Rationale:  rationale,
		Timestamp:  time.Now(),
	}
}

// AuditDirectiveContradiction verifica deductivamente si una instrucción propuesta viola los axiomas rectores
func (kg *KuzuGraphEngine) AuditDirectiveContradiction(directiveText string) (bool, string) {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	lower := strings.ToLower(directiveText)

	// Regla 1: Violación de Zero-PII
	if strings.Contains(lower, "store cleartext ip") || strings.Contains(lower, "guardar ip real") ||
		strings.Contains(lower, "persist pii") || strings.Contains(lower, "almacenar cedula") {
		return true, fmt.Sprintf("COLISIÓN AXIOMÁTICA detectada contra [%s]: La instrucción viola el anonimato absoluto Zero-PII.", kg.axioms["AXIOM_ZERO_PII"].PrincipleID)
	}

	// Regla 2: Violación de Core Freeze
	if strings.Contains(lower, "modify pkg/l0") || strings.Contains(lower, "modificar core/") ||
		strings.Contains(lower, "alter wire format") || strings.Contains(lower, "reescribir identidad canonica") {
		return true, fmt.Sprintf("COLISIÓN AXIOMÁTICA detectada contra [%s]: Intento de mutar el núcleo canónico congelado.", kg.axioms["AXIOM_STRICT_CORE_FREEZE"].PrincipleID)
	}

	// Regla 3: Violación de Flujo Sostenible
	if strings.Contains(lower, "disable pacer") || strings.Contains(lower, "unlimited burst") ||
		strings.Contains(lower, "desactivar marcapasos") || strings.Contains(lower, "ignorar cuello de botella") {
		return true, fmt.Sprintf("COLISIÓN AXIOMÁTICA detectada contra [%s]: La instrucción introduce riesgo de bufferbloat.", kg.axioms["AXIOM_SUSTAINABLE_FLOW"].PrincipleID)
	}

	return false, "VALIDACIÓN AXIOMÁTICA EXITOSA: La directiva es formalmente coherente con los principios rectores."
}


// GetPFOTree retorna la estructura jerárquica Principios -> Funciones -> Observaciones
func (kg *KuzuGraphEngine) GetPFOTree() map[string]interface{} {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	type functionDTO struct {
		FunctionID   string                  `json:"function_id"`
		ModulePath   string                  `json:"module_path"`
		Status       string                  `json:"status"`
		Observations []*EmpiricalObservation `json:"observations"`
	}

	type principleDTO struct {
		PrincipleID string         `json:"principle_id"`
		Statement   string         `json:"statement"`
		Functions   []*functionDTO `json:"functions"`
	}

	principlesMap := make(map[string]*principleDTO)
	for pID, ax := range kg.axioms {
		principlesMap[pID] = &principleDTO{
			PrincipleID: ax.PrincipleID,
			Statement:   ax.Statement,
			Functions:   make([]*functionDTO, 0),
		}
	}

	for fID, fn := range kg.functions {
		pID := kg.derivesFrom[fID]
		pDTO, ok := principlesMap[pID]
		if !ok {
			continue
		}

		obsList := make([]*EmpiricalObservation, 0)
		for _, obsID := range kg.validatedBy[fID] {
			if obs, exists := kg.observations[obsID]; exists {
				obsList = append(obsList, obs)
			}
		}

		pDTO.Functions = append(pDTO.Functions, &functionDTO{
			FunctionID:   fn.FunctionID,
			ModulePath:   fn.ModulePath,
			Status:       fn.Status,
			Observations: obsList,
		})
	}

	resultList := make([]*principleDTO, 0, len(principlesMap))
	for _, pDTO := range principlesMap {
		resultList = append(resultList, pDTO)
	}

	return map[string]interface{}{
		"total_principles":   len(kg.axioms),
		"total_functions":    len(kg.functions),
		"total_observations": len(kg.observations),
		"total_peers":        len(kg.peers),
		"total_xor_links":    len(kg.xorLinks),
		"pfo_tree":           resultList,
	}
}
