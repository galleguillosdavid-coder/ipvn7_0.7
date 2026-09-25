# Registro Histórico de Decisiones de Arquitectura (ADR v0.1 a v0.6)

Archivo de memoria histórica inmutable de IPVN7 Network OS para las decisiones iniciales DEC-001 a DEC-013.
Para las decisiones actuales vigentes de la reingeniería v0.7 en adelante, consultar [`07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md`](./07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md).

---

## Tabla Histórica (DEC-001 a DEC-013)

| ID | Fecha | Título Breve | Axioma Vinculado | Estado |
| :--- | :--- | :--- | :--- | :--- |
| **DEC-001** | 2026-09-16 | Erradicación de Mocks y Simulaciones Artificiales | `Código Real` | **INVIOLABLE** |
| **DEC-002** | 2026-09-16 | Medición Empírica en Vivo en lugar de Constantes Sintéticas | `Fast-Path-First` | **INVIOLABLE** |
| **DEC-003** | 2026-09-16 | Erradicación Total de PII y Hardcoding de Entorno | `Zero-PII` | **INVIOLABLE** |
| **DEC-004** | 2026-09-16 | Higiene de Concurrencia con `context.Context` y Timeouts de Red | `Axioma Concurrencia` | **INVIOLABLE** |
| **DEC-005** | 2026-09-16 | Regla Estricta de Modularidad Atómica ($\le$ 400 líneas) | `Axioma Modularidad` | **INVIOLABLE** |
| **DEC-006** | 2026-09-16 | Desacoplamiento Modular de Auditorías y Diagnósticos | `Living-Lab` / `ALC` | **INVIOLABLE** |
| **DEC-007** | 2026-09-16 | Interoperabilidad Multiplataforma Windows, Linux y macOS | `Autonomía Local` | **INVIOLABLE** |
| **DEC-008** | 2026-09-16 | Erradicación Definitiva de Residuos Sintéticos y Certificación Real | `Código Real` | **INVIOLABLE** |
| **DEC-009** | 2026-09-16 | Auditoría Integral y Sustitución Definitiva de Mocks por Código Real | `Living-Lab` | **INVIOLABLE** |
| **DEC-010** | 2026-09-17 | Reingeniería v0.6: Núcleo Universal y Smart Component Gateway | `Arquitectura v0.6` | **INVIOLABLE** |
| **DEC-011** | 2026-09-17 | Erradicación del Teatro de Simulación y Topología Limpia 1-a-1 | `Realismo Físico` | **INVIOLABLE** |
| **DEC-012** | 2026-09-17 | Modularidad Frontend Obligatoria y Límite de 400 Líneas | `Axioma Modularidad` | **INVIOLABLE** |
| **DEC-013** | 2026-09-21 | Línea de Producción (Pipes & Filters) y FSM Determinista | `Pipes & Filters` | **INVIOLABLE** |

---

## Detalle Histórico DEC-001 a DEC-013

### DEC-001: Erradicación de Mocks y Simulaciones Artificiales
* Se implementa `UserspaceVirtualAdapter` y el enum tipado `AdapterMode` (`ModeKernelVirtual` vs `ModeUserspaceVirtual`), garantizando tráfico real sin privilegios de kernel.

### DEC-002: Medición Empírica en Vivo en lugar de Constantes Sintéticas
* Toda métrica es calculada en nanosegundos empíricos (`time.Now().UnixNano()`) con EWMA. Prohibido hardcodear latencias.

### DEC-003: Erradicación Total de PII y Hardcoding de Entorno
* Ninguna IP local real ni credenciales hardcodeadas; uso exclusivo de RFC 5737 TEST-NET-2 (`198.51.100.0/24`) o `127.0.0.1`.

### DEC-004: Higiene de Concurrencia con `context.Context` y Timeouts de Red
* Toda goroutine de servicio recibe `ctx context.Context` y sockets configuran `SetReadDeadline` para apagado limpio ante SIGINT/SIGTERM.

### DEC-005: Regla Estricta de Modularidad Atómica ($\le$ 400 líneas)
* Límite estricto e infranqueable de 400 líneas por archivo en todo el repositorio.

### DEC-006: Desacoplamiento Modular de Auditorías y Diagnósticos
* Consulta obligatoria de este registro de arquitectura antes de cualquier refactorización estructural.

### DEC-007: Interoperabilidad Multiplataforma Windows, Linux y macOS
* Binarios compilables en Pure Go (`CGO_ENABLED=0`) con paridad multiplataforma nativa.

### DEC-008: Erradicación Definitiva de Residuos Sintéticos y Certificación Real
* Purga de prefijos `*simulate*` y `*synthetic*` en código operativo.

### DEC-009: Auditoría Integral y Sustitución Definitiva de Mocks por Código Real
* La UI orbital se alimenta directamente de los pares reportados por el enrutador en `/api/v1/peers`.

### DEC-010: Reingeniería v0.6: Núcleo Universal y Smart Component Gateway
* Desacople del demonio central (`ipvn7`) enfocado exclusivamente en L0-L2 y bus `/api/v1/*` para servicios satelitales.

### DEC-011: Erradicación del Teatro de Simulación y Topología Limpia 1-a-1
* Tráfico físico real HTTP/UDP con entrega garantizada y purga estricta de self-peering y nodos fantasma.

### DEC-012: Modularidad Frontend Obligatoria y Límite Universal de 400 Líneas
* Descomposición atómica de `web/app.js` en submódulos de dominio específico (`app_core.js`, `app_net.js`, `app_chat.js`, etc.).

### DEC-013: Paradigma de Línea de Producción (Pipes & Filters) y FSM Determinista
* Estructuración secuencial unidireccional de ingesta de datagramas (`LinearPipeline`) gobernada por autómata de estados finitos (`DeterministicNodeFSM`).

### DEC-014: Despacho de Pipeline Zero-Copy Copy-On-Write (0 Allocs/op)
* Paso de datagramas por referencia en hot-path sin duplicaciones de buffers.

### DEC-015: Pooling Reciclable de PacketContext en Ingesta de Sockets
* Reutilización de contextos con `sync.Pool` eliminando presión sobre el GC.

### DEC-016: Ventana Anti-Replay de 1024 Bits Multi-Palabra (RFC 4303)
* Bitmask deslizante de 16 palabras `uint64` inmune a desorden y retransmisiones maliciosas.

### DEC-017: Integridad de Datagramas en Tiempo Constante (BLAKE2s SIMD)
* Verificación criptográfica sin fugas por canales laterales de tiempo.

### DEC-018: Handshake Post-Cuántica Híbrido en 1 RTT (ML-KEM-768 + X25519)
* Negociación criptográfica resistente a ordenadores cuánticos en un único intercambio.

### DEC-019: Modularización Sphinx y Mitigación Anti-DPI mediante Padding Estocástico
* Tramas fijas de 1280B con relleno aleatorio para derrotar análisis estadístico de tráfico.

### DEC-020: Protocolo Gossip Descentralizado para Difusión de BindingRecords UIN
* Propagación epidémica asíncrona de identidades soberanas sin servidor central.

### DEC-021: Perforación NAT P2P Simétrica y Rendezvous Directo ICE-Lite
* Hole punching UDP coordinado con fallback a relays soberanos.

### DEC-022: VPN Corporativa Fricción Cero y Camuflaje TLS 1.3
* SOCKS5 `:10807` y HTTP CONNECT `:10808` en espacio de usuario.

### DEC-023: Gobernanza del Árbitro de Memoria contra DoS (2.8M pps)
* Presupuestos rígidos de memoria con descarte determinista O(1).

### DEC-024: Especificación Formal RFC Canónica para el Protocolo IPVN7-CORE
* Publicación de `docs/rfc/RFC_IPVN7_CORE.md` (156L) con formato estándar IETF.

### DEC-025: Especificación Formal RFC Canónica para el Protocolo IPVN7-SPHINX
* Publicación de `docs/rfc/RFC_IPVN7_SPHINX.md` (120L) bajo estándar IETF.

