# Hoja de Ruta de Fases Futuras — Funcionalidades Rescatadas (v0.7)

**Documento de Planificación Arquitectónica:** Plan estructurado para incorporar progresivamente las capacidades de mayor valor rescatadas de los repositorios históricos de Google Drive (`G:\Mi unidad\...\Versiones_Anteriores`) y disco local (`D:\David\dvd`).  
**Estado:** COMPLETADO Y VERIFICADO (Fases 13, 14, 15 y 16 implementadas bajo /goal con tests físicos 100% PASS).  
**Axiomas de Control:** $\le$ 400 líneas por archivo, Cero Simulación, Realismo Físico y Línea de Producción Desacoplada.

---

## Fase 13: Soberanía de Identidad y Resiliencia Extrema

### TASK-020: Testamento Criptográfico y Clave de Sucesión Offline ($S$)
* **Origen:** `ip7uin_MVP_Spec_v1.1.docx` §4.3 (Google Drive).
* **Problema a Resolver:** En caso de robo o sustracción de la clave privada operativa de un nodo en línea, el legítimo dueño tradicionalmente pierde su identidad y su reputación acumulada en la red.
* **Diseño e Implementación Prevista:**
  1. Durante la inicialización de la identidad, se genera un par de claves de sucesión $S$. La clave privada se resguarda **100% offline** (almacenamiento en frío: pendrive, papel o hardware wallet).
  2. Solo el hash `sha256(S_pub)` se publica en el `BindingRecord` inicial del UIN (`pkg/l1/uin_identity.go`).
  3. Al detectar compromiso, el dueño firma un mensaje de migración con $S$:
     $$\text{root\_id\_nuevo} \parallel \text{timestamp} \parallel \text{"SUCCESSION"}$$
  4. Los nodos de la malla verifican la firma contra el hash registrado, activan una ventana de disputa de 48 horas y, tras confirmación de quórum (3 avales), transfieren el 80% del Trust Score a la nueva identidad, revocando la clave robada.
* **Archivos Objetivo:** `pkg/l1/uin_identity.go` y `pkg/l1/uin_identity_test.go`.

### TASK-021: Modos Operativos Tri-Estado (`MODE_SAFE`, `MODE_DEGRADED`, `MODE_OPEN`)
* **Origen:** `ip7uin_MVP_Spec_v1.1.docx` §4.8 y `01_ARQUITECTURA_NATURAL_DE_IPV8.md`.
* **Problema a Resolver:** Falta de un control operacional de exposición de red en el demonio para alternar entre máxima privacidad, degradación ante congestión y conectividad total.
* **Diseño e Implementación Prevista:**
  - `MODE_SAFE`: Desactiva balizas exteriores de descubrimiento (EBRA/broadcast), descarta conexiones de pares no anclados en disco y opera como nodo oscuro silencioso.
  - `MODE_DEGRADED`: Activación automática o manual ante alta pérdida de paquetes (>30%) o saturación de socket; impone cuotas estrictas y prioriza exclusivamente tráfico de control y salud.
  - `MODE_OPEN`: Modo nominal con descubrimiento adaptativo y enrutamiento Kleinberg completo.
  - Control expuesto vía endpoint `/api/v1/vpn/mode` (Hero Zen button) y subcomando `ipvn7-cli mode`.
* **Archivos Objetivo:** `pkg/core/operating_modes.go`.

---

## Fase 14: Multiplexación de Canales y Plano de Control Declarativo

### TASK-022: Sub-Puertos Lógicos Virtuales (`SubPort uint16`)
* **Origen:** `D:\David\dvd\antiguos\Ipv7-8\transport\subports.go`.
* **Problema a Resolver:** En la actualidad, múltiples aplicaciones sobre el mismo nodo compiten o requieren diferentes puertos físicos UDP, lo que complica el firewall y NAT traversal.
* **Diseño e Implementación Prevista:**
  - Incorporar en el payload canónico un multiplexor de 65,536 sub-puertos virtuales (`SubPort uint16`).
  - Canales canónicos estandarizados:
    - Sub-Puerto `1`: Chat y Mensajería E2EE.
    - Sub-Puerto `2`: Telemetría y Salud en tiempo real.
    - Sub-Puerto `3`: Almacén DAG y Store-and-Forward.
    - Sub-Puerto `4`: Benchmark y diagnóstico RFC 3550.
    - Sub-Puertos `1000..65535`: Aplicaciones dinámicas acopladas al Smart Gateway.
* **Archivos Objetivo:** `pkg/l1/subports.go` y `pkg/l1/subports_test.go`.

### TASK-023: Plano de Control Basado en Intenciones (Intent-Based ZTNA)
* **Origen:** `D:\David\dvd\Ipv7IEU\core\bridge\group_control.go`.
* **Problema a Resolver:** El cortafuegos ZTNA actual filtra por DIDs individuales; en entornos corporativos o con muchos dispositivos se requieren políticas de intención declarativas.
* **Diseño e Implementación Prevista:**
  - Estructura `IntentPolicy`: define quién (`SubjectGroup`), con qué intención declarada (`Intent`), sobre qué destino (`TargetGroup`), con efecto vinculante (`allow` o `deny`).
  - Motor de evaluación en memoria `EvaluateIntent(subjectDID, intent, targetDID)` integrado al pipeline L1.
* **Archivos Objetivo:** `pkg/l3/intent_control.go` y `pkg/l3/intent_control_test.go`.

---

## Fase 15: Conectividad IoT Heterogénea e Interoperabilidad Física

### TASK-024: Componente Satélite Bridge MQTT 3.1.1 Wire-Level
* **Origen:** `D:\David\dvd\Ipv7-4\core\bridge\mqtt_bridge.go`.
* **Problema a Resolver:** Dispositivos de domótica e IoT (Home Assistant, ESP32, Tasmota) se comunican por MQTT y no pueden ejecutar binarios Go completos.
* **Diseño e Implementación Prevista:**
  - Implementación minimalista del protocolo MQTT 3.1.1 en Go puro sobre TCP sin librerías externas de terceros.
  - Conexión a brokers MQTT locales y reenvío bidireccional de tópicos `ipvn7/<DID>/send` y `ipvn7/<DID>/rx` a través del Smart Component Gateway.
* **Archivos Objetivo:** `pkg/components/mqtt_bridge/bridge.go`.

### TASK-025: Componente Satélite Proxy CoAP RFC 7252
* **Origen:** `D:\David\dvd\Ipv7-4\core\bridge\coap_proxy.go`.
* **Problema a Resolver:** Sensores con batería de muy baja potencia no pueden soportar TCP ni CBOR pesado.
* **Diseño e Implementación Prevista:**
  - Proxy UDP en puerto 5683 que traduce peticiones CoAP compactas (GET, POST, Uri-Path) a datagramas de la malla.
* **Archivos Objetivo:** `pkg/components/coap_proxy/proxy.go`.

### TASK-026: Driver Serial para Antenas LoRa UART (SX1276 / Ebyte E22)
* **Origen:** `D:\David\dvd\antiguos\Ipv7-8\transport\lora_serial.go`.
* **Problema a Resolver:** Necesidad de comunicación física offgrid sin Internet ni infraestructura celular.
* **Diseño e Implementación Prevista:**
  - Detección automática de puertos seriales COM / ttyUSB y comunicación directa por UART con módulos LoRa físicos (256 bytes MTU) mediante `io.ReadWriter`.
* **Archivos Objetivo:** `pkg/l1/lora_serial.go`.

---

## Fase 16: Robustez de Plataforma y Red Explicable

### TASK-027: Hardening de Servicio y Tareas Programadas Windows
* **Origen:** `D:\David\dvd\Ipv7-4\core\security_windows.go`.
* **Problema a Resolver:** En Windows, el inicio automático en segundo plano genera alertas de UAC o deja adaptadores Wintun huérfanos tras cierres forzados.
* **Diseño e Implementación Prevista:**
  - Inicio desatendido mediante `schtasks /sc onlogon /rl highest` (bypass limpio de UAC).
  - Rutina de purga preventiva de interfaces Wintun y reglas obsoletas de firewall.
* **Archivos Objetivo:** `pkg/core/platform_windows.go`.

### TASK-028: Motor de Red Explicable en Lenguaje Natural
* **Origen:** `D:\David\dvd\Ipv8\ipv8_natural\05_INNOVACIONES_ESTRATEGICAS_PARA_IPV8.md`.
* **Problema a Resolver:** Los eventos de red y cambios de ruta suelen ser opacos para el usuario.
* **Diseño e Implementación Prevista:**
  - Motor de generación de diagnósticos en lenguaje natural para explicar por qué se eligió un par, por qué se degradó un enlace o por qué se activó el modo de contingencia.
* **Archivos Objetivo:** `pkg/l2/explainable_network.go`.

---

## Fase 25: Conmutación Wire-Speed y Drivers de Kernel Nativos (Bloque B)

### TASK-045: Driver Nativo Wintun L3 en Windows
* **Problema a Resolver:** En Windows, el modo actual utiliza `UserspaceVirtualAdapter`. Se requiere que el tráfico del sistema operativo fluya por interfaces L3 físicas de kernel a velocidad de línea (>5 Gbps) mediante el driver de alto rendimiento Wintun.
* **Diseño e Implementación Prevista:**
  1. Cargar dinámicamente `wintun.dll` y enlazar `WintunCreateAdapter`, `WintunOpenAdapter`, `WintunStartSession`, `WintunReceivePacket`, `WintunReleaseReceivePacket` y `WintunAllocateSendPacket`.
  2. Implementar los anillos circulares de intercambio de buffers L3 en memoria compartida sin llamadas al sistema bloqueantes.
  3. Configurar automáticamente la IP soberana (`10.7.0.X/16` y `fd07::X/64`) y la tabla de enrutamiento con comandos `netsh`.
* **Archivos Objetivo:** `pkg/l1/tun_native_windows.go` y `pkg/l1/tun_native_windows_test.go`.

### TASK-046: Cargador eBPF / XDP Nativo en Linux
* **Problema a Resolver:** La conmutación de datagramas UDP en Linux a través de la pila de red tradicional incurre en cambios de contexto de kernel a userspace.
* **Diseño e Implementación Prevista:**
  1. Implementar el programa C eBPF (`xdp_ipvn7.c`) con inspección de cabecera mágica `IP7V`, desvío $O(1)$ (`XDP_REDIRECT` / `XDP_TX`) y filtrado wire-speed directo en el driver de la tarjeta de red (NIC).
  2. Generar el cargador en Go utilizando la librería estándar `cilium/ebpf` acoplado a `pkg/l0/ebpf_xdp_spec.go`.
* **Archivos Objetivo:** `xdp/xdp_ipvn7.c` y `pkg/l0/ebpf_xdp_loader.go`.

---

## Fase 26: Ecosistema, Despliegue Universal y Satélites (Bloque D)

### TASK-047: Empaquetado Binario y Despliegue en un Solo Comando
* **Problema a Resolver:** Barrera de adopción para usuarios y administradores que requieren compilar desde código fuente.
* **Diseño e Implementación Prevista:**
  1. Script POSIX universal `scripts/install.sh`: Detección automática de arquitectura (amd64, arm64, armv7), descarga del binario verificado con hash SHA-256 e instalación como servicio de fondo (`systemd` / `launchd`).
  2. Manifiesto canónico para Windows Package Manager (`winget install ipvn7`).
* **Archivos Objetivo:** `scripts/install.sh` y `dist/winget/ipvn7.yaml`.

### TASK-048: Componentes Satélites Listos para Usar (SSH PQC & Reverse Proxy)
* **Problema a Resolver:** Facilitar la migración de servicios tradicionales a la malla soberana sin modificar aplicaciones existentes.
* **Diseño e Implementación Prevista:**
  1. **Túnel SSH PQC:** Satélite que encapsula sesiones SSH TCP tradicionales sobre datagramas blindados con ML-KEM-768 a través de un Sub-Puerto L1 dedicado.
  2. **Reverse Proxy Docker:** Conector que expone contenedores web locales mapeándolos a DIDs soberanos sin requerir puertos públicos en el router.
* **Archivos Objetivo:** `pkg/components/ssh_tunnel/` y `pkg/components/docker_proxy/`.

---

## Fase 27: Malla Planetaria Distribuida Multi-Nodo (Bloque C)

### TASK-049: Protocolo de Malla Planetaria Distribuida Multi-Nodo
* **Problema a Resolver:** El laboratorio actual opera certificado en 2 nodos físicos; se requiere validar la convergencia de los 12 anillos de Kleinberg en topologías geodistribuidas complejas ($N \ge 10-50$).
* **Diseño e Implementación Prevista:**
  1. Red de relays soberanos DERP descentralizados para superar NATs simétricos corporativos donde falle el hole punching directo.
  2. Telemetría agregada de convergencia topológica en tiempo real y rebalanceo de anillos ante particiones WAN intercontinentales.
* **Archivos Objetivo:** `pkg/l1/nat_traversal.go` y `pkg/l2/crdt_graph_sync.go`.

---

## Fase 28: Verificación Formal Matemática y Estandarización IETF (Bloque E)

### TASK-050: Verificación Formal Matemática y Publicación RFC
* **Problema a Resolver:** Un estándar mundial exige certificación formal independiente de que los protocolos criptográficos no presentan fallas algebraicas ni ataques de canal lateral o man-in-the-middle.
* **Diseño e Implementación Prevista:**
  1. Modelar el apretón de manos híbrido X25519 + ML-KEM-768 en lenguaje formal **ProVerif** (`ipvn7_pqc.pv`) y ejecutar la prueba de seguridad matemática (Secrecy y Authentication).
  2. Publicar y validar la especificación formal final en [`docs/rfc/RFC_IPVN7_CORE.md`](./rfc/RFC_IPVN7_CORE.md).
* **Archivos Objetivo:** `docs/formal/ipvn7_pqc.pv` y `docs/rfc/RFC_IPVN7_CORE.md`.

