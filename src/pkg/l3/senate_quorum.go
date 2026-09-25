package l3

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// QuorumRule define los umbrales de participación y aprobación para una categoría dada
type QuorumRule struct {
	MinTurnoutPct      float64 `json:"min_turnout_pct"`      // p.ej. 0.50 (50%)
	ApprovalThreshold float64 `json:"approval_threshold"`  // p.ej. 0.66 (supermayoría) o 0.51
	MinUniqueVoters   int     `json:"min_unique_voters"`   // Mínimo de agentes distintos
}

// QuorumEvaluation almacena el dictamen objetivo de un recuento del senado
type QuorumEvaluation struct {
	ProposalID       string         `json:"proposal_id"`
	TotalWeight      float64        `json:"total_weight"`
	CastWeight       float64        `json:"cast_weight"`
	TurnoutPct       float64        `json:"turnout_pct"`
	SupportWeight    float64        `json:"support_weight"`
	OpposeWeight     float64        `json:"oppose_weight"`
	AbstainWeight    float64        `json:"abstain_weight"`
	ApprovalPct      float64        `json:"approval_pct"`
	QuorumReached    bool           `json:"quorum_reached"`
	Passed           bool           `json:"passed"`
	Status           ProposalStatus `json:"status"`
	Reason           string         `json:"reason"`
}

// SenateQuorumEngine gobierna la contabilidad de quórum soberano en L3
type SenateQuorumEngine struct {
	mu           sync.RWMutex
	rules        map[ProposalCategory]QuorumRule
	eligibleKeys map[string]ed25519.PublicKey // DID -> ed25519 PubKey
	voterWeights map[string]float64           // DID -> ProofOfContribution weight
}

// NewSenateQuorumEngine instancia el motor con reglas canónicas de gobernanza
func NewSenateQuorumEngine() *SenateQuorumEngine {
	e := &SenateQuorumEngine{
		rules:        make(map[ProposalCategory]QuorumRule),
		eligibleKeys: make(map[string]ed25519.PublicKey),
		voterWeights: make(map[string]float64),
	}
	// Reglas canónicas
	e.rules[CategoryConstitutionalAudit] = QuorumRule{MinTurnoutPct: 0.60, ApprovalThreshold: 0.6667, MinUniqueVoters: 2}
	e.rules[CategorySecurityPatch] = QuorumRule{MinTurnoutPct: 0.50, ApprovalThreshold: 0.6667, MinUniqueVoters: 2}
	e.rules[CategoryRoutingOptimization] = QuorumRule{MinTurnoutPct: 0.33, ApprovalThreshold: 0.5001, MinUniqueVoters: 1}
	e.rules[CategoryResourcePolicy] = QuorumRule{MinTurnoutPct: 0.40, ApprovalThreshold: 0.5001, MinUniqueVoters: 1}
	return e
}

// RegisterVoter incorpora un par a la asamblea de votación con su peso auditado
func (e *SenateQuorumEngine) RegisterVoter(did string, pubKey ed25519.PublicKey, weight float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eligibleKeys[did] = pubKey
	if weight <= 0 {
		weight = 1.0
	}
	e.voterWeights[did] = weight
}

// VerifyVoteSignature verifica la firma criptográfica Ed25519 de un voto
func (e *SenateQuorumEngine) VerifyVoteSignature(v *AgentVote) error {
	e.mu.RLock()
	pubKey, ok := e.eligibleKeys[v.AgentDID]
	e.mu.RUnlock()

	if !ok || len(pubKey) != ed25519.PublicKeySize {
		return errors.New("voter public key not registered or invalid")
	}

	sigBytes, err := hex.DecodeString(v.SignatureHex)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature hex payload: %w", err)
	}

	msg := []byte(fmt.Sprintf("%s:%s:%s:%.4f", v.ProposalID, v.AgentDID, v.Stance, v.Weight))
	if !ed25519.Verify(pubKey, msg, sigBytes) {
		return errors.New("cryptographic signature verification failed")
	}
	return nil
}

// EvaluateQuorum computa el veredicto final sobre una propuesta del Senado
func (e *SenateQuorumEngine) EvaluateQuorum(p *SenateProposal) QuorumEvaluation {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rule, ok := e.rules[p.Category]
	if !ok {
		rule = QuorumRule{MinTurnoutPct: 0.50, ApprovalThreshold: 0.5001, MinUniqueVoters: 1}
	}

	var totalEligibleWeight float64
	for _, w := range e.voterWeights {
		totalEligibleWeight += w
	}
	if totalEligibleWeight == 0 {
		totalEligibleWeight = 1.0
	}

	var castWeight, supportWeight, opposeWeight, abstainWeight float64
	validVoters := 0

	for did, vote := range p.Votes {
		if vote.HumanVeto {
			return QuorumEvaluation{
				ProposalID: p.ID,
				Status:     StatusVetoed,
				Reason:     fmt.Sprintf("Vetoed by human sovereign: %s", vote.Justification),
			}
		}

		w, exists := e.voterWeights[did]
		if !exists {
			w = vote.Weight
			if w <= 0 {
				w = 1.0
			}
		}

		validVoters++
		castWeight += w
		switch vote.Stance {
		case StanceSupport:
			supportWeight += w
		case StanceOppose:
			opposeWeight += w
		case StanceAbstain:
			abstainWeight += w
		}
	}

	turnout := castWeight / totalEligibleWeight
	activeVotingWeight := supportWeight + opposeWeight
	var approvalPct float64
	if activeVotingWeight > 0 {
		approvalPct = supportWeight / activeVotingWeight
	}

	eval := QuorumEvaluation{
		ProposalID:    p.ID,
		TotalWeight:   totalEligibleWeight,
		CastWeight:    castWeight,
		TurnoutPct:    turnout,
		SupportWeight: supportWeight,
		OpposeWeight:  opposeWeight,
		AbstainWeight: abstainWeight,
		ApprovalPct:   approvalPct,
	}

	// Comprobación de quórum
	if turnout < rule.MinTurnoutPct || validVoters < rule.MinUniqueVoters {
		eval.QuorumReached = false
		eval.Passed = false
		eval.Status = StatusRejected
		eval.Reason = fmt.Sprintf("Quorum NOT reached: turnout %.2f%% < required %.2f%% or voters %d < %d",
			turnout*100, rule.MinTurnoutPct*100, validVoters, rule.MinUniqueVoters)
		return eval
	}

	eval.QuorumReached = true
	if approvalPct >= rule.ApprovalThreshold {
		eval.Passed = true
		eval.Status = StatusApproved
		eval.Reason = fmt.Sprintf("Proposal APPROVED with %.2f%% consensus (threshold %.2f%%)",
			approvalPct*100, rule.ApprovalThreshold*100)
	} else {
		eval.Passed = false
		eval.Status = StatusRejected
		eval.Reason = fmt.Sprintf("Proposal REJECTED with %.2f%% consensus below threshold %.2f%%",
			approvalPct*100, rule.ApprovalThreshold*100)
	}

	return eval
}
