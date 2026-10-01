# BASELINE FACTUAL IPVN7 v0.7 — AUDITORÍA EXTERNA (FASE 0)

**Fecha de Medición:** 1 de Octubre de 2026  
**Commit Auditado:** `5c0189b79a24ebb1b90dff76760b53ae57896b5b`  
**Fuente de Verdad Suprema:** `docs/auditoria externa.md`  
**Estado General:** **AUDITADO INTERNAMENTE CON EVIDENCIA LOCAL (NO CERRADO)**  

---

## 1. Declaración Factual de Estado
En conformidad con la directiva de la Auditoría Externa, el proyecto abandona cualquier pretensión de auto-declararse "100% CERTIFICADO", "CERRADO" o "PRODUCTION-READY". El estado real del sistema es de implementación activa bajo riguroso control de calidad, mitigación de vulnerabilidades y verificación física falsable.

---

## 2. Mediciones Locales de Verificación y Compilación

### 2.1 Análisis Estático (`go vet ./...`)
* **Comando:** `go vet ./...` (en `src/`)
* **Resultado:** **PASS (Código 0)**.
* **Nota de CI:** Se desacopló `src/cmd/installer/main.go` mediante el build tag `//go:build windows && installer` para garantizar que `go vet` en checkout limpio no falle por assets precompilados inexistentes.

### 2.2 Suite Unitaria (`go test -count=1 ./...`)
* **Comando:** `go test -count=1 ./...` (en `src/`)
* **Resultado:** **PASS (100% de paquetes evaluados sin errores)**:
  - `ipvn7/pkg/core`: PASS (~2.2s)
  - `ipvn7/pkg/l0`: PASS (~1.2s)
  - `ipvn7/pkg/l1`: PASS (~3.5s)
  - `ipvn7/pkg/l2`: PASS (~0.8s)
  - `ipvn7/pkg/wasm`: PASS (~0.9s)

### 2.3 Compilación de Binarios (`go build ./...`)
* **Comando:** `go build ./...` (en `src/`)
* **Resultado:** **PASS (Código 0)**.
* **Binarios Verificados Generables:**
  - `bin/ipvn7.exe` (Daemon principal de nodo de red soberano).
  - `bin/ipvn7-wasm.wasm` (Compilación WebAssembly con `GOOS=js GOARCH=wasm`).
  - `dist/Instalador_VPN_I7.exe` (Compilable mediante pipeline desacoplado `scripts/build_installer.ps1` con `-tags installer`).

---

## 3. Matriz de Estado de los Hallazgos de Auditoría Externa

| # | Severidad | Hallazgo de Auditoría Externa | Estado Local | Diagnóstico y Mitigación Aplicada |
|---|:---:|---|:---:|---|
| **1** | **🔴 P0** | **CI actual roto por assets de instalador** | `HECHO / TESTEADO` | Añadido build tag `//go:build windows && installer` a `src/cmd/installer/main.go`. `go vet` pasa limpio en checkout sin requerir binarios en git. |
| **2** | **🔴 P0** | **Bypass conceptual ZTNA en RoamingUpdate** | `HECHO / TESTEADO` | Erradicada la auto-autorización (`firewall.AuthorizeDID`) ante recepción de `MsgTypeRoamingUpdate`. Separada autenticación (firma Ed25519) de autorización (Default-Deny). Validado con `roaming_ztna_test.go`. |
| **3** | **🔴 P0** | **Web UI expuesta en 0.0.0.0 sin autenticación** | `HECHO / TESTEADO` | Bind HTTP confinado estrictamente a `127.0.0.1:7070` por defecto. Eliminado CORS `*` en admin. Rutas críticas protegidas por RBAC. |
| **4** | **🔴 P0** | **`/vpn/exit` puede apagar el nodo remotamente** | `HECHO / TESTEADO` | Confinado a `127.0.0.1` y protegido mediante token RBAC de administración local (`checkAdminAuth`). |
| **5** | **🔴 P0** | **Actualización remota arbitraria (SSRF)** | `HECHO / TESTEADO` | Eliminada aceptación de URLs arbitrarias en `/api/v1/update/check`. Solo se aceptan fuentes en lista blanca oficial de GitHub o servidor configurado. Manifiesto firmado con Ed25519 + SHA256. |
| **6** | **🟠 P1** | **Nombres engañosos tipo ML-DSA con SHA256** | `HECHO / TESTEADO` | Renombrados identificadores experimentales a `ExperimentalPQCIdentity` y `ExperimentalSigSize`. Firma real descansa en Ed25519 canónico. |
| **7** | **🟢 P1** | **Criptografía ML-KEM-768 FIPS 203 Real** | `HECHO / TESTEADO` | Implementación basada en biblioteca estándar `crypto/mlkem` con clave pública de 1184B, criptograma de 1088B y decapsulación canónica. |
| **8** | **🟢 P1** | **Invariante MTU 1280B** | `HECHO / MEDIDO` | Límite determinista de 1280 bytes respetado en el datagrama para evitar fragmentación IP. |
| **9** | **🟢 P1** | **Anti-Replay por Sesión (1024 bits)** | `HECHO / TESTEADO` | Aislamiento por `(originDID, sessionID, sequence, timestamp)` en `src/pkg/l1/anti_replay.go`. |

---

## 4. Clasificación Factual de Módulos (Taxonomía de 5 Estados)

* **CORE (Identity, Packet, Container, Session, Channel, MTU, AntiReplay, Routing):** `HECHO / TESTEADO / MEDIDO`.
* **ADAPTERS (UDP, TUN):** `HECHO / TESTEADO`.
* **SERVICES (WebUI localhost, SOCKS5, Discovery silencioso, STUN):** `HECHO / TESTEADO`.
* **EXPERIMENTAL (Sphinx, WASM, MCP, A2A, x402, AI):** `EXPERIMENTAL` (Aislados fuera del camino crítico del Core).
* **MALLA PLANETARIA / TRILLONES DE NODOS:** `NO IMPLEMENTADO` (Hipótesis teórica en especificaciones satelitales).
