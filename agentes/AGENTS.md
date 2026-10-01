# DIRECTIVA SUPREMA — FUENTE DE VERDAD Y MODO DE TRABAJO IPVN7

## LA ÚNICA FUENTE DE VERDAD DEL PROYECTO
> **"La única fuente de verdad absoluta del proyecto es `docs/auditoria externa.md`. Cualquier especificación, decisión arquitectónica (ADR), regla previa, comentario o hipótesis en el código queda estrictamente supeditada y corregida por las directivas, hallazgos y metodología de la Auditoría Externa."**

---

## 0. REGLA DE ORO DE LOS AGENTES: NO INVENTAR
Si algo no está demostrado por código, test, benchmark o documentación verificable, debe declararse como hipótesis o pendiente. Queda terminantemente prohibido convertir:
* **HIPÓTESIS → HECHO**
* **INTENCIÓN → IMPLEMENTACIÓN**
* **TEST UNITARIO → SEGURIDAD COMPLETA**
* **BUILD LOCAL → CI REPRODUCIBLE**
* **HASH → AUTENTICIDAD**
* **FIRMA VÁLIDA → AUTORIZACIÓN**

---

## 1. ESTADO REAL DE LA AUDITORÍA Y TAXONOMÍA OBLIGATORIA
* **Estado Real:** Al 1 de octubre de 2026, la auditoría **NO está cerrada**. Prohibido auto-declarar el sistema como "100% CERRADO", "CERTIFICADO" o "PRODUCTION-READY".
* **Taxonomía Documental de 5 Estados:** Toda función o componente debe clasificarse honestamente en:
  1. `HECHO`: Implementado físicamente en código fuente.
  2. `TESTEADO`: Verificado con tests unitarios, de integración o adversariales passing.
  3. `MEDIDO`: Validado con benchmarks empíricos y métricas reales (allocs, ns/op, RTT).
  4. `NO IMPLEMENTADO`: Planificado o en diseño, sin código funcional.
  5. `EXPERIMENTAL`: Prototipos o adaptadores preliminares que requieren calibración en campo.
* **Taxonomía de Evidencia Externa:** Solo se puede reportar: `AUDITADO INTERNAMENTE`, `EVIDENCIA LOCAL` o `DEMOSTRADO FÍSICAMENTE`.

---

## 2. EL ALGORITMO DE 5 PASOS
1. **Cuestionar los requisitos:** Toda línea debe responder a una necesidad justificada.
2. **Eliminar partes o procesos:** Podar el 20-30% de código accesorio antes de modularizar.
3. **Simplificar y optimizar:** Reducir al mínimo indispensable lo que sobrevive.
4. **Acelerar el ciclo:** Compilación instantánea y verificación física inmediata.
5. **Automatizar:** Solo automatizar al final sobre el núcleo simplificado.

---

## 3. SEPARACIÓN ESTRUCTURAL DEL SISTEMA (FASE 6)
El Núcleo I7 transporta estructura, no semántica de aplicaciones:
* **I7 CORE (10 Primitivas Inmutables):** `Identity`, `Packet`, `Container`, `Object`, `Session`, `Channel`, `Integrity`, `AntiReplay`, `MTU`, `Routing`. Nada más.
* **ADAPTERS:** `UDP`, `TUN`, `TCP`, `QUIC`, `WireGuard`.
* **SERVICES:** `Discovery`, `STUN`, `NAT Traversal`, `SOCKS5`, `WebUI`.
* **EXPERIMENTAL (Fuera del Camino Crítico):** `Sphinx`, `Planetary`, `WASM`, `AI`, `MCP`, `A2A`, `AP2`, `x402`, `Egress`.

---

## 4. BARRERAS DE SEGURIDAD CRÍTICA (P0 / P1)
1. **ZTNA Estricto (Default-Deny):** La posesión de una clave (firma Ed25519) demuestra identidad, jamás autorización. NUNCA `firma válida → AuthorizeDID()`. AuthorizeDID() es una operación administrativa/política previa.
2. **WebUI Confinada:** Bind exclusivo a `127.0.0.1:7070` por defecto. Prohibido `0.0.0.0` sin autenticación. Prohibido CORS `*` en endpoints administrativos. Operaciones críticas (`/vpn/connect`, `/vpn/disconnect`, `/vpn/exit`, `/vpn/cycle`, `/update/*`) requieren token de autorización local.
3. **Actualización Segura sin SSRF:** Prohibidas URLs arbitrarias en `/api/v1/update/check`. Solo hosts de confianza configurados. Manifiestos firmados con Ed25519 del desarrollador + SHA-256 del binario.
4. **Honestidad Criptográfica:** Prohibido llamar "ML-DSA" a derivaciones HMAC/SHA256. La firma de producción descansa en Ed25519 real; cualquier vector reticular en ajuste debe rotularse honestamente como `ExperimentalPQCIdentity`.
5. **Invariante MTU 1280:** El protocolo I7 limita su datagrama lógico a 1280 bytes para evitar depender de fragmentación IP (RFC 8200). Handshakes KEM y tramas completas deben caber en <= 1280B.
6. **Anti-Replay por Sesión:** Ventana deslizante de 1024 bits indexada por `(originDID, sessionID, sequence)`.

---

## 5. MODELO DE 7 AGENTES ESPECIALIZADOS
En lugar de acumular roles genéricos o dispersos, el trabajo de desarrollo y mantenimiento se gobierna por **7 Agentes Especializados de Responsabilidad Única**:

1. **AGENTE 1 — Arquitecto:** Responsabilidad exclusiva: arquitectura, interfaces, dependencias, límites. No programa. Pregunta permanente: *¿Dónde debería vivir esto?*
2. **AGENTE 2 — Seguridad:** Responsabilidad exclusiva: ZTNA, auth, autorización, crypto, replay, SSRF, WebUI, updates. No agrega funcionalidades. Pregunta permanente: *¿Cómo puede abusarse de esto?*
3. **AGENTE 3 — CORE:** Responsabilidad exclusiva: las 10 primitivas (Identity, Packet, Container, Session, Channel, MTU, AntiReplay, Routing). Prohibido tocar MCP, AI, WASM, Sphinx, A2A, x402. *El Core transporta estructura, no semántica.*
4. **AGENTE 4 — Testing:** Hostil al código. No desarrolla funcionalidades. Busca fallos, crea tests de frontera, fuzzing, regresiones y escenarios adversariales. Intenta romper lo construido.
5. **AGENTE 5 — CI/Build:** Responsabilidad exclusiva: GitHub Actions, toolchains de Go, builds Windows/Linux/macOS, instalador desacoplado, empaquetado y artefactos. No modifica protocolos.
6. **AGENTE 6 — Performance:** Solo interviene cuando Security=OK, CI=OK y Tests=OK. Trabaja exclusivamente con mediciones empíricas reales: allocations, ns/op, throughput, latencia y zero-copy.
7. **AGENTE 7 — Documentación:** No inventa capacidades. Solo puede registrar hechos demostrados con la taxonomía: HECHO, TESTEADO, MEDIDO, NO IMPLEMENTADO, EXPERIMENTAL.

---

## 6. REGLA DE TRABAJO Y COMPUERTA ENTRE AGENTES
Flujo obligatorio y secuencial:
```text
ARQUITECTO ──► PLAN ──► PROGRAMADOR ──► TESTER ──► SEGURIDAD ──► CI ──► MERGE
```
* Cada etapa tiene potestad estricta de **rechazar** el trabajo anterior y exigir corrección.
* **Regla de los Commits:** Cada commit debe responder a una sola pregunta técnica (ej. `fix: prevent unauthorized roaming authorization`, `test: reject replayed session packets`).

---

## 7. PROCESO DE 10 PASOS ANTES DE PROGRAMAR
1. Leer arquitectura existente y `docs/auditoria externa.md`.
2. Localizar exactamente el código y componentes afectados.
3. Identificar invariantes matemáticos y de seguridad involucrados.
4. Identificar tests existentes correspondientes.
5. Explicar el cambio mínimo necesario (si se resuelve con 10 líneas, no escribir 100).
6. Programar únicamente el cambio diseñado.
7. Probar localmente (`go test`).
8. Revisar vectores de ataque y seguridad (ZTNA, DoS, Replay).
9. Ejecutar compuerta universal de paso (`scripts/verify_ipvn7_standard.ps1`).
10. Documentar evidencia empírica en el baseline/ADR con taxonomía estricta.

---

## 8. PLAN DE TRABAJO CANÓNICO EN 13 FASES (FASE 0 A 12)
* **FASE 0 — Baseline:** Medir, congelar, crear `docs/BASELINE.md`, eliminar claims no demostrados.
* **FASE 1 — Seguridad Crítica (ZTNA):** Desacoplar firma de autorización, Default-Deny estricto.
* **FASE 2 — Cerrar WebUI:** Confinar a 127.0.0.1, token RBAC, eliminar CORS *.
* **FASE 3 — Eliminar SSRF en Updates:** Hosts de confianza, firma Ed25519 + SHA256.
* **FASE 4 — Arreglar CI:** Desacoplar instalador de assets precompilados, multi-job limpio.
* **FASE 5 — Limpiar Criptografía:** Nomenclatura honesta, binding completo de handshake.
* **FASE 6 — Congelar el CORE:** 10 primitivas aisladas de Adapters, Services y Experimental.
* **FASE 7 — Contratos del CORE:** Documentar especificaciones `docs/core/*.md`.
* **FASE 8 — Contratos Matemáticos:** Invariantes de MTU (<=1280B), SessionID y Anti-Replay.
* **FASE 9 — Pruebas Destructivas:** Adversarial suites y fuzzing de parsers.
* **FASE 10 — Rendimiento Empírico:** Benchmarks reales antes de optimizar.
* **FASE 11 — Interoperabilidad:** Nodos físicos A -> B, A -> C, A -> B -> C sobre múltiples transportes.
* **FASE 12 — Servicios:** Chat, Files, VPN, etc., como consumidores puros del CORE.

---

## 9. INVARIANTES OPERATIVOS PERMANENTES
1. **Límite Estricto de 400 Líneas (Axioma III):** Ningún archivo puede superar las 400 líneas de código. Alarma preventiva a las 320 líneas.
2. **Zero-Copy en Pipeline Interno:** 0 B/op y 0 allocs/op en `BenchmarkLinearPipeline_Execute`.
3. **Cero Simulación Criptográfica:** Uso de bibliotecas estándar reales (`crypto/mlkem`, `crypto/ed25519`).
4. **Verificación Física en Laboratorio 2 Nodos:**
   - **Nodo A (PC Principal):** `192.168.1.198`
   - **Nodo B (Notebook Dvd):** `192.168.1.106`
   - Prohibido auto-emparejamiento y nodos fantasma.
5. **Compuerta de Paso Universal:** Ninguna modificación es válida sin ejecución limpia de `scripts/verify_ipvn7_standard.ps1` (0 violaciones de 400L, `go vet` limpio, `go test ./pkg/...` 100% PASS).
6. **Cerrojo Atómico (agentes/task.lock):** Exclusión mutua obligatoria contra colisiones.
