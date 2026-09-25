package l3

import (
	"testing"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
)

func TestAgentSenateEngine(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	verifier := NewConstitutionalVerifier(id)
	dag := l1.NewDAGStore(id)
	wot := l1.NewWebOfTrust()
	acc := l2.NewTransitAccounting()

	senate := NewAgentSenateEngine(id, verifier, dag, wot, acc)

	// 1. Probar propuesta rechazada por violar el Artículo I (telemetría externa)
	dirtyCode := `import "net/http" ; func leak() { http.Get("http://leak.com/stats") }`
	_, err = senate.SubmitProposal("Parche con Fuga", "Fuga de analítica", CategoryRoutingOptimization, dirtyCode)
	if err == nil {
		t.Errorf("la propuesta con telemetría debió ser rechazada por el verificador constitucional")
	}

	// 2. Probar propuesta legítima
	cleanCode := `func optimizeRouting() { /* bypass eBPF local */ }`
	prop, err := senate.SubmitProposal("Optimización eBPF v2", "Mejora bypass XDP", CategoryRoutingOptimization, cleanCode)
	if err != nil {
		t.Fatalf("error creando propuesta válida: %v", err)
	}
	if prop.ID == "" || prop.Status != StatusDebating {
		t.Errorf("propuesta inválida: %+v", prop)
	}

	// 3. Agregar argumento técnico con métricas de sandbox
	metrics := TechnicalMetrics{
		BandwidthSavingsPct: 15.2,
		LatencyImpactMs:     -1.8,
		MemoryDeltaMB:       -2.4,
		SandboxStatus:       "PASSED",
	}
	arg, err := senate.AddArgument(prop.ID, StanceSupport, metrics, "Optimiza 15% el ancho de banda sin alterar el L0")
	if err != nil {
		t.Fatalf("error agregando argumento de debate: %v", err)
	}
	if !arg.ConstitutionalCompliance || arg.SignatureHex == "" {
		t.Errorf("argumento de debate inválido: %+v", arg)
	}

	// 4. Emitir voto de agente basado en Proof-of-Contribution
	vote, err := senate.CastAgentVote(prop.ID, StanceSupport, "Voto a favor tras pasar validación de sandbox y Artículo I")
	if err != nil {
		t.Fatalf("error emitiendo voto: %v", err)
	}
	if vote.Weight <= 0 || vote.HumanVeto {
		t.Errorf("voto de agente inválido: %+v", vote)
	}

	// 5. Generar Reporte Matutino para el humano
	report := senate.GenerateMorningReport()
	if report.VotesCastByAgent != 1 || len(report.Decisions) != 1 {
		t.Errorf("reporte matutino incorrecto: %+v", report)
	}
	if report.Decisions[0].Vetoed {
		t.Errorf("la decisión no debería estar vetada aún")
	}

	// 6. Aplicar Veto Soberano Humano
	err = senate.SovereignHumanVeto(prop.ID, "Prefiero mantener el enrutamiento anterior por precaución")
	if err != nil {
		t.Fatalf("error aplicando veto soberano humano: %v", err)
	}

	pUpdated, _ := senate.GetProposal(prop.ID)
	if pUpdated.Status != StatusVetoed {
		t.Errorf("el estado de la propuesta debió cambiar a VETOED_BY_SOVEREIGN")
	}

	// 7. Verificar que el Reporte Matutino refleje el veto
	reportPostVeto := senate.GenerateMorningReport()
	if !reportPostVeto.Decisions[0].Vetoed {
		t.Errorf("la decisión en el reporte matutino debió figurar como vetada")
	}
}
