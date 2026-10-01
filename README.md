# IPVN7 — Universal Sovereign Mesh & Post-Quantum Datapath (v0.7.0)

[![CI & Security Gate](https://github.com/galleguillosdavid-coder/ipvn7_0.7/actions/workflows/ci.yml/badge.svg)](https://github.com/galleguillosdavid-coder/ipvn7_0.7/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/galleguillosdavid-coder/ipvn7_0.7)](https://goreportcard.com/report/github.com/galleguillosdavid-coder/ipvn7_0.7)
[![Zero-Copy](https://img.shields.io/badge/Datapath-Zero--Copy%20(0%20B%2Fop)-brightgreen)](#invariantes-del-núcleo)
[![MTU](https://img.shields.io/badge/MTU-1280%20Bytes%20Deterministic-blue)](#invariantes-del-núcleo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**IPVN7** es un protocolo soberano de malla entre pares (P2P), enrutamiento descentralizado y datapath de ultra-bajo retardo resistente a la computación cuántica, construido bajo el paradigma **Zero-Trust Network Access (ZTNA)** y memoria **Zero-Copy**.

---

## 📊 Matriz Factual de Realidad Técnica

Siguiendo el principio de **Honestidad Técnica Radical** consagrado en nuestra [Fuente de Verdad](docs/FUENTE_DE_VERDAD.md) y la Constitución de Autogobernanza, cada componente se clasifica de acuerdo con su estado empírico demostrado:

| Componente | Clasificación Factual | Estado de Implementación | Evidencia Técnica |
| :--- | :--- | :--- | :--- |
| **Ed25519 (RFC 8032)** | Estándar de Producción | **HECHO / TESTEADO** | Autenticación y no-repudio canónico de nodos. |
| **ML-KEM-768 (`crypto/mlkem`)** | Post-Cuántica Canónica | **HECHO / TESTEADO** | KEM FIPS 203 nativo de Go (1184B pub, 1088B ct). |
| **ZTNA Default-Deny** | Control Plane / Autorización | **HECHO / TESTEADO** | Aislamiento estricto: autenticidad $\neq$ autorización. |
| **CBOR Wire Framing (RFC 8949)** | Datapath Framing | **HECHO / TESTEADO** | Serialización binaria canónica determinista. |
| **Anti-Replay 1024-bit** | Filtro de Ventana Deslizante | **HECHO / TESTEADO** | Ventana bitwise con aislamiento por DID y sesión. |
| **Presupuesto MTU 1280B** | Resiliencia de Transporte | **HECHO / TESTEADO** | Límite matemático estricto $\forall \text{ seq} \in [0, 2^{64}-1]$. |
| **UPnP Anti-SSRF** | Mapeo Automatizado de Puertos | **HECHO / TESTEADO** | Validación rigurosa RFC 1918 / loopback y bloqueo de redirects. |
| **Actualizador Seguro** | Gestión de Ciclo de Vida | **HECHO / TESTEADO** | SemVer anti-rollback + SHA-256 + firma Ed25519. |
| **X-Wing KEM** | Híbrido ML-KEM + X25519 | **EXPERIMENTAL** | Compatible con `draft-ietf-cfrg-xwing` (combiner SHA3-256). |
| **Vector Reticular de Firma** | Compromiso HMAC | **EXPERIMENTAL** | Vector complementario determinista (NO es FIPS 204). |
| **ML-DSA (FIPS 204)** | Firmas Post-Cuánticas | **NO IMPLEMENTADO** | Planificado conceptualmente; no presente en v0.7.0. |
| **Instalador Windows** | Despliegue de Sistema | **HECHO (Arquitectura B)** | Desacoplado sin `go:embed` en el árbol Git; firewall UDP 7777. |
| **Daemon Supervisor** | Autogobernanza Persistente | **HECHO / TESTEADO** | Orquestación reproducible en `sistema/bin/` (14/14 tests PASS). |
| **Red Física WAN / NAT Real** | Topología Distribuida | **PENDIENTE DE AUDITORÍA FÍSICA** | Validado en localhost y CI; pendiente prueba física multi-sitio. |

---

## ⚡ Los 7 Invariantes del Núcleo I7

1. **Axioma I — El Core solo transporta estructura, jamás semántica:** Cero lógica de aplicación dentro del camino crítico de paquetes.
2. **Axioma II — Las 10 Primitivas Nucleares Inmutables:** `Identity`, `Packet`, `Container`, `Object`, `Session`, `Channel`, `Integrity`, `AntiReplay`, `MTU`, `Routing`.
3. **Axioma III — Límite de 400 Líneas por Archivo:** Modularidad estricta para garantizar mantenibilidad y legibilidad humana y sintética.
4. **Axioma IV — Memoria Zero-Copy (0 B/op):** El pipeline lineal recicla búferes en ring-buffers prealocados (18.53 ns/op, 0 alocaciones de heap).
5. **Axioma V — MTU Determinista de 1280 Bytes:** Diseñado para interoperar sin fragmentación sobre cualquier túnel o enlace IPv6 (RFC 8200).
6. **Axioma VI — ZTNA Estricto (Default-Deny Bidireccional):** Ningún paquete se procesa sin autorización explícita previa en el firewall local.
7. **Axioma VII — Enrutador Kleinberg Canónico de 16 Anillos:** Búsqueda métrica XOR logarítmica determinista en tiempo $O(\log N)$.

---

## 🚀 Inicio Rápido

### Requisitos Previos
* **Go:** 1.24+ con soporte para `crypto/mlkem`.
* **PowerShell:** 7+ (para ejecución de suites en Windows/Linux).
* **Python:** 3.10+ (para módulos de gobernanza y supervisión en `sistema/`).

### 1. Compilación del Nodo IPVN7
```bash
cd src
go build -trimpath -ldflags="-s -w" -o ../bin/ipvn7.exe ./cmd/ipvn7
```

### 2. Ejecución Local (Espacio de Usuario / Zero-Admin)
```bash
# Lanzamiento en modo estándar con WebUI en localhost:7070 y SOCKS5 en 10807
.\bin\ipvn7.exe -no-elevate
```
Abre tu navegador en: [http://127.0.0.1:7070](http://127.0.0.1:7070) para interactuar con la WebUI local.

### 3. Compilación del Instalador Desacoplado (Arquitectura B)
```powershell
powershell -ExecutionPolicy Bypass -File scripts/build_installer.ps1
```
Genera `dist/Instalador_VPN_I7.exe` listo para distribución independiente.

---

## 🛡️ Verificación de Calidad y Compuertas de Paso

IPVN7 cuenta con una suite integral de verificación que evalúa estáticamente y dinámicamente todo el repositorio:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/verify_ipvn7_standard.ps1
```

La suite certifica de forma automática:
- Invariante estricto de 400 líneas por archivo.
- Análisis estático con `go vet ./...`.
- Detección exhaustiva de carreras con `-race`.
- Rendimiento Zero-Copy (0 B/op en `BenchmarkLinearPipeline_Execute`).
- Batería de Fuzzing determinista de paquetes y fronteras de secuencia.
- Pruebas unitarias de todos los paquetes (100% PASS).
- Compilación fresca y validada de binarios de producción.

---

## 📚 Documentación Canónica

* 📜 [FUENTE_DE_VERDAD.md](docs/FUENTE_DE_VERDAD.md) — Norma técnica suprema y compendio de invariantes.
* ⚖️ [CONSTITUCION.md](sistema/CONSTITUCION.md) — Ley fundamental del ecosistema de autogobernanza.
* 🎯 [INTENCION.md](frondabrick_01/INTENCION.md) — Bitácora operativa en tiempo real del agente rector FrondaBrick_01.
* 🔬 [COMPENDIO_INVESTIGACION_PQC_Y_REDES.md](docs/research/COMPENDIO_INVESTIGACION_PQC_Y_REDES.md) — Investigación tecnológica sobre PQC, NAT y datapath.

---

## 📄 Licencia

Este proyecto se distribuye bajo los términos de la Licencia MIT. Consulta [LICENSE](LICENSE) para más detalles.
