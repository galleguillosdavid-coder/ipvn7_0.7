package l2

import (
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
)

func TestSentinelImmunologyEngine(t *testing.T) {
	sentinelID1, _ := l0.GenerateIdentity()
	sentinelID2, _ := l0.GenerateIdentity()
	offenderID, _ := l0.GenerateIdentity()

	fw := l1.NewZTNAFirewall(true)
	acc := NewTransitAccounting()
	wot := l1.NewWebOfTrust()

	sie := NewSentinelImmunologyEngine(sentinelID1, fw, acc, wot)

	// 1. Probar emisión de desafío de auditoría
	chal, err := sie.IssueAuditChallenge(offenderID.DID())
	if err != nil {
		t.Fatalf("error emitiendo desafío: %v", err)
	}
	if chal.ChallengeID == "" || chal.ExpectedSandboxHash == "" {
		t.Errorf("desafío inválido: %+v", chal)
	}

	// 2. Probar respuesta fraudulenta (debe fallar y emitir alerta)
	badResp := &AuditResponse{
		ChallengeID: chal.ChallengeID,
		TargetDID:   offenderID.DID(),
		AnswerHash:  "hash_corrompido_o_alucinado",
	}
	passed, err := sie.VerifyAuditResponse(chal.ChallengeID, badResp)
	if err != nil {
		t.Fatalf("error verificando respuesta: %v", err)
	}
	if passed {
		t.Errorf("respuesta fraudulenta NO debió ser aprobada")
	}

	// Comprobar que se generó una alerta
	alerts := sie.GetActiveAlerts()
	if len(alerts) == 0 {
		t.Fatalf("debió generarse una alerta inmunológica tras auditoría fallida")
	}
	alert := alerts[0]
	if alert.Incident != IncidentDishonestAI || alert.OffenderDID != offenderID.DID() {
		t.Errorf("alerta con datos incorrectos: %+v", alert)
	}
	if alert.QuorumReached {
		t.Errorf("el quórum no debería estar completo con sólo 1 firma (umbral 2)")
	}

	// 3. Respaldo (Endorsement) por segundo centinela independiente
	alertUpdated, err := sie.EndorseAlert(alert.AlertID, sentinelID2)
	if err != nil {
		t.Fatalf("error respaldando alerta: %v", err)
	}
	if !alertUpdated.QuorumReached || !alertUpdated.Slashed {
		t.Errorf("con 2 centinelas debió alcanzarse quórum y ejecutar slashing: %+v", alertUpdated)
	}

	// 4. Verificar que el agresor esté en la lista negra y neutralizado
	if !sie.IsSlashed(offenderID.DID()) {
		t.Errorf("el DID agresor debió figurar como neutralizado")
	}

	// Verificar bloqueo en cortafuegos ZTNA
	policy, ok := fw.GetPolicy(offenderID.DID())
	if !ok || policy.AllowInbound || policy.AllowOutbound {
		t.Errorf("el cortafuegos debió revocar y denegar todo tráfico para el agresor")
	}

	// Verificar degradación real en contabilidad Tit-for-Tat
	tier := acc.GetPeerTier(offenderID.DID())
	if tier != TierThrottled {
		t.Errorf("el tier contable debió ser degradado a TierThrottled, obtenido: %v", tier)
	}

	// 5. Verificar estadísticas
	stats := sie.GetStats()
	if stats.TotalAlertsEmitted == 0 || len(stats.NeutralizedDIDs) != 1 {
		t.Errorf("estadísticas inconsistentes: %+v", stats)
	}
}
