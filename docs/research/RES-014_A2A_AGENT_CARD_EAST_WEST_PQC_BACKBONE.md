# RES-014: PROTOCOLO AGENT2AGENT (A2A), MANIFIESTOS AGENT CARD Y TRANSPORTE SOBERANO ESTE-OESTE EN IPVN7

## 1. Contexto y Pregunta de Investigación
Tras la consolidación del stack agéntico por la **Agentic AI Foundation (AAIF)** (Linux Foundation, 2026), la industria ha dividido formalmente la interoperabilidad de IA en dos planos:
* **Conectividad Norte-Sur (Herramientas y Datos):** Estandarizada mediante **MCP (Model Context Protocol)**, permitiendo que los agentes invoquen APIs y recursos locales (implementado nativamente en IPvN7 `/mcp`).
* **Conectividad Este-Oeste (Colaboración Inter-Agentes):** Estandarizada mediante el protocolo **A2A (Agent2Agent Protocol v1.0.0)**, permitiendo que agentes autónomos de diferentes proveedores (Google, Anthropic, OpenAI, agentes locales) se descubran, negocien tareas y deleguen subtareas horizontalmente.

Sin embargo, la especificación A2A v1.0.0 deja sin resolver la capa de transporte perimetral y la autorización criptográfica, alertando sobre riesgos críticos de **suplantación de agentes (agent impersonation)**, **envenenamiento de contexto (context poisoning)** y dependencia de proxies de nube centralizados. ¿Cómo puede IPvN7 capitalizar A2A para convertirse en el canal de transporte Este-Oeste seguro, post-cuántico y P2P definitivo?

## 2. Hallazgos en Estándares Mundiales de Frontera (Agosto–Septiembre 2026)
* **Arquitectura de Descubrimiento A2A (`Agent Card`):** Los agentes anuncian sus habilidades mediante un manifiesto JSON canónico ubicado en `/.well-known/agent-card.json`. Este manifiesto declara la identidad del agente, descripción, habilidades soportadas, endpoints de servicio y esquemas de autenticación.
* **Modelo de Comunicación A2A:** Emplea JSON-RPC 2.0 sobre HTTP(S) con soporte para peticiones síncronas, streaming vía Server-Sent Events (SSE) y callbacks asíncronos para tareas de larga duración.
* **La Vulnerabilidad Central de A2A:** Al ser un protocolo de nivel de aplicación desacoplado de la red, los agentes dependen de certificados TLS convencionales o tokens de portador (bearer tokens) fácilmente interceptables o falsificables, exponiéndose a ataques Man-in-the-Middle y spoofing si el transporte no está autenticado punto a punto.

## 3. Oportunidades y Beneficios Estratégicos para IPvN7
1. **Blindaje Criptográfico contra Suplantación (DID-Bound A2A):**
   - En IPvN7, cada nodo y agente cuenta con una identidad soberana `did:ipvn7:<pubkey_hash>`. Las peticiones A2A entre agentes se autentican mutuamente mediante firmas Ed25519 y llaves de sesión post-cuánticas (ML-KEM-768), erradicando de raíz la suplantación de agentes sin requerir CAs (Certificate Authorities) centralizadas.
2. **Descubrimiento Cero-Fricción vía `/.well-known/agent-card.json` Nativo:**
   - Exponer un manifiesto estático ultra-compacto (<20 líneas de JSON en `pkg/core/web_ui.go`) que declare a IPvN7 como un nodo de infraestructura soberana con capacidades de red: enrutamiento P2P, perforación NAT, pasarela IoT ZTNA y telemetría wire-speed.
   - Cualquier agente A2A que explore la red local o la malla reconocerá a IPvN7 automáticamente sin configuración manual.
3. **Transporte Este-Oeste P2P con Latencia de Microsegundos:**
   - Mientras las implementaciones comerciales de A2A envían el tráfico inter-agentes a través de nubes corporativas centralizadas (agregando 100-300 ms de latencia), IPvN7 enruta los mensajes A2A directamente de socket a socket sobre UDP/PQC con perforación NAT STUN (RFC 5389), alcanzando latencias de microsegundos en LAN y mínimas de tránsito en WAN.
4. **Protección ZTNA Default-Deny contra Envenenamiento de Contexto:**
   - La tabla de rutas de Kleinberg en L1 filtra qué nodos pueden enviar tráfico de tareas a qué agentes locales. Un agente malicioso externo no puede enviar cargas maliciosas si su DID no está formalmente emparejado.

## 4. Filtro Antihumo Estricto y Presupuesto de Código
* **Cero Complejidad Accesoria:** Para habilitar compatibilidad con A2A no se requiere crear un motor complejo de orquestación ni parser de tareas. Solo se necesita exponer el endpoint de descubrimiento canónico `/.well-known/agent-card.json` y permitir que las peticiones JSON-RPC reutilicen el multiplexor HTTP existente.
* **Presupuesto $\le 400$ Líneas:** Se mantendrá estrictamente el límite de líneas en el núcleo HTTP sin agregar dependencias dinámicas.

## 5. Decisión Técnica Propuesta
* Registrar **DEC-123**: Adopción del estándar Agent2Agent (A2A) v1.0.0 para conectividad Este-Oeste y exposición del manifiesto canónico `/.well-known/agent-card.json` en IPvN7.

## 6. Fuentes Primarias
* Agentic AI Foundation (AAIF) & Linux Foundation: Agent2Agent (A2A) Protocol Specification v1.0.0 (2026).
* W3C Decentralized Identifiers (DIDs) v1.1 Standard.
* NIST FIPS 203 (ML-KEM) & IETF RFC 8446 (TLS 1.3).
