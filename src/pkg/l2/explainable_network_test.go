package l2

import (
	"strings"
	"testing"
)

func TestExplainableNetworkEngine_SelectAndExplain(t *testing.T) {
	engine := NewExplainableNetworkEngine()

	// 1. Caso sin candidatos
	_, err := engine.SelectAndExplainRoute("did:ipvn7:target", nil)
	if err == nil {
		t.Fatal("expected error on empty candidates, got nil")
	}

	// 2. Caso con candidato único
	single := []CandidateMetrics{
		{
			PeerDID:         "did:ipvn7:node-b",
			LatencyMs:       12.0,
			RFC3550JitterMs: 0.05,
			PacketLossPct:   0.0,
			BatteryPct:      95,
			TrustTier:       3,
		},
	}
	explSingle, err := engine.SelectAndExplainRoute("did:ipvn7:target", single)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if explSingle.SelectedPeer != "did:ipvn7:node-b" {
		t.Fatalf("expected node-b selected, got: %s", explSingle.SelectedPeer)
	}
	if !strings.Contains(explSingle.HumanExplanation, "unico salto disponible") {
		t.Fatalf("explanation mismatch: %s", explSingle.HumanExplanation)
	}

	// 3. Caso competitivo: Nodo B (enlace limpio y seguro) vs Nodo C (enlace ruidoso y saturado)
	candidates := []CandidateMetrics{
		{
			PeerDID:         "did:ipvn7:node-c-noisy",
			LatencyMs:       85.0,
			RFC3550JitterMs: 14.5,
			PacketLossPct:   8.0,
			BatteryPct:      20,
			TrustTier:       1,
		},
		{
			PeerDID:         "did:ipvn7:node-b-clean",
			LatencyMs:       15.0,
			RFC3550JitterMs: 0.04,
			PacketLossPct:   0.0,
			BatteryPct:      90,
			TrustTier:       3,
		},
	}

	expl, err := engine.SelectAndExplainRoute("did:ipvn7:target", candidates)
	if err != nil {
		t.Fatalf("SelectAndExplainRoute failed: %v", err)
	}

	if expl.SelectedPeer != "did:ipvn7:node-b-clean" {
		t.Fatalf("expected node-b-clean to win, got: %s", expl.SelectedPeer)
	}

	if expl.SelectedScore <= expl.EvaluatedCandidates[1].TotalScore {
		t.Fatalf("winning score (%.1f) should be strictly greater than second place (%.1f)",
			expl.SelectedScore, expl.EvaluatedCandidates[1].TotalScore)
	}

	if !strings.Contains(expl.HumanExplanation, "Ventaja competitiva") {
		t.Fatalf("expected competitive explanation, got: %s", expl.HumanExplanation)
	}
}
