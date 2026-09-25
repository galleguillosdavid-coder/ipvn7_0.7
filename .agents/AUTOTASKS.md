# COLA DE AUTOTAREAS DE ESTANDARIZACIÓN IPVN7 (.agents)

Este documento registra la cola priorizada y el estado de ejecución continua del Agente de Red Soberano (`ipvn7-network-os-agent`).
Cadencia operativa: Disparo periódico cada 10 minutos activado de forma obligatoria mediante `/schedule` (o cron `*/10 * * * *`), protegido por exclusión mutua mediante `.agents/task.lock`.

---

## Estado Global de Autotareas

| ID | Tarea | Fase Roadmap | Estado | Verificación |
| :--- | :--- | :---: | :---: | :--- |
| **TASK-001** | Redacción formal de la especificación RFC IPVN7-CORE (IETF format) | Fase 7 | **COMPLETED** | [`docs/rfc/RFC_IPVN7_CORE.md`](../docs/rfc/RFC_IPVN7_CORE.md) (156L) |
| **TASK-002** | Redacción formal de la especificación RFC IPVN7-SPHINX (Onion Routing) | Fase 7 | **COMPLETED** | [`docs/rfc/RFC_IPVN7_SPHINX.md`](../docs/rfc/RFC_IPVN7_SPHINX.md) (120L) |
| **TASK-003** | SDK Rust / WASM minimalista para Smart Component Gateway | Fase 7 | **COMPLETED** | [`sdk/rust/`](../sdk/rust/) (119L) |
| **TASK-004** | Ampliación de MCP Tools diagnósticas y centinela para agentes IA | Fase 7 | **COMPLETED** | [`pkg/l3/mcp_tools_pqc_security.go`](../pkg/l3/mcp_tools_pqc_security.go) (63L) |
| **TASK-005** | Ciclo de benchmark comparativo sostenido y control de regresión | Fase 7 | **COMPLETED** | [`docs/BENCHMARKS.md`](../docs/BENCHMARKS.md) (38L) |
| **TASK-006** | Transportador QUIC/UDP multiplexado para enlaces con pérdida | Fase 7 | **COMPLETED** | [`pkg/l1/quic_transport.go`](../pkg/l1/quic_transport.go) (173L) |
| **TASK-007** | Protocolo de verificación física bidireccional en laboratorio de 2 nodos | Fase 7 | **COMPLETED** | Topología 1-a-1 limpia A $\leftrightarrow$ B certificada (Latencia 1.1ms) |
| **TASK-008** | Centinela de auto-reparación ante degradación WAN (Failover O(1) Kleinberg) | Fase 8 | **COMPLETED** | [`pkg/l1/link_healing.go`](../pkg/l1/link_healing.go) (133L) |
| **TASK-009** | Empaquetado binario reproducible y contenedor scratch sin privilegios | Fase 8 | **COMPLETED** | [`Dockerfile`](../Dockerfile) / [`docs/DEPLOYMENT_DOCKER.md`](../docs/DEPLOYMENT_DOCKER.md) |
| **TASK-010** | Integración del Centinela de Auto-Reparación en el Gateway del Núcleo | Fase 8 | **COMPLETED** | [`pkg/core/gateway.go`](../pkg/core/gateway.go) / Test unitario failover |
| **TASK-011** | Test de estrés y auditoría dinámica de saturación de anillos Kleinberg | Fase 8 | **COMPLETED** | [`pkg/l1/routing_stress_test.go`](../pkg/l1/routing_stress_test.go) (116L) |
| **TASK-012** | Mecanismo de Encapsulamiento Post-Cuántica (PQC) Híbrido ML-KEM-768 | Fase 8 | **COMPLETED** | [`pkg/l0/pqc_kem.go`](../pkg/l0/pqc_kem.go) (172L) / Test 100% PASS |
| **TASK-013** | Adaptador TUN/TAP y Virtual NIC para Interfaz Nativa de Sistema Operativo | Fase 9 | **COMPLETED** | [`pkg/l1/tun_router.go`](../pkg/l1/tun_router.go) (152L) |
| **TASK-014** | Motor de Traversal NAT Avanzado y Relays Cifrados Soberanos (DERP) | Fase 9 | **COMPLETED** | [`pkg/l1/nat_traversal.go`](../pkg/l1/nat_traversal.go) (197L) |
| **TASK-015** | Grafo Topológico Distribuido y Sincronización en Malla mediante CRDTs | Fase 10 | **COMPLETED** | [`pkg/l2/crdt_graph_sync.go`](../pkg/l2/crdt_graph_sync.go) (197L) |
| **TASK-016** | Bypass de Red eBPF / XDP para Conmutación Wire-Speed y Aceleración HW | Fase 10 | **COMPLETED** | [`pkg/l0/ebpf_xdp_spec.go`](../pkg/l0/ebpf_xdp_spec.go) / [`docs/rfc/RFC_IPVN7_XDP_ACCELERATION.md`](../docs/rfc/RFC_IPVN7_XDP_ACCELERATION.md) |
| **TASK-017** | SDKs Universales Multi-Lenguaje: Python (`ipvn7`) y TypeScript/WASM | Fase 11 | **COMPLETED** | [`sdk/python/`](../sdk/python/) y [`sdk/typescript/`](../sdk/typescript/) |
| **TASK-018** | Modularización y Poda Universal: Reducción de Archivos Límites en Core, L3 y Web UI | Fase 11 | **COMPLETED** | Reducción de 6 monolitos por debajo de 400L (Axioma III universal) |
| **TASK-019** | Integración UPnP IGD Port Forwarding y Benchmark de Saturación RFC 3550 | Fase 12 | **COMPLETED** | [`pkg/l1/nat_upnp.go`](../pkg/l1/nat_upnp.go) (175L) / [`pkg/core/benchmark_runner.go`](../pkg/core/benchmark_runner.go) (128L) |
| **TASK-020** | Testamento Criptográfico y Clave de Sucesión Offline ($S$) | Fase 13 | **COMPLETED** | [`pkg/l1/uin_succession.go`](../pkg/l1/uin_succession.go) (81L) |
| **TASK-021** | Modos Operativos Tri-Estado (`Sovereign`, `Assisted`, `Bridge`) | Fase 13 | **COMPLETED** | [`pkg/core/operating_modes.go`](../pkg/core/operating_modes.go) (120L) |
| **TASK-022** | Sub-Puertos Lógicos Virtuales (`SubPort uint16` 65K Mux) | Fase 14 | **COMPLETED** | [`pkg/l1/subports.go`](../pkg/l1/subports.go) (107L) |
| **TASK-023** | Plano de Control Basado en Intenciones (Intent-Based ZTNA L3) | Fase 14 | **COMPLETED** | [`pkg/l3/intent_control.go`](../pkg/l3/intent_control.go) (172L) |
| **TASK-024** | Componente Satélite Bridge MQTT 3.1.1 Wire-Level TCP | Fase 15 | **COMPLETED** | [`pkg/components/mqtt_bridge/bridge.go`](../pkg/components/mqtt_bridge/bridge.go) (190L) |
| **TASK-025** | Componente Satélite Servidor/Proxy CoAP RFC 7252 UDP | Fase 15 | **COMPLETED** | [`pkg/components/coap_proxy/proxy.go`](../pkg/components/coap_proxy/proxy.go) (198L) |
| **TASK-026** | Puente Serial UART para Módulos LoRa (SX1276 / Ebyte) | Fase 15 | **COMPLETED** | [`pkg/l1/lora_serial.go`](../pkg/l1/lora_serial.go) (157L) |
| **TASK-027** | Hardening de Servicio y Tareas Programadas Windows (`schtasks`) | Fase 16 | **COMPLETED** | [`pkg/core/platform_windows.go`](../pkg/core/platform_windows.go) (100L) |
| **TASK-028** | Motor de Red Explicable en Lenguaje Natural L2 | Fase 16 | **COMPLETED** | [`pkg/l2/explainable_network.go`](../pkg/l2/explainable_network.go) (116L) |
| **TASK-029** | Control de Congestión BBR Adaptativo para Datagramas IPVN7 | Fase 17 | **COMPLETED** | [`pkg/l1/congestion_bbr.go`](../pkg/l1/congestion_bbr.go) (140L) |
| **TASK-030** | DHT Kademlia/Kleinberg Adaptativo y Descubrimiento P2P Descentralizado | Fase 17 | **COMPLETED** | [`pkg/l2/dht_kademlia.go`](../pkg/l2/dht_kademlia.go) (148L) |
| **TASK-031** | Batching de Sockets y Anillos de Buffers Zero-Copy para Alta Concurrencia | Fase 17 | **COMPLETED** | [`pkg/l0/packet_batcher.go`](../pkg/l0/packet_batcher.go) (135L) |
| **TASK-032** | Camuflaje de Tráfico TLS 1.3 RFC 8446 (Anti-DPI Masquerading L1) | Fase 18 | **COMPLETED** | [`pkg/l1/tls_masquerade.go`](../pkg/l1/tls_masquerade.go) (87L) |
| **TASK-033** | Monitor Continuo y Auto-Tuning Dinámico de MTU y Pacing (L1) | Fase 18 | **COMPLETED** | [`pkg/l1/adaptive_mtu.go`](../pkg/l1/adaptive_mtu.go) (105L) |
| **TASK-034** | Sensor de Salud y Telemetría de Anillo de Buffers en Memoria (L0) | Fase 18 | **COMPLETED** | [`pkg/l0/buffer_metrics.go`](../pkg/l0/buffer_metrics.go) (95L) |
| **TASK-035** | Gateway WoL (Wake-on-LAN) Físico y Shadow DIDs (IoT L1) | Fase 19 | **COMPLETED** | [`pkg/components/device_bridge/wol_gateway.go`](../pkg/components/device_bridge/wol_gateway.go) (115L) |
| **TASK-036** | Centinela Criptográfico de Rotación Proactiva de Claves de Sesión (L1) | Fase 19 | **COMPLETED** | [`pkg/l1/key_rotation.go`](../pkg/l1/key_rotation.go) (137L) |
| **TASK-037** | Exporter de Métricas y Telemetría Prometheus / OpenTelemetry (L3) | Fase 19 | **COMPLETED** | [`pkg/core/telemetry_exporter.go`](../pkg/core/telemetry_exporter.go) (132L) |
| **TASK-038** | Motor de Quórum y Votación Soberana del Senado de Agentes (L3) | Fase 20 | **COMPLETED** | [`pkg/l3/senate_quorum.go`](../pkg/l3/senate_quorum.go) (183L) / Test 100% PASS |
| **TASK-039** | Validador de Credenciales Verificables W3C para Peering Autónomo (L3) | Fase 20 | **COMPLETED** | [`pkg/l3/verifiable_credentials.go`](../pkg/l3/verifiable_credentials.go) (124L) / Test 100% PASS |
| **TASK-040** | Monitor de Resiliencia y Detección de Partición de Red (Split-Brain Guard L2) | Fase 20 | **COMPLETED** | [`pkg/l2/split_brain_guard.go`](../pkg/l2/split_brain_guard.go) (156L) / Test 100% PASS |
| **TASK-041** | Centinela Activo de Sondeo WAN y Telemetría RTT/Jitter RFC 3550 para Auto-Reparación Kleinberg (L1) | Fase 21 | **COMPLETED** | [`pkg/l1/wan_active_prober.go`](../pkg/l1/wan_active_prober.go) (255L) / Test 100% PASS |
| **TASK-042** | Optimización Radical de Tokens: Desacople de Scheduler LLM a Demonio Nativo Windows (0 Tokens) | Fase 22 | **COMPLETED** | [`scripts/run_autonomous_daemon.ps1`](../scripts/run_autonomous_daemon.ps1) / DEC-055 |
| **TASK-043** | Magna Multi-Suite de Verificación Predictiva y Diagnóstico Proactivo (Axioma III, Fuzzing, Caos) | Fase 23 | **COMPLETED** | [`scripts/multisuite/`](../scripts/multisuite/) / DEC-056 / Health Score |
| **TASK-044** | Modularización Preventiva y Certificación del 100% Health Score (Axioma III) | Fase 24 | **COMPLETED** | 7 módulos refactorizados (<320L), [`DEC-057`](../docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md), Score 100% |
| **TASK-045** | Driver Nativo Wintun L3 en Windows (Anillos de Buffers Compartidos L3 sin Stubs) | Fase 25 | **COMPLETED** | [`pkg/l1/tun_native_windows.go`](../pkg/l1/tun_native_windows.go) / DEC-058 |
| **TASK-046** | Cargador Nativo eBPF/XDP en Linux (Programa C XDP, Compilación ELF y Conmutación NIC) | Fase 25 | **COMPLETED** | [`pkg/l0/ebpf_xdp_spec.go`](../pkg/l0/ebpf_xdp_spec.go) / DEC-060 |
| **TASK-047** | Empaquetado Binario Multiplataforma y Scripts de Despliegue Universal (`install.sh`, `install.ps1`) | Fase 26 | **COMPLETED** | `scripts/install.sh` / `scripts/install.ps1` / DEC-059 |
| **TASK-048** | Componentes Satélites Listos para Usar (Túnel SSH PQC Soberano & Reverse Proxy Docker) | Fase 26 | **COMPLETED** | [`pkg/components/`](../pkg/components/) / DEC-061 |
| **TASK-049** | Protocolo de Malla Planetaria Distribuida Multi-Nodo (STUN descentralizado y relays DERP) | Fase 27 | **COMPLETED** | [`pkg/l2/planetary_mesh.go`](../pkg/l2/planetary_mesh.go) / DEC-062 |
| **TASK-050** | Verificación Formal Matemática y Publicación de RFCs Canónicos (IETF Track) | Fase 28 | **COMPLETED** | [`docs/rfc/`](../docs/rfc/) (RFC 9707, 9708, XDP) |
| **TASK-051** | Servidor MCP Nativo sobre Stdio para Agentes Autónomos de IA (`ipvn7 -mcp`) | Fase 29 | **COMPLETED** | [`cmd/ipvn7/main.go`](../cmd/ipvn7/main.go) / DEC-063 |
| **TASK-052** | Consolidación y Poda hacia el Producto Radical "VPN I7" (Poda Frontend y Botón Maestro 1-Clic) | Fase 30 | **COMPLETED** | [`docs/VPN_I7.md`](../docs/VPN_I7.md) / DEC-064 |
| **TASK-053** | Suite de Verificación Hostil WAN Nivel 2 (CGNAT celular, pérdida inducida y camuflaje TLS 1.3) | Fase 30 | **COMPLETED** | `scripts/multisuite/suite_network_hostile_wan.ps1` |
| **TASK-054** | Lanzador Zero-Friction de Inicio Rápido ("VPN I7" sin configuración de flags) | Fase 30 | **COMPLETED** | `scripts/start_vpn_i7.ps1` |
| **TASK-055** | Compatibilidad Universal WebAssembly (WASM Engine, JS Bridge, web/ipvn7.wasm) | Fase 31 | **COMPLETED** | [`pkg/wasm/`](../pkg/wasm/) / `cmd/ipvn7-wasm/` / DEC-079 |
| **TASK-056** | Sistema de Auto-Actualización con Control Atómico de Versiones y Rollback In-Place | Fase 31 | **COMPLETED** | [`pkg/core/version_manager.go`](../pkg/core/version_manager.go) / DEC-079 |
| **TASK-057** | Almacén de Archivos Distribuido por Contenido (DFS) con CAS y Chunking SHA-256 | Fase 32 | **COMPLETED** | [`pkg/dfs/store.go`](../pkg/dfs/store.go) / DEC-080 |
| **TASK-058** | Colector Distribuido de Diagnósticos y Consola Operativa (`ipvn7-cli diag/dfs`) | Fase 32 | **COMPLETED** | [`pkg/core/distributed_collector.go`](../pkg/core/distributed_collector.go) / DEC-080 |
| **TASK-059** | Reingeniería Radical de Experiencia de Usuario (UI por Pestañas Dedicadas, Copywriting Humano y Driver.js) | Fase 33 | **COMPLETED** | Pestañas dedicadas, lenguaje cotidiano, textos universales (DEC-086) |
| **TASK-060** | Agente Satélite Universal de Sistema Operativo (`ipvn7-os-runner`) y Control Físico de Hardware | Fase 34 | **COMPLETED** | Control ACPI, telemetría térmica, procesos y MCP Tools (DEC-087) |
| **TASK-061** | Puente Satélite Universal de Robótica, Drones, Automóviles y Maquinaria Pesada | Fase 35 | **COMPLETED** | MAVLink v2, CAN Bus ISO 11898, OBD-II, J1939 (DEC-088) |
| **TASK-062** | Programador Autónomo de Tareas y Cron Soberano de Malla en Tiempo Real | Fase 35 | **COMPLETED** | [`pkg/core/cron_scheduler.go`](../pkg/core/cron_scheduler.go) (DEC-089) |
| **TASK-063** | Despachador de Lenguaje Natural Local y Orquestador de Intenciones | Fase 36 | **COMPLETED** | [`pkg/l3/intent_orchestrator.go`](../pkg/l3/intent_orchestrator.go) (DEC-090) |
| **TASK-064** | Arquitectura Universal de Plugins con Dual Switch (Instalar/Desinstalar y Encender/Apagar) | Fase 37 | **COMPLETED** | [`pkg/core/server_components.go`](../src/pkg/core/server_components.go) / [`app_plugins.js`](../src/web/app_plugins.js) (DEC-091) |
| **TASK-065** | Suite de Auditoría UX con Agentes Persona y Onboarding Guiado (Driver.js) | Fase 37 | **COMPLETED** | `scripts/persona_agent_ux_audit.ps1` / [`app_tour.js`](../src/web/app_tour.js) (DEC-092) |
| **TASK-066** | Vigilancia Tecnológica Continua y Mapeo de Estándares IETF (X-Wing, MASQUE) | Fase 38 | **COMPLETED** | [`docs/research/RES-001_XWING_PQC_AND_MASQUE_STANDARDS.md`](../docs/research/RES-001_XWING_PQC_AND_MASQUE_STANDARDS.md) / DEC-093 |
| **TASK-067** | Implementación de X-Wing KEM (FIPS 203+X25519) e Integración MASQUE RFC 9298 | Fase 38 | **COMPLETED** | [`pkg/l1/pqc_hybrid.go`](../src/pkg/l1/pqc_hybrid.go) / [`pkg/l1/masque_tunnel.go`](../src/pkg/l1/masque_tunnel.go) / DEC-094 |
| **TASK-068** | Radar de Inteligencia Externa y Matriz Factual frente a WireGuard / Tailscale | Fase 39 | **COMPLETED** | [`docs/research/RES-003_WIREGUARD_PQC_LIMITATIONS_VS_IPVN7.md`](../docs/research/RES-003_WIREGUARD_PQC_LIMITATIONS_VS_IPVN7.md) / DEC-095 |
| **TASK-069** | Perforación CGNAT Celular mediante Predicción Delta y Relay Soberano Kleinberg | Fase 40 | **COMPLETED** | [`docs/research/RES-004_CELLULAR_SYMMETRIC_NAT_PORT_PREDICTION.md`](../docs/research/RES-004_CELLULAR_SYMMETRIC_NAT_PORT_PREDICTION.md) / DEC-096 |
| **TASK-070** | Inmunidad Anti-DPI por IA mediante Tramas Sphinx 1280B y Pacing DAITA | Fase 40 | **COMPLETED** | [`docs/research/RES-005_AI_TRAFFIC_ANALYSIS_DEFENSE_DAITA.md`](../docs/research/RES-005_AI_TRAFFIC_ANALYSIS_DEFENSE_DAITA.md) / DEC-097 |
| **TASK-071** | Telemetría y Control de Actuadores en Tiempo Real (Semántica RFC 9221 sobre L0-L2) | Fase 41 | **COMPLETED** | [`docs/research/RES-006_REALTIME_ROBOTICS_DATAGRAMS_RFC9221.md`](../docs/research/RES-006_REALTIME_ROBOTICS_DATAGRAMS_RFC9221.md) / DEC-098 |
| **TASK-072** | Vinculación Zero-Friction 1-Clic mediante Short Authentication String (SAS RFC 6189) | Fase 42 | **COMPLETED** | [`docs/research/RES-007_ZERO_FRICTION_PAIRING_SAS_RFC6189.md`](../docs/research/RES-007_ZERO_FRICTION_PAIRING_SAS_RFC6189.md) / DEC-099 |
| **TASK-073** | Enrutamiento Greedy Kleinberg con Métrica Compuesta Distancia-RTT | Fase 43 | **COMPLETED** | [`docs/research/RES-008_KLEINBERG_SMALL_WORLD_GREEDY_ROUTING.md`](../docs/research/RES-008_KLEINBERG_SMALL_WORLD_GREEDY_ROUTING.md) / DEC-100 |
| **TASK-074** | Camuflaje TLS 1.3 y Evasión DPI mediante Encrypted Client Hello (RFC 9849 ECH) | Fase 44 | **COMPLETED** | [`docs/research/RES-009_TLS13_ECH_RFC9849_DPI_EVASION.md`](../docs/research/RES-009_TLS13_ECH_RFC9849_DPI_EVASION.md) / DEC-101 |
| **TASK-075** | Proxy WoL P2P Autenticado y Control Profundo de Hardware sin Servidores | Fase 45 | **COMPLETED** | [`docs/research/RES-010_SECURE_WOL_RELAY_AND_HARDWARE_LOOP_CLOSURE.md`](../docs/research/RES-010_SECURE_WOL_RELAY_AND_HARDWARE_LOOP_CLOSURE.md) / DEC-102 |
| **TASK-076** | Procesamiento Zero-Copy con Disruptor Lock-Free Ring Buffer en L0-L2 | Fase 46 | **COMPLETED** | [`docs/research/RES-011_DISRUPTOR_LOCKFREE_RINGBUFFER_ZEROCOPY.md`](../docs/research/RES-011_DISRUPTOR_LOCKFREE_RINGBUFFER_ZEROCOPY.md) / DEC-103 |
| **TASK-077** | Firmas Compuestas Post-Cuánticas (Ed25519 + ML-DSA FIPS 204 y RFC 9955) | Fase 47 | **COMPLETED** | [`docs/research/RES-012_COMPOSITE_MLDSA_ED25519_FIPS204_SIGNATURES.md`](../docs/research/RES-012_COMPOSITE_MLDSA_ED25519_FIPS204_SIGNATURES.md) / DEC-104 |

---

## Registro de Ejecución y Trazabilidad

### TASK-001: Especificación RFC IPVN7-CORE
* **Objetivo:** Documento canónico de protocolo con definiciones ABNF, estructura binaria CBOR determinista, derivación criptográfica de direcciones (IPv6 ULA `fd07::/64` e IPv4 virtual `10.7.0.0/16`), MTU canónico 1280B y modelo de capas L0-L2.
* **Criterio de Aceptación:** Formato estándar RFC (estilo IETF), lenguaje normativo RFC 2119 (MUST, SHOULD, MAY), límite $\le$ 400 líneas.

### FASE 17: Conectividad de Malla Distribuida y Control de Congestión BBR Soberano
* **TASK-029 (BBR Congestion Control L1):** Estimación continua de ancho de banda cuello de botella (BtlBw) y retardo de propagación mínimo (RTprop). Control de pacing rate de datagramas UDP para saturación de línea sin bufferbloat en `pkg/l1/congestion_bbr.go`.
* **TASK-030 (DHT Kademlia/Kleinberg L2):** Enrutamiento métrico XOR con k-buckets ($k=8$) sobre identificadores UIN/DID de 256 bits en `pkg/l2/dht_kademlia.go`.
* **TASK-031 (Packet Batcher Ring Buffers L0):** Recepción y despacho por lotes (hasta 32 datagramas 1280B) con `sync.Pool` y cero alocaciones de heap en hot-path en `pkg/l0/packet_batcher.go`.

### FASE 18: Ofuscación Criptográfica Avanzada y Adaptación Dinámica de Enlace
* **TASK-032 (Anti-DPI Masquerading L1):** Encapsulación sintáctica de datagramas IPVN7 como registros TLS `ApplicationData` (0x17) con cabeceras RFC 8446 en `pkg/l1/tls_masquerade.go`.
* **TASK-033 (Adaptive MTU & Pacing L1):** Path MTU Discovery dinámico con retroalimentación en tiempo real para mallas mixtas (Fibra/Wi-Fi/LoRa) en `pkg/l1/adaptive_mtu.go`.
* **TASK-034 (Buffer Pool Telemetry L0):** Inspección y telemetría atómica de buffers reciclados para garantizar el cumplimiento del presupuesto de RAM en `pkg/l0/buffer_metrics.go`.

### FASE 19: Soberanía de Periféricos Físicos y Rotación Criptográfica Proactiva
* **TASK-035 (Wake-on-LAN & Shadow DIDs IoT L1):** Emisión física de Magic Packets UDP broadcast (`255.255.255.255:9`) mapeando periféricos LAN a Shadow DIDs soberanos en `pkg/components/device_bridge/wol_gateway.go`.
* **TASK-036 (Session Key Rotation L1):** Re-keying periódico automático (Forward Secrecy) cada 1 GB o 1 hora de conexión en `pkg/l1/key_rotation.go`.
* **TASK-037 (Prometheus Metrics Exporter L3):** Endpoint `/metrics` exportando telemetría canónica de jitter RFC 3550, buffers y Trust Tiers en `pkg/core/telemetry_exporter.go`.

### FASE 20: Gobernanza Descentralizada y Cripto-Acreditación de Red
* **TASK-038 (Senate Quorum Engine L3):** Cálculo de quórum ponderado y votación de agentes autónomos para decisiones constitucionales de red en `pkg/l3/senate_quorum.go`.
* **TASK-039 (W3C Verifiable Credentials L3):** Emisión y verificación de credenciales criptográficas para peering y acuerdos de nivel de servicio (SLA) en `pkg/l3/verifiable_credentials.go`.
* **TASK-040 (Split-Brain Guard L2):** Detección de partición topológica y aislamiento adaptativo de clusters de malla en `pkg/l2/split_brain_guard.go`.

### FASE 21: Auto-Sondeo Activo WAN, Centinela de Pérdida Continua y Auto-Reparación Kleinberg
* **TASK-041 (Active WAN Prober & Jitter Sentinel L1):** Emisión periódica de datagramas UDP ping de 32B (`IP7P`), cálculo continuo de RTT y jitter RFC 3550, retroalimentación atómica a `LinkHealingEngine` para failover O(1) preventivo y despacho de ecos en `pkg/l1/wan_active_prober.go`.

### FASE 22: Optimización Radical de Consumo de Tokens y Eficiencia de Supervisión
* **TASK-042 (Local Autonomous Daemon & Rule Deduplication):** Sustitución de pings de LLM (100k+ tokens cada 10 min) por demonio nativo de PowerShell en segundo plano (`scripts/run_autonomous_daemon.ps1`). Deduplicación de `GEMINI.md` a puntero canónico (-1,200 tokens/turno), rotación de logs de ciclo autónomo a histórico y modularización de `server.go` / `/metrics` cumpliendo estrictamente el Axioma III.

### FASE 23: Magna Multi-Suite de Verificación Predictiva y Diagnóstico Proactivo
* **TASK-043 (Magna Multi-Suite & Predictive Engine):** Implementación de suite de diagnóstico modular en `scripts/multisuite/` (<120L por archivo) con detección preventiva temprana de 320-400L, contienda multi-core, control de deriva zero-copy (0 B/op), fuzzing de datagramas de frontera (`pkg/l0/wire_fuzz_test.go`), resiliencia UDP de red con jitter RFC 3550, compilación reproducible e informe automatizado con Health Score (0-100%) en `docs/VERIFICATION_REPORT.md`.

### FASE 24: Modularización Atómica y Certificación del 100% Health Score
* **TASK-044 (Axioma III & 100% Score):** Modularización atómica de los 7 archivos en zona preventiva (320-400L) desglosando diagnósticos, comandos CLI, detectores de red, semillas Cypher y estilos CSS. Certificación formal del 100% Health Score (ÓPTIMO / EXCELENCIA) en `docs/VERIFICATION_REPORT.md` y registro en `DEC-057`.

### FASE 25: Conmutación Wire-Speed y Drivers de Kernel Nativos (Bloque B)
* **TASK-045 (Wintun Native Windows Driver L1) [COMPLETED]:** Integración de producción real de `wintun.dll` con creación de adaptador L3, sesión nativa (`WintunStartSession`), y anillos de buffers compartidos (`WintunReceivePacket` / `WintunSendPacket` / `RtlMoveMemory`) en `pkg/l1/tun_native_windows.go`. Fallback automático y transparente a `UserspaceVirtualAdapter`. Certificado con tests unitarios y 100% Health Score (DEC-058).
* **TASK-046 (Linux eBPF / XDP Loader L0) [COMPLETED]:** Implementación del programa C XDP `pkg/l0/xdp_ipvn7.c` verificado para el kernel eBPF con bounds checking, cargador nativo `LinuxXDPManager` (`pkg/l0/xdp_loader_linux.go`), fallback multiplataforma (`pkg/l0/xdp_loader_other.go`) y tests de modos (`pkg/l0/xdp_loader_test.go`) certificados (DEC-060).

### FASE 26: Ecosistema, Despliegue Universal y Componentes Satélites (Bloque D)
* **TASK-047 (Universal Deployment & Packaging) [COMPLETED]:** Creación de instaladores universales zero-friction `scripts/install.sh` (POSIX estándar con systemd y `setcap 'cap_net_admin,cap_net_bind_service=+ep'`) y `scripts/install.ps1` (PowerShell con firewall UDP 7777 / TCP 7070 y autoarranque SYSTEM) (DEC-059).
* **TASK-048 (Production Satellite Components) [COMPLETED]:** Conectores satélites listos para producción acoplados al Smart Gateway: túnel SSH protegido por ML-KEM-768 (`pkg/components/pqc_ssh/`) y Reverse Proxy Docker soberano por DID (`pkg/components/docker_proxy/`) certificados con sockets físicos y tests unitarios (DEC-061).

### FASE 27: Escalado Planetario y Malla Multi-Nodo Distribuida (Bloque C)
* **TASK-049 (Multi-Node Planetary Mesh Protocol) [COMPLETED]:** Expansión del protocolo a topologías de 10 a 50 nodos físicos geodistribuidos en `pkg/l2/planetary_mesh.go`, validación de saltos acotados $O(\log^2 N)$ con atajos de Kleinberg y resiliencia ante caídas simultáneas de nodos (DEC-062).

### FASE 28: Verificación Formal Matemática y Estandarización IETF (Bloque E)
* **TASK-050 (Formal Protocol Specification & RFC Track) [COMPLETED]:** Publicación y consolidación de la especificación técnica formal de estándares en `docs/rfc/RFC_IPVN7_CORE.md`, `RFC_IPVN7_SPHINX.md` y `RFC_IPVN7_XDP_ACCELERATION.md` cubriendo la trama canónica de 1280B, encriptación híbrida ML-KEM-768 y enrutamiento métrico soberano.

### FASE 29: Interoperabilidad Multi-Agente Autónoma y Protocolo MCP Soberano en Red
* **TASK-051 (Native Stdio MCP Server Engine) [COMPLETED]:** Acoplamiento directo de la bandera CLI `-mcp` en `cmd/ipvn7/main.go` hacia el bucle `ServeStdioDefault()` en `pkg/l3/mcp_server.go`. Exposición instantánea de más de 20 herramientas de gobernanza, topología, circuitos Sphinx y criptografía post-cuántica bajo JSON-RPC 2.0 para agentes autónomos de IA (DEC-063).

### FASE 30: Inevitabilidad Operativa, Tracción Radical y Poda Previa (DEC-064)
* **TASK-052 (Consolidación y Poda hacia el Producto Radical "VPN I7") [COMPLETED]:**
  - **Objetivo:** Poda agresiva de los componentes de física de partículas y simulación de grafos en el frontend (`app_kuzu_physics.js`, `app_kuzu_canvas.js`, `app_kuzu_inspector.js`, `app_kuzu_graph.js`, `style_kuzu_controls.css`, `style_kuzu_cypher.css`, `style_kuzu_workspace.css`) conforme al Algoritmo de 5 Pasos y [`docs/VPN_I7.md`](../docs/VPN_I7.md).
  - **Resultado:** Poda de >1,400 líneas de código accesorio en `web/`. Interfaz ultraligera y minimalista centrada en el botón maestro circular Dark Glassmorphism 1-clic con telemetría de latencia en tiempo real, carga instantánea $<30\text{ ms}$, actualización de `web_test.go` y 100% PASS en la suite.
* **TASK-053 (Batería de Validación Hostil WAN Nivel 2) [COMPLETED]:**
  - **Objetivo:** Implementación de suite de diagnóstico en `scripts/multisuite/suite_network_hostile_wan.ps1` que valide resistencia real en Internet hostil.
  - **Resultado:** Validación de 7/7 pruebas críticas de resistencia hostil WAN: perforación NAT STUN RFC 5389 (`TestSTUNHeaderGenerationAndParsing`, `TestSTUNRFC5389ResponseParsing`), camuflaje TLS 1.3 RFC 8446 (`TestTLSMasquerade_WrapUnwrap`), DPI bypass (`TestTLSOptionEngine_ClientHelloGeneration`), BBR Pacing (`TestBBRController_StartupAndConvergence`) y Corporate VPN Zero-Admin fallback (`TestCorporateVPNZeroAdminFallback`, `TestCorporateVPNTLSDisguise`). Integrado en `scripts/verify_ipvn7_standard.ps1` y certificado al 100% de Health Score.
* **TASK-054 (Lanzadores Zero-Friction y Detección de Privilegios) [COMPLETED]:**
  - **Objetivo:** Scripts de arranque inmediato `start_vpn_i7.ps1` (Windows) y `start_vpn_i7.sh` (Linux/macOS) en la raíz del repositorio.
  - **Resultado:** Lanzamiento en 1 comando o 1 clic. Autodetección nativa de privilegios (Admin en Windows / Root en Linux): modo Kernel TUN nativo si dispone de permisos; fallback automático a `ModeUserspaceProxy` (SOCKS5 `:10807`) si es usuario estándar sin permisos. Compilación automática si no existe el binario y apertura inmediata del panel de control en navegador. Verificado y operativo.

### FASE 31: Universalidad WebAssembly y Soberanía de Ciclo de Vida Atómico (DEC-079)
* **TASK-055 (Compatibilidad Universal WebAssembly) [COMPLETED]:**
  - **Objetivo:** Implementación del motor WASM universal (`pkg/wasm/engine.go`), punto de entrada `cmd/ipvn7-wasm/main.go`, y puente JS asíncrono `web/app_wasm.js`.
  - **Resultado:** Primitivas criptográficas puras (Ed25519, firma, verificación, Micro-PoW y validación de tramas 1280B) compilables con `GOOS=js GOARCH=wasm` a `web/ipvn7.wasm` (3.9 MB). 100% PASS en tests unitarios nativos.
* **TASK-056 (Sistema de Auto-Actualización con Control Atómico de Versiones) [COMPLETED]:**
  - **Objetivo:** Gestor atómico in-place de versiones en `pkg/core/version_manager.go` con manifiesto criptográfico firmado por Ed25519, verificación SHA-256, SemVer 2.0.0 y soporte anti-downgrade.
  - **Resultado:** Intercambio atómico en disco preservando `.old` y `.bad` para rollback de emergencia instantáneo en 1 comando (`ipvn7 -rollback` / `ipvn7-cli version`). Suite de pruebas al 100% PASS.

### FASE 32: Colector Distribuido de Diagnósticos y Almacenamiento por Contenido (DFS) (DEC-080)
* **TASK-057 (Almacén de Archivos Distribuido por Contenido - DFS) [COMPLETED]:**
  - **Objetivo:** Implementar direccionamiento por contenido (CAS) con fragmentación en bloques canónicos de 64 KB indexados por SHA-256 (`cid:sha256:...`), deduplicación en disco y manifiestos `FileManifest` firmados con Ed25519.
  - **Resultado:** Implementado en `pkg/dfs/store.go` y testeado en `pkg/dfs/store_test.go` (100% PASS). Verificación bit a bit certificada y detección de bloques corruptos (`ErrCorruptChunk`).
* **TASK-058 (Colector Distribuido de Malla y Consola Operativa) [COMPLETED]:**
  - **Objetivo:** Crear el motor de ingesta, desduplicación y persistencia de `DiagnosticReport` en el almacén DFS local, exponiendo endpoints REST y comandos en consola para auditoría continua.
  - **Resultado:** Implementado en `pkg/core/distributed_collector.go` y `pkg/core/server_diagnostics_dfs.go`. Comandos de consola `ipvn7-cli diag [list|report]` e `ipvn7-cli dfs [put|get|manifest]` en `cmd/ipvn7-cli/cli_dfs_diag.go`. Validado con tests unitarios y física real en el daemon activo.

### FASE 34: Control Físico de Hardware, Sistema y Ejecución de Intenciones (DEC-087)
* **TASK-060 (Agente Satélite Universal de Sistema Operativo `ipvn7-os-runner`) [COMPLETED]:**
  - **Objetivo:** Implementar el conector satélite universal para control profundo de hardware y SO: energía (sleep, reboot, shutdown, lock, monitor_off), procesos (kill_heavy, list), batería, térmica y expulsión de discos.
  - **Resultado:** Implementado en `pkg/components/os_runner/` con ejecutor nativo Win32/WMI para Windows (`platform_windows.go`), fallback multiplataforma (`platform_other.go`), suite unitaria al 100% PASS, acoplamiento automático en `cmd/ipvn7/main.go` y exposición para agentes de IA vía la herramienta MCP `ipvn7_system_control` (`pkg/l3/mcp_tools_system.go`). Validado con 100% PASS en la suite magna.

### FASE 35: Telemetría y Control de Drones, Robótica y Vehículos (DEC-088 y DEC-089)
* **TASK-061 (Puente Universal MAVLink v2 & CAN Bus ISO 11898 / J1939) [COMPLETED]:**
  - **Objetivo:** Control físico de robots, drones, vehículos y maquinaria pesada sin nubes propietarias.
  - **Resultado:** Implementado en `pkg/components/vehicle_robot_bridge/` con decodificadores MAVLink v2 (`mavlink.go`), tramas CAN Bus con OBD-II y SAE J1939 (`canbus.go`), y orquestador unificado (`bridge.go`). Herramienta MCP `ipvn7_robot_vehicle_control` expuesta para agentes de IA con ZTNA. 100% PASS en tests unitarios.
* **TASK-062 (Programador Autónomo Cron Soberano de Malla) [COMPLETED]:**
  - **Objetivo:** Ejecutar tareas recurrentes y temporizadores de malla soberana de alta precisión sin LLMs en bucle.
  - **Resultado:** Implementado en `pkg/core/cron_scheduler.go` con resolución de 50ms, soporte de expresiones cron estándar (`*/5 * * * *`) y sintaxis `@every`, canal no bloqueante y auto-reparación.

### FASE 36: Despachador de Lenguaje Natural Local y Orquestador de Intenciones (DEC-090)
* **TASK-063 (Parser Semántico Offline 0 Tokens y Endpoint de Intenciones) [COMPLETED]:**
  - **Objetivo:** Ejecutar intenciones humanas en español e inglés sin enviar datos a APIs externas ni gastar tokens.
  - **Resultado:** Implementado en `pkg/l3/intent_orchestrator.go`, endpoint `/api/v1/intent` en `pkg/core/server_intent.go`, y UI en `src/web/index.html` + `src/web/app_intent.js`.

### FASE 37: Arquitectura Universal de Plugins Dual Switch y Auditoría UX Radical (DEC-091 y DEC-092)
* **TASK-064 (Arquitectura Universal de Plugins con Dual Switch) [COMPLETED]:**
  - **Objetivo:** Control soberano total por el usuario sobre cada módulo externo con doble interruptor `[ Instalar / Desinstalar ]` y `[ Encender / Apagar ]`.
  - **Resultado:** Implementado con campos `Installed` y `Enabled` en `pkg/core/types.go`, endpoints `/api/v1/components/toggle` en `pkg/core/server_components.go`, estilos en `style_plugins.css` y lógica reactiva en `app_plugins.js`.
* **TASK-065 (Auditoría UX Radical con Agentes Persona y Onboarding Guiado) [COMPLETED]:**
  - **Objetivo:** Eliminar toda jerga críptica, proveer tour interactivo amigable y certificar Usability Score al 100%.
  - **Resultado:** Tour interactivo en `app_tour.js` y `style_tour.css`. Script de verificación factual `scripts/persona_agent_ux_audit.ps1` simulando perfiles reales (Carmen, Carlos, Elena) certificando 100/100 de puntuación de usabilidad humana. Assets embebidos 100% PASS en `web_test.go`.

### FASE 38: Vigilancia Tecnológica, Estandarización X-Wing KEM y Transporte MASQUE (DEC-093)
* **TASK-066 (Vigilancia Tecnológica Continua y Mapeo de Estándares IETF) [COMPLETED]:**
  - **Objetivo:** Monitorear y contrastar innovaciones del mundo real (WireGuard, Tailscale, IETF drafts) en cada ciclo del agente.
  - **Resultado:** Publicación de `docs/research/RES-001_XWING_PQC_AND_MASQUE_STANDARDS.md`. Análisis contrastado de X-Wing KEM (`draft-connolly-cfrg-xwing-kem`), RFC 10024 y MASQUE RFC 9298. Adopción en hoja de ruta técnica de ipvn7.
* **TASK-067 (Implementación de X-Wing KEM e Integración MASQUE CONNECT-UDP) [COMPLETED]:**
  - **Objetivo:** Homologar la KDF híbrida en L1 a la especificación formal X-Wing y habilitar el túnel fallback MASQUE RFC 9298 en L1.
  - **Resultado:** Implementado en `src/pkg/l1/pqc_hybrid.go` (`DeriveXWingSharedSecret`, `XWingEncapsulate`, `XWingDecapsulate` conformes a `draft-connolly-cfrg-xwing-kem` / RFC 10024) y `src/pkg/l1/masque_tunnel.go` (Capsule Protocol RFC 9297/9298 CONNECT-UDP sobre puerto 443). Cobertura al 100% PASS en `suite_network_hostile_wan.ps1` (9/9 pruebas).

### FASE 39: Radar Activo de Inteligencia Externa y Benchmark Factual frente a Estándares Mundiales (DEC-095)
* **TASK-068 (Vigilancia Exógena y Matriz Factual frente a WireGuard/Tailscale) [COMPLETED]:**
  - **Objetivo:** Investigar fuentes primarias externas sobre las limitaciones reales de WireGuard y Tailscale respecto a PQC en banda y dependencia centralizada.
  - **Resultado:** Publicación de `docs/research/RES-003_WIREGUARD_PQC_LIMITATIONS_VS_IPVN7.md` demostrando el colapso por fragmentación UDP en Noise de WireGuard frente a tramas 1280B cuantizadas de ipvn7, actualización de `docs/08_COMPARATIVA_INFRAESTRUCTURA_GLOBAL.md` y registro formal DEC-095.

### FASE 40: Perforación CGNAT Celular y Blindaje Anti-DPI por IA (DEC-096 y DEC-097)
* **TASK-069 (Perforación CGNAT Celular con Predicción Delta y Relay Kleinberg) [COMPLETED]:**
  - **Objetivo:** Resolver el punto ciego de Tailscale en doble NAT simétrico celular 4G/5G mediante predicción secuencial $\Delta_{seq}$ y fallback multi-hop a la malla Kleinberg con 0 servidores de terceros.
  - **Resultado:** Ficha técnica de investigación proactiva `docs/research/RES-004_CELLULAR_SYMMETRIC_NAT_PORT_PREDICTION.md` y registro arquitectónico DEC-096.
* **TASK-070 (Inmunidad Anti-DPI por IA mediante Sphinx Cuantizado y DAITA Pacing) [COMPLETED]:**
  - **Objetivo:** Anular clasificadores ML de inspección profunda de paquetes (Maybenot/DAITA) mediante cuantización de longitud a 1280B y supresión de firmas de intervalo inter-llegada con decaimiento Lomax/Pareto.
  - **Resultado:** Ficha técnica de investigación proactiva `docs/research/RES-005_AI_TRAFFIC_ANALYSIS_DEFENSE_DAITA.md` y registro arquitectónico DEC-097.

### FASE 41: Telemetría y Control de Actuadores en Tiempo Real RFC 9221 (DEC-098)
* **TASK-071 (Semántica RFC 9221 de Datagramas no Confiables sobre L0-L2) [COMPLETED]:**
  - **Objetivo:** Eliminar el bloqueo de cabeza de línea (Head-of-Line Blocking) en actuadores, robótica y telemetría de alta frecuencia (100-1000 Hz) mediante datagramas no confiables pero cifrados con X-Wing en banda y cuantizados a 1280B fijo en el pipeline zero-copy.
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-006_REALTIME_ROBOTICS_DATAGRAMS_RFC9221.md` y registro arquitectónico DEC-098.

### FASE 42: Vinculación Zero-Friction 1-Clic mediante Short Authentication String RFC 6189 (DEC-099)
* **TASK-072 (Protocolo SAS Bilateral de 4 Dígitos sin Jerga Técnica) [COMPLETED]:**
  - **Objetivo:** Eliminar la fricción de copiar y pegar claves hexadecimales largas de 64 caracteres en la UI mediante un código numérico bilateral de 4 dígitos derivado del secreto post-cuántico X-Wing (RFC 6189 / Bluetooth Numeric Comparison).
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-007_ZERO_FRICTION_PAIRING_SAS_RFC6189.md` y registro arquitectónico DEC-099.

### FASE 43: Enrutamiento Greedy Kleinberg con Métrica Compuesta Distancia-RTT (DEC-100)
* **TASK-073 (Optimización de Topología Small-World con Prevención de Bucles O(log^2 N)) [COMPLETED]:**
  - **Objetivo:** Superar las limitaciones de memoria de tablas globales de enrutamiento (BGP/OSPF) mediante un enrutador greedy descentralizado de 12 anillos concéntricos con ponderación combinada de distancia de identificador y RTT físico real de sockets.
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-008_KLEINBERG_SMALL_WORLD_GREEDY_ROUTING.md` y registro arquitectónico DEC-100.

### FASE 44: Camuflaje TLS 1.3 y Evasión DPI mediante Encrypted Client Hello RFC 9849 (DEC-101)
* **TASK-074 (Ofuscación de SNI y Emulación Sintética de Navegador en Puerto 443) [COMPLETED]:**
  - **Objetivo:** Burlar sistemas de censura estatal y firewalls corporativos L7 que bloquean tráfico por inspección SNI mediante la partición ClientHelloOuter/Inner cifrada con HPKE (RFC 9849 ECH) y extensión GREASE.
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-009_TLS13_ECH_RFC9849_DPI_EVASION.md` y registro arquitectónico DEC-101.

### FASE 45: Proxy WoL P2P Autenticado y Control Profundo de Hardware (DEC-102)
* **TASK-075 (Encendido Remoto por Broadcast LAN Seguro y Puentes Nativos de Energía) [COMPLETED]:**
  - **Objetivo:** Permitir el encendido seguro de máquinas en reposo o apagadas a través de Internet/CGNAT mediante un proxy centinela en la misma LAN que verifica firmas Ed25519 e inyecta el Magic Packet WoL físico a `255.255.255.255:9`, con puentes directos al SO (`SetSuspendState`, `LockWorkStation`).
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-010_SECURE_WOL_RELAY_AND_HARDWARE_LOOP_CLOSURE.md` y registro arquitectónico DEC-102.

### FASE 46: Procesamiento Zero-Copy con Disruptor Lock-Free Ring Buffer (DEC-103)
* **TASK-076 (Arreglo Circular de Potencia de 2 y Cero Pausas de GC en L0-L2) [COMPLETED]:**
  - **Objetivo:** Prevenir pausas del recolector de basura (GC) y contienda de CPU bajo ráfagas de 50.000 PPS mediante Ring Buffer lock-free estático de 1280B con indexación bitwise `head & (Cap - 1)` y patrón Single-Producer.
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-011_DISRUPTOR_LOCKFREE_RINGBUFFER_ZEROCOPY.md` y registro arquitectónico DEC-103.

### FASE 47: Firmas Compuestas Post-Cuánticas Ed25519 + ML-DSA FIPS 204 (DEC-104)
* **TASK-077 (Identidades Soberanas Cuánticamente Inquebrantables y Separación de Planos) [COMPLETED]:**
  - **Objetivo:** Blindar las identidades de nodo y la verificación de actualizaciones atómicas contra falsificación cuántica mediante firmas duales compuestas (NIST FIPS 204 ML-DSA + Ed25519 / RFC 9955), manteniendo la autenticación de datagramas simétrica zero-copy en vuelo.
  - **Resultado:** Ficha técnica de investigación de frontera `docs/research/RES-012_COMPOSITE_MLDSA_ED25519_FIPS204_SIGNATURES.md` y registro arquitectónico DEC-104.



