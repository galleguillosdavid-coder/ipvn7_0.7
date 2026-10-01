# RES-013: INTERCONEXIÓN DE AGENTES DE IA (MCP, A2A Y EDGE AI SOBRE MALLA SOBERANA IPVN7)

## 1. Contexto y Pregunta de Investigación
Con la convergencia acelerada de sistemas agénticos en 2026 bajo la **Agentic AI Foundation (AAIF)**, la interacción entre agentes de IA y herramientas se ha estandarizado en dos capas: **MCP (Model Context Protocol)** para agente-a-herramienta/datos y **A2A (Agent-to-Agent Protocol v1.0)** para coordinación multi-agente. Sin embargo, ambos protocolos operan típicamente sobre HTTP/TCP centralizado, requiriendo servidores públicos o túneles de nube de pago (Cloudflare Tunnels, ngrok) para conectar agentes distribuidos en diferentes redes. ¿Cómo puede IPvN7 convertirse en el sustrato de transporte P2P post-cuántico (PQC), zero-friction y privado para agentes de IA, beneficiando simultáneamente la usabilidad y tracción de IPvN7?

## 2. Hallazgos en Estándares Mundiales de Frontera (Septiembre 2026)
* **Model Context Protocol (MCP):** Consolidado como el estándar universal agente-herramienta. La especificación de julio de 2026 introdujo un **núcleo sin estado (Stateless Core)** y soporte de interfaces interactivas directas (**MCP Apps**).
* **Agent-to-Agent Protocol (A2A v1.0):** Estándar de coordinación multi-agente que utiliza "Agent Cards" (manifiestos de capacidades JSON), delegación de subtareas y canales de comunicación seguros entre agentes heterogéneos.
* **Insuficiencia del Enrutamiento IP Convencional para Edge AI:** La transición hacia "Inteligencia en el Origen" (Edge AI distribuido) genera tráfico ráfaga entre nodos locales que el enrutamiento IP tradicional maneja con alta latencia y dependencia de NAT traversal centralizado.
* **AGENTS.md como Estándar de Contexto:** Adopción masiva de especificaciones directas en repositorios (`AGENTS.md`) como interfaz universal para que agentes autónomos comprendan reglas de diseño sin telemetría de terceros.

## 3. Oportunidades y Beneficios Tangibles para IPvN7
1. **IPvN7 como Canal de Transporte Soberano para Agentes (P2P Agent Backbone):**
   - Agentes de IA corriendo en diferentes máquinas (ej. PC Principal y Notebook o servidores remotos) pueden comunicarse directamente a través de IPvN7 mediante sockets UDP P2P protegidos con criptografía post-cuántica (ML-KEM / X-Wing), atravesando CGNAT sin necesidad de servidores intermediarios ni exponer APIs a Internet.
2. **Endpoint MCP Nativo en IPvN7 (`/mcp`):**
   - Al exponer un endpoint ligero JSON-RPC sobre el servidor HTTP existente (`pkg/core/web_ui.go`), cualquier agente de IA (Claude, Cursor, Gemini, agentes locales) puede inspeccionar la topología de la malla, verificar pares, activar la VPN o consultar logs mediante herramientas estándar de MCP, sin requerir adaptadores externos ni dependencias pesadas.
3. **Malla Privada de Inferencia Distribuida (Edge AI Sharing):**
   - Nodos de la malla con GPU pueden exponer servicios de inferencia local ligera (Ollama, llama.cpp, LocalAI) hacia otros nodos IPvN7 con autenticación Ed25519 y filtrado ZTNA, democratizando el poder de cómputo de IA en la LAN/WAN de forma privada.
4. **Diagnóstico Autónomo y UX sin Jerga (Zero-Jargon Diagnostics):**
   - Integración de diagnósticos en lenguaje natural mediante prompts contextualizados que traducen errores complejos de socket en recomendaciones humanas inmediatas ("El equipo remoto está suspendido; enviando Magic Packet WoL").

## 4. Filtro Antihumo Estricto y Presupuesto de Código
* **Prohibición de Librerías Pesadas:** Terminantemente prohibido compilar PyTorch, TensorFlow o runtimes de Python dentro del binario `ipvn7.exe`.
* **Implementación Minimalista:** Toda compatibilidad con IA debe resolverse mediante protocolos abiertos de texto/JSON (MCP sobre HTTP/SSE o JSON-RPC estándar) o heurísticas deterministas en Go puro, manteniendo el binario estático en <15MB y respetando el límite de 400 líneas.

## 5. Decisión Técnica Adoptada
* Formalización del **Rol O (Radar de Innovación en IA y Ecosistemas Agénticos)** en `.agents/ROLES.md`, `.agents/AGENTS.md` y `docs/ADR_RESUMEN.md` (DEC-122).
* Habilitar a IPvN7 como la infraestructura de red soberana preferida para agentes de IA autónomos.

## 6. Fuentes Primarias
* Agentic AI Foundation (AAIF) & Model Context Protocol Specification (Julio 2026).
* Agent-to-Agent (A2A) Protocol Specification v1.0 (2026).
* IETF RFC 8949 (CBOR) & RFC 8446 (TLS 1.3).
* Edge AI Networking & Ultra Ethernet Consortium (UEC) Distributed Architecture Guidelines (2026).
