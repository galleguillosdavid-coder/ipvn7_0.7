# ipvn7 Network OS — Universal Sovereign Core (v0.7.0)

```text
██╗██████╗ ██╗   ██╗███╗   ██╗███████╗
██║██╔══██╗██║   ██║████╗  ██║╚════██║
██║██████╔╝██║   ██║██╔██╗ ██║    ██╔╝
██║██╔═══╝ ╚██╗ ██╔╝██║╚██╗██║   ██╔╝ 
██║██║      ╚████╔╝ ██║ ╚████║   ██║  
╚═╝╚═╝       ╚═══╝  ╚═╝  ╚═══╝   ╚═╝  
   Universal Sovereign Core - ipvn7 NOS v0.7.0
```

**ipvn7** es un Sistema Operativo de Red Autónomo, Descentralizado y Soberano (Network Operating System - NOS) que opera como una malla superpuesta (*overlay mesh*) de escala planetaria.

---

> [!IMPORTANT]
> ### 🤖 DIRECTIVA DE ASUNCIÓN DE ROL Y AUTODISPARADOR CONTINUO (agentes)
> Toda entidad (desarrollador, agente de inteligencia artificial o lector) que interactúe con este proyecto **debe asumir inmediatamente el rol y normas configuradas en [`agentes/`](agentes/README.md)** y la **[`Constitución del Sistema`](../sistema/CONSTITUCION.md)**.
> **Obligación operativa:** Actuar en estricta conformidad con los estándares de ingeniería pragmática, gobernanza por intenciones y verificación continua.

---

## 🏛️ Arquitectura v0.7: El Núcleo Funcional Universal

En la versión **0.7**, el sistema consolida un modelo análogo al de la **Internet moderna pero inteligente**:

> En la Internet tradicional, el núcleo de red solo resuelve **IP + ICMP + UDP/TCP**, mientras que los servidores web, bases de datos, aplicaciones de streaming y clientes se conectan externamente a través de sockets.  
> En **ipvn7 v0.7**, el demonio central (`ipvn7`) es un **Núcleo Funcional Universal** ultra-liviano, hiper-estable y 100% portable, enfocado exclusivamente en transporte soberano, identidad y enrutamiento en malla. Cualquier funcionalidad adicional (chat, inteligencia artificial, almacenamiento, interfaces web o utilitarios) se acopla externamente como un **Componente Inteligente Satelital** a través del **Smart Component Gateway**.

---

## ⚡ Pilares del Núcleo Universal (v0.7)

1. **Identidad Criptográfica Soberana (L0):** Desacoplamiento total de direcciones IP físicas. Todo nodo es exclusivamente su par de claves asimétricas Ed25519 (`did:ipvn7:<pubkey>`), con derivación determinista de direcciones IPv6 soberanas (`fd07::/64`) e IPv4 virtuales (`10.7.0.0/16`).
2. **Formato Canónico Determinado (L0):** Tramas CBOR deterministas firmadas (RFC 8949) y MTU canónico estricto de 1280 bytes para eludir cualquier fragmentación en la Internet física.
3. **Enrutador de Mundo Pequeño de Kleinberg (L1):** Tablas de memoria acotadas estrictamente a 120 pares en 12 anillos concéntricos logarítmicos, con reenvío voraz $O(\log N)$ por distancia XOR y soporte de Roaming IP sin cortes de sesión.
4. **Reserva de Memoria Zero-Copy (L1):** Preasignación y reciclaje atómico de búferes en 3 niveles (64B / 1500B / 64KB) con invariante estricto de **0 B/op** y **0 allocs/op**.
5. **Cortafuegos ZTNA Default-Deny (L1):** Micro-segmentación nativa bidireccional por DID soberano sin confianza implícita.
6. **Centinela WAN y Failover O(1) (L1):** Sondeo continuo de enlaces (`IP7P`) y conmutación automática de rutas ante degradación o pérdidas.
7. **Telemetría Sincronizada (L2):** Ring buffer de observabilidad concurrente protegido con `sync.RWMutex` para registro atómico de métricas en tiempo real.
8. **Smart Component Gateway (El Bus Inteligente v0.7):** Sustrato de integración mediante WebSockets, Server-Sent Events (SSE), IPC local y API REST para acoplar componentes externos sin tocar el código fuente del núcleo.
9. **Pasarela de Salida Soberana Dinámica (Egress L1/L4):** Salida experimental al Internet público con selección del nodo de mayor velocidad (DEC-130), histéresis anti-flapping y MSS clamping a 1220B.

---

## 🚀 Despliegue y Ejecución Inmediata (Multiplataforma)

El Núcleo v0.7 se ejecuta de forma nativa en **Windows, Linux y macOS** sin dependencias externas:

```bash
# 1. Compilar el binario soberano (desde la raíz o src/)
cd src && go build -o ../bin/ipvn7.exe ./cmd/ipvn7

# 2. Ejecutar diagnóstico de arranque en frío
./bin/ipvn7.exe --diagnostics

# 3. Lanzar el nodo soberano (1-clic o comando)
./bin/ipvn7.exe --port 7777 --web-port 7070
```

### 🛡️ Modelo de Ejecución: Auto-Elevación con Fallback Determinista
Al ejecutar `ipvn7.exe` (por terminal o doble clic):
1. **Intento de Auto-Elevación (UAC):** Solicita permisos de Administrador de Windows de forma autoejecutada (`RunAs`).
2. **Si se concede Administrador:** Inicializa la interfaz virtual **Wintun L3** en el kernel. Habilita tráfico total TCP + UDP nativo (soporta tests de velocidad, WebRTC, DNS y streaming a máxima tasa).
3. **Si se rechaza o cancela el UAC:** El nodo **NO falla ni se interrumpe**; conmuta automáticamente a **Modo Usuario (Universal Gateway HTTP CONNECT + SOCKS5 en 127.0.0.1:10807)**, permitiendo navegación web segura sin privilegios.
4. **Protección Anti-Cortes:** Un centinela independiente supervisa el PID del proceso para restaurar la conexión a Internet directa en menos de 1 segundo si el programa se cierra.

El panel de control interactivo estará disponible en tu navegador en:  
👉 **[http://localhost:7070](http://localhost:7070)**

---

## 💻 Consola CLI `ipvn7-cli` (v0.7.0)

```bash
# Estado de identidad DID y prefijos IPv6 / IPv4
./bin/ipvn7-cli status

# Inspección de componentes inteligentes acoplados al gateway
./bin/ipvn7-cli components

# Lista de pares en los 12 anillos de Kleinberg
./bin/ipvn7-cli peers

# FSM de salud de pares
./bin/ipvn7-cli fsm
```

---

## 📚 Estructura de Documentación Canónica

### 🏛️ Documentos Cardinales (Raíz de `docs/`)

| Documento | Descripción |
|---|---|
| [`docs/FUENTE_DE_VERDAD.md`](FUENTE_DE_VERDAD.md) | **Fuente Suprema de Verdad:** Carta Magna Técnica, invariantes inmutables y matriz de auditoría. |
| [`docs/README.md`](README.md) | Portal maestro y visión integral del NOS universal v0.7. |
| [`docs/ARQUITECTURA.md`](ARQUITECTURA.md) | Visión canónica, modelo de capas L0-L2 y contratos raíz. |
| [`docs/BASELINE.md`](BASELINE.md) | Estado empírico medido bajo taxonomía estricta de 5 estados. |

---

### 📂 Carpetas Temáticas Especializadas

| Subdirectorio | Contenido Principal | Enlace Clave |
|---|---|---|
| [`docs/specs/`](specs/) | Decisiones arquitectónicas (ADRs), Smart Gateway API y especificaciones satelitales. | [`docs/specs/ADR_RESUMEN.md`](specs/ADR_RESUMEN.md) / [`ESPECIFICACIONES`](specs/ESPECIFICACIONES_SATELITALES_Y_ARQUITECTURA.md) |
| [`docs/guides/`](guides/) | Guías de usuario, instalación multiplataforma, CLI y VPN 1-clic. | [`docs/guides/GUIA_INSTALACION_Y_DESPLIEGUE.md`](guides/GUIA_INSTALACION_Y_DESPLIEGUE.md) / [`MANUAL CLI/VPN`](guides/MANUAL_DE_USO_CLI_Y_VPN.md) |
| [`docs/ops/`](ops/) | Reportes multi-suite, pruebas físicas de 2 nodos y bitácora de ciclos autónomos. | [`docs/ops/VERIFICATION_REPORT.md`](ops/VERIFICATION_REPORT.md) / [`LAB 2 NODOS`](ops/VERIFICATION_PHYSICAL_2NODE.md) |
| [`docs/plans/`](plans/) | Historial consolidado de planes de evolución y alineación técnica. | [`docs/plans/HISTORIAL_PLANES_FASES.md`](plans/HISTORIAL_PLANES_FASES.md) |
| [`docs/audit/`](audit/) | Informe de respuesta y cierre de auditoría más antecedente histórico íntegro. | [`docs/audit/INFORME_RESPUESTA_Y_CIERRE.md`](audit/INFORME_RESPUESTA_Y_CIERRE.md) |
| [`docs/legacy/`](legacy/) | Bitácoras de génesis, conceptos legacy y lecciones aprendidas de versiones previas. | [`docs/legacy/MEMORIA_HISTORICA_Y_LECCIONES.md`](legacy/MEMORIA_HISTORICA_Y_LECCIONES.md) |
| [`docs/core/`](core/) | Especificaciones de las 10 primitivas inmutables del Core I7. | [`docs/core/`](core/) |
| [`docs/rfc/`](rfc/) | Especificaciones formales estilo IETF (Core, Sphinx Onion Routing, XDP). | [`docs/rfc/`](rfc/) |
| [`docs/research/`](research/) | Compendio consolidado de investigación tecnológica (X-Wing PQC, NAT, DAITA, WASI). | [`docs/research/COMPENDIO_INVESTIGACION_PQC_Y_REDES.md`](research/COMPENDIO_INVESTIGACION_PQC_Y_REDES.md) |


