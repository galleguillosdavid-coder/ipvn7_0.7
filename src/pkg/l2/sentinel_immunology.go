package l2

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

type IncidentType string

const (
	IncidentCodeInjection IncidentType = "CODE_INJECTION_ATTEMPT"
	IncidentSybilAttack   IncidentType = "SYBIL_ATTACK_DETECTED"
	IncidentDAGTampering  IncidentType = "DAG_IMMUTABILITY_BREACH"
	IncidentDishonestAI   IncidentType = "DISHONEST_AI_AGENT"
	IncidentMetricFraud   IncidentType = "METRIC_FALSIFICATION"
	IncidentL0Breach      IncidentType = "UNAUTHORIZED_L0_MUTATION"
)

// AuditChallenge emitido por un centinela para auditar a un vecino
type AuditChallenge struct {
	ChallengeID          string    `json:"challenge_id"`
	SentinelDID          string    `json:"sentinel_did"`
	TargetDID            string    `json:"target_did"`
	NonceHex             string    `json:"nonce_hex"`
	Timestamp            time.Time `json:"timestamp"`
	ExpectedSandboxHash  string    `json:"expected_sandbox_hash"`
}

// AuditResponse devuelto por el nodo auditado
type AuditResponse struct {
	ChallengeID  string    `json:"challenge_id"`
	TargetDID    string    `json:"target_did"`
	AnswerHash   string    `json:"answer_hash"`
	ExecutionMs  float64   `json:"execution_ms"`
	Timestamp    time.Time `json:"timestamp"`
	SignatureHex string    `json:"signature"`
}

// ImmunologicalAlert representa una alerta celular propagada en el enjambre de centinelas
type ImmunologicalAlert struct {
	AlertID             string            `json:"alert_id"`
	Incident            IncidentType      `json:"incident"`
	OffenderDID         string            `json:"offender_did"`
	ReporterSentinelDID string            `json:"reporter_sentinel_did"`
	EvidenceHash        string            `json:"evidence_hash"`
	EvidencePayload     string            `json:"evidence_payload"`
	Timestamp           time.Time         `json:"timestamp"`
	Signatures          map[string]string `json:"signatures"` // SentinelDID -> Ed25519 Sig
	QuorumReached       bool              `json:"quorum_reached"`
	Slashed             bool              `json:"slashed"`
}

// SentinelStats resume la salud inmunológica del nodo y de la red
type SentinelStats struct {
	TotalAuditsExecuted uint64   `json:"total_audits_executed"`
	ActiveSentinels     int      `json:"active_sentinels"`
	TotalAlertsEmitted  uint64   `json:"total_alerts_emitted"`
	QuorumThreshold     int      `json:"quorum_threshold"`
	NeutralizedDIDs     []string `json:"neutralized_dids"`
}

// SentinelImmunologyEngine gestiona la inmunidad celular descentralizada (docs/GOBERNANZA.md)
type SentinelImmunologyEngine struct {
	mu              sync.RWMutex
	Identity        *l0.Identity
	Firewall        *l1.ZTNAFirewall
	Accounting      *TransitAccounting
	WoT             *l1.WebOfTrust
	QuorumThreshold int

	ActiveChallenges map[string]*AuditChallenge
	Alerts           map[string]*ImmunologicalAlert
	Neutralized      map[string]time.Time

	totalAudits atomic.Uint64
	totalAlerts atomic.Uint64
}

// NewSentinelImmunologyEngine inicializa el motor inmunológico
func NewSentinelImmunologyEngine(
	id *l0.Identity,
	fw *l1.ZTNAFirewall,
	acc *TransitAccounting,
	wot *l1.WebOfTrust,
) *SentinelImmunologyEngine {
	return &SentinelImmunologyEngine{
		Identity:         id,
		Firewall:         fw,
		Accounting:       acc,
		WoT:              wot,
		QuorumThreshold:  2, // Al menos 2 centinelas independientes para slashing
		ActiveChallenges: make(map[string]*AuditChallenge),
		Alerts:           make(map[string]*ImmunologicalAlert),
		Neutralized:      make(map[string]time.Time),
	}
}

// IssueAuditChallenge emite un desafío aleatorio a un nodo vecino para verificar honestidad técnica
func (sie *SentinelImmunologyEngine) IssueAuditChallenge(targetDID string) (*AuditChallenge, error) {
	sie.mu.Lock()
	defer sie.mu.Unlock()

	var nonce [16]byte
	_, _ = rand.Read(nonce[:])
	nonceHex := hex.EncodeToString(nonce[:])

	// Desafío criptográfico determinista de integridad en ejecución (Proof-of-Computation)
	expectedInput := fmt.Sprintf("AUDIT:%s:%s", targetDID, nonceHex)
	h := sha256.Sum256([]byte(expectedInput))
	expectedHash := hex.EncodeToString(h[:])

	chalID := fmt.Sprintf("chal:%s", nonceHex[:12])
	chal := &AuditChallenge{
		ChallengeID:         chalID,
		SentinelDID:         sie.Identity.DID(),
		TargetDID:           targetDID,
		NonceHex:            nonceHex,
		Timestamp:           time.Now().UTC(),
		ExpectedSandboxHash: expectedHash,
	}

	sie.ActiveChallenges[chalID] = chal
	sie.totalAudits.Add(1)
	return chal, nil
}

// VerifyAuditResponse audita la respuesta entregada por el vecino
func (sie *SentinelImmunologyEngine) VerifyAuditResponse(chalID string, resp *AuditResponse) (bool, error) {
	sie.mu.Lock()
	defer sie.mu.Unlock()

	chal, exists := sie.ActiveChallenges[chalID]
	if !exists {
		return false, errors.New("desafío no encontrado o expirado")
	}
	delete(sie.ActiveChallenges, chalID)

	// Comprobar coincidencia matemática de ejecución
	if resp.AnswerHash != chal.ExpectedSandboxHash {
		// La respuesta es fraudulenta o la IA/nodo está manipulado
		_, _ = sie.emitAlertInternal(IncidentDishonestAI, chal.TargetDID, fmt.Sprintf("Hash inconsistente: esperaba %s, recibió %s", chal.ExpectedSandboxHash, resp.AnswerHash))
		return false, nil
	}

	return true, nil
}

// EmitImmunologicalAlert crea y firma una alerta inmunológica inicial
func (sie *SentinelImmunologyEngine) EmitImmunologicalAlert(incident IncidentType, offenderDID string, evidence string) (*ImmunologicalAlert, error) {
	sie.mu.Lock()
	defer sie.mu.Unlock()

	return sie.emitAlertInternal(incident, offenderDID, evidence)
}

func (sie *SentinelImmunologyEngine) emitAlertInternal(incident IncidentType, offenderDID string, evidence string) (*ImmunologicalAlert, error) {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", incident, offenderDID, evidence)))
	evidenceHash := hex.EncodeToString(h[:])
	alertID := fmt.Sprintf("cid:alert:%s", evidenceHash[:16])

	sigData := []byte(fmt.Sprintf("ALERT:%s:%s:%s", alertID, incident, offenderDID))
	sig := sie.Identity.Sign(sigData)

	alert := &ImmunologicalAlert{
		AlertID:             alertID,
		Incident:            incident,
		OffenderDID:         offenderDID,
		ReporterSentinelDID: sie.Identity.DID(),
		EvidenceHash:        evidenceHash,
		EvidencePayload:     evidence,
		Timestamp:           time.Now().UTC(),
		Signatures: map[string]string{
			sie.Identity.DID(): hex.EncodeToString(sig),
		},
		QuorumReached: false,
		Slashed:       false,
	}

	sie.Alerts[alertID] = alert
	sie.totalAlerts.Add(1)

	// Si el quórum es 1 (modo estricto) o ya se alcanza, ejecutar slashing de inmediato
	if len(alert.Signatures) >= sie.QuorumThreshold {
		sie.executeSlashing(alert)
	}

	return alert, nil
}

// EndorseAlert permite a otro centinela independiente respaldar la alerta con su firma
func (sie *SentinelImmunologyEngine) EndorseAlert(alertID string, endorserID *l0.Identity) (*ImmunologicalAlert, error) {
	sie.mu.Lock()
	defer sie.mu.Unlock()

	alert, exists := sie.Alerts[alertID]
	if !exists {
		return nil, errors.New("alerta inmunológica no encontrada")
	}

	sigData := []byte(fmt.Sprintf("ALERT:%s:%s:%s", alertID, alert.Incident, alert.OffenderDID))
	sig := endorserID.Sign(sigData)
	alert.Signatures[endorserID.DID()] = hex.EncodeToString(sig)

	if len(alert.Signatures) >= sie.QuorumThreshold && !alert.Slashed {
		sie.executeSlashing(alert)
	}

	return alert, nil
}

// executeSlashing castiga y neutraliza para siempre al DID agresor en toda la pila
func (sie *SentinelImmunologyEngine) executeSlashing(alert *ImmunologicalAlert) {
	alert.QuorumReached = true
	alert.Slashed = true
	offender := alert.OffenderDID

	// 1. Marcar como neutralizado en el registro inmunológico
	sie.Neutralized[offender] = time.Now().UTC()

	// 2. Expulsión en el cortafuegos ZTNA: Revocar permisos y bloquear
	if sie.Firewall != nil {
		sie.Firewall.RevokeDID(offender)
		sie.Firewall.AuthorizeDID(&l1.DIDPolicy{
			DID:           offender,
			AllowInbound:  false,
			AllowOutbound: false,
			AllowRelay:    false,
		})
	}

	// 3. Castigo Tit-for-Tat: Degradación inmediata y persistente a estrangulamiento total (THROTTLED)
	if sie.Accounting != nil {
		sie.Accounting.SlashPeer(offender)
	}
}

// IsSlashed verifica si una identidad ha sido expulsada de por vida
func (sie *SentinelImmunologyEngine) IsSlashed(did string) bool {
	sie.mu.RLock()
	defer sie.mu.RUnlock()

	_, ok := sie.Neutralized[did]
	return ok
}

// GetActiveAlerts retorna todas las alertas registradas
func (sie *SentinelImmunologyEngine) GetActiveAlerts() []*ImmunologicalAlert {
	sie.mu.RLock()
	defer sie.mu.RUnlock()

	res := make([]*ImmunologicalAlert, 0, len(sie.Alerts))
	for _, a := range sie.Alerts {
		res = append(res, a)
	}
	return res
}

// GetNeutralizedDIDs entrega la lista de DIDs expulsados
func (sie *SentinelImmunologyEngine) GetNeutralizedDIDs() []string {
	sie.mu.RLock()
	defer sie.mu.RUnlock()

	res := make([]string, 0, len(sie.Neutralized))
	for did := range sie.Neutralized {
		res = append(res, did)
	}
	return res
}

// GetStats consolida las métricas del sistema inmunológico
func (sie *SentinelImmunologyEngine) GetStats() SentinelStats {
	sie.mu.RLock()
	defer sie.mu.RUnlock()

	neutralizedList := make([]string, 0, len(sie.Neutralized))
	for did := range sie.Neutralized {
		neutralizedList = append(neutralizedList, did)
	}

	return SentinelStats{
		TotalAuditsExecuted: sie.totalAudits.Load(),
		ActiveSentinels:     1, // Este nodo
		TotalAlertsEmitted:  sie.totalAlerts.Load(),
		QuorumThreshold:     sie.QuorumThreshold,
		NeutralizedDIDs:     neutralizedList,
	}
}
