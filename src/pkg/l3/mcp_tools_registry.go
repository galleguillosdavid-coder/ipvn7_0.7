package l3

// GetSupportedTools retorna el catálogo de herramientas para el modelo de IA (Dimensión 6)
func (s *MCPServer) GetSupportedTools() []MCPTool {
	tools := []MCPTool{
		{
			Name:        "get_node_status",
			Description: "Obtiene el estado general del nodo ipvn7, DID soberano, IPs y métricas en tiempo real",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "list_peers",
			Description: "Lista todos los pares activos en los 12 anillos del enrutador geométrico Kleinberg",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "route_packet",
			Description: "Calcula el siguiente salto de enrutamiento voraz hacia un DID de destino",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dest_did": map[string]interface{}{
						"type":        "string",
						"description": "DID soberano del nodo destino",
					},
				},
				"required": []string{"dest_did"},
			},
		},
		{
			Name:        "ipvn7_firewall_rule",
			Description: "Consulta o configura políticas ZTNA Default-Deny para un DID soberano",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"description": "'get' para consultar o 'set' para autorizar",
						"enum":        []string{"get", "set"},
					},
					"did": map[string]interface{}{
						"type":        "string",
						"description": "DID objetivo a evaluar o autorizar",
					},
					"allow_inbound": map[string]interface{}{
						"type":        "boolean",
						"description": "Habilitar tráfico entrante (para action 'set')",
					},
					"allow_outbound": map[string]interface{}{
						"type":        "boolean",
						"description": "Habilitar tráfico saliente (para action 'set')",
					},
					"allow_relay": map[string]interface{}{
						"type":        "boolean",
						"description": "Habilitar retransmisión de tránsito (para action 'set')",
					},
				},
				"required": []string{"action", "did"},
			},
		},
		{
			Name:        "ipvn7_dag_put",
			Description: "Almacena datos inmutables en el DAG direccionado por contenido (CID) y firma con Ed25519",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"payload": map[string]interface{}{
						"type":        "string",
						"description": "Contenido de texto o datos a almacenar en el bloque",
					},
					"target_did": map[string]interface{}{
						"type":        "string",
						"description": "DID de destino opcional para encolado asíncrono DTN",
					},
				},
				"required": []string{"payload"},
			},
		},
		{
			Name:        "ipvn7_wot_vouch",
			Description: "Emite un aval criptográfico de confianza en la Web-of-Trust y calcula reputación",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"subject_did": map[string]interface{}{
						"type":        "string",
						"description": "DID del par que se está avalando",
					},
					"trust_level": map[string]interface{}{
						"type":        "number",
						"description": "Nivel de confianza de 0.0 a 1.0",
					},
					"reason": map[string]interface{}{
						"type":        "string",
						"description": "Justificación textual del aval",
					},
				},
				"required": []string{"subject_did", "trust_level"},
			},
		},
		{
			Name:        "ipvn7_ddns_resolve",
			Description: "Resuelve un petname mnemotécnico local (.ipv7) a DID soberano y direcciones IP",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre memorable (ej. notebook.ipv7)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "ipvn7_sas_derive",
			Description: "Deriva el código SAS (6 dígitos y 4 emojis) para emparejamiento visual seguro OOB",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did": map[string]interface{}{
						"type":        "string",
						"description": "DID del nodo con el cual verificar el emparejamiento",
					},
				},
				"required": []string{"peer_did"},
			},
		},
		{
			Name:        "ipvn7_ai_diagnose",
			Description: "Ejecuta diagnóstico autónomo con el Copiloto de IA y aplica auto-curación",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"recent_latency_ms": map[string]interface{}{
						"type":        "number",
						"description": "Latencia reciente observada (ms)",
					},
					"drops_count": map[string]interface{}{
						"type":        "integer",
						"description": "Conteo de descartes observados",
					},
				},
			},
		},
		{
			Name:        "ipvn7_query_kuzu",
			Description: "Ejecuta consultas en lenguaje openCypher sobre la base de datos de grafos de topología y pipeline PFO en KùzuDB",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Consulta openCypher (ej. MATCH (p:Peer)-[l:XOR_LINK]->(m:Peer) RETURN p.did, l.latency_ms LIMIT 100)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "ipvn7_pqc_status",
			Description: "Obtiene el estado de la criptografía híbrida cuántico-resistente (ML-DSA-65 firma dual y ML-KEM-768 encapsulación de clave)",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_sphinx_circuit",
			Description: "Construye y encapsula un circuito cebolla Sphinx de 3 saltos con tramas de 1280 bytes fijos y desprendimiento iterativo de capas (peeling)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"payload": map[string]interface{}{
						"type":        "string",
						"description": "Mensaje o payload a encapsular en el circuito cebolla (opcional)",
					},
				},
			},
		},
		{
			Name:        "ipvn7_socks5_status",
			Description: "Consulta el estado operativo del gateway universal SOCKS5 (RFC 1928), puertos e interceptación de tráfico comercial",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_nat_probe",
			Description: "Ejecuta un diagnóstico reflexivo STUN RFC 5389 para descubrir la IP pública, puerto y tipo de NAT/CGNAT",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_guardian_health",
			Description: "Inspecciona la salud del sistema, estado de memoria (Heap), descriptores y estado de Circuit Breaker",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_explain_decision",
			Description: "Explica en lenguaje natural (inspirado en IPv8 Natural) el por qué de una decisión de enrutamiento, ZTNA o PFO",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_did": map[string]interface{}{
						"type":        "string",
						"description": "DID soberano involucrado en la decisión",
					},
					"decision_type": map[string]interface{}{
						"type":        "string",
						"description": "Tipo de decisión: 'routing', 'ztna_firewall', 'pfo_audit' o 'nat_traversal'",
					},
				},
				"required": []string{"target_did"},
			},
		},
		{
			Name:        "ipvn7_uin_resolve_entity",
			Description: "Resuelve un root_id de 256 bits o DID soberano a su Pasaporte Digital UIN y claves subordinadas activas",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_memory_arbiter_status",
			Description: "Inspecciona el presupuesto de memoria RAM por clase (Replay 20%, QoS 30%, Trust 20%, Bindings 20%) y descartes por saturación OOM",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_validate_intent",
			Description: "Valida si una intención de transmisión declarada por un nodo o agente de IA está autorizada según su clase (0 a 4)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"did": map[string]interface{}{
						"type":        "string",
						"description": "DID soberano del nodo emisor o agente de IA",
					},
					"intent": map[string]interface{}{
						"type":        "string",
						"description": "Intención declarada: 'general', 'ai_agent', 'settlement', 'telemetry' o 'actuator'",
					},
				},
				"required": []string{"did", "intent"},
			},
		},
		{
			Name:        "ipvn7_pqc_status",
			Description: "Audita el estado de la suite criptográfica post-cuántica (ML-KEM-768, ML-DSA-65, 1-RTT handshake)",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_sphinx_status",
			Description: "Inspecciona la topología de circuitos cebolla Sphinx 3-hop, MTU 1280B y padding estocástico anti-DPI",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "ipvn7_memory_arbiter_status",
			Description: "Monitorea cuotas de RAM y métricas de resiliencia del Árbitro Global de Memoria contra ataques DoS",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
	}
	tools = append(tools, getSystemTools()...)
	tools = append(tools, getRobotVehicleTools()...)
	return append(tools, getGovernanceTools()...)
}

// HandleRequest procesa una llamada JSON-RPC del cliente MCP
