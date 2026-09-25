# ipvn7 — Gobernanza Digital, Senado de Agentes y Derecho Computable

**Versión:** 0.7.0  
**Subsistemas:** [`pkg/l3/agent_senate.go`](../pkg/l3/agent_senate.go), [`pkg/l3/computational_law.go`](../pkg/l3/computational_law.go), [`pkg/l2/sentinel_immunology.go`](../pkg/l2/sentinel_immunology.go)

---

## 1. Carta Magna y Axiomas Universales (Artículo I)
Toda regla de red debe subordinarse estrictamente a 4 postulados blindados criptográficamente:
1. **Soberanía Criptográfica:** Dominio indelegable del par de claves DID (`pkg/l0`). Cero intermediarios.
2. **No Agresión Estructural:** Prohibición absoluta de DoS, inyección maliciosa o sabotaje de paquetes.
3. **Neutralidad Semántica:** Enrutamiento libre guiado por condiciones de red, sin censura política ni comercial.
4. **Inmunidad Ética Radical:** Rechazo ciego de malware y ataques mediante *Private Set Intersection* (PSI) sin romper el cifrado de extremo a extremo.

> [!IMPORTANT]
> **Invariante Core Freeze L0:** Ninguna mayoría del Senado ni algoritmo de IA tiene potestad para mutar las primitivas criptográficas de `pkg/l0`.

---

## 2. Senado de Agentes y Democracia Líquida
El gobierno ordinario de la red se debate entre agentes de software vinculados a identidades soberanas:
* **Proof-of-Contribution (PoC):** El peso de votación no depende de dinero ni de IPs (anti-Sybil, anti-plutocracia):
  $$\text{Weight} = \text{Base} + (\text{WoT\_Reputation} \times 0.4) + \min(30, \text{Transit\_MB})$$
* **Informe Matutino:** El agente sintetiza las decisiones de cada ciclo para el usuario humano.
* **Veto Soberano Humano (1-Click Veto):** El usuario retiene la autoridad suprema inalienable para anular cualquier voto de su agente en 1 clic o comando CLI.

---

## 3. Gramática Institucional ADICO y Derecho Computable
Las políticas de subred se estructuran mediante 5 componentes deónticos formales:
$$\mathcal{N}_{\text{local}} \sqsubseteq \mathcal{N}_{\text{core}} \quad [\text{A}][\text{D}][\text{I}][\text{C}][\text{O}]$$
* **[A] Atributo:** Sujetos normados.
* **[D] Deóntico:** $\mathcal{O}$ (Obligación), $\mathcal{P}$ (Permisión), $\mathcal{F}$ (Prohibición).
* **[I] Objetivo:** Acción regulada.
* **[C] Condiciones:** Límites de latencia, reputación o topología.
* **[O] Consecuencia (*Or Else*):** Sanción, degradación de ancho de banda o fianza.

### Reglas de Prelación Jurisdiccional
1. **Lex Superior:** La Carta Magna universal prevalece siempre sobre cualquier política local ($\bot$).
2. **Lex Specialis:** La norma contextual calificada prevalece sobre la norma genérica.
3. **Lex Posterior:** La directiva ratificada más recientemente deroga a la anterior en caso de paridad.

---

## 4. Inmunología Celular de Centinelas
Auditoría cruzada descentralizada entre pares:
1. **Desafíos Criptográficos:** Los nodos emiten periódicamente `AuditChallenge` para verificar ejecución idéntica en sandbox.
2. **Quórum Slashing:** Con $N \ge 2$ firmas de centinelas independientes ante anomalías dolosas:
   * Revocación inmediata en cortafuegos ZTNA (*Drop Default-Deny* perpetuo).
   * Estrangulamiento económico en contabilidad Tit-for-Tat (*THROTTLED*).

---

## 5. Matriz Operativa CLI y API

| Subsistema | CLI (`ipvn7-cli`) | Gateway REST API | MCP (JSON-RPC) |
|---|---|---|---|
| **Constitución** | `constitution view`, `verify <file>` | `/api/constitution/text`, `/verify` | `ipvn7_constitution_verify` |
| **Senado** | `senate list`, `propose`, `vote`, `veto` | `/api/senate/proposals`, `/vote`, `/veto` | `ipvn7_senate_propose`, `_vote` |
| **Centinelas** | `sentinel status`, `audit <did>`, `alerts` | `/api/sentinel/status`, `/audit`, `/alerts` | `ipvn7_sentinel_audit` |
| **Derecho Computable** | `law validate <norm>`, `antinomies` | `/api/law/norms`, `/api/law/evaluate` | `ipvn7_law_validate` |
