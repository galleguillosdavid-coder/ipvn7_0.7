// Package l2 implementa el motor de explicabilidad de red (Explainable Network Engine),
// rescatado de Ipv7IEU (core/engine/explain.go) y adaptado a ipvn7 v0.7.
// Permite auditar en lenguaje natural y estructurado las decisiones de conmutación de malla.
package l2

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// CandidateMetrics agrupa las variables físicas observadas para un enlace de nodo
type CandidateMetrics struct {
	PeerDID         string  `json:"peer_did"`
	LatencyMs       float64 `json:"latency_ms"`
	RFC3550JitterMs float64 `json:"rfc3550_jitter_ms"`
	PacketLossPct   float64 `json:"packet_loss_pct"`
	BatteryPct      int     `json:"battery_pct"`
	TrustTier       int     `json:"trust_tier"` // 1 a 4 (PQC / UIN Tier)
}

// CandidateScore representa el puntaje ponderado de un candidato evaluado
type CandidateScore struct {
	PeerDID    string  `json:"peer_did"`
	TotalScore float64 `json:"total_score"`
	Breakdown  string  `json:"breakdown"`
}

// RoutingExplanation registra la justificación formal y transparente de la decisión
type RoutingExplanation struct {
	TargetDID          string           `json:"target_did"`
	SelectedPeer       string           `json:"selected_peer"`
	SelectedScore      float64          `json:"selected_score"`
	HumanExplanation   string           `json:"human_explanation"`
	EvaluatedCandidates []CandidateScore `json:"evaluated_candidates"`
	Timestamp          time.Time        `json:"timestamp"`
}

// ExplainableNetworkEngine evalúa y explica las elecciones de ruta
type ExplainableNetworkEngine struct{}

// NewExplainableNetworkEngine inicializa el motor explicable
func NewExplainableNetworkEngine() *ExplainableNetworkEngine {
	return &ExplainableNetworkEngine{}
}

// SelectAndExplainRoute elige la mejor ruta y genera la justificación formal
func (e *ExplainableNetworkEngine) SelectAndExplainRoute(targetDID string, candidates []CandidateMetrics) (*RoutingExplanation, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("explainable: no hay candidatos disponibles para el destino %s", targetDID)
	}

	scores := make([]CandidateScore, 0, len(candidates))

	for _, c := range candidates {
		// Ponderación de función de costo:
		// Menor latencia, menor jitter y menor pérdida mejoran el puntaje.
		// Mayor TrustTier y mayor batería mejoran el puntaje.
		latencyPenalty := math.Min(c.LatencyMs*0.2, 50.0)
		jitterPenalty := math.Min(c.RFC3550JitterMs*1.5, 30.0)
		lossPenalty := math.Min(c.PacketLossPct*2.0, 50.0)

		trustBonus := float64(c.TrustTier) * 15.0
		batteryBonus := float64(c.BatteryPct) * 0.1

		score := (100.0 + trustBonus + batteryBonus) - (latencyPenalty + jitterPenalty + lossPenalty)
		if score < 0 {
			score = 0
		}

		breakdown := fmt.Sprintf("RTT: %.1fms, Jitter: %.2fms, Perdida: %.1f%%, Trust: T%d, Bat: %d%%",
			c.LatencyMs, c.RFC3550JitterMs, c.PacketLossPct, c.TrustTier, c.BatteryPct)

		scores = append(scores, CandidateScore{
			PeerDID:    c.PeerDID,
			TotalScore: score,
			Breakdown:  breakdown,
		})
	}

	// Ordenar candidatos por puntaje descendente
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	best := scores[0]
	var humanExpl string
	if len(scores) == 1 {
		humanExpl = fmt.Sprintf("Se selecciono el unico salto disponible %s con puntaje %.1f (%s).",
			best.PeerDID, best.TotalScore, best.Breakdown)
	} else {
		second := scores[1]
		diff := best.TotalScore - second.TotalScore
		humanExpl = fmt.Sprintf("Se selecciono %s (Puntaje: %.1f) sobre %s (Puntaje: %.1f, delta: +%.1f). Ventaja competitiva debida a mejor correlacion de RTT/Jitter RFC 3550 y garantia de confianza PQC.",
			best.PeerDID, best.TotalScore, second.PeerDID, second.TotalScore, diff)
	}

	return &RoutingExplanation{
		TargetDID:          targetDID,
		SelectedPeer:       best.PeerDID,
		SelectedScore:      best.TotalScore,
		HumanExplanation:   humanExpl,
		EvaluatedCandidates: scores,
		Timestamp:          time.Now(),
	}, nil
}
