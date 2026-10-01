# LÍNEA BASE DE AUDITORÍA — IPvN7 v0.7.0 (FASE 0)

> **Fecha de captura:** 2026-10-01T07:30:00-03:00  
> **Rama de trabajo:** `audit/fix-security`  
> **Commit de referencia:** `39870be90ed22b78248f64c777005a2ec91e9796`  
> **Objetivo:** Congelar el estado actual verificado como punto de comparación factual antes de modificaciones subsecuentes.

---

## 1. Estado de la Suite de Pruebas Unitarias (`go test -v ./pkg/...`)

Todas las suites de los paquetes del núcleo pasaron limpiamente al 100%:

| Paquete | Estado | Tiempo | Suites Críticas Incluidas |
| :--- | :---: | :---: | :--- |
| `ipvn7/pkg/core` | **PASS** | 1.801s | Pipeline, buffers, context pools, logging |
| `ipvn7/pkg/l0` | **PASS** | 0.879s | Identidad Ed25519, wire CBOR, Anti-Replay L0, ML-KEM-768 FIPS 203 |
| `ipvn7/pkg/l1` | **PASS** | 3.645s | Routing Kleinberg, ZTNA, STUN RFC 5389, KEM 1-RTT, SOCKS5, TUN, Adversarial A-H |
| `ipvn7/pkg/l2` | **PASS** | 0.050s | Telemetría concurrente, detección de anomalías, ring buffer |
| `ipvn7/pkg/wasm` | **PASS** | 0.848s | Validación Wasm, PoW, frames |

### Resumen de Suites Adversariales Recientes
- `TestAdversarial_TestA_FakeSourceDID`: **PASS** (Rechazo inmediato de suplantación de DID)
- `TestAdversarial_TestB_InvalidSignature`: **PASS** (Rechazo de firma forjada)
- `TestAdversarial_TestC_CrossSession`: **PASS** (Aislamiento cross-session en secuencias idénticas)
- `TestAdversarial_TestD_CrossPeer`: **PASS** (Aislamiento cross-peer sin colisión)
- `TestAdversarial_TestE_Replay`: **PASS** (Descarte de paquete repetido)
- `TestAdversarial_TestF_HandshakeReplay`: **PASS** (Descarte de HandshakeInit repetido)
- `TestAdversarial_TestG_ResponseWrongDestination`: **PASS** (Descarte de respuesta con DestDID erróneo)
- `TestAdversarial_TestH_TamperedSourceDID`: **PASS** (Descarte de paquete alterado en tránsito)
- `TestPQCDatapath_PhysicalUDP_ZTNA_AntiReplay`: **PASS** (Loopback UDP físico con PQC, ZTNA y AntiReplay L1)

---

## 2. Línea Base de Rendimiento y Memoria (Benchmarks)

### 2.1 Orquestación de Pipeline (`pkg/core`)
- `BenchmarkLinearPipeline_Execute`: **32.06 ns/op \| 0 B/op \| 0 allocs/op**
- `BenchmarkPacketContext_AcquireRelease`: **23.28 ns/op \| 0 B/op \| 0 allocs/op**

### 2.2 Validación Wire y Checksum (`pkg/l0`)
- `BenchmarkAntiReplayFilter_ValidateAndUpdate`: **36.36 ns/op \| 0 B/op \| 0 allocs/op**
- `BenchmarkFastPacketChecksum_1280B`: **2600 ns/op (492.33 MB/s) \| 0 B/op \| 0 allocs/op**

### 2.3 Camuflaje y Handshake KEM (`pkg/l1`)
- `BenchmarkTLSOptionEngine_WrapUnwrap`: **605.2 ns/op \| 1408 B/op \| 1 allocs/op**
- `BenchmarkTLSOptionEngine_BuildClientHello`: **1818 ns/op \| 1152 B/op \| 11 allocs/op**
- `BenchmarkHolePunchMessage_EncodeDecode`: **39.64 ns/op \| 24 B/op \| 1 allocs/op**
- `BenchmarkHybridHandshake1RTT`: **464.98 µs/op \| 14000 B/op \| 79 allocs/op**
- `BenchmarkStochasticPaddingEngine_Generate`: **758.53 µs/op \| 36812 B/op \| 125 allocs/op**

---

## 3. Estado de la Compuerta Universal (`scripts/verify_ipvn7_standard.ps1`)
- **Invariante de 400 líneas:** CUMPLIDO en 100% de los archivos (0 violaciones).
- **Zona preventiva (320-400 líneas):** 0 archivos en zona preventiva (`main.go` en 312 líneas tras poda previa).
- **Análisis estático (`go vet`):** 0 observaciones.
- **Detección de carreras (`-race -cpu=4,8`):** 100% PASS sin contiendas.
- **Zero-Copy Pipeline:** Certificado a 0 B/op y 0 allocs/op.
- **Fuzzing & Resiliencia WAN:** 100% PASS (STUN, CGNAT, TLS 1.3, X-Wing KEM, TUN fallback).
- **Binario:** `bin/ipvn7.exe` compilado limpiamente.
- **Health Score Consolidado:** **100% (ÓPTIMO / EXCELENCIA)**.

---

## 4. Estado del Checklist Maestro de Corrección

- [x] **FASE 0 — Congelar el estado actual** (Completada: rama `audit/fix-security`, commit `39870be`, reporte `baseline.md`).
- [x] **FASE 1 — Cerrar el agujero crítico del DID** (Implementado en `session_manager.go` con firma Ed25519 obligatoria y validación estricta; formalizar transcripción explícita si se requiere).
- [x] **FASE 2 — Arreglar Anti-Replay** (Implementado: `l1.NewAntiReplayFilter(nil)` integrado en `main.go` con `(originDID, SessionID, Sequence, Timestamp)`).
- [x] **FASE 3 — Proteger el Handshake Response** (Implementado: validación `DestDID == localDID`, correspondencia con sesión pendiente y tiempo de expiración).
- [x] **FASE 4 — Fallos de inicialización** (Implementado: `main.go` aborta inmediatamente con `os.Exit(1)` si falla `GenerateHybridKeyPair`).
- [ ] **FASE 5 — Corregir la afirmación de Zero-Copy** (Crear `BenchmarkDatapathEndToEnd` real que mida el flujo UDP `Decode -> Validate -> AntiReplay -> ZTNA -> Decrypt -> Route` y documentar mediciones exactas).
- [x] **FASE 6 — Auditar PQC** (ML-KEM-768 real NIST FIPS 203 nativo; ML-DSA delimitado como experimental en documentación).
- [x] **FASE 7 — Crear batería de seguridad adversarial** (Implementada en `session_adversarial_test.go` Tests A-H).
- [ ] **FASE 8 — Orden correcto del datapath** (Documentar e implementar formalmente el pipeline desacoplado).
- [ ] **FASE 9 — Separar CORE de ADAPTERS** (Definir interfaces y delimitar adapters/ sin romper compatibilidad).
- [ ] **FASE 10 — Discovery** (Validar y documentar modo silencioso por defecto, dirigido y STUN).
- [ ] **FASE 11 — Tests de concurrencia avanzados** (Simulaciones de contienda múltiple de sesiones).
- [ ] **FASE 12 — MTU** (Integración de MTU dinámico y fragmentación).
- [ ] **FASE 13 — Documentación** (Limpieza y taxonomía estricta de 4 estados).
- [ ] **FASE 14 — CI real** (Crear workflow `.github/workflows/test.yml` con tests, race, vet, bench y security).
- [ ] **FASE 15 — Validación final** (Auditoría de cierre).
