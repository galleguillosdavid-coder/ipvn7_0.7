# Archivo Histórico de Decisiones Arquitectónicas (ADR - DEC-014 a DEC-035)

Este documento preserva formalmente las decisiones de arquitectura de IPVN7 Network OS correspondientes a las fases de optimización Zero-Copy, Handshake PQC, Gobernanza de Memoria y Estandarización Universal (v0.6 a v0.7), segregadas de [`docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md`](./07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md) para garantizar el cumplimiento universal del Axioma III ($\le$ 400 líneas por archivo).

---

### DEC-014: Despacho de Pipeline Zero-Copy con Copy-On-Write (0 Allocs/op)
* **Fecha:** 2026-09-22
* **Contexto:** En el bucle de alta frecuencia de `LinearPipeline.Execute`, se realizaba una copia defensiva asignando un nuevo slice en el heap (`make([]interfaces.PipelineStage, len(p.stages))`) por cada paquete individual entrante.
* **Error / Anti-Patrón Prohibido:** Asignar memoria dinámica en el heap (allocations) en la ruta crítica de procesamiento y retransmisión de datagramas.
* **Decisión Adoptada:** Implementar semántica Copy-On-Write (COW) en `AddStage`, garantizando que la lista de estaciones sea inmutable durante la ejecución de datagramas. Leer la referencia inmutable bajo `p.mu.RLock()` en `Execute` sin clonar slices.

### DEC-015: Pooling Reciclable de PacketContext (Zero Allocations en Ingesta de Socket)
* **Fecha:** 2026-09-22
* **Contexto:** En el bucle de recepción de sockets UDP, cada datagrama recibido instanciaba dinámicamente un contexto en el heap, saturando el recolector de basura.
* **Error / Anti-Patrón Prohibido:** Instanciar estructuras de contexto de paquete con `new` o literales en heap dentro del bucle de escucha de sockets.
* **Decisión Adoptada:** Implementar `PacketContext.Reset()` para reinicialización en $O(1)$ y `sync.Pool` en `pkg/core/pipeline.go`. Benchmark: 25.48 ns/op, 0 B/op y 0 allocs/op.

### DEC-016: Ventana Anti-Replay de 1024 Bits Multi-Palabra (Tolerancia a Jitter Extremo)
* **Fecha:** 2026-09-22
* **Contexto:** Ventana reducida de 64 posiciones provocaba descartes falsos de paquetes desordenados bajo jitter WAN elevado.
* **Error / Anti-Patrón Prohibido:** Emplear ventanas anti-repetición reducidas en un sistema operativo de redes global.
* **Decisión Adoptada:** Ventana deslizante de 1024 bits (`[16]uint64`) RFC 4303 lock-free mediante `shiftLeft`. Benchmark: 34.64 ns/op, 0 allocs/op.

### DEC-017: Integridad de Datagramas en Tiempo Constante mediante Checksum SIMD (BLAKE2s)
* **Fecha:** 2026-09-22
* **Contexto:** Paquetes corruptos forzaban el parseo CBOR y validación de firmas asimétricas antes de validar integridad física.
* **Error / Anti-Patrón Prohibido:** Deserializar estructuras complejas sobre datagramas cuya integridad física no ha sido comprobada en $O(1)$.
* **Decisión Adoptada:** `FastPacketChecksum` y `VerifyPacketChecksum` en `pkg/l0/wire.go` con BLAKE2s SIMD y tiempo constante. Benchmark: 2593 ns/op, 493.65 MB/s, 0 allocs/op.

### DEC-018: Handshake Post-Cuántica Híbrido en 1 RTT (ML-KEM-768 + X25519)
* **Fecha:** 2026-09-22
* **Contexto:** Noise tradicional requería 1.5 RTT y dependía únicamente de curvas elípticas clásicas.
* **Error / Anti-Patrón Prohibido:** Negociar claves de sesión sin protección contra ataques cuánticos futuros.
* **Decisión Adoptada:** Handshake híbrido en 1 RTT en `pkg/l1/pqc_handshake.go` combinando ML-KEM-768 y X25519. Benchmark: 251 μs por handshake completo.

### DEC-019: Modularización de Sphinx y Mitigación Anti-DPI mediante Padding Estocástico
* **Fecha:** 2026-09-22
* **Contexto:** `pkg/l1/sphinx_onion.go` alcanzaba 371 líneas y requería generador de tráfico dummy para mitigar análisis de tráfico (DPI).
* **Error / Anti-Patrón Prohibido:** Monolitos > 400 líneas y flujos sin protección contra correlación temporal.
* **Decisión Adoptada:** Modularización en `sphinx_builder.go` y motor `StochasticPaddingEngine` en `sphinx_padding.go`.

### DEC-020: Protocolo Gossip Descentralizado para Difusión de BindingRecords UIN
* **Fecha:** 2026-09-22
* **Contexto:** Necesidad de propagación de identidades delegadas sin servidores centrales.
* **Error / Anti-Patrón Prohibido:** Dependencia de directorios centralizados para propagación de claves.
* **Decisión Adoptada:** `UINGossipManager` en `pkg/l1/uin_gossip.go` con deduplicación en $O(1)$ y fanout aleatorizado hacia anillos Kleinberg.

### DEC-021: Perforación NAT P2P Simétrica y Rendezvous Directo ICE-Lite
* **Fecha:** 2026-09-22
* **Contexto:** Aislamiento de nodos tras NAT simétrica sin dependencias de STUN/TURN de terceros propietarios.
* **Error / Anti-Patrón Prohibido:** Depender de infraestructura propietaria para rendezvous.
* **Decisión Adoptada:** `NATHolePunchEngine` con mensajes binarios canónicos (`IP7H`).

### DEC-022: VPN Corporativa Fricción Cero Multi-OS y Camuflaje TLS 1.3 (RFC 8446)
* **Fecha:** 2026-09-22
* **Contexto:** Sistemas sin privilegios de kernel e inspección profunda (DPI) corporativa.
* **Error / Anti-Patrón Prohibido:** Exigir root para operar la VPN corporativa o enviar tráfico opaco sin camuflaje RFC.
* **Decisión Adoptada:** Conmutación automática a `ModeUserspaceProxy` (SOCKS5/HTTP CONNECT) y camuflaje con `TLSOptionEngine`.

### DEC-023: Gobernanza del Árbitro de Memoria y Resistencia a Inundación DoS (2.8M pps)
* **Fecha:** 2026-09-22
* **Contexto:** Peligro de OOM-killer por ataques de agotamiento de memoria.
* **Error / Anti-Patrón Prohibido:** Asignación indiscriminada en heap sin control de cuotas rígidas.
* **Decisión Adoptada:** `GlobalMemoryArbiter` con particionamiento de RAM por clases (2.84M pps sostenidos).

### DEC-026: SDK Soberano en Rust para el Smart Component Gateway
* **Fecha:** 2026-09-22
* **Contexto:** Interoperabilidad con ecosistemas nativos de alto rendimiento y WASM.
* **Error / Anti-Patrón Prohibido:** Desarrollar clientes con dependencias pesadas sin validación estricta de tramas de 1280B.
* **Decisión Adoptada:** Implementar `sdk/rust/` con `Ipvn7Client`, Serde completo, validación de MTU 1280B y cero bloatware (119L).

### DEC-027: Ampliación de Herramientas MCP para Auditoría PQC, Sphinx y Árbitro de Memoria
* **Fecha:** 2026-09-22
* **Contexto:** Agentes IA requerían visibilidad sobre criptografía cuántica, anonimato y cuotas de RAM.
* **Error / Anti-Patrón Prohibido:** Acumular lógica en monolitos o expandir `mcp_tools_execution.go` sobre 400L.
* **Decisión Adoptada:** Crear `pkg/l3/mcp_tools_pqc_security.go` (63L) con herramientas `ipvn7_pqc_status`, `ipvn7_sphinx_status` e `ipvn7_memory_arbiter_status`.

### DEC-028: Suite de Micro-Benchmarks Empíricos y Reporte Comparativo
* **Fecha:** 2026-09-22
* **Contexto:** Validación formal del estándar frente a IPv4/IPv6 sin simulaciones.
* **Error / Anti-Patrón Prohibido:** Afirmar superioridad sin benchmarks instrumentados en nanosegundos reales.
* **Decisión Adoptada:** Instrumentar `scripts/benchmark_comparative.go` y generar informe en `docs/BENCHMARKS.md` (38L).

### DEC-029: Motor de Transporte Multiplexado QUIC/UDP para Enlaces Inestables
* **Fecha:** 2026-09-22
* **Contexto:** Bloqueo de cabeza de línea en mallas P2P móviles o satelitales con alta pérdida.
* **Error / Anti-Patrón Prohibido:** Dependencias pesadas externas o exceder MTU canónico de 1280 octetos.
* **Decisión Adoptada:** `QUICTransportEngine` en `pkg/l1/quic_transport.go` (173L) con SACK de 64 bits.

### DEC-030: Certificación de Verificación Física en Laboratorio de 2 Nodos
* **Fecha:** 2026-09-22
* **Contexto:** Certificar en hardware real la comunicación entre Nodo A (`192.168.1.198`) y Nodo B (`192.168.1.106`).
* **Error / Anti-Patrón Prohibido:** Simular conectividad con bucles locales o tolerar nodos fantasma.
* **Decisión Adoptada:** Emparejamiento recíproco real con entrega confirmada A <-> B (1.1 ms) y listas 1-a-1 sin fantasmas.

### DEC-031: Centinela de Auto-Reparación de Enlaces WAN y Failover O(1) Kleinberg
* **Fecha:** 2026-09-22
* **Contexto:** Fluctuaciones de enlace y degradación de ISP antes de expirar por timeout.
* **Error / Anti-Patrón Prohibido:** Esperar pasivamente desconexión total o alocaciones de heap al conmutar pares.
* **Decisión Adoptada:** `LinkHealingEngine` en `pkg/l1/link_healing.go` (133L) con failover $O(1)$ tras 3 pérdidas o latencia >500ms.

### DEC-032: Empaquetado Binario Reproducible y Contenedor Scratch sin Privilegios
* **Fecha:** 2026-09-22
* **Contexto:** Despliegue en Kubernetes y edge sin dependencias de SO.
* **Error / Anti-Patrón Prohibido:** Imágenes base pesadas con CVEs o ejecución como root.
* **Decisión Adoptada:** Compilación multi-etapa en `Dockerfile` sobre imagen vacía `scratch` con usuario sin privilegios `10007:10007`.

### DEC-033: Integración del Centinela de Auto-Reparación en el Gateway
* **Fecha:** 2026-09-22
* **Contexto:** `SmartComponentGateway.SendDatagram` requería conmutación dinámica ante degradación.
* **Error / Anti-Patrón Prohibido:** Desconectar centinelas de salud de los puntos de entrada/salida de paquetes.
* **Decisión Adoptada:** Integrar `healing *l1.LinkHealingEngine` en el gateway con conmutación en $O(1)$ hacia pares saludables.

### DEC-034: Auditoría Dinámica de Enrutamiento Kleinberg y Saturación de Anillos
* **Fecha:** 2026-09-22
* **Contexto:** Topología elástica bajo saturación masiva y contienda concurrente.
* **Error / Anti-Patrón Prohibido:** Degradación a búsqueda lineal $O(N)$ o deadlocks en roaming concurrente.
* **Decisión Adoptada:** Suite de estrés en `pkg/l1/routing_stress_test.go` (116L) con 100 identidades y 50 goroutines concurrentes.

### DEC-035: Intercambio de Claves Híbrido Post-Cuántica ML-KEM-768 + X25519
* **Fecha:** 2026-09-22
* **Contexto:** Blindaje de sesiones L0/L1 contra ataques prospectivos de computación cuántica.
* **Error / Anti-Patrón Prohibido:** CGO opaco o descartar la criptografía clásica comprobada.
* **Decisión Adoptada:** `MLKEM768Adapter` en `pkg/l0/pqc_kem.go` (172L) y derivación híbrida determinista `DeriveHybridSecret`.

### DEC-036: Formalización de Directivas de Cerrojo Mutex, Invariante Zero-Copy y Compuerta de Paso
* **Fecha:** 2026-09-22
* **Contexto:** Tras múltiples ciclos sostenidos de autoejecución y verificación continua, era necesario codificar normativamente las lecciones empíricas aprendidas para prevenir colisiones asíncronas, regresiones de memoria y acumulación excesiva de líneas.
* **Error / Anti-Patrón Prohibido:** Lanzar ciclos paralelos sin cerrojo (`.agents/task.lock`), alocaciones en canalización L0-L2 o dar tareas por concluidas sin la compuerta.
* **Decisión Adoptada:** Enriquecer directivas con Secciones 9 (Cerrojo), 10 (Poda ADR a 320L), 11 (Invariante zero-copy 0 B/op) y 12 (Compuerta `verify_ipvn7_standard.ps1`).

### DEC-037: Enrutador de Interfaz Virtual de Red (TUNRouter) y Mapeo Determinista IP-DID
* **Fecha:** 2026-09-22
* **Contexto:** Las aplicaciones existentes del sistema operativo requerían interoperar de forma transparente con IPVN7 mediante sockets estándar IPv4/IPv6 sin reescribir software.
* **Error / Anti-Patrón Prohibido:** Forzar a todas las aplicaciones cliente a integrar librerías de SDK o utilizar exclusivamente endpoints HTTP/WebSocket para enviar tráfico de red.
* **Decisión Adoptada:** Implementar `TUNRouter` en `pkg/l1/tun_router.go` (152L) con enrutamiento bidireccional y resolución determinista `ResolveIPToDID` para tramas IPv4 `10.7.0.0/16` e IPv6 `fd07::/64`.

### DEC-038: Motor de Traversal NAT Avanzado y Relays Cifrados Soberanos (DERP)
* **Fecha:** 2026-09-22
* **Contexto:** Enlaces entre nodos residenciales y corporativos bloqueados por CGNAT o topologías de NAT simétrico donde el rendezvous tradicional falla.
* **Error / Anti-Patrón Prohibido:** Depender de servidores STUN/TURN de terceros o abortar la conexión ante escenarios de NAT simétrica doble.
* **Decisión Adoptada:** Implementar `NATTraversalEngine` en `pkg/l1/nat_traversal.go` con clasificación RFC 3489/5389, hole-punching UDP coordinado y fallback transparente a relays soberanos cifrados extremo a extremo (`RouteViaDERPRelay`).

### DEC-039: Grafo Topológico Distribuido y Convergencia Determinista mediante CRDTs
* **Fecha:** 2026-09-22
* **Contexto:** La topología de red indexada en memoria local requería propagarse de manera distribuida y asíncrona entre todos los nodos sin un servidor centralizado ni consensos pesados (Raft/Paxos).
* **Error / Anti-Patrón Prohibido:** Forzar consenso síncrono bloqueante o tolerar inconsistencias/bifurcaciones de estado topológico en particiones de red.
* **Decisión Adoptada:** Implementar `CRDTTopologyGraph` en `pkg/l2/crdt_graph_sync.go` con semántica LWW (Last-Write-Wins), relojes lógicos Lamport y tumbas (tombstones) para borrado de aristas y nodos.

### DEC-040: Especificación y Simulador de Bypass de Red eBPF / XDP para Conmutación Wire-Speed
* **Fecha:** 2026-09-22
* **Contexto:** Para entornos de backbone y centros de datos de 40/100 GbE, el procesamiento de datagramas IPVN7 en userspace genera sobrecarga de cambio de contexto del kernel.
* **Error / Anti-Patrón Prohibido:** Quedar atado a la pila de red tradicional del sistema operativo en backbones de ultra-alta velocidad.
* **Decisión Adoptada:** Publicar especificación formal en `docs/rfc/RFC_IPVN7_XDP_ACCELERATION.md` con código C canónico eBPF/XDP y cargador en `pkg/l0/ebpf_xdp_spec.go`. Descarte de datagramas corruptos a velocidad de línea en el anillo de la NIC.

### DEC-041: SDKs Universales Multi-Lenguaje: Python (`ipvn7`) y TypeScript/WASM
* **Fecha:** 2026-09-22
* **Contexto:** Frameworks de agentes de IA y aplicaciones web/Node.js requerían conectividad de primera clase a la malla sin dependencias CGO complejas.
* **Error / Anti-Patrón Prohibido:** Limitar la integración a lenguajes de sistemas, dejando aislados a los ecosistemas dominantes de IA y frontend.
* **Decisión Adoptada:** Desarrollar SDK Python en `sdk/python/` (con validación canónica de MTU 1280B y empaquetado pip `pyproject.toml`) y SDK TypeScript en `sdk/typescript/`.

### DEC-042: Modularización Proactiva y Poda de Monolitos de Alta Densidad (Algoritmo de 5 Pasos)
* **Fecha:** 2026-09-22
* **Contexto:** Módulos centrales en L3 y frontend se encontraban en el rango de 370-399 líneas, al borde del límite estricto de 400 líneas (Axioma III universal).
* **Error / Anti-Patrón Prohibido:** Permitir que los archivos crezcan acumulando responsabilidades dispares.
* **Decisión Adoptada:** Aplicar poda y modularización atómica en `mcp_tools_governance.go`, `computational_law.go`, `agent_senate.go`, `app_kuzu_physics.js` e `index.html`.

### DEC-043: Interfaz Soberana Hero Zen (Botón Tricolor P2P/Internet) y Panel Desacoplado
* **Fecha:** 2026-09-22
* **Contexto:** Simplificación radical a un solo control circular táctil con gradientes de textura reflejando estado de protección de Internet (Verde/Amarillo/Rojo).
* **Error / Anti-Patrón Prohibido:** Interfaces sobrecargadas como pantalla inicial o simulación sin socket real de proxy.
* **Decisión Adoptada:** Botón circular Hero Zen en `web/style_hero_button.css` y `web/app_vpn_button.js`, endpoints en Go `/api/v1/vpn/status` y `/api/v1/vpn/toggle` gobernando SOCKS5 (`127.0.0.1:10807`).

### DEC-044: Integración UPnP IGD Port Forwarding y Herramienta Empírica de Benchmark RFC 3550
* **Fecha:** 2026-09-22
* **Contexto:** Mapeo automático de puertos en routers residenciales (UPnP IGD vía SSDP + SOAP) y benchmark empírico de saturación con cálculo de jitter RFC 3550.
* **Error / Anti-Patrón Prohibido:** Mantener `time.Sleep` simulando sondeos de red o afirmar throughput sin herramienta nativa CLI.
* **Decisión Adoptada:** Implementar `UPnPMapper` en `pkg/l1/nat_upnp.go`, sustituir simulación con ráfagas físicas de datagramas UDP, y crear `pkg/core/benchmark_runner.go` y `cmd/ipvn7-cli/cli_bench.go`.

### DEC-045: UIN Testamento Criptográfico / Sucesión S y Modos Tri-State Operacionales
* **Fecha:** 2026-09-22
* **Contexto:** En caso de caída de un nodo soberano, la identidad UIN requería un mecanismo de sucesión criptográfica verificable sin servidor central.
* **Error / Anti-Patrón Prohibido:** Permitir suplantación de identidad o depender de CAs centralizadas para transferir la clave UIN.
* **Decisión Adoptada:** Implementar `UINWill` en `pkg/l1/uin_succession.go` con firmas Ed25519 y `OperatingModeManager` en `pkg/core/operating_modes.go`.
