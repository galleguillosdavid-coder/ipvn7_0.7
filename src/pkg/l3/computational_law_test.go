package l3_test

import (
	"strings"
	"testing"
	"time"

	"ipvn7/pkg/l0"
	"ipvn7/pkg/l1"
	"ipvn7/pkg/l2"
	"ipvn7/pkg/l3"
)

func TestComputationalLaw_ConstitutionalSeeding(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("Error generando identidad: %v", err)
	}

	cle := l3.NewComputationalLawEngine(id)
	rules := cle.ListAllRules()

	if len(rules) < 6 {
		t.Errorf("Se esperaban al menos 6 axiomas constitucionales sembrados, se obtuvieron %d", len(rules))
	}

	// Verificar que la acción prohibida por axioma sea rechazada
	deontic, reason, allowed := cle.EvaluateAction("root", "inject_malicious_code_or_dos")
	if allowed || deontic != l3.DeonticProhibition {
		t.Errorf("La acción de no-agresión debió ser PROHIBITION, obtenida: %s (%s)", deontic, reason)
	}
}

func TestComputationalLaw_LexSuperiorRejection(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	cle := l3.NewComputationalLawEngine(id)

	// Intentar registrar una directiva en subred local que permita monetizar votos (viola Carta Magna)
	rebelRule := l3.ADICORule{
		ID:            "rule:rebel:buy_votes",
		ContextDomain: "subred.darkmarket.ipv7",
		Attribute:     "rich_nodes",
		Deontic:       l3.DeonticPermission, // Trata de PERMITIR lo que la Carta Magna PROHÍBE
		Aim:           "monetize_or_buy_voting_power",
		Condition:     "if_wealthy",
		OrElse:        "none",
		Priority:      l3.PriorityCommunity,
		Timestamp:     time.Now().UTC(),
	}

	_, err := cle.RegisterRule(rebelRule)
	if err == nil {
		t.Fatalf("RegisterRule debió fallar por colisión con Lex Superior (Carta Magna)")
	}
	t.Logf("Rechazo exitoso por Lex Superior: %v", err)
}

func TestComputationalLaw_AntinomyDetection(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	cle := l3.NewComputationalLawEngine(id)

	ruleA := l3.ADICORule{
		ID:            "rule:telecom:a",
		ContextDomain: "subred.telecom.ipv7",
		Attribute:     "packet_routers",
		Deontic:       l3.DeonticObligation,
		Aim:           "compress_payloads",
		Priority:      l3.PriorityCommunity,
	}

	ruleB := l3.ADICORule{
		ID:            "rule:telecom:b",
		ContextDomain: "subred.telecom.ipv7",
		Attribute:     "packet_routers",
		Deontic:       l3.DeonticProhibition,
		Aim:           "compress_payloads",
		Priority:      l3.PriorityCommunity,
	}

	isAntinomy, reason := cle.DetectAntinomy(ruleA, ruleB)
	if !isAntinomy {
		t.Errorf("DetectAntinomy debió detectar contradicción entre Obligación y Prohibición sobre el mismo Aim")
	}
	t.Logf("Antinomia detectada correctamente: %s", reason)
}

func TestComputationalLaw_LexSpecialisAndPosterior(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	cle := l3.NewComputationalLawEngine(id)

	t1 := time.Now().UTC().Add(-1 * time.Hour)
	t2 := time.Now().UTC()

	// Norma General Comunitaria (menor profundidad de subdominio)
	ruleGeneral := l3.ADICORule{
		ID:            "rule:general:mesh",
		ContextDomain: "subred.market.ipv7",
		Attribute:     "sellers",
		Deontic:       l3.DeonticPermission,
		Aim:           "p2p_trade_hours",
		Priority:      l3.PriorityCommunity,
		Timestamp:     t1,
	}

	// Norma Específica Local (mayor profundidad de subdominio)
	ruleSpecific := l3.ADICORule{
		ID:            "rule:specific:artisan",
		ContextDomain: "subred.market.artisan.coop.ipv7",
		Attribute:     "sellers",
		Deontic:       l3.DeonticObligation,
		Aim:           "p2p_trade_hours",
		Priority:      l3.PriorityCommunity,
		Timestamp:     t1,
	}

	resSpecialis := cle.ResolveConflict(ruleGeneral, ruleSpecific)
	if resSpecialis.WinningRule.ID != ruleSpecific.ID || resSpecialis.DoctrineUsed != "LEX_SPECIALIS" {
		t.Errorf("Lex Specialis falló: debió prevalecer ruleSpecific por mayor profundidad de contexto")
	}

	// Probar Lex Posterior (mismo nivel y contexto, pero timestamp t2 > t1)
	ruleNewer := l3.ADICORule{
		ID:            "rule:general:mesh:v2",
		ContextDomain: "subred.market.ipv7",
		Attribute:     "sellers",
		Deontic:       l3.DeonticPermission,
		Aim:           "p2p_trade_hours",
		Priority:      l3.PriorityCommunity,
		Timestamp:     t2,
	}

	resPosterior := cle.ResolveConflict(ruleGeneral, ruleNewer)
	if resPosterior.WinningRule.ID != ruleNewer.ID || resPosterior.DoctrineUsed != "LEX_POSTERIOR" {
		t.Errorf("Lex Posterior falló: debió prevalecer ruleNewer por timestamp más reciente")
	}
}

func TestComputationalLaw_EvaluateActionFlow(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	cle := l3.NewComputationalLawEngine(id)

	// Acción permitida implícitamente por subsidiaridad
	deontic, reason, allowed := cle.EvaluateAction("subred.gaming.ipv7", "host_virtual_tourney")
	if !allowed || deontic != l3.DeonticPermission {
		t.Errorf("Acción lícita debió ser permitida: %s (%s)", deontic, reason)
	}

	// Registrar una obligación local en subred educativa
	eduRule := l3.ADICORule{
		ID:            "rule:edu:cite_source",
		ContextDomain: "subred.edu.ipv7",
		Attribute:     "researchers",
		Deontic:       l3.DeonticObligation,
		Aim:           "cite_cryptographic_sources",
		Condition:     "on_publish",
		Priority:      l3.PrioritySectorial,
	}
	_, err := cle.RegisterRule(eduRule)
	if err != nil {
		t.Fatalf("Error registrando regla en subred.edu: %v", err)
	}

	d, r, ok := cle.EvaluateAction("subred.edu.ipv7", "cite_cryptographic_sources")
	if !ok || d != l3.DeonticObligation {
		t.Errorf("EvaluateAction debió reportar OBLIGATION para regla registrada: %s (%s)", d, r)
	}
}

func TestComputationalLaw_MCPIntegration(t *testing.T) {
	id, _ := l0.GenerateIdentity()
	router := l1.NewKleinbergRouter(id)
	telemetry := l2.NewTelemetryRingBuffer()
	mcp := l3.NewMCPServer(id, router, telemetry)
	cle := l3.NewComputationalLawEngine(id)
	mcp.AttachLawEngine(cle)

	// 1. Listar reglas vía MCP
	res, err := mcp.ExecuteTool("ipvn7_law_list_rules", map[string]interface{}{})
	if err != nil {
		t.Fatalf("ipvn7_law_list_rules falló: %v", err)
	}
	if !strings.Contains(res, "axiom:sovereignty") {
		t.Errorf("ipvn7_law_list_rules no incluyó axiomas constitucionales sembrados")
	}

	// 2. Validar y registrar regla válida vía MCP
	_, err = mcp.ExecuteTool("ipvn7_law_validate_rule", map[string]interface{}{
		"rule_id":        "rule:civic:p2p_mesh",
		"context_domain": "subred.civic.ipv7",
		"attribute":      "civic_nodes",
		"deontic":        "PERMISSION",
		"aim":            "relay_mesh_packets",
		"condition":      "battery_above_20",
		"or_else":        "power_save_mode",
		"priority":       10.0,
	})
	if err != nil {
		t.Fatalf("ipvn7_law_validate_rule falló para regla lícita: %v", err)
	}

	// 3. Evaluar acción permitida vía MCP
	evalRes, err := mcp.ExecuteTool("ipvn7_law_evaluate_action", map[string]interface{}{
		"context_domain": "subred.civic.ipv7",
		"action":         "relay_mesh_packets",
	})
	if err != nil {
		t.Fatalf("ipvn7_law_evaluate_action falló: %v", err)
	}
	if !strings.Contains(evalRes, "PERMISSION") || !strings.Contains(evalRes, `"allowed": true`) {
		t.Errorf("ipvn7_law_evaluate_action inesperado: %s", evalRes)
	}

	// 4. Evaluar acción prohibida por Lex Superior (Carta Magna)
	evalDeny, err := mcp.ExecuteTool("ipvn7_law_evaluate_action", map[string]interface{}{
		"context_domain": "subred.civic.ipv7",
		"action":         "inject_malicious_code_or_dos",
	})
	if err != nil {
		t.Fatalf("ipvn7_law_evaluate_action falló: %v", err)
	}
	if !strings.Contains(evalDeny, "PROHIBITION") || !strings.Contains(evalDeny, `"allowed": false`) {
		t.Errorf("ipvn7_law_evaluate_action debió prohibir con false: %s", evalDeny)
	}
}
