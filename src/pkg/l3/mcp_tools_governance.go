package l3

import (
	"encoding/json"
	"fmt"

	"ipvn7/pkg/l2"
)

// ExecuteGovernanceTool ejecuta las herramientas del Senado, Centinelas y Ley Computable
func (s *MCPServer) ExecuteGovernanceTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "ipvn7_constitution_verify":
		code, _ := args["code"].(string)
		targetID, _ := args["target_id"].(string)
		if code == "" {
			return "", fmt.Errorf("parámetro 'code' requerido")
		}
		if s.Constitution == nil {
			s.Constitution = NewConstitutionalVerifier(s.Identity)
		}
		report, err := s.Constitution.VerifyCode(code, targetID)
		if err != nil {
			return "", fmt.Errorf("error verificando código: %w", err)
		}
		data, _ := json.MarshalIndent(report, "", "  ")
		return string(data), nil

	case "ipvn7_senate_propose":
		if s.Senate == nil {
			return "", fmt.Errorf("subsistema del Senado de Agentes no inicializado")
		}
		title, _ := args["title"].(string)
		desc, _ := args["description"].(string)
		catStr, _ := args["category"].(string)
		code, _ := args["source_code"].(string)
		if title == "" || code == "" {
			return "", fmt.Errorf("parámetros 'title' y 'source_code' requeridos")
		}
		cat := ProposalCategory(catStr)
		if cat == "" {
			cat = CategoryRoutingOptimization
		}
		prop, err := s.Senate.SubmitProposal(title, desc, cat, code)
		if err != nil {
			return "", fmt.Errorf("propuesta rechazada: %w", err)
		}
		data, _ := json.MarshalIndent(prop, "", "  ")
		return string(data), nil

	case "ipvn7_senate_vote":
		if s.Senate == nil {
			return "", fmt.Errorf("subsistema del Senado de Agentes no inicializado")
		}
		propID, _ := args["proposal_id"].(string)
		stanceStr, _ := args["stance"].(string)
		just, _ := args["justification"].(string)
		if propID == "" || stanceStr == "" {
			return "", fmt.Errorf("parámetros 'proposal_id' y 'stance' requeridos")
		}
		vote, err := s.Senate.CastAgentVote(propID, StanceType(stanceStr), just)
		if err != nil {
			return "", fmt.Errorf("error emitiendo voto: %w", err)
		}
		data, _ := json.MarshalIndent(vote, "", "  ")
		return string(data), nil

	case "ipvn7_senate_morning_report":
		if s.Senate == nil {
			return "", fmt.Errorf("subsistema del Senado de Agentes no inicializado")
		}
		report := s.Senate.GenerateMorningReport()
		data, _ := json.MarshalIndent(report, "", "  ")
		return string(data), nil

	case "ipvn7_sentinel_audit":
		if s.Sentinel == nil {
			return "", fmt.Errorf("subsistema de Inmunología Centinela no inicializado")
		}
		targetDID, _ := args["target_did"].(string)
		if targetDID == "" {
			return "", fmt.Errorf("parámetro 'target_did' requerido")
		}
		chal, err := s.Sentinel.IssueAuditChallenge(targetDID)
		if err != nil {
			return "", fmt.Errorf("error emitiendo desafío: %w", err)
		}
		data, _ := json.MarshalIndent(chal, "", "  ")
		return string(data), nil

	case "ipvn7_sentinel_report_incident":
		if s.Sentinel == nil {
			return "", fmt.Errorf("subsistema de Inmunología Centinela no inicializado")
		}
		incStr, _ := args["incident"].(string)
		offender, _ := args["offender_did"].(string)
		evid, _ := args["evidence"].(string)
		if offender == "" || evid == "" {
			return "", fmt.Errorf("parámetros 'offender_did' y 'evidence' requeridos")
		}
		alert, err := s.Sentinel.EmitImmunologicalAlert(l2.IncidentType(incStr), offender, evid)
		if err != nil {
			return "", fmt.Errorf("error emitiendo alerta inmunológica: %w", err)
		}
		data, _ := json.MarshalIndent(alert, "", "  ")
		return string(data), nil

	case "ipvn7_law_validate_rule":
		if s.LawEngine == nil {
			return "", fmt.Errorf("motor de Derecho Computable no inicializado")
		}
		ctxDomain, _ := args["context_domain"].(string)
		attr, _ := args["attribute"].(string)
		deonticStr, _ := args["deontic"].(string)
		aim, _ := args["aim"].(string)
		cond, _ := args["condition"].(string)
		orElse, _ := args["or_else"].(string)
		ruleID, _ := args["rule_id"].(string)
		prioFloat, _ := args["priority"].(float64)

		if ctxDomain == "" || attr == "" || deonticStr == "" || aim == "" {
			return "", fmt.Errorf("parámetros 'context_domain', 'attribute', 'deontic' y 'aim' son obligatorios")
		}

		rule := ADICORule{
			ID:            ruleID,
			ContextDomain: ctxDomain,
			Attribute:     attr,
			Deontic:       DeonticOp(deonticStr),
			Aim:           aim,
			Condition:     cond,
			OrElse:        orElse,
			Priority:      int(prioFloat),
		}

		registered, err := s.LawEngine.RegisterRule(rule)
		if err != nil {
			return "", fmt.Errorf("error al validar/registrar regla ADICO: %w", err)
		}
		data, _ := json.MarshalIndent(registered, "", "  ")
		return string(data), nil

	case "ipvn7_law_evaluate_action":
		if s.LawEngine == nil {
			return "", fmt.Errorf("motor de Derecho Computable no inicializado")
		}
		ctxDomain, _ := args["context_domain"].(string)
		action, _ := args["action"].(string)
		if ctxDomain == "" || action == "" {
			return "", fmt.Errorf("parámetros 'context_domain' y 'action' son obligatorios")
		}
		deontic, justification, allowed := s.LawEngine.EvaluateAction(ctxDomain, action)
		res := map[string]interface{}{
			"context_domain": ctxDomain,
			"action":         action,
			"deontic":        deontic,
			"justification":  justification,
			"allowed":        allowed,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		return string(data), nil

	case "ipvn7_law_list_rules":
		if s.LawEngine == nil {
			return "", fmt.Errorf("motor de Derecho Computable no inicializado")
		}
		filterCtx, _ := args["context_domain"].(string)
		allRules := s.LawEngine.ListAllRules()
		filtered := make([]*ADICORule, 0)
		for _, r := range allRules {
			if filterCtx == "" || r.ContextDomain == filterCtx {
				filtered = append(filtered, r)
			}
		}
		data, _ := json.MarshalIndent(filtered, "", "  ")
		return string(data), nil

	default:
		return "", fmt.Errorf("herramienta de gobernanza desconocida: %s", name)
	}
}

