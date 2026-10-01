# COLA DE AUTOTAREAS DE AUDITORÍA EXTERNA IPVN7 (agentes)

Este documento registra la planificación y el estado factual de ejecución continua del Agente de Red Soberano (`ipvn7-network-os-agent`), organizado estrictamente bajo el **Plan Canónico de 13 Fases de la Auditoría Externa** (`docs/auditoria externa.md`).

**Taxonomía Factual de 5 Estados:** `HECHO` (código presente), `TESTEADO` (tests passing), `MEDIDO` (benchmarks verificados), `EXPERIMENTAL` (en calibración), `NO IMPLEMENTADO` (pendiente).

---

## 📊 Matriz de Fases Canónicas de la Auditoría Externa

| Fase | Título de la Fase | Responsable | Estado Factual | Evidencia Verificable |
| :---: | :--- | :---: | :---: | :--- |
| **FASE 0** | **Baseline y Congelamiento Inicial** | Agente 7 (Docs) | **HECHO / TESTEADO** | `docs/BASELINE.md`, mediciones `go vet`, `go test` y compuerta limpia |
| **FASE 1** | **Seguridad Crítica: ZTNA Default-Deny** | Agente 2 (Seguridad) | **HECHO / TESTEADO** | `src/pkg/l1/roaming_ztna_test.go` (Firma válida NO autoriza DID) |
| **FASE 2** | **Cerrar WebUI y Proteger API Admin** | Agente 2 (Seguridad) | **HECHO / TESTEADO** | `src/pkg/core/web_ui_handlers.go`, bind `127.0.0.1`, auth RBAC, CORS local |
| **FASE 3** | **Eliminar SSRF en Actualizaciones** | Agente 2 (Seguridad) | **HECHO / TESTEADO** | `version_manager.go`, whitelist oficial, firma Ed25519 + SHA256 |
| **FASE 4** | **Arreglar CI y Desacoplar Instalador** | Agente 5 (CI/Build) | **HECHO / TESTEADO** | Build tag `windows && installer`, `go vet ./...` pasa en checkout limpio |
| **FASE 5** | **Limpiar Criptografía y Nombres Honestos** | Agente 2 (Seguridad) | **HECHO / TESTEADO** | `crypto/mlkem` real, sustitución de falsos ML-DSA por `ExperimentalPQCIdentity` |
| **FASE 6** | **Congelar el CORE (10 Primitivas Puras)** | Agente 1 (Arquitecto) | **HECHO / TESTEADO** | Desacoplamiento de Core vs Adapters, Services y Experimental |
| **FASE 7** | **Crear Contratos del CORE (`docs/core/`)** | Agente 1 (Arquitecto) | **HECHO** | `docs/core/*.md` (Identity, Packet, Container, Session, Channel, Routing, MTU, AntiReplay) |
| **FASE 8** | **Contratos Matemáticos e Invariantes** | Agente 3 (Core) | **TESTEADO / MEDIDO** | Invariante MTU <= 1280B, SessionID único, ventana anti-replay 1024b |
| **FASE 9** | **Pruebas Destructivas (Adversarial & Fuzzing)**| Agente 4 (Testing) | **TESTEADO** | `session_adversarial_test.go`, fuzzing determinista de parsers |
| **FASE 10**| **Rendimiento Empírico y Telemetría Vital** | Agente 6 (Performance)| **MEDIDO** | 33 ns/op, 0 B/op, 0 allocs/op en pipeline, ECG a 60 FPS reactivo |
| **FASE 11**| **Interoperabilidad Multi-Nodo Física** | Agente 4 (Testing) | **TESTEADO** | Laboratorio 2 nodos (PC `192.168.1.198` <-> Notebook `192.168.1.106`) |
| **FASE 12**| **Servicios Desacoplados sobre el CORE** | Agente 1 (Arquitecto) | **EXPERIMENTAL** | VPN 1-clic, SOCKS5, DNS Mesh y servicios satélites como consumidores del Core |

---

## 📝 Detalle de Tareas por Fase Canónica

### FASE 0: BASELINE Y CONGELAMIENTO
* **TASK-F0-1 [HECHO/TESTEADO]:** Medición y registro de línea base en `docs/BASELINE.md`: verificación de commit, compilación, `go vet` y tests limpios.
* **TASK-F0-2 [HECHO]:** Purga documental de términos absolutistas ("100% CERRADO", "CERTIFICADO"). Estado oficial: AUDITADO INTERNAMENTE CON EVIDENCIA LOCAL.

### FASE 1: SEGURIDAD CRÍTICA ZTNA
* **TASK-F1-1 [HECHO/TESTEADO]:** Erradicar invocación `firewall.AuthorizeDID` al recibir `MsgTypeRoamingUpdate` en `main.go` y `pipeline_stages.go`. La firma válida Ed25519 solo autentica identidad, jamás otorga autorización.
* **TASK-F1-2 [HECHO/TESTEADO]:** Suite de pruebas en `src/pkg/l1/roaming_ztna_test.go`:
  - `TestRoamingValidSignatureUnauthorizedDID`: descarta y bloquea.
  - `TestRoamingValidSignatureAuthorizedDID`: acepta si fue pre-autorizado.
  - `TestRoamingInvalidSignature`: descarta por firma forjada.
  - `TestRoamingUnknownDID`: descarta por DID desconocido.

### FASE 2: CERRAR WEBUI Y PROTEGER RUTAS ADMIN
* **TASK-F2-1 [HECHO/TESTEADO]:** Bind HTTP confinado a `127.0.0.1:7070` por defecto en `web_ui.go`. Prohibido `0.0.0.0` sin autorización explícita.
* **TASK-F2-2 [HECHO/TESTEADO]:** Supresión de CORS `*` en panel administrativo (reemplazado por loopback restringido).
* **TASK-F2-3 [HECHO/TESTEADO]:** Protección RBAC (`checkAdminAuth`) en operaciones críticas: `/vpn/connect`, `/vpn/disconnect`, `/vpn/exit`, `/vpn/cycle` y `/update/*`. Validado con `web_ui_admin_auth_test.go`.

### FASE 3: ELIMINAR SSRF EN ACTUALIZACIÓN
* **TASK-F3-1 [HECHO/TESTEADO]:** Erradicación de URLs arbitrarias en `/api/v1/update/check?url=...`. Solo se admiten endpoints en lista blanca oficial de GitHub o servidor configurado de confianza.
* **TASK-F3-2 [HECHO/TESTEADO]:** Cadena de confianza estricta para auto-actualización: Manifiesto -> Firma digital Ed25519 del desarrollador -> Hash SHA-256 -> Binario.

### FASE 4: ARREGLAR CI Y DESACOPLAR INSTALADOR
* **TASK-F4-1 [HECHO/TESTEADO]:** Incorporación del build tag `//go:build windows && installer` en `src/cmd/installer/main.go`, permitiendo que `go vet ./...` y `go test ./...` pasen sin requerir `assets/ipvn7.exe` ni `assets/wintun.dll` en el checkout base.
* **TASK-F4-2 [HECHO/TESTEADO]:** Pipeline de compilación desacoplado en `scripts/build_installer.ps1`: 1) Build de `ipvn7.exe`, 2) Descarga/copia de wintun, 3) Compilación del instalador con tags.

### FASE 5: LIMPIAR CRIPTOGRAFÍA
* **TASK-F5-1 [HECHO/TESTEADO]:** Preservar el núcleo post-cuántico real: NIST FIPS 203 ML-KEM-768 (`crypto/mlkem`), X25519, Ed25519, HKDF y ChaCha20-Poly1305.
* **TASK-F5-2 [HECHO/TESTEADO]:** Erradicar pseudónimos que sugieran ML-DSA en implementaciones simuladas con SHA-256; renombrados honestamente a `ExperimentalPQCIdentity` y `ExperimentalSigSize`.
* **TASK-F5-3 [HECHO/TESTEADO]:** Handshake binding completo en 1-RTT y aislamiento estricto de sesión en `session_manager.go`.

### FASE 6: CONGELAR EL CORE IPV7
* **TASK-F6-1 [HECHO]:** Delimitación nítida del CORE en 10 primitivas estructurales en `src/pkg/l0` y `src/pkg/core`.
* **TASK-F6-2 [HECHO]:** Aislamiento de adaptadores de transporte (UDP, TUN), servicios de red (SOCKS5, STUN, WebUI) y extensiones experimentales (Sphinx, WASM, MCP, AI).

### FASE 7: CONTRATOS DEL CORE
* **TASK-F7-1 [HECHO]:** Publicación de las especificaciones canónicas de contrato en `docs/core/`:
  - `identity.md`, `packet.md`, `container.md`, `session.md`, `channel.md`, `routing.md`, `mtu.md`, `anti_replay.md`.

### FASE 8: CONTRATOS MATEMÁTICOS E INVARIANTES
* **TASK-F8-1 [TESTEADO/MEDIDO]:** Invariante determinista `payload + header <= 1280` bytes demostrado en tests de frontera (`l0_test.go`, `session_adversarial_test.go`).
* **TASK-F8-2 [TESTEADO]:** Ventana anti-replay de 1024 bits indexada por `(originDID, sessionID, sequence)` en `src/pkg/l1/anti_replay.go`.

### FASE 9: PRUEBAS DESTRUCTIVAS Y ADVERSARIALES
* **TASK-F9-1 [TESTEADO]:** Batería adversarial completa en `session_adversarial_test.go`: suplantación de origen, firmas corruptas, repetición de tramas, cruce de sesiones y mitigación de DoS.
* **TASK-F9-2 [TESTEADO]:** Fuzzing determinista sobre parsers CBOR y límites de datagrama en `wire_fuzz_test.go`.

### FASE 10: RENDIMIENTO Y TELEMETRÍA VITAL
* **TASK-F10-1 [MEDIDO]:** Certificación empírica de 0 B/op y 0 allocs/op en orquestación de pipeline interno (`BenchmarkLinearPipeline_Execute`).
* **TASK-F10-2 [MEDIDO]:** Datapath criptográfico completo validado en ~3.6 µs/op (~270,000 datagramas/s por hilo).
* **TASK-F10-3 [HECHO/TESTEADO]:** Monitor bio-mórfico de ritmo cardíaco (ECG a 60 FPS reactivo al flujo real de datos en `templates.go`).

### FASE 11: INTEROPERABILIDAD FÍSICA
* **TASK-F11-1 [TESTEADO]:** Validación bidireccional física en topología 1-a-1 limpia: PC Principal `192.168.1.198` <-> Notebook `192.168.1.106` sin nodos fantasma ni auto-emparejamiento.

### FASE 12: SERVICIOS SOBRE EL CORE
* **TASK-F12-1 [EXPERIMENTAL]:** Consumo del Core por parte de servicios superiores (VPN I7 1-clic, SOCKS5 proxy userspace, DNS anti-fugas) sin contaminar las 10 primitivas.
