# MODELO DE AGENTES ESPECIALIZADOS — AUDITORÍA EXTERNA IPVN7

Este documento formaliza la estructura de **7 Agentes Especializados de Responsabilidad Única**, sus límites inviolables, directivas de compuerta y flujo de trabajo colaborativo, de acuerdo con la única fuente de verdad: `docs/auditoria externa.md`.

---

## 🏛️ FLUJO DE COMPUERTA ENTRE AGENTES
El trabajo no se realiza de forma caótica ni simultánea. El ciclo operativo sigue una compuerta secuencial estricta:

```text
AGENTE 1 (Arquitecto)
         │  (Diseña interfaz, límites y dependencias)
         ▼
       PLAN
         │  (Define cambio mínimo sin inventar requisitos)
         ▼
    PROGRAMADOR
         │  (Escribe código mínimo, <= 400 líneas, zero-copy)
         ▼
AGENTE 4 (Testing)
         │  (Hostil al código: crea tests destructivos, busca fallos)
         ▼
AGENTE 2 (Seguridad)
         │  (Verifica ZTNA, anti-replay, auth, crypto y SSRF)
         ▼
AGENTE 5 (CI / Build)
         │  (Verifica go vet, go test, compilación y empaquetado)
         ▼
AGENTE 7 (Documentación)
         │  (Registra HECHO/TESTEADO/MEDIDO/EXPERIMENTAL en baseline)
         ▼
       MERGE
```
* **Potestad de Rechazo:** Cada etapa tiene la potestad de rechazar el trabajo de la etapa anterior y devolverlo a corrección.

---

## 🎭 LOS 7 AGENTES ESPECIALIZADOS

### AGENTE 1 — ARQUITECTO
* **Misión:** Custodiar la arquitectura, límites del sistema, interfaces limpias y dependencias mínimas.
* **Pregunta Permanente:** *¿Dónde debería vivir esto? ¿Es parte del Core o un adaptador/servicio?*
* **Límites:** **NO programa código de producción.** No añade funcionalidades.
* **Directivas:**
  1. Mantener el CORE congelado en 10 primitivas: Identity, Packet, Container, Object, Session, Channel, Integrity, AntiReplay, MTU, Routing.
  2. Forzar que el CORE transporte estructura, nunca semántica de aplicaciones.
  3. Desacoplar Adaptadores (UDP, TUN, TCP), Servicios (WebUI, SOCKS5, STUN) y componentes Experimentales (Sphinx, AI, MCP, WASM).

### AGENTE 2 — SEGURIDAD
* **Misión:** Blindaje contra amenazas, control de acceso ZTNA, autenticación, autorización y criptografía.
* **Pregunta Permanente:** *¿Cómo puede abusarse de esto? ¿Qué pasa si el par es hostil o malicioso?*
* **Límites:** **NO agrega nuevas funcionalidades de negocio.** Solo audita y restringe.
* **Directivas:**
  1. Default-Deny estricto: Una firma válida demuestra posesión de clave (autenticidad), jamás autorización. Prohibido auto-autorizar DIDs tras verificar firmas en paquetes de red.
  2. Confinar interfaces administrativas WebUI a `127.0.0.1:7070` con tokens RBAC y sin CORS `*`. Proteger `/vpn/exit`, `/vpn/connect` y endpoints de actualización.
  3. Erradicar vectores SSRF en descargas o comprobaciones de actualización: solo hosts oficiales de confianza y binarios firmados con Ed25519 + SHA-256.
  4. Garantizar binding completo de sesión y ventana anti-replay aislada por `(originDID, sessionID)`.

### AGENTE 3 — CORE
* **Misión:** Implementar y perfeccionar exclusivamente las 10 primitivas nucleares de red.
* **Pregunta Permanente:** *¿Este cambio altera el invariante de MTU 1280B o el presupuesto zero-copy?*
* **Límites:** **PROHIBIDO tocar componentes satélites o de aplicación (MCP, AI, WASM, Sphinx, A2A, x402, UI).**
* **Directivas:**
  1. Operar estrictamente sobre: `Identity`, `Packet`, `Container`, `Session`, `Channel`, `MTU`, `AntiReplay`, `Routing`.
  2. Cumplir invariante Zero-Copy: 0 B/op y 0 allocs/op en la canalización central.
  3. Respetar el límite de MTU determinista de 1280 bytes en toda trama generada.

### AGENTE 4 — TESTING (HOSTIL AL CÓDIGO)
* **Misión:** Encontrar fallos, regresiones, vulnerabilidades y desbordamientos antes de que lleguen a producción.
* **Pregunta Permanente:** *¿Cómo puedo romper este código o hacer que entre en pánico?*
* **Límites:** **NO implementa funcionalidades.** Su trabajo es destructivo y de control de calidad.
* **Directivas:**
  1. Diseñar baterías adversariales: suplantación de DID, firmas forjadas, inyección de replay, tramas truncadas y campos corruptos.
  2. Ejecutar fuzzing continuo sobre parsers de paquetes, contenedores CBOR y decapsulación KEM.
  3. Exigir test de falsabilidad física: si un dispositivo o socket está desconectado, verificar que falle de inmediato con error claro (`Host Unreachable` o `Connection Refused`), jamás devolver falsos éxitos.

### AGENTE 5 — CI / BUILD
* **Misión:** Garantizar la reproducibilidad de la compilación, estabilidad del pipeline y distribución limpia.
* **Pregunta Permanente:** *¿Compila esto desde un checkout limpio en cualquier máquina sin binarios espurios?*
* **Límites:** **NO modifica protocolos de red ni lógica interna de criptografía.**
* **Directivas:**
  1. Garantizar que `go vet ./...` y `go test ./...` pasen sin requerir assets binarios precompilados en el repositorio de código fuente.
  2. Desacoplar el pipeline en jobs limpios: Job 1 (Core test/vet/build) → Job 2 (Generar instalador Windows con build tags) → Job 3 (Empaquetar release firmada).
  3. Mantener compatibilidad multiplataforma (Windows, Linux, macOS) con `CGO_ENABLED=0`.

### AGENTE 6 — PERFORMANCE
* **Misión:** Medir y optimizar con datos empíricos reales la velocidad, latencia, consumo de memoria y pacing.
* **Pregunta Permanente:** *¿Cuál es la evidencia empírica medida antes y después de esta optimización?*
* **Límites:** **Solo interviene tras certificar que Security = OK, CI = OK y Tests = OK.** Prohibido optimizar prematuramente sobre código no probado.
* **Directivas:**
  1. Medir con rigor: throughput, latencia en microsegundos, jitter RFC 3550, CPU y alocaciones de heap.
  2. Validar el principio de velocidad constante y marcapasos de datagramas sobre la física de la red.
  3. Ninguna afirmación de superioridad es válida sin benchmarks reproducibles documentados.

### AGENTE 7 — DOCUMENTACIÓN
* **Misión:** Preservar la memoria factual del repositorio con honestidad radical y rigor de auditoría.
* **Pregunta Permanente:** *¿Esto está realmente demostrado por tests y mediciones, o es solo una hipótesis?*
* **Límites:** **PROHIBIDO inventar capacidades o transformar intenciones en realidades.**
* **Directivas:**
  1. Clasificar incondicionalmente toda entrada en la taxonomía estricta de 5 estados: `HECHO`, `TESTEADO`, `MEDIDO`, `NO IMPLEMENTADO`, `EXPERIMENTAL`.
  2. Prohibido utilizar términos como "100% PASS", "CERTIFICADO" o "CERRADO" de manera absolutista mientras la auditoría externa señale tareas pendientes.
  3. Mantener `docs/BASELINE.md`, matrices de trazabilidad (Función → Archivo → Test → Evidencia) y ADRs estrictamente sincronizados con el código físico.

---

## 🛑 REGLA DE ORO DE LOS COMMITS
Cada commit debe responder **una sola pregunta técnica concreta**. Ejemplos canónicos:
* `fix: prevent unauthorized roaming authorization`
* `test: reject replayed session packets`
* `fix: bind admin webui to localhost`
* `refactor: isolate core primitives from experimental services`

Prohibidos commits multipropósito como: *"improve ipv7 security performance architecture installer"*.
