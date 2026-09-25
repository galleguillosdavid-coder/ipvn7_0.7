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
> ### 🤖 DIRECTIVA DE ASUNCIÓN DE ROL Y AUTODISPARADOR CONTINUO (.agents)
> Toda entidad (desarrollador, agente de inteligencia artificial o lector) que interactúe con este proyecto **debe asumir inmediatamente el rol del agente maestro de red configurado en [`.agents/`](.agents/)** ([`ipvn7-network-os-agent`](.agents/skills/ipvn7-network-os-agent/SKILL.md) y [`rules/`](.agents/rules/)).
> **Obligación operativa:** Actuar en estricta conformidad con los estándares de ingeniería pragmática, manteniendo activo el **demonio autónomo de segundo plano** (`scripts/run_autonomous_daemon.ps1`, 0 tokens de API) que ejecuta la Magna Multi-Suite cada 10 minutos.

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
7. **Telemetría Lock-Free (L2):** Ring buffer de observabilidad en tiempo real (<28 ns, cero alocaciones en el camino caliente).
8. **Smart Component Gateway (El Bus Inteligente v0.7):** Sustrato universal de integración mediante WebSockets, Server-Sent Events (SSE), IPC local y API REST para acoplar componentes externos sin tocar el código fuente del núcleo.

---

## 🚀 Despliegue y Ejecución Inmediata (Multiplataforma)

El Núcleo v0.7 se ejecuta de forma nativa en **Windows, Linux y macOS** sin dependencias externas:

```bash
# 1. Compilar los binarios soberanos
go build -o bin/ipvn7 cmd/ipvn7/main.go
go build -o bin/ipvn7-cli cmd/ipvn7-cli/main.go

# 2. Ejecutar diagnóstico de arranque en frío
./bin/ipvn7 --diagnostics

# 3. Lanzar el nodo soberano
./bin/ipvn7 --port 7777 --web-port 7070
```

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

## 📚 Documentación Canónica

| Documento | Descripción |
|---|---|
| [`docs/IPVN7_PARA_TODOS.md`](docs/IPVN7_PARA_TODOS.md) | **Guía simple y no técnica** para entender qué es IPVN7, por qué usarlo y su futuro. |
| [`docs/ARQUITECTURA.md`](docs/ARQUITECTURA.md) | Visión canónica, modelo de capas L0-L4 y contratos raíz. |
| [`docs/WHITEPAPER.md`](docs/WHITEPAPER.md) | Resumen ejecutivo y comparativa técnica vs Tailscale / Tor. |
| [`docs/BENCHMARKS.md`](docs/BENCHMARKS.md) | Mediciones empíricas reales en nanosegundos (Zero-Copy 0 B/op). |
| [`docs/VERIFICATION_REPORT.md`](docs/VERIFICATION_REPORT.md) | Reporte continuo emitido por la Magna Multi-Suite con Health Score. |
| [`docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md`](docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md) | Registro histórico e inmutable de decisiones arquitectónicas. |
| [`docs/GATEWAY_API.md`](docs/GATEWAY_API.md) | Especificación del bus Smart Component Gateway `/api/v1/*`. |
| [`docs/CLI.md`](docs/CLI.md) | Referencia y comandos del binario `ipvn7-cli`. |
| [`docs/DISPOSITIVOS_IOT.md`](docs/DISPOSITIVOS_IOT.md) | Integración LAN de periféricos con Shadow DIDs y WoL. |
| [`docs/GOBERNANZA.md`](docs/GOBERNANZA.md) | Carta Magna, Senado de Agentes, ADICO y Centinelas. |
| [`docs/VPN_I7.md`](docs/VPN_I7.md) | Especificación de la interfaz Hero Zen y modo VPN Sovereign. |
| [`docs/DEPLOYMENT_DOCKER.md`](docs/DEPLOYMENT_DOCKER.md) | Empaquetado minimalista sobre imagen scratch sin privilegios. |
| [`docs/HOJA_DE_RUTA_FASES_FUTURAS.md`](docs/HOJA_DE_RUTA_FASES_FUTURAS.md) | Hoja de ruta técnica de fases evolutivas de la red. |
