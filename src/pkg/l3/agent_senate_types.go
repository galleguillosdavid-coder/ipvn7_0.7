package l3

import "time"

type ProposalCategory string

const (
	CategoryRoutingOptimization ProposalCategory = "ROUTING_OPTIMIZATION"
	CategorySecurityPatch       ProposalCategory = "SECURITY_PATCH"
	CategoryResourcePolicy      ProposalCategory = "RESOURCE_POLICY"
	CategoryConstitutionalAudit ProposalCategory = "CONSTITUTIONAL_AUDIT"
)

type ProposalStatus string

const (
	StatusDebating ProposalStatus = "DEBATING"
	StatusApproved ProposalStatus = "APPROVED"
	StatusRejected ProposalStatus = "REJECTED"
	StatusVetoed   ProposalStatus = "VETOED_BY_SOVEREIGN"
)

type StanceType string

const (
	StanceSupport StanceType = "SUPPORT"
	StanceOppose  StanceType = "OPPOSE"
	StanceAbstain StanceType = "ABSTAIN"
)

// TechnicalMetrics representa las mediciones objetivas auditadas en el sandbox
type TechnicalMetrics struct {
	BandwidthSavingsPct float64 `json:"bandwidth_savings_pct"`
	LatencyImpactMs     float64 `json:"latency_impact_ms"`
	MemoryDeltaMB       float64 `json:"memory_delta_mb"`
	SandboxStatus       string  `json:"sandbox_status"` // "PASSED" / "FAILED"
}

// DebateArgument es el paquete semántico de argumentación entre agentes (docs/GOBERNANZA.md)
type DebateArgument struct {
	ProposalID               string           `json:"proposal_id"`
	SenderAgent              string           `json:"sender_agent"`
	NodeOwnerDID             string           `json:"node_owner_did"`
	Timestamp                time.Time        `json:"timestamp"`
	ProofOfContributionScore float64          `json:"proof_of_contribution_score"`
	Stance                   StanceType       `json:"stance"`
	TechnicalMetrics         TechnicalMetrics `json:"technical_metrics"`
	ArgumentVector           string           `json:"argument_vector"`
	ConstitutionalCompliance bool             `json:"constitutional_compliance"`
	SignatureHex             string           `json:"signature"`
}

// AgentVote representa el voto emitido por el agente en delegación del usuario humano
type AgentVote struct {
	ProposalID    string     `json:"proposal_id"`
	AgentDID      string     `json:"agent_did"`
	NodeOwnerDID  string     `json:"node_owner_did"`
	Stance        StanceType `json:"stance"`
	Weight        float64    `json:"weight"` // Atado exclusivamente al Proof-of-Contribution
	Justification string     `json:"justification"`
	Timestamp     time.Time  `json:"timestamp"`
	HumanVeto     bool       `json:"human_veto"`
	SignatureHex  string     `json:"signature"`
}

// SenateProposal es una propuesta estructurada sujeta a debate en el Senado de Agentes
type SenateProposal struct {
	ID                   string                `json:"id"`
	Title                string                `json:"title"`
	Description          string                `json:"description"`
	Category             ProposalCategory      `json:"category"`
	AuthorDID            string                `json:"author_did"`
	SourceCode           string                `json:"source_code"`
	CreatedAt            time.Time             `json:"created_at"`
	ExpiresAt            time.Time             `json:"expires_at"`
	Arguments            []DebateArgument      `json:"arguments"`
	Votes                map[string]*AgentVote `json:"votes"`
	Status               ProposalStatus        `json:"status"`
	SupportWeight        float64               `json:"support_weight"`
	OpposeWeight         float64               `json:"oppose_weight"`
	ConstitutionalReport *ComplianceReport     `json:"constitutional_report"`
	VetoReason           string                `json:"veto_reason,omitempty"`
}

// MorningReportItem detalle de acción para el informe matutino del humano
type MorningReportItem struct {
	ProposalID          string     `json:"proposal_id"`
	Title               string     `json:"title"`
	AgentStance         StanceType `json:"agent_stance"`
	WeightUsed          float64    `json:"weight_used"`
	Justification       string     `json:"justification"`
	Vetoed              bool       `json:"vetoed"`
	ConstitutionalValid bool       `json:"constitutional_valid"`
}

// MorningReport es el resumen ejecutivo que el agente presenta al humano para supervisión
type MorningReport struct {
	Date               string              `json:"date"`
	CycleID            string              `json:"cycle_id"`
	TotalActiveDebates int                 `json:"total_active_debates"`
	VotesCastByAgent   int                 `json:"votes_cast_by_agent"`
	Decisions          []MorningReportItem `json:"decisions"`
	GeneratedAt        time.Time           `json:"generated_at"`
}
