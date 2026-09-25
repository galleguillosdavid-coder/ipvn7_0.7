package l3

import (
	"testing"
	"time"
)

func TestIntentEngine_ZTNAEvaluation(t *testing.T) {
	ie := NewIntentEngine()

	// 1. Probar Default Deny sin políticas
	reqDefault := IntentRequest{
		SubjectDID:   "did:ipvn7:guest-node",
		Action:       IntentActionChatSend,
		ContextFlags: []ContextFlag{ContextFlagFieldOps},
		Timestamp:    time.Now(),
	}
	dec := ie.EvaluateIntent(reqDefault)
	if dec.Allowed {
		t.Fatal("expected default deny for unregistered policy")
	}

	// 2. Registrar política para nodo A hacia Telemetría
	policyNodeA := IntentPolicy{
		PolicyID:             "pol-node-a-telem",
		SubjectDID:           "did:ipvn7:node-a",
		AllowedActions:       []IntentAction{IntentActionTelemetryPush, IntentActionDAGSync},
		RequiredContextFlags: []ContextFlag{ContextFlagMissionCritical},
		MaxRateLimitPerSec:   2,
	}
	if err := ie.AddPolicy(policyNodeA); err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}

	// 3. Probar solicitud sin el contexto requerido
	reqMissingCtx := IntentRequest{
		SubjectDID:   "did:ipvn7:node-a",
		Action:       IntentActionTelemetryPush,
		ContextFlags: []ContextFlag{ContextFlagFieldOps}, // Falta ContextFlagMissionCritical
		Timestamp:    time.Now(),
	}
	dec = ie.EvaluateIntent(reqMissingCtx)
	if dec.Allowed {
		t.Fatal("expected deny due to missing required context flag")
	}

	// 4. Probar solicitud válida con contexto completo
	reqValid := IntentRequest{
		SubjectDID:   "did:ipvn7:node-a",
		Action:       IntentActionTelemetryPush,
		ContextFlags: []ContextFlag{ContextFlagMissionCritical, ContextFlagFieldOps},
		Timestamp:    time.Now(),
	}
	dec = ie.EvaluateIntent(reqValid)
	if !dec.Allowed {
		t.Fatalf("expected allowed intent, got: %s", dec.Reason)
	}

	// 5. Probar límite de tasa (2 req/s)
	dec2 := ie.EvaluateIntent(reqValid)
	if !dec2.Allowed {
		t.Fatalf("expected 2nd request allowed, got: %s", dec2.Reason)
	}
	dec3 := ie.EvaluateIntent(reqValid)
	if dec3.Allowed {
		t.Fatal("expected rate limit exceeded on 3rd request")
	}

	// 6. Probar expiración de política
	expiredPolicy := IntentPolicy{
		PolicyID:       "pol-expired",
		SubjectDID:     "did:ipvn7:temp-node",
		AllowedActions: []IntentAction{IntentActionBenchmarkRun},
		ExpiresAt:      time.Now().Add(-1 * time.Hour), // Ya expirada
	}
	_ = ie.AddPolicy(expiredPolicy)

	reqExpired := IntentRequest{
		SubjectDID: "did:ipvn7:temp-node",
		Action:     IntentActionBenchmarkRun,
	}
	decExpired := ie.EvaluateIntent(reqExpired)
	if decExpired.Allowed {
		t.Fatal("expected deny for expired policy")
	}

	// 7. Probar eliminación de política
	ie.RemovePolicy("pol-node-a-telem")
	decRemoved := ie.EvaluateIntent(reqValid)
	if decRemoved.Allowed {
		t.Fatal("expected deny after policy was removed")
	}
}
