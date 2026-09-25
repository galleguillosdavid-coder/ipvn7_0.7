package l3

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"ipvn7/pkg/l0"
)

// DeonticOp representa los operadores de la lógica deóntica formal
type DeonticOp string

const (
	DeonticObligation  DeonticOp = "OBLIGATION"  // \mathcal{O}: Conducta preceptiva
	DeonticPermission  DeonticOp = "PERMISSION"  // \mathcal{P}: Conducta facultativa / lícita
	DeonticProhibition DeonticOp = "PROHIBITION" // \mathcal{F}: Conducta vedada / ilícita
)

// Constantes de Jerarquía Axiomática (Lex Superior)
const (
	PriorityConstitutional = 100 // Carta Magna Universal (Artículo I e Invariantes)
	PrioritySectorial      = 50  // Directivas de Bloque / Subred Institucional
	PriorityCommunity      = 10  // Acuerdos locales / Omnistore P2P
)

// ADICORule representa una directiva institucional formalmente estructurada (Crawford & Ostrom)
type ADICORule struct {
	ID            string    `json:"id"`
	ContextDomain string    `json:"context_domain"` // e.g. "root", "subred.edu.ipv7", "market.p2p.ipv7"
	Attribute     string    `json:"attribute"`      // Sujeto/rol (e.g. "all", "civic_agent", "transit_node")
	Deontic       DeonticOp `json:"deontic"`        // OBLIGATION, PERMISSION, PROHIBITION
	Aim           string    `json:"aim"`            // Acción regulada (e.g. "telemetry_tracking", "p2p_trade")
	Condition     string    `json:"condition"`      // Expresión lógica / circunstancia
	OrElse        string    `json:"or_else"`        // Sanción o consecuencia graduada
	Priority      int       `json:"priority"`       // Nivel jerárquico (10, 50, 100)
	Timestamp     time.Time `json:"timestamp"`      // Tiempo de ratificación para Lex Posterior
	IssuerDID     string    `json:"issuer_did"`
	SignatureHex  string    `json:"signature"`
}

// ConflictResolutionResult detalla la decisión tomada por el motor de derecho computable
type ConflictResolutionResult struct {
	WinningRule   ADICORule `json:"winning_rule"`
	LosingRule    ADICORule `json:"losing_rule"`
	DoctrineUsed  string    `json:"doctrine_used"` // "LEX_SUPERIOR", "LEX_SPECIALIS", "LEX_POSTERIOR"
	Justification string    `json:"justification"`
}

// ComputationalLawEngine implementa el Derecho Computable y la resolución de antinomias en ipvn7
type ComputationalLawEngine struct {
	mu           sync.RWMutex
	Identity     *l0.Identity
	rules        map[string]*ADICORule            // ruleID -> Rule
	contextRules map[string]map[string]*ADICORule // contextDomain -> aim -> Rule
}

// NewComputationalLawEngine inicializa el motor cargado con los 6 axiomas de la Carta Magna
func NewComputationalLawEngine(id *l0.Identity) *ComputationalLawEngine {
	engine := &ComputationalLawEngine{
		Identity:     id,
		rules:        make(map[string]*ADICORule),
		contextRules: make(map[string]map[string]*ADICORule),
	}

	// Sembrar axiomas constitucionales inmutables (Prioridad 100, Ámbito "root")
	engine.seedConstitutionalAxioms()
	return engine
}

// RegisterRule evalúa e inscribe una nueva directiva institucional previa verificación de no-antinomia
func (cle *ComputationalLawEngine) RegisterRule(rule ADICORule) (*ADICORule, error) {
	cle.mu.Lock()
	defer cle.mu.Unlock()

	if rule.ContextDomain == "" {
		return nil, errors.New("el context_domain no puede ser vacío")
	}
	if rule.Aim == "" {
		return nil, errors.New("el aim (objetivo) de la regla no puede ser vacío")
	}
	if rule.Deontic != DeonticObligation && rule.Deontic != DeonticPermission && rule.Deontic != DeonticProhibition {
		return nil, fmt.Errorf("operador deóntico inválido: %s", rule.Deontic)
	}

	if rule.ID == "" {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s:%d", rule.ContextDomain, rule.Attribute, rule.Deontic, rule.Aim, time.Now().UnixNano())))
		rule.ID = fmt.Sprintf("rule:%s", hex.EncodeToString(h[:12]))
	}
	if rule.Timestamp.IsZero() {
		rule.Timestamp = time.Now().UTC()
	}
	if rule.Priority <= 0 {
		rule.Priority = PriorityCommunity
	}

	// 1. Verificación de Lex Superior contra la Carta Magna Universal (root)
	rootRules := cle.contextRules["root"]
	if rootRules != nil {
		if rootRule, exists := rootRules[rule.Aim]; exists {
			isAntinomy, reason := cle.isDirectAntinomy(*rootRule, rule)
			if isAntinomy {
				// Lex Superior anula la regla en origen: N_core |= ~phi => Inv(N_local(phi))
				return nil, fmt.Errorf("invalidez absoluta en origen por colisión con Lex Superior (Carta Magna): %s", reason)
			}
		}
	}

	// 2. Verificación de Antinomias en el mismo contexto
	if cle.contextRules[rule.ContextDomain] == nil {
		cle.contextRules[rule.ContextDomain] = make(map[string]*ADICORule)
	}

	if existing, exists := cle.contextRules[rule.ContextDomain][rule.Aim]; exists {
		isAntinomy, _ := cle.isDirectAntinomy(*existing, rule)
		if isAntinomy {
			// Resolver por Lex Posterior (la más reciente deroga a la anterior en el mismo nivel)
			res := cle.resolveConflictInternal(*existing, rule)
			if res.WinningRule.ID != rule.ID {
				return nil, fmt.Errorf("regla rechazada: prevalece la norma existente '%s' por doctrina %s", existing.ID, res.DoctrineUsed)
			}
			// La nueva regla gana y deroga a la antigua
			delete(cle.rules, existing.ID)
		}
	}

	// Firmar y almacenar
	if rule.IssuerDID == "" {
		rule.IssuerDID = cle.Identity.DID()
	}
	if rule.SignatureHex == "" {
		sig := cle.Identity.Sign([]byte(fmt.Sprintf("%s:%s:%s", rule.ID, rule.Deontic, rule.Aim)))
		rule.SignatureHex = hex.EncodeToString(sig)
	}

	ruleCopy := rule
	cle.rules[rule.ID] = &ruleCopy
	cle.contextRules[rule.ContextDomain][rule.Aim] = &ruleCopy

	return &ruleCopy, nil
}

// DetectAntinomy comprueba si dos reglas colisionan ontológicamente
func (cle *ComputationalLawEngine) DetectAntinomy(ruleA, ruleB ADICORule) (bool, string) {
	cle.mu.RLock()
	defer cle.mu.RUnlock()

	return cle.isDirectAntinomy(ruleA, ruleB)
}

func (cle *ComputationalLawEngine) isDirectAntinomy(ruleA, ruleB ADICORule) (bool, string) {
	if ruleA.Aim != ruleB.Aim {
		return false, ""
	}

	// Antinomia canónica 1: Obligación vs Prohibición (O(phi) ^ F(phi) => bottom)
	if (ruleA.Deontic == DeonticObligation && ruleB.Deontic == DeonticProhibition) ||
		(ruleA.Deontic == DeonticProhibition && ruleB.Deontic == DeonticObligation) {
		return true, fmt.Sprintf("Contradicción irreconciliable: la acción '%s' es definida simultáneamente como Obligatoria y Prohibida", ruleA.Aim)
	}

	// Antinomia canónica 2: Permisión vs Prohibición en el mismo contexto
	if (ruleA.Deontic == DeonticPermission && ruleB.Deontic == DeonticProhibition) ||
		(ruleA.Deontic == DeonticProhibition && ruleB.Deontic == DeonticPermission) {
		return true, fmt.Sprintf("Incompatibilidad normativa: la acción '%s' es declarada Prohibida y Permitida bajo el mismo objetivo", ruleA.Aim)
	}

	return false, ""
}

// ResolveConflict dirime una colisión normativa aplicando la jerarquía computacional:
// 1. Lex Superior -> 2. Lex Specialis -> 3. Lex Posterior
func (cle *ComputationalLawEngine) ResolveConflict(ruleA, ruleB ADICORule) ConflictResolutionResult {
	cle.mu.RLock()
	defer cle.mu.RUnlock()

	return cle.resolveConflictInternal(ruleA, ruleB)
}

func (cle *ComputationalLawEngine) resolveConflictInternal(ruleA, ruleB ADICORule) ConflictResolutionResult {
	// 1. Lex Superior: Primacía jerárquica incuestionable
	if ruleA.Priority > ruleB.Priority {
		return ConflictResolutionResult{
			WinningRule:   ruleA,
			LosingRule:    ruleB,
			DoctrineUsed:  "LEX_SUPERIOR",
			Justification: fmt.Sprintf("Prevalece '%s' (Prioridad %d) sobre '%s' (Prioridad %d)", ruleA.ID, ruleA.Priority, ruleB.ID, ruleB.Priority),
		}
	} else if ruleB.Priority > ruleA.Priority {
		return ConflictResolutionResult{
			WinningRule:   ruleB,
			LosingRule:    ruleA,
			DoctrineUsed:  "LEX_SUPERIOR",
			Justification: fmt.Sprintf("Prevalece '%s' (Prioridad %d) sobre '%s' (Prioridad %d)", ruleB.ID, ruleB.Priority, ruleA.ID, ruleA.Priority),
		}
	}

	// 2. Lex Specialis: Prevalencia del contexto más específico sobre el genérico
	// Ejemplo: "market.p2p.ipv7" es más específico que "root"
	subdomainsA := len(strings.Split(ruleA.ContextDomain, "."))
	subdomainsB := len(strings.Split(ruleB.ContextDomain, "."))

	if ruleA.ContextDomain != "root" && ruleB.ContextDomain == "root" {
		return ConflictResolutionResult{
			WinningRule:   ruleA,
			LosingRule:    ruleB,
			DoctrineUsed:  "LEX_SPECIALIS",
			Justification: fmt.Sprintf("Prevalece '%s' por ámbito específico local sobre el precepto general", ruleA.ID),
		}
	} else if ruleB.ContextDomain != "root" && ruleA.ContextDomain == "root" {
		return ConflictResolutionResult{
			WinningRule:   ruleB,
			LosingRule:    ruleA,
			DoctrineUsed:  "LEX_SPECIALIS",
			Justification: fmt.Sprintf("Prevalece '%s' por ámbito específico local sobre el precepto general", ruleB.ID),
		}
	} else if subdomainsA > subdomainsB {
		return ConflictResolutionResult{
			WinningRule:   ruleA,
			LosingRule:    ruleB,
			DoctrineUsed:  "LEX_SPECIALIS",
			Justification: fmt.Sprintf("Prevalece '%s' por mayor profundidad contextual (%d niveles vs %d)", ruleA.ID, subdomainsA, subdomainsB),
		}
	} else if subdomainsB > subdomainsA {
		return ConflictResolutionResult{
			WinningRule:   ruleB,
			LosingRule:    ruleA,
			DoctrineUsed:  "LEX_SPECIALIS",
			Justification: fmt.Sprintf("Prevalece '%s' por mayor profundidad contextual (%d niveles vs %d)", ruleB.ID, subdomainsB, subdomainsA),
		}
	}

	// 3. Lex Posterior: Sucesión temporal criptográfica (la más reciente deroga a la anterior)
	if ruleA.Timestamp.After(ruleB.Timestamp) {
		return ConflictResolutionResult{
			WinningRule:   ruleA,
			LosingRule:    ruleB,
			DoctrineUsed:  "LEX_POSTERIOR",
			Justification: fmt.Sprintf("Prevalece '%s' por ratificación más reciente (%s vs %s)", ruleA.ID, ruleA.Timestamp.Format(time.RFC3339), ruleB.Timestamp.Format(time.RFC3339)),
		}
	}

	return ConflictResolutionResult{
		WinningRule:   ruleB,
		LosingRule:    ruleA,
		DoctrineUsed:  "LEX_POSTERIOR",
		Justification: fmt.Sprintf("Prevalece '%s' por ratificación más reciente (%s vs %s)", ruleB.ID, ruleB.Timestamp.Format(time.RFC3339), ruleA.Timestamp.Format(time.RFC3339)),
	}
}

// EvaluateAction dictamina la calificación deóntica de una acción semántica en un dominio contextual
func (cle *ComputationalLawEngine) EvaluateAction(contextDomain, action string) (DeonticOp, string, bool) {
	cle.mu.RLock()
	defer cle.mu.RUnlock()

	// 1. Comprobar prohibiciones absolutas en la Carta Magna (root)
	if rootRules, ok := cle.contextRules["root"]; ok {
		if rule, found := rootRules[action]; found {
			if rule.Deontic == DeonticProhibition {
				return DeonticProhibition, fmt.Sprintf("[LEX SUPERIOR]: Acción prohibida terminantemente por Carta Magna (%s): %s", rule.ID, rule.OrElse), false
			}
		}
	}

	// 2. Comprobar reglas específicas del dominio contextual
	if localRules, ok := cle.contextRules[contextDomain]; ok {
		if rule, found := localRules[action]; found {
			allowed := rule.Deontic != DeonticProhibition
			return rule.Deontic, fmt.Sprintf("[LEX SPECIALIS - %s]: %s (Condición: %s)", contextDomain, rule.Deontic, rule.Condition), allowed
		}
	}

	// 3. Por defecto en la red soberana ipvn7: Lo que no está expresamente prohibido está permitido
	return DeonticPermission, "Permiso implícito por subsidiaridad (sin prohibiciones específicas registradas)", true
}

// ListAllRules entrega todas las reglas vigentes en el ordenamiento jurídico de ipvn7
func (cle *ComputationalLawEngine) ListAllRules() []*ADICORule {
	cle.mu.RLock()
	defer cle.mu.RUnlock()

	res := make([]*ADICORule, 0, len(cle.rules))
	for _, r := range cle.rules {
		rCopy := *r
		res = append(res, &rCopy)
	}
	return res
}
