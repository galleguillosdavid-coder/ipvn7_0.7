# FUENTE DE VERDAD CANÓNICA — IPVN7 v0.7.0 (UNIVERSAL SOVEREIGN CORE)

> **Estado:** VIGENTE / NORMA TÉCNICA SUPREMA  
> **Versión Normativa:** 1.0.0 (Consolidación Post-Auditoría v0.7.0)  
> **Fecha de Emisión:** Octubre 2026  
> **Autoridad:** Sistema de Autogobernanza IPVN7 / Agente Rector [FrondaBrick_01](../agentes/Frondabrick01/AGENTE.md)  
> **Antecedente Histórico:** [`docs/audit/AUDITORIA_EXTERNA_HISTORICA_ee56f89.md`](audit/AUDITORIA_EXTERNA_HISTORICA_ee56f89.md)  

---

## 🏛️ CAPÍTULO 1: LA REGLA SUPREMA Y PRECEDENCIA CONSTITUCIONAL

### 1.1 Declaración de Fuente Suprema de Verdad
Este documento constituye la **única fuente de verdad absoluta y normativa del proyecto IPVN7**.
Cualquier diseño arquitectónico anterior (ADR), documentación de versiones preliminares, issue, comentario en el código o hipótesis técnica queda **estrictamente supeditada, corregida y gobernada** por las directivas, invariantes y taxonomía de esta Fuente de Verdad.

### 1.2 El Axioma de No Invención (Regla Cero)
Queda terminantemente prohibido a cualquier desarrollador, operador o agente autónomo de inteligencia artificial asumir capacidades inexistentes o convertir expectativas en realidades sin demostración física comprobable:
* ❌ **HIPÓTESIS $\to$ HECHO** (Toda hipótesis requiere un experimento diseñado y medido).
* ❌ **INTENCIÓN $\to$ IMPLEMENTACIÓN** (Una orden en chat no es código hasta superar tests y compuertas).
* ❌ **TEST UNITARIO $\to$ SEGURIDAD COMPLETA** (Un test verde demuestra un caso de prueba, jamás inmunidad adversarial).
* ❌ **BUILD LOCAL $\to$ CI REPRODUCIBLE** (Compilar localmente no garantiza ausencia de dependencias ocultas o assets espurios).
* ❌ **HASH $\to$ AUTENTICIDAD** (Un hash sin firma asimétrica verificada no prueba autoría).
* ❌ **FIRMA VÁLIDA $\to$ AUTORIZACIÓN** (La firma demuestra identidad/posesión de clave; jamás confiere acceso ZTNA).

---

## 📊 CAPÍTULO 2: TAXONOMÍA EPISTEMOLÓGICA Y ESTADOS DE AUDITORÍA

### 2.1 Taxonomía Obligatoria de 5 Estados
Toda función, paquete, módulo o capacidad descrita en el repositorio debe clasificarse formalmente en una de las siguientes cinco categorías:

1. **`HECHO`:** Implementado físicamente en código fuente en `src/`.
2. **`TESTEADO`:** Respaldado por tests unitarios, de integración o suites adversariales Passing sin excepciones.
3. **`MEDIDO`:** Cuantificado con métricas empíricas en CPU real (latencia ns/op, allocs/op, jitter RFC 3550, tasa de transferencia).
4. **`NO IMPLEMENTADO`:** Planificado conceptualmente o en diseño, pero sin código funcional en producción.
5. **`EXPERIMENTAL`:** Prototipos, servicios satelitales o adaptadores preliminares que operan fuera del camino crítico del Core.

### 2.2 Rigor en la Taxonomía de Evidencia
Prohibido el uso de términos absolutistas desregulados (*"100% Certificado para Producción"*, *"Auditoría Cerrada"* o *"Zero-Bugs"*). La evidencia solo puede reportarse en tres niveles:
* **`EVIDENCIA LOCAL`:** Ejecución de pruebas y análisis en el entorno de desarrollo del operador.
* **`CI REPRODUCIBLE`:** Pipeline de GitHub Actions ejecutado en contenedores limpios desde cero.
* **`DEMOSTRADO FÍSICAMENTE`:** Verificación en laboratorio de 2 nodos físicos distintos con sockets reales (ej. Nodo A `192.168.1.198` $\leftrightarrow$ Nodo B `192.168.1.106`).

---

## ⚡ CAPÍTULO 3: LOS 7 INVARIANTES Y AXIOMAS INMUTABLES DEL CORE I7

El núcleo IPVN7 descansa sobre siete invariantes matemáticos y de ingeniería inquebrantables:

### Axioma I — El Core solo transporta estructura, jamás semántica
El núcleo de red resuelve exclusivamente identidad criptográfica, encapsulación, integridad, control de repetición y enrutamiento en malla. Ninguna lógica de aplicación (chat, archivos, streaming, inteligencia artificial, modelos MCP o transacciones de pago) reside en el Core; todo se acopla externamente vía sockets o el Smart Component Gateway.

### Axioma II — Las 10 Primitivas Nucleares Inmutables
El Core está congelado estrictamente en 10 primitivas fundamentales:
1. `Identity` (Ed25519 DID determinista).
2. `Packet` (Datagrama binario CBOR determinista RFC 8949).
3. `Container` (Estructura de empaquetado L0/L1).
4. `Object` (Payload binario desacoplado de semántica).
5. `Session` (Canal criptográfico efímero derivado con KEM).
6. `Channel` (Multiplexación lógica de flujos).
7. `Integrity` (Autenticación criptográfica Poly1305 / HMAC-SHA256).
8. `AntiReplay` (Ventana deslizante de 1024 bits indexada por origen y sesión).
9. `MTU` (Presupuesto máximo de datagrama de 1280 bytes).
10. `Routing` (Enrutamiento voraz en malla de Kleinberg por distancia XOR).

### Axioma III — Límite Estricto de 400 Líneas por Archivo (Axioma III)
Ningún archivo de código fuente puede superar las 400 líneas de código. Se activa alarma preventiva de refactorización a las 320 líneas. La modularidad radical previene monolitos y reduce la complejidad ciclomática.

### Axioma IV — Invariante de Memoria Zero-Copy (0 B/op)
En el camino crítico de transporte (`LinearPipeline`), los búferes se reciclan atómicamente a través de ring-buffers preasignados en 3 tiers (64B, 1500B, 64KB).
* **Métrica inviolable:** **0 B/op** y **0 allocs/op** certificados por benchmarks reproducibles (`BenchmarkLinearPipeline_Execute`).

### Axioma V — Invariante MTU Determinista de 1280 Bytes
Para eludir de forma absoluta la fragmentación de paquetes en la Internet pública (RFC 8200) y mitigar caídas de MTU en túneles VPN:
* Todo paquete I7 (encabezado + payload + tag KEM + firma) se garantiza menor o igual a **1280 bytes**.

### Axioma VI — ZTNA Estricto (Default-Deny Bidireccional)
* La verificación de una firma digital Ed25519 demuestra posesión de clave privada (autenticidad e identidad), **jamás autorización de red**.
* Prohibida la auto-autorización en paquetes de enlace.
* Toda sesión entrante es descartada silenciosamente (`Default-Deny`) a menos que el DID emisor se encuentre explícitamente autorizado en la política de control de acceso local (`Firewall.IsAllowed`).

### Axioma VII — Enrutador Kleinberg Canónico de 16 Anillos
El enrutador de mundo pequeño opera formalmente sobre **16 anillos concéntricos logarítmicos deterministas** (`CanonicalKleinbergRings = 16`, límite de 120 pares en memoria). La selección de saltos se calcula en tiempo $O(\log N)$ por distancia métrica XOR.

---

## 🏗️ CAPÍTULO 4: DELIMITACIÓN ARQUITECTÓNICA DE DOMINIOS

```text
┌─────────────────────────────────────────────────────────────────┐
│               EXPERIMENTAL & APLICACIONES                       │
│    (Sphinx Onion, MCP AI Agents, A2A, WASM, Sovereign Egress)   │
└────────────────────────────────┬────────────────────────────────┘
                                 │ Smart Component Gateway (REST/WS)
┌────────────────────────────────┴────────────────────────────────┐
│               SERVICIOS DE SISTEMA (SERVICES)                   │
│         (Discovery, STUN RFC 5389, SOCKS5, WebUI 127.0.0.1)     │
└────────────────────────────────┬────────────────────────────────┘
                                 │
┌────────────────────────────────┴────────────────────────────────┐
│               ADAPTADORES DE TRANSPORTE (ADAPTERS)              │
│               (UDP Nativo, Wintun L3, TCP Fallback)             │
└────────────────────────────────┬────────────────────────────────┘
                                 │
┌────────────────────────────────┴────────────────────────────────┐
│               NÚCLEO SOBERANO I7 (CORE CONGELADO)              │
│       [Las 10 Primitivas: Identity, Packet, MTU 1280, ZTNA...]   │
└─────────────────────────────────────────────────────────────────┘
```

* **Core:** Inmutable, 0 dependencias externas, cero alocaciones de heap en hot-path.
* **Adapters:** Traducen datagramas del SO hacia el Core y viceversa.
* **Services:** Servicios utilitarios de valor agregado que consumen el Core.
* **WebUI:** Confinada estrictamente a `127.0.0.1:7070` con token de autorización local (CORS restringido, prohibido bind `0.0.0.0` sin auth).

---

## 📋 CAPÍTULO 5: MATRIZ FACTUAL DE CIERRE DE AUDITORÍA EXTERNA (ee56f89)

La auditoría externa independiente realizada sobre el commit `ee56f89` formuló un conjunto de observaciones críticas. Esta matriz documenta formalmente su estado de resolución comprobable en código:

| ID | Área Auditada | Severidad Original | Hallazgo Inicial | Estado Actual | Evidencia y Verificación |
|---|---|---|---|---|---|
| **AUD-01** | Infraestructura CI | 🔴 **P0** | Renombrado incorrecto de carpetas `.github` y `.vscode`. | **`HECHO / TESTEADO`** | Restituido a `.github/` y `.vscode/`. Pipeline multi-job limpio sin assets binarios precompilados. |
| **AUD-02** | ZTNA Firewall | 🔴 **P0** | Auto-autorización indebida de DIDs tras validar firma en Handshake (`session_manager.go`). | **`HECHO / TESTEADO`** | Default-Deny estricto restaurado. Removido `firewall.AuthorizeDID()` automático. Test adversarial passing: `TestAdversarial_TestI_UninvitedPeerHandshakeRejected_DefaultDeny`. |
| **AUD-03** | Criptografía PQC | 🔴 **P0** | Combiner X-Wing difería del borrador IETF (SHA-256 vs SHA3-256, etiqueta errónea). | **`HECHO / TESTEADO`** | Actualizado a SHA3-256 canónico y etiqueta canónica de 6 bytes `\.//^\` (`0x5c, 0x2e, 0x2f, 0x2f, 0x5e, 0x5c`). |
| **AUD-04** | Enrutamiento | 🟠 **P1** | Discrepancia entre documentación (12 anillos) y código (16 anillos). | **`HECHO / TESTEADO`** | Unificado formalmente a 16 anillos canónicos (`CanonicalKleinbergRings = 16`). |
| **AUD-05** | WebUI / Admin | 🟠 **P1** | Potencial exposición de interfaz administrativa y riesgo CORS `*`. | **`HECHO / TESTEADO`** | Bind local confinado a `127.0.0.1:7070` con token RBAC para operaciones críticas de VPN y corte. |
| **AUD-06** | Actualizaciones | 🟠 **P1** | Riesgo de SSRF al consultar URLs de actualización arbitrarias. | **`HECHO / TESTEADO`** | Restringido a hosts de confianza preconfigurados con verificación de firma Ed25519 y SHA-256. |
| **AUD-07** | Matriz Release | 🟡 **P2** | Rotulado engañoso de release universal para macOS conteniendo solo ARM64. | **`HECHO / TESTEADO`** | Pipeline de `.github/workflows/release.yml` actualizado con compilación dual explícita para `darwin/amd64` y `darwin/arm64`. |

---

## 👑 CAPÍTULO 6: MODELO DE AUTOGOBERNANZA Y ORQUESTACIÓN

El desarrollo, evolución y auditoría continua del repositorio IPVN7 sigue una jerarquía estricta:

```text
                  OPERADOR HUMANO
                         │
                         ▼
        👑 AGENTE PRINCIPAL (FrondaBrick_01)
         (Interfaz Conversacional de Gobernanza)
                         │
                         ▼
          LÍMITES CONSTITUCIONALES & SCOPE LOCK
                         │
     ┌───────────────────┴───────────────────┐
     ▼                                       ▼
AGENTE 1 (Arquitecto)                   AGENTE 4 (Atacante / Tester)
  Diseño, Interfaces, Límites             Fuzzing, Tests Hostiles
     │                                       ▲
     ▼                                       │
AGENTE 3 (Implementador) ────────────────────┘
  Código mínimo, <= 400L, Zero-Copy
     │
     ▼
AGENTE 2 (Seguridad ZTNA) ──► AGENTE 5 (CI/Build) ──► AGENTE 6 (Performance)
                                                            │
                                                            ▼
                                                      AGENTE 7 (Auditor/Docs)
```

### Reglas de Oro de Gobernanza
1. **Invocación Universal:** Toda entidad (humana o sintética) que interactúe con el repositorio asume el rol rector de Frondabrick para preservar disciplina, contexto e invariantes.
2. **Un Solo Propósito por Commit:** Cada commit debe responder exclusivamente a una pregunta técnica concreta (ej. `fix: reject uninvited peer handshakes under default-deny`).
3. **Compuerta Previa Obligatoria:** Ningún cambio se integra a `main` sin la ejecución con 100% PASS de la Magna Multi-Suite (`scripts/verify_ipvn7_standard.ps1`).

---

## 🔄 CAPÍTULO 7: CONTINUIDAD Y VERIFICACIÓN AUTÓNOMA

El repositorio cuenta con dos mecanismos de verificación continua:
1. **Compuerta Local Inmediata (0 Tokens):** `scripts/verify_ipvn7_standard.ps1` valida de forma proactiva:
   - Invariante de 400 líneas por archivo.
   - Análisis estático con `go vet ./...`.
   - Concurrencia y carreras con `-race`.
   - Presupuesto de memoria Zero-Copy (0 B/op).
   - Resiliencia de red y sockets UDP físicos.
   - Pruebas unitarias de todos los paquetes Go (100% PASS).
   - Compilación limpia del binario de producción `bin/ipvn7.exe`.
2. **Daemon Centinela Autónomo:** Proceso supervisor en segundo plano (`scripts/run_autonomous_daemon.ps1 -IntervalMinutes 30`) que ejecuta periódicamente la suite, mantiene telemetría y previene colisiones mediante el cerrojo atómico `agentes/task.lock`.
