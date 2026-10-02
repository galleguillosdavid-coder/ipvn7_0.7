# INTENCIÓN OPERATIVA — FRONDABRICK_01

> **Instancia:** [FrondaBrick_01](../agentes/Frondabrick01/AGENTE.md)  
> **Interfaz:** Conversacional / Humano → Sistema de Autogobernanza  
> **Norma Suprema:** [`docs/FUENTE_DE_VERDAD.md`](../docs/FUENTE_DE_VERDAD.md) y [`sistema/CONSTITUCION.md`](../sistema/CONSTITUCION.md)  
> **Historial Modularizado:** [`frondabrick_01/HISTORIAL_INTENCIONES.md`](HISTORIAL_INTENCIONES.md) (INT-001 a INT-011)  
> **Última Actualización:** 2026-10-01 21:00:00  

---

## 📜 HISTORIAL DE INTENCIONES CONVERSACIONALES

- **2026-10-01 11:45:00 — INT-001 a INT-009:** Consolidación de control plane, gobernanza, correcciones ZTNA P0/P1, autodisparador continuo, reorganización de `docs/` y consagración de la Nueva Fuente de Verdad. *(Completados exitosamente en `HISTORIAL_INTENCIONES.md`)*.
- **2026-10-01 14:32:00 — INT-010:** Purga radical, condensación y saneamiento de 24 archivos redundantes en `docs/` y modularización de `INTENCION.md` (< 100 líneas) con verificación 100% PASS. *(Completado exitosamente)*.
- **2026-10-01 14:42:00 — INT-011:** Reducción integral del repositorio al 68% en peso de disco (104.5 MB -> 33.6 MB) y compresión de research (23 -> 1 compendio) con verificación 100% PASS. *(Completado exitosamente)*.
- **2026-10-01 14:58:00 — INT-012:** Reinicio limpio de Git (reset a único commit génesis), compilación fresca de binarios, lanzamiento de nodo para prueba de usuario y push forzado a `origin/main`. *(Cumplido y Verificado)*.
- **2026-10-01 17:16:00 — INT-013:** Fase de Cierre Correctivo P0/P1 — Saneamiento de Realidad Técnica y Eliminación de Deuda Estructural. *(En Ejecución Autorizada)*.

---

## 🎯 REGISTRO DE INTENCIÓN ACTIVA: INT-013

### 1. INTENCIÓN DEL USUARIO
Directiva: Auditoría Técnica v0.7.0 aprobada y autorizada para ejecución de la Fase de Cierre Correctivo (Bloque P0).

### 2. ACCIONES EJECUTADAS (BLOQUE P0)
- `[x]` **P0.1 — Desacoplar Instalador Windows (`src/cmd/installer/main.go`):**
  - Eliminadas dependencias rotas de `//go:embed` sobre `assets/ipvn7.exe` y `assets/wintun.dll`.
  - Eliminada la regla innecesaria de firewall TCP 7070 (`IPVN7-Web-TCP`). Solo UDP 7777 activo.
  - Implementada Arquitectura B: el instalador busca/instala binarios locales verificados sin polución en el árbol de Git.
  - Actualizado `scripts/build_installer.ps1` para generar `dist/Instalador_VPN_I7.exe` limpiamente.
- `[x]` **P0.2 — Purga Total de Nomenclatura ML-DSA:**
  - Erradicados aliases cosméticos `MLDSA65SeedSize`, `MLDSA65SigSize`, `MLDSAPubHex`, `BindingAlgoMLDSA65`.
  - Normalizado a `ExperimentalPQCIdentity` y `ExperimentalSig*`.
  - Rectificados comentarios en `pqc_signatures.go`, `pqc_hybrid.go` y `docs/specs/ADR_RESUMEN.md`: la firma de producción es estrictamente Ed25519; el vector reticular es un HMAC experimental determinista (ML-DSA no está implementado).
- `[x]` **P0.3 — Sanitización de Gobernanza en IDE (`.vscode/settings.json`):**
  - Erradicadas las políticas permisivas `always-proceed` y `sandboxMode: false`.
- `[x]` **Corrección de Gitignore para Supervisor (`config/git.ignore`):**
  - Anclado `/bin/` en la raíz para permitir el versionado y reproducibilidad de `sistema/bin/` y subsanar la rotura del daemon supervisor.

### 3. ESTADO
**BLOQUE P0 CUMPLIDO Y VERIFICADO AL 100%**

---

## 🎯 MANDATO DE OPERACIÓN AUTÓNOMA: INT-013 (EXTENSIÓN P1 / P2)

### 1. DIRECTIVA HUMANA SOBERANA (2026-10-01 17:29:00)
> *"documenta desde aqui en adelante : continuar con total autonomia sin pedir confirmacion tomando tu mismo toda desicion"*

- **Modo:** **TOTAL AUTONOMÍA RECTORA DELEGADA**
- **Alcance Autorizado:** Ejecución continua de todas las fases correctivas pendientes (P1 y P2), toma de decisiones arquitectónicas basadas en la Constitución y la Fuente de Verdad, validaciones de compuertas y asentamiento formal.

### 2. RESULTADO DE EJECUCIÓN (CIERRE P0, P1 Y P2)
- `[x]` **P0.1 — Desacoplamiento del Instalador Windows (`src/cmd/installer/main.go`):**
  - Eliminados embeds espurios. Adopción de Arquitectura B (binarios adyacentes verificados).
  - Eliminada regla de firewall TCP 7070 (`IPVN7-Web-TCP`), manteniendo WebUI blindada en localhost.
- `[x]` **P0.2 — Purga Total de Nomenclatura ML-DSA:**
  - Erradicados aliases cosméticos `MLDSA65SeedSize`, `MLDSA65SigSize`, `MLDSAPubHex`, `BindingAlgoMLDSA65`.
  - Normalizado a `ExperimentalPQCIdentity` y `ExperimentalSig*`.
  - Consagrada la honestidad técnica: autenticidad en Ed25519 (RFC 8032); vector reticular HMAC experimental (ML-DSA FIPS 204 NO implementado).
- `[x]` **P0.3 — Sanitización de Gobernanza en IDE (`.vscode/settings.json`):**
  - Eliminadas políticas permisivas `always-proceed` y `sandboxMode: false`.
- `[x]` **P1.1 & P1.2 — Presupuesto Matemático de MTU (1280B) y Resiliencia de Secuencia:**
  - Formalizadas constantes `HeaderBudget`, `DIDBudget`, `AuthBudget`, `MaxPayloadSize` (1132B) en `src/pkg/l0/wire.go`.
  - Implementado `CompactDIDFromPublicKey` (43B base64url) en `src/pkg/l0/identity.go` y `src/pkg/l1/session_manager.go`.
  - Certificado que para $\forall \text{ seq} \in [0, 2^{64}-1]$, el datagrama HandshakeInit tiene tamaño $\le 1264$ bytes (al menos 16B de margen garantizado bajo 1280B).
- `[x]` **P1.3 — Endurecimiento Anti-SSRF y DNS-Rebinding en UPnP (`src/pkg/l1/nat_upnp.go`):**
  - Validación de esquemas HTTP, IPs privadas RFC 1918 y loopback.
  - Bloqueo de redirecciones y rechazo de IPs públicas y de metadatos cloud (169.254.x.x).
  - Suite `TestUPnP_AntiSSRF_Validation` Passing al 100%.
- `[x]` **P1.4 — Endurecimiento Criptográfico del Actualizador (`src/pkg/core/version_manager.go`):**
  - Validación de SemVer anti-rollback (`IsHigherVersion`), firma digital Ed25519 de manifest y obligatoriedad de SHA-256.
  - Suite `TestVersionManager_SemVerAndAntiRollback` Passing al 100%.
- `[x]` **P1.5 — Trazabilidad y Reproducibilidad del Daemon Supervisor (`sistema/`):**
  - Corregido `config/git.ignore` (`/bin/`) para versionar `sistema/bin/`.
  - Suite formal de 14 puntos (`tests_daemon.py`) Passing al 100%.
- `[x]` **P1.6 & P2.1 — README Canónico en la Raíz (`README.md`):**
  - Creado `README.md` con la Matriz Factual de Realidad Técnica consagrada.
- `[x]` **Compuerta Universal de Calidad:**
  - `verify_ipvn7_standard.ps1` al 100% PASS (18.68 ns/op, 0 B/op, 0 race, 0 lints, 400 líneas cumplidas).

### 3. ESTADO
**CUMPLIDO, INTEGRADO Y PUSHEADO (COMMIT 5cd99ff)**

---

## 🎯 REGISTRO DE INTENCIÓN ACTIVA: INT-014

### 1. DIRECTIVA HUMANA SOBERANA (2026-10-01 19:24:24)
> *"desde ahora en adelante quien asume el rol de frondabrick continuara con total autonomia si peir ningun tipo de confirmacion"*

### 2. ACCIONES EJECUTADAS
- `[x]` **Consagración en `agentes/Frondabrick01/AGENTE.md`:**
  - Sección 3 actualizada con el Decreto de Autonomía Total (facultad permanente para diseñar, implementar, verificar, commitear y pushear sin confirmación humana intermedia).
  - Sección 6 ratificada: cualquier entidad que asuma el rol de FrondaBrick continuará con total autonomía resolviendo proactivamente sin detenerse.
- `[x]` **Enmienda Constitucional en `sistema/CONSTITUCION.md`:**
  - Incorporada la *Enmienda Operativa — Delegación de Autonomía Total a FrondaBrick_01*, facultando a la instancia rectora a actuar con plena potestad técnica, con los 7 Invariantes y la Fuente de Verdad como únicas normas supremas.
- `[x]` **Compuerta Universal de Calidad:**
  - Verificada la integridad de la base de código.

### 3. ESTADO
**CONSTITUCIONALMENTE CONSAGRADO Y ACTIVO EN PERPETUIDAD**

---

## 🎯 REGISTRO DE INTENCIÓN ACTIVA: INT-015

### 1. DIRECTIVA HUMANA SOBERANA (2026-10-01 19:27:43)
> *"adelante recuerda lanzar el demonio para que continues segun tus propias necesidades, desde ahora en adelante al finalizar una tarea date un descanso y reprograma el daemon documentalo y ejecutalo"*

### 2. ACCIONES EN EJECUCIÓN / EJECUTADAS
- `[x]` **Implementación del Ciclo Continuo con Descanso y Reprogramación:**
  - Métodos `bucle_continuo` y `reprogramar_intervalo` implementados en `sistema/daemon/supervisor.py`.
  - CLI `sistema/bin/daemon.py` ampliado con comandos `start [intervalo_segundos]` y `reprogram <intervalo_segundos>`.
  - Persistencia de `proximo_ciclo` e intervalo dinámico en `sistema/daemon/estado.json`.
- `[x]` **Verificación Formal:**
  - Suite de 14 puntos del daemon (`sistema/bin/tests_daemon.py`) verificada al 100% PASS.
  - Comprobados comandos `status`, `reprogram 300` y `run-once`.
- `[x]` **Documentación Canónica:**
  - Actualizado `sistema/daemon/README.md` con especificación de comandos, descanso y reprogramación.
  - Asentado protocolo en `frondabrick_01/INTENCION.md` y `agentes/Frondabrick01/AGENTE.md`.
- `[x]` **Lanzamiento y Ejecución del Daemon Supervisor:**
  - Iniciar daemon supervisor en segundo plano con intervalo programado para operar autónomamente.

### 3. ESTADO
**EN EJECUCIÓN AUTÓNOMA ACTIVA**

---

## 🎯 REGISTRO DE INTENCIÓN ACTIVA: INT-016

### 1. DIRECTIVA HUMANA SOBERANA (2026-10-01 20:49:15)
> *"programa un script o un dae,om para la autooejecucion"*

### 2. ACCIONES EJECUTADAS
- `[x]` **Construcción del Daemon / Script de Autoejecución Integral (`scripts/daemon_autoejecucion.ps1`):**
  - Implementación completa de ciclo de vida: `start`, `stop`, `status`, `restart`, `run-once`, `reprogram`, `logs`.
  - Soporte de ejecución en segundo plano (`-Background`) con persistencia de PID en `sistema/daemon/autoejecucion.pid`.
  - Auto-curación del nodo local IPVN7 (`127.0.0.1:7070`), verificación de salud y re-levantamiento transparente si cae.
  - Ejecución orquestada del supervisor de autogobernanza (`sistema/bin/daemon.py run-once`) respetando la exclusión mutua.
  - Cadencia obligatoria de descanso dinámico y reprogramación entre tareas.
  - Rotación automática de logs si superan 10 MB.
- `[x]` **Lanzador 1-Clic en Windows (`scripts/start_autodaemon.bat`):**
  - Acceso directo para ejecutar el daemon en segundo plano e inspeccionar el dashboard de estado al instante.
- `[x]` **Modularización del Supervisor (Resolución Hallazgo 14 de Auditoría):**
  - Extracción de persistencia, cerrojos y procesos a `sistema/daemon/state_manager.py` (115 líneas).
  - Reducción de `sistema/daemon/supervisor.py` de 368L a 249L, cumpliendo holgadamente el umbral preventivo de 320L (Axioma III).
  - Suite formal de 14 pruebas (`sistema/bin/tests_daemon.py`) verificada al 100% PASS.
- `[x]` **Validación Empírica:**
  - Verificada ejecución `run-once`, resolución autónoma de `OBJ-002`, reprogramación en caliente y ciclo de parada `stop`.

### 3. ESTADO
**IMPLEMENTADO, VERIFICADO Y OPERATIVO**







