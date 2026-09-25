package l3

// getGovernanceTools retorna el catálogo MCP de herramientas del Senado, Centinelas y Ley Computable
func getGovernanceTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "ipvn7_constitution_verify",
			Description: "Audita estáticamente código, parches o scripts contra el Artículo I de la Constitución Digital de ipvn7 (Privacidad, Cero Telemetría, Anti-Plutocracia y Core Freeze L0)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Código fuente o fragmento de parche a auditar",
					},
					"target_id": map[string]interface{}{
						"type":        "string",
						"description": "Identificador o nombre de la propuesta/parche",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "ipvn7_senate_propose",
			Description: "Ingresa una propuesta técnica o de optimización al Senado de Agentes de la red (Democracia Líquida)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Título descriptivo de la moción",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Justificación técnica o de infraestructura",
					},
					"proposer_did": map[string]interface{}{
						"type":        "string",
						"description": "DID del proponente (opcional, usa el nodo local por defecto)",
					},
					"payload_action": map[string]interface{}{
						"type":        "string",
						"description": "Acción técnica a ejecutar en la red si es aprobada",
					},
				},
				"required": []string{"title", "description", "payload_action"},
			},
		},
		{
			Name:        "ipvn7_senate_vote",
			Description: "Emite un voto ponderado por reputación WOT o delega soberanía en un par para una propuesta del Senado",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"proposal_id": map[string]interface{}{
						"type":        "string",
						"description": "ID de la propuesta a votar",
					},
					"choice": map[string]interface{}{
						"type":        "string",
						"description": "Elección de voto: 'YES', 'NO', 'ABSTAIN', 'DELEGATE'",
						"enum":        []string{"YES", "NO", "ABSTAIN", "DELEGATE"},
					},
					"delegate_to": map[string]interface{}{
						"type":        "string",
						"description": "DID del par al que se delega el voto (requerido si choice es 'DELEGATE')",
					},
				},
				"required": []string{"proposal_id", "choice"},
			},
		},
		{
			Name:        "ipvn7_senate_morning_report",
			Description: "Retorna el resumen legislativo de las últimas 24h: propuestas activas, quórum alcanzado y balance de consensos",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_sentinel_audit",
			Description: "Ejecuta una auditoría de integridad en tiempo real sobre la cadena inmunológica del nodo local o pares conocidos",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did": map[string]interface{}{
						"type":        "string",
						"description": "DID del par a auditar (opcional, por defecto el nodo local)",
					},
				},
			},
		},
		{
			Name:        "ipvn7_sentinel_report_incident",
			Description: "Registra un incidente de seguridad o violación de directiva en la cadena de eventos forenses inmutables",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"incident": map[string]interface{}{
						"type":        "string",
						"description": "Tipo de incidente: 'CODE_INJECTION_ATTEMPT', 'SYBIL_ATTACK_DETECTED', 'DAG_IMMUTABILITY_BREACH', 'DISHONEST_AI_AGENT', 'METRIC_FALSIFICATION', 'UNAUTHORIZED_L0_MUTATION'",
					},
					"offender_did": map[string]interface{}{
						"type":        "string",
						"description": "DID de la entidad atacante",
					},
					"evidence": map[string]interface{}{
						"type":        "string",
						"description": "Prueba o traza del incidente",
					},
				},
				"required": []string{"incident", "offender_did", "evidence"},
			},
		},
		{
			Name:        "ipvn7_law_validate_rule",
			Description: "Valida y registra una norma ADICO en el motor de Derecho Computable (docs/GOBERNANZA.md), verificando antinomias y precedencia Lex Superior / Specialis / Posterior",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"rule_id": map[string]interface{}{
						"type":        "string",
						"description": "Identificador único de la norma (opcional, generado automáticamente si se omite)",
					},
					"context_domain": map[string]interface{}{
						"type":        "string",
						"description": "Dominio de contexto o jurisdicción (e.g. 'root', 'subred.edu.ipv7', 'market.p2p.ipv7')",
					},
					"attribute": map[string]interface{}{
						"type":        "string",
						"description": "Sujeto de la norma o rol afectado (e.g. 'all_nodes', 'civic_agent', 'transit_node')",
					},
					"deontic": map[string]interface{}{
						"type":        "string",
						"description": "Operador deóntico formal: 'OBLIGATION' (deber), 'PERMISSION' (facultad), 'PROHIBITION' (vedado)",
					},
					"aim": map[string]interface{}{
						"type":        "string",
						"description": "Acción semántica regulada u objetivo institucional",
					},
					"condition": map[string]interface{}{
						"type":        "string",
						"description": "Condición o circunstancia de aplicación lógica de la norma",
					},
					"or_else": map[string]interface{}{
						"type":        "string",
						"description": "Sanción o consecuencia graduada ante incumplimiento",
					},
					"priority": map[string]interface{}{
						"type":        "number",
						"description": "Nivel de prioridad jerárquico (100 = Constitucional, 50 = Sectorial, 10 = Comunitario)",
					},
				},
				"required": []string{"context_domain", "attribute", "deontic", "aim"},
			},
		},
		{
			Name:        "ipvn7_law_evaluate_action",
			Description: "Evalúa si una acción planificada por una entidad o agente IA está permitida, obligada o prohibida según la jerarquía normativa",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"context_domain": map[string]interface{}{
						"type":        "string",
						"description": "Dominio contextual o jurisdicción en la que se ejecuta la acción (e.g. 'root', 'subred.edu.ipv7')",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Acción semántica que se desea evaluar",
					},
				},
				"required": []string{"context_domain", "action"},
			},
		},
		{
			Name:        "ipvn7_law_list_rules",
			Description: "Lista las directivas institucionales y normas ADICO vigentes en una jurisdicción o en todo el ordenamiento de ipvn7",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"context_domain": map[string]interface{}{
						"type":        "string",
						"description": "Filtro opcional por dominio de contexto (e.g. 'root', 'subred.edu.ipv7'). Si se omite, retorna todas",
					},
				},
			},
		},
	}
}
