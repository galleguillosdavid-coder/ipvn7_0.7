package l3

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

// AgentSenateEngine orquesta el Senado de Agentes y la Democracia Líquida
type AgentSenateEngine struct {
	mu         sync.RWMutex
	Identity   *l0.Identity
	Verifier   *ConstitutionalVerifier
	DAGStore   *l1.DAGStore
	WoT        *l1.WebOfTrust
	Accounting *l2.TransitAccounting
	Proposals  map[string]*SenateProposal
	MyVotes    map[string]*AgentVote
}

// NewAgentSenateEngine inicializa el motor del Senado de Agentes
func NewAgentSenateEngine(
	id *l0.Identity,
	verifier *ConstitutionalVerifier,
	dag *l1.DAGStore,
	wot *l1.WebOfTrust,
	acc *l2.TransitAccounting,
) *AgentSenateEngine {
	return &AgentSenateEngine{
		Identity:   id,
		Verifier:   verifier,
		DAGStore:   dag,
		WoT:        wot,
		Accounting: acc,
		Proposals:  make(map[string]*SenateProposal),
		MyVotes:    make(map[string]*AgentVote),
	}
}

// CalculateProofOfContributionScore calcula el peso cívico del nodo (anti-plutocracia)
func (s *AgentSenateEngine) CalculateProofOfContributionScore(did string) float64 {
	baseScore := 50.0 // Reputación inicial base por hardware funcional

	if s.WoT != nil {
		wotScore, _ := s.WoT.CalculateReputation(s.Identity.DID(), did)
		baseScore += wotScore * 0.4
	}

	if s.Accounting != nil {
		if stats, ok := s.Accounting.GetPeerStats(did); ok {
			transitMB := float64(stats.BytesTx+stats.BytesRx) / (1024 * 1024)
			if transitMB > 30.0 {
				transitMB = 30.0
			}
			baseScore += transitMB
		}
	}

	if baseScore > 100.0 {
		baseScore = 100.0
	}
	return baseScore
}

// SubmitProposal ingresa una propuesta técnica al Senado con validación constitucional previa
func (s *AgentSenateEngine) SubmitProposal(title, desc string, cat ProposalCategory, sourceCode string) (*SenateProposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generar ID determinista basado en contenido
	raw := fmt.Sprintf("%s:%s:%s", title, cat, sourceCode)
	sum := sha256.Sum256([]byte(raw))
	propID := fmt.Sprintf("cid:prop:%s", hex.EncodeToString(sum[:16]))

	// 1. Auditoría Constitucional Estricta previa
	report, err := s.Verifier.VerifyCode(sourceCode, propID)
	if err != nil {
		return nil, fmt.Errorf("fallo durante verificación constitucional: %w", err)
	}

	if !report.Certified {
		return nil, fmt.Errorf("propuesta RECHAZADA por violar la Constitución Digital (Artículo I): %s", report.Violations[0].Description)
	}

	prop := &SenateProposal{
		ID:                   propID,
		Title:                title,
		Description:          desc,
		Category:             cat,
		AuthorDID:            s.Identity.DID(),
		SourceCode:           sourceCode,
		CreatedAt:            time.Now().UTC(),
		ExpiresAt:            time.Now().UTC().Add(48 * time.Hour),
		Arguments:            make([]DebateArgument, 0),
		Votes:                make(map[string]*AgentVote),
		Status:               StatusDebating,
		ConstitutionalReport: report,
	}

	// Persistir en DAG Store inmutable si está disponible
	if s.DAGStore != nil {
		_, _ = s.DAGStore.PutBlock([]byte(fmt.Sprintf("PROPOSAL:%s:%s", propID, title)), nil, "")
	}

	s.Proposals[propID] = prop
	return prop, nil
}

// AddArgument inyecta un argumento técnico con justificación semántica en el debate
func (s *AgentSenateEngine) AddArgument(proposalID string, stance StanceType, metrics TechnicalMetrics, vector string) (*DebateArgument, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop, exists := s.Proposals[proposalID]
	if !exists {
		return nil, errors.New("propuesta no encontrada en el Senado")
	}

	if prop.Status != StatusDebating {
		return nil, fmt.Errorf("la propuesta no está en debate (estado: %s)", prop.Status)
	}

	pocScore := s.CalculateProofOfContributionScore(s.Identity.DID())
	compliant := metrics.SandboxStatus == "PASSED" && prop.ConstitutionalReport.Certified

	sigData := []byte(fmt.Sprintf("%s:%s:%s:%.2f", proposalID, stance, vector, pocScore))
	sig := s.Identity.Sign(sigData)

	arg := DebateArgument{
		ProposalID:               proposalID,
		SenderAgent:              "agent:" + s.Identity.DID()[:18],
		NodeOwnerDID:             s.Identity.DID(),
		Timestamp:                time.Now().UTC(),
		ProofOfContributionScore: pocScore,
		Stance:                   stance,
		TechnicalMetrics:         metrics,
		ArgumentVector:           vector,
		ConstitutionalCompliance: compliant,
		SignatureHex:             hex.EncodeToString(sig),
	}

	prop.Arguments = append(prop.Arguments, arg)
	return &arg, nil
}

// CastAgentVote emite el voto automatizado de la IA en representación del nodo
func (s *AgentSenateEngine) CastAgentVote(proposalID string, stance StanceType, justification string) (*AgentVote, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop, exists := s.Proposals[proposalID]
	if !exists {
		return nil, errors.New("propuesta no encontrada en el Senado")
	}

	if prop.Status != StatusDebating {
		return nil, fmt.Errorf("la propuesta no admite votos (estado: %s)", prop.Status)
	}

	weight := s.CalculateProofOfContributionScore(s.Identity.DID())
	sigData := []byte(fmt.Sprintf("VOTE:%s:%s:%.2f", proposalID, stance, weight))
	sig := s.Identity.Sign(sigData)

	vote := &AgentVote{
		ProposalID:    proposalID,
		AgentDID:      "agent:" + s.Identity.DID()[:18],
		NodeOwnerDID:  s.Identity.DID(),
		Stance:        stance,
		Weight:        weight,
		Justification: justification,
		Timestamp:     time.Now().UTC(),
		HumanVeto:     false,
		SignatureHex:  hex.EncodeToString(sig),
	}

	// Registrar voto y actualizar conteo ponderado
	prop.Votes[s.Identity.DID()] = vote
	s.MyVotes[proposalID] = vote

	s.recalculateTally(prop)
	return vote, nil
}

// SovereignHumanVeto ejerce el veto soberano inalienable del usuario humano sobre una decisión de su IA
func (s *AgentSenateEngine) SovereignHumanVeto(proposalID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prop, exists := s.Proposals[proposalID]
	if !exists {
		return errors.New("propuesta no encontrada")
	}

	vote, voted := s.MyVotes[proposalID]
	if !voted {
		return errors.New("el agente local no ha emitido votos sobre esta propuesta")
	}

	// Marcar veto y anular el peso del voto del agente local
	vote.HumanVeto = true
	vote.Justification = fmt.Sprintf("[VETO SOBERANO HUMANO]: %s (Original: %s)", reason, vote.Justification)
	prop.VetoReason = reason
	prop.Status = StatusVetoed

	s.recalculateTally(prop)
	return nil
}

func (s *AgentSenateEngine) recalculateTally(prop *SenateProposal) {
	var support, oppose float64
	for _, v := range prop.Votes {
		if v.HumanVeto {
			continue // Voto vetado por el humano queda sin peso
		}
		if v.Stance == StanceSupport {
			support += v.Weight
		} else if v.Stance == StanceOppose {
			oppose += v.Weight
		}
	}
	prop.SupportWeight = support
	prop.OpposeWeight = oppose

	// Umbral simple de aprobación (100 puntos de consenso neto y más soporte que oposición)
	if support+oppose >= 100.0 && prop.Status == StatusDebating {
		if support > oppose*1.5 {
			prop.Status = StatusApproved
		} else if oppose > support*1.5 {
			prop.Status = StatusRejected
		}
	}
}

// GenerateMorningReport consolida el resumen ejecutivo de supervisión para el usuario humano
func (s *AgentSenateEngine) GenerateMorningReport() *MorningReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var decisions []MorningReportItem
	for propID, vote := range s.MyVotes {
		title := propID
		valid := true
		if p, ok := s.Proposals[propID]; ok {
			title = p.Title
			valid = p.ConstitutionalReport != nil && p.ConstitutionalReport.Certified
		}

		decisions = append(decisions, MorningReportItem{
			ProposalID:          propID,
			Title:               title,
			AgentStance:         vote.Stance,
			WeightUsed:          vote.Weight,
			Justification:       vote.Justification,
			Vetoed:              vote.HumanVeto,
			ConstitutionalValid: valid,
		})
	}

	return &MorningReport{
		Date:               time.Now().UTC().Format("2006-01-02"),
		CycleID:            fmt.Sprintf("cycle-%d", time.Now().UTC().Day()),
		TotalActiveDebates: len(s.Proposals),
		VotesCastByAgent:   len(decisions),
		Decisions:          decisions,
		GeneratedAt:        time.Now().UTC(),
	}
}

// ListProposals lista todas las propuestas activas e históricas del Senado
func (s *AgentSenateEngine) ListProposals() []*SenateProposal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*SenateProposal, 0, len(s.Proposals))
	for _, p := range s.Proposals {
		res = append(res, p)
	}
	return res
}

// GetProposal obtiene una propuesta específica por su ID
func (s *AgentSenateEngine) GetProposal(id string) (*SenateProposal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.Proposals[id]
	return p, ok
}
