# Plan Maestro de Alineación y Purga de Realidad con la Auditoría Externa

> **"Eliminar el teatro de simulación, desacoplar el exceso y consolidar el Núcleo Mínimo I7 con verificación física demostrable."**
> Basado estrictamente en los hallazgos de [`auditoria externa.md`](../auditoria%20externa.md).

---

## Estado Global del Plan

- [x] **Fase 0: Resoluciones Críticas P0 Inmediatas**
  - [x] **P0.1 Criptografía PQC Real:** Eliminar simulación HMAC de ML-KEM-768 en `src/pkg/l0/pqc_kem.go` e implementar NIST FIPS 203 real con `crypto/mlkem` oficial de Go 1.24/1.26.
  - [x] **P0.2 Pipeline de Release:** Reconstruir `.github/workflows/release.yml` para compilar dentro de `src/cmd/ipvn7` y versionar `v0.7.0`.
  - [x] **P0.3 Taxonomía Factual de 4 Estados:** Formalizar en `AGENTS.md` (Regla 1) los 4 estados: *DEMOSTRADO*, *IMPLEMENTADO*, *EXPERIMENTAL*, *NO DEMOSTRADO*.

- [x] **Fase 1: Correcciones de Robustez y Criptografía en L1 (P1)**
  - [x] **P1.1 BufferPool sin Panic:** Corregir `BufferPool.Acquire` en `src/pkg/l1/buffer_pool.go` para tramas $>64$ KB (Jumbo) sin desbordar memoria ni lanzar panic de slice bounds.
  - [x] **P1.2 Sincronización de Telemetría:** Añadir `sync.RWMutex` a las entradas de `TelemetryRingBuffer` en `src/pkg/l2/telemetry.go` y eliminar afirmación no contextualizada "<28 ns".
  - [x] **P1.3 Handshake 1-RTT Post-Cuántica Real:** Integrar `crypto/mlkem` real en `src/pkg/l1/pqc_handshake.go` utilizando `Encapsulate()` con criptograma completo de 1088 bytes en lugar del fallback simulado.
  - [x] **P1.4 NAT Hole Punching con Confirmación Real:** Eliminar falso éxito al solo emitir UDP en `src/pkg/l1/nat_traversal.go`. Exigir recepción de datagrama de confirmación bidireccional (`PUNCH_ACK`) antes de declarar éxito.
  - [x] **P1.5 Forwarding Físico en SmartComponentGateway:** Conectar `SendDatagram()` en `src/pkg/core/gateway.go` con un transporte físico real mediante UDP socket `WriteToUDP`, verificado con tests reales de socket a socket.

- [x] **Fase 2: Delimitación de Kademlia, Pacing y Núcleo Mínimo I7 (P1/P2)**
  - [x] **P2.1 Claridad Factual en Kademlia:** Documentar en `src/pkg/l2/dht_kademlia.go` que opera como tabla de enrutamiento XOR / K-bucket local y añadir manejo de cola de reemplazo ante buckets llenos.
  - [x] **P2.2 Honestidad en Pacing Token Bucket:** Clarificar en `src/pkg/l1/pacing.go` que el cuello de botella es un parámetro configurado / estimado preliminar y no un estimador BBR automático demostrado.
  - [x] **P2.3 Desacoplar Núcleo Mínimo I7 de Satélites:** Formalizar en DEC-137 la arquitectura estricta (Core: Identity, Object, Session, Channel, Path, MTU, Integrity vs Adapters: UDP/TUN vs Plugins: DNS/SOCKS5/WebUI).

- [x] **Fase 3: Verificación Física Extremo a Extremo (Loop Closure)**
  - [x] **P3.1 Test Físico Bidireccional A -> I7 -> UDP -> I7 -> B:** Crear suite de prueba de transporte físico loopback bidireccional (`TestPQC_PhysicalUDPLoopback_Bidirectional`) con aserción bit a bit de carga útil cifrada con PQC FIPS 203 nativo y ChaCha20-Poly1305.
  - [x] **P3.2 Ejecución Limpia de la Compuerta Universal:** Pasar `scripts/verify_ipvn7_standard.ps1` con 100% de tests unitarios, concurrencia e invariante zero-copy.
