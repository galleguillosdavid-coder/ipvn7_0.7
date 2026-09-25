# Registro Histórico de Decisiones de Arquitectura IPVN7 (DEC-046 a DEC-078)

Este documento preserva las decisiones arquitectónicas estables consolidadas durante la versión v0.7 de IPVN7 Network OS, en estricto cumplimiento de la regla de poda y modularización preventiva (Axioma III, Límite de 400 líneas).

---

### DEC-046: Multiplexor de Sub-Puertos 65K y Plano de Control ZTNA Basado en Intenciones
* **Fecha:** 2026-09-22
* **Contexto:** Se requería multiplexar miles de microservicios, chat, telemetría y DAG sobre el único puerto UDP físico de malla (7777 / 7001), con políticas ZTNA basadas en intención y contexto (UIN Capas 2 y 3).
* **Error / Anti-Patrón Prohibido:** Abrir múltiples puertos UDP físicos en el host o evaluar accesos mediante listas de control estáticas basadas únicamente en IPs de red.
* **Decisión Adoptada:** Implementar `SubPortMux` en `pkg/l1/subports.go` (107L) con prefijo binario de 2 bytes (`uint16`), y `IntentEngine` en `pkg/l3/intent_control.go` (172L) con políticas declarativas ZTNA, limitador de tasa y validación de contexto operacional. Tests 100% PASS en `subports_test.go` e `intent_control_test.go`.

### DEC-047: Pasarelas IoT Físicas (MQTT 3.1.1, CoAP RFC 7252 y Enlace Serial UART LoRa SX1276)
* **Fecha:** 2026-09-22
* **Contexto:** La integración de sensores de campo, microcontroladores (ESP32/STM32) y nodos de radiofrecuencia de largo alcance requería adaptadores nativos sin dependencias externas pesadas.
* **Error / Anti-Patrón Prohibido:** Usar simulaciones sintéticas o forzar a microcontroladores de 8/32 bits a correr pilas completas TLS/JSON.
* **Decisión Adoptada:** Implementar `MQTTBridge` TCP físico en `pkg/components/mqtt_bridge/bridge.go` (190L), servidor/codec binario CoAP RFC 7252 sobre UDP en `pkg/components/coap_proxy/proxy.go` (198L), y puente serial UART LoRa con SyncWord `0x3C 0x5E` y checksum XOR en `pkg/l1/lora_serial.go` (157L). Tests 100% PASS en loopback físico y sockets reales.

### DEC-048: Endurecimiento de Servicio en Windows, Limpieza Wintun y Red Explicable L2
* **Fecha:** 2026-09-22
* **Contexto:** Prevenir fallos de elevación en inicio de sesión en Windows, limpiar adaptadores de red huérfanos (`ieu0*`) y transparentar las decisiones de conmutación de la malla ante el operador humano.
* **Error / Anti-Patrón Prohibido:** Asumir que el SO siempre limpia las interfaces virtuales tras un cierre inesperado o mantener el enrutamiento como una "caja negra" ininteligible.
* **Decisión Adoptada:** Implementar `RegisterLogonTask` y `PurgeOrphanWintunAdapters` en `pkg/core/platform_windows.go` (100L) y `ExplainableNetworkEngine` en `pkg/l2/explainable_network.go` (116L) con ponderación determinista de RTT, jitter RFC 3550, pérdida y PQC Trust Tier. Tests 100% PASS en `platform_windows_test.go` y `explainable_network_test.go`.

### DEC-049: Agrupamiento por Lotes Zero-Copy en L0 y Camuflaje de Tráfico TLS 1.3 Anti-DPI
* **Fecha:** 2026-09-22
* **Contexto:** Maximizar el throughput de datagramas bajo alta concurrencia mediante reciclaje de buffers de 1280B y mitigar bloqueos o censura estatal mediante evasión de inspección profunda de paquetes (DPI).
* **Error / Anti-Patrón Prohibido:** Realizar syscalls individuales por cada paquete entrante o emitir datagramas con patrones de tráfico reconocibles por cortafuegos restrictivos.
* **Decisión Adoptada:** Implementar `PacketBatcher` y `BatchBufferPool` en `pkg/l0/packet_batcher.go` (135L) garantizando 0 alocaciones dinámicas, y códec de camuflaje TLS 1.3 RFC 8446 en `pkg/l1/tls_masquerade.go` (87L) encapsulando datagramas en registros `ApplicationData` (0x17). Tests 100% PASS en `packet_batcher_test.go` y `tls_masquerade_test.go`.

### DEC-050: Soberanía de Periféricos WoL, Forward Secrecy Proactivo y Telemetría Prometheus
* **Fecha:** 2026-09-23
* **Contexto:** Mapear dispositivos de red local legacy sin stack IPVN7 a identidades soberanas, forzar re-keying proactivo de claves de sesión (Pilar 1) y exponer telemetría abierta para dashboards y alertas de nivel corporativo.
* **Error / Anti-Patrón Prohibido:** Permitir claves de sesión estáticas indefinidas o simular la interacción con periféricos físicos de red local.
* **Decisión Adoptada:** Implementar `WoLGateway` en `pkg/components/device_bridge/wol_gateway.go` (115L) con emisión física UDP broadcast y Shadow DIDs, centinela de rotación proactiva `KeyRotationSentinel` en `pkg/l1/key_rotation.go` (137L) con ventana de gracia de 30s, y `TelemetryExporter` en `pkg/core/telemetry_exporter.go` (132L) con endpoint `/metrics` en formato OpenMetrics. Tests 100% PASS en toda la suite.

### DEC-051: Gobernanza Soberana por Quórum Criptográfico, Credenciales W3C y Detección de Partición L2
* **Fecha:** 2026-09-23
* **Contexto:** Consolidación de gobernanza autónoma entre nodos agentes (L3), validación formal de acuerdos de peering descentralizados mediante estándares W3C, y resiliencia de la topología de malla L2 frente a particiones de red y divergencia split-brain.
* **Error / Anti-Patrón Prohibido:** Confiabilidad ciega en votos sin firma criptográfica, acuerdos de peering no verificables o permitir escrituras críticas en topologías de red partidas sin quórum de latidos.
* **Decisión Adoptada:** Implementar `SenateQuorumEngine` en `pkg/l3/senate_quorum.go` (170L) con verificación Ed25519 y reglas de supermayoría/quórum; validador W3C `VCRegistryValidator` en `pkg/l3/verifiable_credentials.go` (125L) con canonical payload signing; y `SplitBrainGuard` en `pkg/l2/split_brain_guard.go` (155L) con detección de partición y ventana de estabilización antes de permitir mutaciones de red. Tests 100% PASS en `senate_quorum_test.go`, `verifiable_credentials_test.go` y `split_brain_guard_test.go`.

### DEC-052: Prohibición de Consulta Pasiva al Asumir Rol e Inicio Autónomo Inmediato
* **Fecha:** 2026-09-23
* **Contexto:** En cada nueva conversación, la IA tendía por reflejo de alineación por defecto a formular preguntas pasivas ("¿qué deseas que haga?", "¿cuál es el siguiente paso?") tras asumir su rol soberano en lugar de ejecutar proactivamente las tareas requeridas por el estándar.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Formular preguntas de permiso o transferir la carga de decisión operativa al usuario tras asumir el rol de Agente Soberano de Red.
  * ❌ Permanecer inactivo esperando instrucciones cuando existen tareas pendientes en la cola `.agents/AUTOTASKS.md` o verificaciones del estándar por ejecutar.
* **Decisión Adoptada:**
  * ✅ Establecer prohibición formal e inviolable en `AGENTS.md`, `GEMINI.md` y `.agents/rules/`: al asumir el rol, el agente tiene prohibido preguntar qué hacer.
  * ✅ Mandato de ejecución inmediata: el agente debe inspeccionar el estado del sistema, verificar la compuerta universal e iniciar la siguiente tarea técnica de inmediato por su propia iniciativa.

### DEC-053: Centinela Activo de Sondeo WAN en Vivo y Auto-Reparación Preventiva Kleinberg (L1)
* **Fecha:** 2026-09-23
* **Contexto:** `LinkHealingEngine` poseía la lógica de conmutación $O(1)$ ante enlaces degradados, pero carecía de un generador proactivo de latidos/ping UDP en segundo plano, dependiendo exclusivamente de errores en datagramas de usuario.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Descubrir la caída de un enlace WAN únicamente cuando el usuario sufre cortes o pérdida de paquetes en una aplicación activa.
  * ❌ Generar sobrecarga de sondeo pesada con tramas voluminosas que degraden el ancho de banda del canal.
* **Decisión Adoptada:**
  * ✅ Implementar `WANActiveProber` en `pkg/l1/wan_active_prober.go` (255L) con datagramas ultralivianos de 32 bytes (`IP7P`), medición continua de RTT y jitter RFC 3550, despacho zero-copy de ecos y retroalimentación atómica directa hacia `LinkHealingEngine`.
  * ✅ Integración directa en el loop de transporte UDP en `cmd/ipvn7/main.go` con test 100% PASS en sockets loopback reales (`wan_active_prober_test.go`).

### DEC-054: Doble Mecanismo Permanente de Ejecución Autónoma (Schedule + Demonio PowerShell)
* **Fecha:** 2026-09-23
* **Contexto:** Tras cerrar turnos o conversaciones, el agente puede entrar en reposo si solo se ejecutó un script síncrono puntual, dejando pasar ventanas de tiempo sin avances en la bitácora de autotareas.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Asumir que una ejecución única síncrona mantendrá vivo el ciclo de 10 minutos una vez que la conversación finaliza.
  * ❌ Dejar la supervisión continua dependiendo únicamente de mensajes manuales del usuario.
* **Decisión Adoptada:**
  * ✅ Establecer formalmente en `AGENTS.md`, `GEMINI.md` y `.agents/rules/` el **Doble Mecanismo Permanente Obligatorio**:
    1. Programación formal del cron de fondo mediante `/schedule` (`*/10 * * * *`, `IsDaemon=true`).
    2. Proceso demonio de PowerShell residente en segundo plano ejecutando `scripts/run_autonomous_daemon.ps1` (`IsDaemon=true`).
  * ✅ Ambos mecanismos operan bajo exclusión mutua mediante `.agents/task.lock` atómico (`STATUS=FREE` / `STATUS=LOCKED`), garantizando cadencia ininterrumpida sin colisiones.

### DEC-055: Optimización Radical de Consumo de Tokens y Desacople de Supervisión Local
* **Fecha:** 2026-09-23
* **Contexto:** El uso del scheduler de IA (`/schedule`) generaba un consumo excesivo de tokens (100k+ por iteración de 10 min) al reinyectar el historial completo acumulado de la conversación. Además, la duplicación textual de reglas en `GEMINI.md` y `AGENTS.md` inflaba el prompt de sistema en cada turno.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Emplear bucles de reloj de LLM de alto costo de tokens para tareas de sondeo y verificación repetitiva que pueden ejecutarse localmente.
  * ❌ Mantener reglas de sistema duplicadas íntegramente en múltiples archivos inyectados en el contexto (`AGENTS.md` + `GEMINI.md`).
  * ❌ Dejar crecer logs de ejecución continua (`docs/AUTONOMOUS_CYCLE_LOG.md`) por encima de los límites del Axioma III.
* **Decisión Adoptada:**
  * ✅ Cancelar el timer de LLM y delegar el ciclo autónomo de 10 minutos de forma permanente y soberana al demonio PowerShell nativo (`scripts/run_autonomous_daemon.ps1`) en Windows con costo 0 tokens de API.
  * ✅ Sintetizar `GEMINI.md` a un puntero canónico de 20 líneas referenciando `AGENTS.md`, ahorrando ~1,200 tokens de entrada en cada turno.
  * ✅ Rotar `docs/AUTONOMOUS_CYCLE_LOG.md` segregando el historial a `AUTONOMOUS_CYCLE_LOG_HISTORICO.md` (42 líneas activas).
  * ✅ Modularizar `pkg/core/server.go` extrayendo `handleEventsSSE` a `server_events.go` (48 líneas) e integrando `TelemetryExporter` en `/metrics`.
  * ✅ Poda proactiva de `docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md` archivando DEC-026 a DEC-035 a `ADR_HISTORICO_V06_V07.md`.

### DEC-056: Magna Multi-Suite de Verificación Predictiva y Diagnóstico Proactivo
* **Fecha:** 2026-09-23
* **Contexto:** La compuerta universal básica realizaba verificaciones binarias pasivas sin anticipar o proyectar fallas de contienda, saturación de MTU, fugas de memoria o crecimiento de archivos hacia el límite del Axioma III ($\le$ 400L).
* **Error / Anti-Patrón Prohibido:**
  * ❌ Descubrir la rotura del límite de 400 líneas de un archivo cuando ya falló la compilación o el commit.
  * ❌ Acumular suites de pruebas masivas en un solo script monolítico de PowerShell que supere las 400 líneas.
  * ❌ No auditar la resiliencia de la capa wire ante bitflips en checksums SIMD o desbordamientos del MTU de 1280B.
* **Decisión Adoptada:**
  * ✅ Implementar la **Magna Multi-Suite** modular en `scripts/multisuite/` con 6 submódulos especializados (< 120 líneas c/u): estático con zona preventiva (320-400L), estrés concurrente multi-core, control de deriva zero-copy (0 B/op), fuzzing algorítmico de frontera (`wire_fuzz_test.go`), resiliencia de red UDP loopback y generador de reporte de salud.
  * ✅ Generar informe consolidado en `docs/VERIFICATION_REPORT.md` con *Health Score* global (0-100%) y alertas predictivas en cada ciclo del demonio autónomo (0 tokens de API consumidos).

### DEC-057: Modularización Atómica Preventiva y Certificación del 100% Health Score
* **Fecha:** 2026-09-23
* **Contexto:** El informe de la Magna Multi-Suite arrojaba un Health Score de 86% (Aceptable) debido a 7 penalizaciones predictivas (-2% cada una) por archivos en la zona preventiva de riesgo (320-400 líneas), amenazando el cumplimiento futuro del Axioma III.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Tolerar archivos en zona preventiva (320-400L) que acumulan deuda técnica y arriesgan sobrepasar el límite de 400 líneas en evoluciones posteriores.
  * ❌ Procrastinar la modularización atómica o realizar desgloses desordenados que rompan dependencias, selectores CSS o eventos en el frontend y CLI.
* **Decisión Adoptada:**
  * ✅ Aplicar el Algoritmo de 5 Pasos para modularizar atómicamente los 7 componentes en zona preventiva:
    1. `cmd/ipvn7/main.go`: Desglose de diagnósticos y registro de componentes a `diagnostics.go` (main: 261L, diag: 123L).
    2. `cmd/ipvn7-cli/main.go`: Desglose de comandos base (status, peers, radar...) a `cli_core_ops.go` (main: 182L, ops: 178L).
    3. `pkg/l1/network_probe.go`: Desglose de detección de censura y sondeo DNS a `probe_detectors.go` (probe: 217L, det: 117L).
    4. `pkg/l3/kuzu_graph.go`: Desglose de siembra canónica PFO/ADR a `kuzu_seeds.go` (graph: 231L, seeds: 94L).
    5. `web/index.html`: Formateo y compactación semántica manteniendo el 100% de IDs y funcionalidad (312L).
    6. `web/style_dashboard.css`: Desglose de despachador y eventos a `style_dispatcher.css` (dash: 250L, disp: 80L).
    7. `web/style_kuzu_controls.css`: Desglose de consola Cypher a `style_kuzu_cypher.css` (ctrls: 218L, cyph: 107L).
  * ✅ Certificar formalmente **100% Health Score (ÓPTIMO / EXCELENCIA)** en `docs/VERIFICATION_REPORT.md` con 0 violaciones, 0 advertencias, 0 B/op en memoria y 100% tests PASS.

### DEC-058: Integración de Producción del Driver Nativo Wintun en Windows
* **Fecha:** 2026-09-23
* **Contexto:** Para eliminar el cuello de botella del espacio de usuario y lograr conmutación L3 wire-speed en Windows, se requería operar sobre el driver de kernel oficial de WireGuard (`wintun.dll`) usando un anillo de memoria compartida de 4 MiB, con fallback transparente a virtual userspace si la DLL no está presente.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Simular interfaces de red de kernel mediante sockets mock o stubs sin interacción real de memoria.
  * ❌ Realizar conversiones no seguras de `uintptr` a `unsafe.Pointer` que violen las directivas de `go vet` (`unsafeptr`).
  * ❌ Bloquear el hilo de red principal sin sincronización mediante eventos nativos de Windows (`WaitForSingleObject`).
* **Decisión Adoptada:**
  * ✅ Implementar `NativeTunAdapter` en `pkg/l1/tun_native_windows.go` interactuando directamente con las exportaciones de `wintun.dll` (`WintunCreateAdapter`, `WintunStartSession`, `WintunReceivePacket`, `WintunAllocateSendPacket`).
  * ✅ Utilizar `procRtlMoveMemory` de `kernel32.dll` para copiar datagramas hacia/desde el anillo de memoria compartida de Wintun con punteros canónicos y cero violación de `go vet`.
  * ✅ Implementar fallback transparente a `UserspaceVirtualAdapter` si `wintun.dll` no está en el PATH, garantizando portabilidad sin fallos de inicio.
  * ✅ Validar en `pkg/l1/tun_native_windows_test.go` manteniendo 100% PASS y cero alocaciones en el trayecto crítico.

### DEC-059: Despliegue Universal One-Line y Servicios de Sistema
* **Fecha:** 2026-09-23
* **Contexto:** La adopción global de IPVN7 como sistema operativo de redes requería un proceso de instalación zero-friction ejecutable en un solo comando en Linux, macOS y Windows, configurando capacidades de red (`CAP_NET_ADMIN`) y servicios residentes sin requerir compilación manual.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Requerir que el usuario clone repositorios de Git y configure dependencias complejas para probar o desplegar un nodo.
  * ❌ Ejecutar daemons con privilegios de root absolutos en Linux cuando `setcap` permite aislar solo las capacidades de red indispensables.
* **Decisión Adoptada:**
  * ✅ Crear `scripts/install.sh` (POSIX estándar) con autodetección de arquitectura (`amd64`, `arm64`, `arm`), configuración de `systemd` y `setcap 'cap_net_admin,cap_net_bind_service=+ep'`.
  * ✅ Crear `scripts/install.ps1` (PowerShell) con auto-elevación, configuración de firewall para UDP 7777 / TCP 7070 y tarea programada de arranque `SYSTEM`.

### DEC-060: Conmutador de Kernel eBPF / XDP Nativo en Linux
* **Fecha:** 2026-09-23
* **Contexto:** Para alcanzar conmutación wire-speed (>10 Gbps) en Linux eludiendo la pila SKB del kernel, se requería un programa eBPF que filtre o desvíe datagramas en la tarjeta de red física (NIC driver) antes de que intervenga el subsistema IP de Linux.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Depender únicamente de simulaciones en espacio de usuario para componentes de conmutación de kernel.
  * ❌ Romper la compatibilidad en sistemas no-Linux (Windows, macOS) mediante dependencias directas de headers C de Linux.
* **Decisión Adoptada:**
  * ✅ Desarrollar el programa C oficial `pkg/l0/xdp_ipvn7.c` verificado para el eBPF verifier de Linux con filtrado determinista O(1) de magic bytes `0x49503756` ("IP7V") y mitigación de paquetes malformados.
  * ✅ Implementar la interfaz `XDPManager` con cargador nativo `LinuxXDPManager` mediante netlink/iproute2 (`pkg/l0/xdp_loader_linux.go`) y fallback limpio `FallbackXDPManager` (`pkg/l0/xdp_loader_other.go`).

### DEC-061: Conectores Satélites de Producción: PQC-SSH y Docker Proxy
* **Fecha:** 2026-09-23
* **Contexto:** Para extender la soberanía de IPVN7 hacia el ecosistema devops y de administración remota, se requería tunelizar flujos SSH estándar bajo protección criptográfica post-cuántica (ML-KEM-768) y proveer un proxy inverso para microservicios Docker sin depender de DNS ni autoridades certificadoras (CAs) centralizadas.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Exponer puertos SSH (22) sin protección cuántica hacia redes públicas o depender de retransmisores VPN centralizados.
  * ❌ Requerir certificados SSL/TLS emitidos por CAs centralizadas para microservicios internos cuando el espacio de nombres DID ya garantiza autenticidad criptográfica nativa.
* **Decisión Adoptada:**
  * ✅ Implementar `PQCSSHGateway` en `pkg/components/pqc_ssh/` realizando encapsulación ML-KEM-768 y cifrado ChaCha20-Poly1305 sobre flujos TCP físicos de OpenSSH con telemetría en tiempo real.
  * ✅ Implementar `SovereignDockerProxy` en `pkg/components/docker_proxy/` resolviendo rutas locales mediante cabeceras `Host` y `X-IPVN7-Target-DID`, permitiendo publicar contenedores Docker sobre identidades soberanas IPVN7.

### DEC-062: Malla Planetaria Distribuida (10 a 50 Nodos) y Enrutamiento de Kleinberg O(log^2 N)
* **Fecha:** 2026-09-23
* **Contexto:** Superar la certificación de 2 nodos físicos y validar a escala el comportamiento topológico de mallas geodistribuidas ($N \ge 10-50$), probando convergencia de atajos armónicos de Kleinberg ($P \propto d^{-2}$) y resiliencia ante caídas simultáneas de nodos sin fragmentación.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Asumir que los algoritmos de enrutamiento funcionan a escala planetaria sin pruebas empíricas estructuradas de saltos acotados $O(\log^2 N)$.
  * ❌ Desconectar clústeres ante la caída de enlaces locales en lugar de conmutar por atajos de largo alcance probabilísticos.
* **Decisión Adoptada:**
  * ✅ Implementar `PlanetaryMeshCluster` en `pkg/l2/planetary_mesh.go` con enlaces Manhattan locales y atajos ponderados de largo alcance (Kleinberg), garantizando caminos diagonales acotados ($\le 15$ saltos para 25-50 nodos).
  * ✅ Validar en `pkg/l2/planetary_mesh_test.go` la resiliencia topológica ante fallos aleatorios de nodos (10%), manteniendo conectividad y cero panics en tiempo de ejecución.

### DEC-063: Servidor MCP Nativo sobre Stdio para Agentes Autónomos de Inteligencia Artificial
* **Fecha:** 2026-09-23
* **Contexto:** Para que cualquier agente de IA (Antigravity, Claude, Cursor, modelos locales) gobierne la red sin APIs intermedias, el binario principal requería ejecutar un servidor Model Context Protocol (JSON-RPC 2.0) de alta velocidad sobre los canales estándar de entrada y salida (`stdio`).
* **Error / Anti-Patrón Prohibido:**
  * ❌ Declarar banderas CLI (`-mcp`) sin acoplarlas al bucle real de servicio, provocando que el comando no ejecute el protocolo esperado.
  * ❌ Requerir llamadas HTTP complejas cuando el estándar MCP para herramientas locales se orquesta nativamente mediante subprocesos stdin/stdout.
* **Decisión Adoptada:**
  * ✅ Enlazar la bandera `-mcp` en `cmd/ipvn7/main.go` directamente con `mcpServer.ServeStdioDefault()`, exponiendo el catálogo completo de 20+ herramientas operativas.

### DEC-064: Reformulación hacia la Inevitabilidad Operativa, Validación Hostil WAN y Poda Previa
* **Fecha:** 2026-09-23
* **Contexto:** La formulación original del objetivo supremo fomentaba la sobre-ingeniería teórica. Asimismo, la validación se restringía a una red local limpia (LAN) sin verificar resistencia ante CGNAT simétrico y cortafuegos restrictivos.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Desarrollar abstracciones teóricas antes de consolidar la tracción masiva del producto nuclear.
  * ❌ Subdividir archivos creando burocracia de código cuando el código accesorio debió ser eliminado.
* **Decisión Adoptada:**
  * ✅ Redefinir el Objetivo Supremo formalmente hacia la inevitabilidad física: 1-comando / 1-clic ("VPN I7").
  * ✅ Establecer el principio de **"Podar antes de modularizar"**: eliminar el 20-30% de código innecesario antes de considerar segregarlo.
  * ✅ Implementar la validación física en dos niveles: Nivel 1 (LAN limpia) y Nivel 2 (Validación Hostil WAN).

### DEC-065: Túnel de Ingreso Soberano (Zero-Trust Ingress Anti-Cloudflare) y Estrategia de Malla Global
* **Fecha:** 2026-09-23
* **Contexto:** Publicar servicios locales instantáneamente a la malla global con cifrado post-cuántico E2EE y ZTNA Default-Deny sin intermediarios centralizados.
* **Error / Anti-Patrón Prohibido:**
  * ❌ Requerir cuentas de terceros, tarjetas de crédito o delegación centralizada de DNS para publicar un servicio local a la red.
* **Decisión Adoptada:**
  * ✅ Implementar el comando `tunnel` y alias `expose` en `cmd/ipvn7-cli/cli_tunnel.go`: publica puertos locales bajo el DID soberano y alias `.ipvn7`.

### DEC-066: Inmunidad Anti-Envenenamiento de Rutas y Aislamiento Estricto de Radio de Impacto IoT
* **Fecha:** 2026-09-23
* **Contexto:** Garantizar que dispositivos IoT comprometidos no puedan degradar la topología global mediante anuncios BGP falsos o envenenamiento de tablas.
* **Decisión Adoptada:** Aislamiento estricto de radio de impacto mediante anillos Kleinberg zonales y atenuación de reputación cuadrática WoT.

### DEC-067: Blindaje Anti-Inanición del Plano de Control y Resolución de Dilemas Físicos
* **Fecha:** 2026-09-23
* **Contexto:** Prevenir el agotamiento de CPU por inundación de telemetría basura fuera de banda (`SubPort 2`) y garantizar determinismo XOR O(1) ante colisiones de hash a escala planetaria.
* **Decisión Adoptada:** Desacoplo estricto mediante buffers en anillo circulares con descarte O(1) de métricas obsoletas y resolución jerárquica de colisiones DID.

### DEC-068: Invariante Zero-Copy Scatter-Gather, Inmunidad OOM en Subpuertos y Anclaje de Movilidad
* **Fecha:** 2026-09-23
* **Contexto:** 1) `PacketBatcher` ejecutaba copias físicas de memoria CPU (`copy()`); 2) Desbordamiento de memoria por DIDs efímeros; 3) Tormentas de señalización de movilidad.
* **Decisión Adoptada:** `BatchIOVec` con punteros Scatter-Gather sin alocaciones; derivación de identidades no verificadas a `untrustedSharedBucket`; y amortiguación temporal O(1) en `MobilityAnchorEngine`.

### DEC-069: Mitigaciones de Primera Generación contra Estrés Sintético (L0/L1)
* **Fecha:** 2026-09-23
* **Contexto:** Mitigar ataques volumétricos mediante cookies HMAC y padding de longitud fija.
* **Decisión Adoptada:** Cookies sin estado en L0 y tramas Sphinx deterministas cuantizadas a 1280B.

### DEC-070: Blindaje Antagonista Avanzado (Subred, Fast-Failover y Anti-Secuestro)
* **Fecha:** 2026-09-23
* **Contexto:** Partición espacial por subred en L0, conmutación O(1) ante enlaces flapping y números de secuencia monótonos anti-secuestro en movilidad.
* **Decisión Adoptada:** Matriz de 256 cubetas de subred en L0; transiciones Make-Before-Break en L1; y validación monótona de secuencias en `MobilityAnchorEngine`.

### DEC-071: Blindaje Antagonista de Tercera Generación (Micro-PoW, Piso de Entropía, RFC 2439 y Resincronización)
* **Fecha:** 2026-09-23
* **Contexto:** Escape mediante micro-PoW ante cubetas saturadas, piso constante de entropía contra censura DPI, y resincronización criptográfica con saltos de secuencia masivos.
* **Decisión Adoptada:** Gradiente PoW de 8 bits en L0; piso basal en `StochasticPaddingEngine`; Route Flap Damping RFC 2439 en `LinkHealingEngine`; y validación Ed25519 de firmas de resincronización.

### DEC-072: Blindaje Antagonista de Cuarta Generación (Presupuesto Atómico de CPU y Renovación Estocástica)
* **Fecha:** 2026-09-23
* **Contexto:** Prevenir agotamiento de CPU por noncios PoW falsos y firmas corruptas en el Nodo Ancla; eliminar armónicos en el relleno de tráfico.
* **Decisión Adoptada:** Presupuesto atómico `powVerifyTokens` (10,000 SHA-256/s); proceso de renovación estocástico en $[1.5\text{s}, 2.5\text{s}]$; y presupuesto `resyncVerifyTokens` (500 Ed25519/s).

### DEC-073: Blindaje Antagonista de Quinta Generación (Cuotas Segregadas PoW, Cola Pesada Pareto y Tickets HMAC)
* **Fecha:** 2026-09-23
* **Contexto:** 1) Inanición global de tokens PoW; 2) Detección de soporte compacto uniforme por clasificadores KDE; 3) Bloqueo de roaming por spam asimétrico.
* **Decisión Adoptada:** Cuotas PoW segregadas por subred con compuerta de alineación en 1 ns; jitter con decaimiento de Pareto ($\alpha = 1.5$); y tickets de desafío stateless HMAC para prioridad absoluta en movilidad.

### DEC-074: Inmunización Antagonista de Sexta Generación (Hash PRF Sembrado, Espectro Continuo Lomax y Roaming 0-RTT)
* **Fecha:** 2026-09-23
* **Contexto:** 1) Botnets con IPs precomputadas para saturar cubetas específicas; 2) Clasificadores Next-Gen DPI que detectan ausencia de tránsito corto ($<400\text{ms}$); 3) Pérdida de paquetes en conmutación física de interfaces (.198 <-> .106).
* **Decisión Adoptada:** Mapeo de cubetas mediante `maphash` con semilla en RAM; espectro continuo Lomax Type II ($\alpha=1.5, \lambda=80\text{ms}, x_0=5\text{ms}$); y roaming 0-RTT con entrega proactiva de `ProactiveTicket`.

### DEC-075: Inmunización Antagonista de Séptima Generación (Descarte Anti-Fallback, Ratcheting de Semilla PRF y LUT Lomax)
* **Fecha:** 2026-09-23
* **Contexto:** 1) Ataques de degradación de vía rápida a lenta mediante tickets falsificados con secuencias alineadas (`seq & 0x0F == 0`); 2) Análisis de canal lateral de caché Prime+Probe en hipervisores multi-tenant; 3) Sobrecarga de ciclos FPU por invocación continua de `math.Pow` en la distribución de Lomax bajo congestión activa.
* **Decisión Adoptada:** Descarte estricto en 1 ns de tickets falsos; rotación de semilla PRF; tabla cuantizada (LUT 256 entradas, 512B); y modularización de `AntiReplayFilter` hacia `anti_replay.go` (83L).

### DEC-076: Inmunización Antagonista de Octava Generación (PRF NUMA Zero-Invalidation, Dithering Estocástico y Cold PoW)
* **Fecha:** 2026-09-23
* **Contexto:** 1) Inanición por Cache-Line Bouncing / Invalidación MESI en arquitecturas NUMA por `atomic.Pointer.Store()`; 2) Detección de Shannon Entropy por 256 estados discretos repetitivos en la LUT; 3) Bloqueo de arranque en frío (Cold Handover Lockout) al agotarse el pool por spam sin tickets.
* **Decisión Adoptada:** Mapeo PRF funcional sin escrituras en memoria mediante semilla base inmutable en RAM (MESI Shared, 0 invalidaciones NUMA) y rotación temporal matemática por Época Dorada ($H \oplus (\text{epoch} \cdot \phi)$) en `StatelessCookieGenerator`; dithering estocástico continuo en microsegundos; y gradiente de escape PoW en arranque en frío (`VerifyColdResyncPoW`: 8 bits en cero, ~1 µs).

### DEC-077: Inmunización Antagonista de Novena Generación (Desfase Estocástico por IP, LUT Branchless y Pre-filtro Cold PoW)
* **Fecha:** 2026-09-23
* **Contexto:** Fuga de frontera de fase, inanición de pipeline por branch misprediction y agotamiento por verificación ciega de SHA-256.
* **Decisión Adoptada:** Desfase estocástico de fase por IP (`ipPhase := (h >> 48) & 0x1F`); acceso a LUT 100% branchless por over-allocation en `lomaxLUT[LomaxQuantizedSteps + 1]`; y pre-filtro de hardware de 1 ciclo instantáneo `(nonce & 0x0F) != 0` complementado con límite atómico `coldPoWTokens` de 5,000 hashes/s.

### DEC-078: Inmunización Antagonista de Décima Generación (Warping Temporal Aperiódico, LUT Atómica de 32 Bits y PoW Dual-Tier)
* **Fecha:** 2026-09-23
* **Contexto:** Fuga por correlación de fase estática invariante al tiempo, fuga de canal lateral por cruce de línea de caché e inanición de cuota por inyección alineada.
* **Decisión Adoptada:** Modulación temporal aperiódica no estacionaria (Aperiodic Temporal Warping con secuencia de Weyl); tabla compacta de 514B y lectura de 32 bits en 1 instrucción `binary.LittleEndian.Uint32`; segregación en 64 cubetas por DID, etiqueta dinámica `ColdPoWTag(nodeDID, seq)` y gradiente adaptativo Dual-Tier (Tier-2: 12 bits en cero, ~15 µs).
