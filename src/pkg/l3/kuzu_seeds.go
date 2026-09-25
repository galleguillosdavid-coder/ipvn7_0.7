package l3

import "time"

// seedCanonicalAxioms siembra los principios axiomáticos inquebrantables de genesis.md
func (kg *KuzuGraphEngine) seedCanonicalAxioms() {
	kg.axioms["AXIOM_ZERO_PII"] = &AxiomPrinciple{
		PrincipleID: "AXIOM_ZERO_PII",
		Statement:   "Prohibición estricta de registrar o persistir IPs reales, geolocalizaciones o metadatos personales fuera de contenedores CBOR cifrados.",
		Immutable:   true,
	}
	kg.axioms["AXIOM_STRICT_CORE_FREEZE"] = &AxiomPrinciple{
		PrincipleID: "AXIOM_STRICT_CORE_FREEZE",
		Statement:   "Inmutabilidad del núcleo L0 (core/). Toda extensión debe implementarse en capas periféricas sin alterar el formato de cable canónico.",
		Immutable:   true,
	}
	kg.axioms["AXIOM_SUSTAINABLE_FLOW"] = &AxiomPrinciple{
		PrincipleID: "AXIOM_SUSTAINABLE_FLOW",
		Statement:   "Subordinación del emisor al cuello de botella de la ruta. Erradicación del bufferbloat mediante Packet Pacing determinista.",
		Immutable:   true,
	}
	kg.axioms["AXIOM_FAST_PATH_FIRST"] = &AxiomPrinciple{
		PrincipleID: "AXIOM_FAST_PATH_FIRST",
		Statement:   "Toda decisión de filtrado y reenvío debe ejecutarse en el camino crítico de kernel (eBPF/XDP) o en buffers lock-free (<28ns).",
		Immutable:   true,
	}
	kg.axioms["AXIOM_SMALL_WORLD_RINGS"] = &AxiomPrinciple{
		PrincipleID: "AXIOM_SMALL_WORLD_RINGS",
		Statement:   "Enrutamiento voraz en espacio métrico XOR de 256 bits estructurado en 12 anillos logarítmicos de Kleinberg.",
		Immutable:   true,
	}
}

// seedCanonicalFunctions siembra los módulos operativos y sus vínculos de derivación PFO
func (kg *KuzuGraphEngine) seedCanonicalFunctions() {
	funcs := []struct{ fID, path, pID, status string }{
		{"L0_SOVEREIGN_IDENTITY", "pkg/l0/identity.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L0_DETERMINISTIC_WIRE", "pkg/l0/wire.go", "AXIOM_STRICT_CORE_FREEZE", "VERIFIED"},
		{"L1_ZTNA_FIREWALL", "pkg/l1/firewall.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L1_PACKET_PACER", "pkg/l1/pacing.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L1_WDRR_SCHEDULER", "pkg/l1/wdrr_scheduler.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L1_KLEINBERG_ROUTER", "pkg/l1/routing.go", "AXIOM_SMALL_WORLD_RINGS", "VERIFIED"},
		{"L1_DARK_NODE_MACRO", "pkg/l1/dark_node_profile.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L1_USERSPACE_ADAPTER", "pkg/l1/tun_adapter.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L1_MULTIPATH", "pkg/l1/multipath.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L1_PQC_HYBRID", "pkg/l1/pqc_hybrid.go", "AXIOM_STRICT_CORE_FREEZE", "VERIFIED"},
		{"L1_SPHINX_ONION", "pkg/l1/sphinx_onion.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L1_SOCKS5_GATEWAY", "pkg/l1/socks5_gateway.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L1_NAT_STUN", "pkg/l1/nat_stun.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L1_SILENT_DISPATCHER", "pkg/l1/silent_call_response.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L1_UIN_IDENTITY", "pkg/l1/uin_identity.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L1_BUFFER_POOL", "pkg/l1/buffer_pool.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L1_CASCADE_MULTICAST", "pkg/l1/cascade_multicast.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L1_CORPORATE_VPN", "pkg/l1/corporate_vpn.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L1_BLACKOUT_RECOVERY", "pkg/l1/blackout_recovery.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L2_TELEMETRY_RING", "pkg/l2/telemetry.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L2_ANOMALY_FINGERPRINTER", "pkg/l2/anomaly_fingerprinter.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L2_RESILIENCE_GUARDIAN", "pkg/l2/resilience_guardian.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L2_SENTINEL_IMMUNOLOGY", "pkg/l2/sentinel_immunology.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L3_AI_COPILOT", "pkg/l3/ai_copilot.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L3_KUZU_TOPOLOGY", "pkg/l3/kuzu_graph.go", "AXIOM_SMALL_WORLD_RINGS", "VERIFIED"},
		{"L3_AGENT_SENATE", "pkg/l3/agent_senate.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L3_CONSTITUTIONAL_VERIFIER", "pkg/l3/constitutional_verifier.go", "AXIOM_STRICT_CORE_FREEZE", "VERIFIED"},
		{"L3_COMPUTATIONAL_LAW", "pkg/l3/computational_law.go", "AXIOM_ZERO_PII", "VERIFIED"},
		{"L4_WEB_DASHBOARD", "pkg/l4/web_server.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L4_REMOTE_DESKTOP", "pkg/l4/remote_desktop.go", "AXIOM_FAST_PATH_FIRST", "VERIFIED"},
		{"L4_LIVE_STREAM", "pkg/l4/live_stream.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
		{"L4_AUDIO_RADIO", "pkg/l4/audio_radio.go", "AXIOM_SUSTAINABLE_FLOW", "VERIFIED"},
	}

	for _, fn := range funcs {
		kg.functions[fn.fID] = &OperationalFunction{
			FunctionID: fn.fID,
			ModulePath: fn.path,
			Status:     fn.status,
		}
		kg.derivesFrom[fn.fID] = fn.pID
	}
}

// seedCanonicalDecisions siembra las decisiones arquitectónicas registradas en docs/ADR.md
func (kg *KuzuGraphEngine) seedCanonicalDecisions() {
	now := time.Now().UTC()
	decs := []struct{ id, rat string }{
		{"DEC-001", "Erradicación de Mocks; uso de UserspaceVirtualAdapter y AdapterMode."},
		{"DEC-002", "Medición Empírica en Vivo en lugar de Constantes Sintéticas (EWMA nanosegundo real)."},
		{"DEC-003", "Erradicación Total de PII y Hardcoding (RFC 5737 TEST-NET y variables de entorno)."},
		{"DEC-004", "Higiene de Concurrencia con context.Context y timeouts de red."},
		{"DEC-005", "Regla Estricta de Modularidad Atómica (<= 400 líneas por archivo)."},
		{"DEC-006", "Kùzu Graph como Navegador Lógico-Visual para Auditorías y SSOT."},
		{"DEC-007", "Interoperabilidad Multiplataforma Windows, WSL2 y Linux."},
		{"DEC-008", "Erradicación Definitiva de Residuos Sintéticos y Certificación de Código 100% Real."},
		{"DEC-009", "Auditoría Integral y Sustitución Definitiva de Mocks por Código Real."},
		{"DEC-010", "Entorno de Lanzamiento Canónico Ubuntu WSL2 y Memoria Kùzu DB para la IA."},
	}
	for _, d := range decs {
		kg.decisions[d.id] = &EngineeringDecision{DecisionID: d.id, Rationale: d.rat, Timestamp: now}
	}
}
