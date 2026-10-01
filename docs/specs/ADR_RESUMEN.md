# ADR Resumido - Decisiones de Arquitectura IPVN7

**Propósito:** Memoria persistente de decisiones arquitectónicas. Anti-reincidencia obligatoria.

## Históricos (DEC-001 → DEC-078)
Las decisiones históricas DEC-001 a DEC-078 residen archivadas en el historial inmutable del repositorio Git tras la condensación canónica al núcleo mínimo (DEC-107 y DEC-112).

## Decisiones Activas (RESUMEN)

### Compatibilidad y Despliegue
**DEC-079:** WASM Universal + Auto-Actualización Atómica  
- WASM puro para navegadores/Cloudflare Workers sin CGO
- Version manager con rollback Ed25519 + SHA-256
- ❌ Prohibido: Reemplazar binarios sin backup criptográfico

**DEC-080:** Colector Distribuido + DFS  
- Telemetría descentralizada sin nubes SaaS
- Almacenamiento CAS SHA-256 de 64KB
- ❌ Prohibido: Enviar diagnósticos a terceros

**DEC-081:** Poda Kùzu + VPN 1-Clic  
- Eliminados 800+ líneas de código muerto
- CSS paralelo + flag nativo --vpn
- ❌ Prohibido: Código muerto + descargas seriales CSS

**DEC-084:** Instalación Zero-Admin (30s)  
- Auto-detección de privilegios + fallback userspace
- Instalación en LOCALAPPDATA sin admin
- ❌ Prohibido: Exigir admin cuando existe modo userspace

### UX e Interfaz
**DEC-086:** UX Radical + Copywriting Humano  
- Lenguaje cotidiano sin jerga técnica
- Pestañas dedicadas + textos universales
- ❌ Prohibido: Términos técnicos sin explicación

**DEC-091:** Plugins Dual Switch  
- Switch independiente: Instalar/Desinstalar + Encender/Apagar
- Soberanía de usuario total
- ❌ Prohibido: Acoplar instalación a ejecución

**DEC-092:** Auditoría UX + Onboarding  
- Agentes Persona para validación
- Tour guiado Driver.js
- ❌ Prohibido: Validar con suposiciones de ingenieros

### Seguridad y Criptografía
**DEC-093:** Vigilancia Tecnológica  
- Radar IETF/NIST + fichas RES-XXX
- Estándares X-Wing PQC + MASQUE
- ❌ Prohibido: Diseñar en cámara de eco

**DEC-094:** X-Wing KEM + MASQUE RFC 9298  
- Híbrido ML-KEM-768 + X25519
- Túneles MASQUE sobre HTTPS 443
- ❌ Prohibido: Protocolos propietarios

**DEC-095:** Inteligencia Externa + PQC Nativo  
- Matriz factual vs WireGuard/Tailscale
- PQC NIST Nivel 3 nativo
- ❌ Prohibido: Aislamiento intelectual

**DEC-097:** Anti-DPI IA (Sphinx 1280B)  
- Tramas fijas + pacing DAITA
- Defensa contra análisis de tráfico
- ❌ Prohibido: Patrones de tráfico identificables

**DEC-099:** Vinculación 1-Clic SAS RFC 6189  
- Short Authentication String
- PQC P2P sin fricción
- ❌ Prohibido: Pairing complejo multi-paso

**DEC-101:** TLS 1.3 Evasión ECH RFC 9849  
- Encrypted Client Hello
- Anti-DPI corporativo
- ❌ Prohibido: TLS sin camuflaje

**DEC-104:** Firmas Compuestas (Ed25519 + Vector Experimental)  
- Ed25519 (RFC 8032 canónico) + Vector Reticular Experimental (ML-DSA no implementado en v0.7.0)
- Identidad soberana de transición post-cuántica
- ❌ Prohibido: Afirmar cumplimiento FIPS 204 sin implementación formal


### Red y Rendimiento
**DEC-085:** SSDP/UPnP Estable  
- IDs fijos sin puertos efímeros
- Derivación determinista shadowDID
- ❌ Prohibido: Puertos aleatorios como identidad

**DEC-096:** CGNAT Celular (Predicción Delta)  
- Birthday paradox para perforación
- Predicción de puertos simétricos
- ❌ Prohibido: Fallo en CGNAT móvil

**DEC-098:** Telemetría Real-Time RFC 9221  
- Semantics sobre L0-L2
- Control actuadores <28ns
- ❌ Prohibido: Telemetría lenta

**DEC-100:** Kleinberg Greedy RTT  
- Métrica compuesta distancia + RTT
- Enrutamiento small-world
- ❌ Prohibido: RTT estático

**DEC-102:** WoL P2P + Hardware HIL  
- Control físico autenticado
- Loop closure obligatorio
- ❌ Prohibido: Simulación de hardware

**DEC-103:** Zero-Copy Disruptor  
- Ring buffer lock-free
- 0 B/op en hot-path
- ❌ Prohibido: Copias de buffers

### Arquitectura
**DEC-090:** Lenguaje Natural Local  
- Orquestador de intenciones
- 0 tokens de API
- ❌ Prohibido: Dependencia de LLM cloud

**DEC-110:** Auto-Elevación a Administrador con Fallback Determinista a Modo Usuario  
- Mandato: Privilegios de administrador autoejecutados (AGENTS.md)
- Intento inmediato de auto-elevación RunAs para activar Wintun L3 (soporte total TCP + UDP, speedtests, gaming)
- Fallback automático transparente a Modo Usuario (Universal Gateway HTTP CONNECT + SOCKS5) si se cancela UAC
- ❌ Prohibido: Bloquear la ejecución si no hay admin; prohibido omitir el intento de elevación para Wintun L3

**DEC-111:** Logging Persistente en Disco y Ventana Interactiva Permanente (`cmd.exe /k`)  
- Consola de ejecución elevada persistente e interactiva mediante `cmd.exe /k` con título dedicado: previene ventanas que se cierran solas o ejecuciones invisibles.
- Registro dual simultáneo (consola interactiva + archivo `data/ipvn7.log`) con marcas de tiempo milimétricas.
- Banderas operativas `-debug` y `-logfile`, y visor en vivo `scripts/view_logs.ps1 -Follow`.
- ❌ Prohibido: Ejecutar procesos elevados sin ventana de consola interactiva; prohibido descartar logs de depuración al cerrar la consola.

**DEC-112:** Purga de Código Muerto y Reemplazo por Especificaciones en Lenguaje Natural  
- Eliminación de 134 archivos (~4.6 MB) de prototipos históricos obsoletos en `.archived/`.
- Condensación de la memoria técnica y arquitectónica en especificaciones de lenguaje natural de alta densidad (`docs/ESPECIFICACIONES_SATELITALES.md`).
- ❌ Prohibido: Acumular carpetas de código muerto o archivos de descarte en el árbol de trabajo; Git preserva el historial cronológico.

**DEC-113:** Modelo Agéntico Basado en Roles Operativos (7 Sub-Roles)  
- Estructuración del Agente Maestro (`ipvn7-network-os-agent`) en 7 sub-roles especializados (`agentes/ROLES.md`): Funcionalidades, Simplificación, Custodia Docs/Rutas, Radar Exógeno, Diagnóstico/Logs/Memoria, Compacidad $\le 400$L y Cero Basura.
- Cero tolerancia a parches tramposos en tests o violaciones de invariantes de raíz pura.
- ❌ Prohibido: Asumir tareas complejas sin invocar al rol de diagnóstico y verificación física correspondiente.

**DEC-114:** Expansión del Modelo Agéntico a 10 Sub-Roles Especializados (Roles H, I, J)  
- Incorporación formal de 3 sub-roles operativos de alta rigurosidad:
  1. **Rol H (Auditor UX Radical):** Erradicación de jerga técnica y comprensión en <5s por usuarios no técnicos (Regla 18).
  2. **Rol I (Adversario Caos WAN):** Falsabilidad HIL en red hostil, 20% pérdida de paquetes y cortes bruscos (Reglas 2 y 16).
  3. **Rol J (Guardián Benchmark Zero-Copy):** Invariante 0 B/op, 0 allocs/op y latencia wire-speed vs WireGuard (Regla 11).
**DEC-115:** Auditoría Integral Predictiva y Certificación de Lanzamiento en Vivo
- Validación física y estática mediante la Magna Multi-Suite (`scripts/verify_ipvn7_standard.ps1`): Health Score alcanzado en 100%.
- Aserción de sockets reales UDP (7777), SOCKS5 (10807), WebUI HTTP (7070) con respuestas JSON de estado estructuradas.
- ❌ Prohibido: Lanzar el sistema sin verificación previa de compuerta universal e integridad física de sockets.

**DEC-116:** Poda Preventiva en Pasarela Universal SOCKS5 (Axioma III / Regla 6)
- Reducción y poda del 15% de sobrecarga y parsing redundante en `src/pkg/l1/socks5_gateway.go` (de 394L a 335L).
- Preservación íntegra de contratos RFC 1928, tunelización CONNECT para navegadores y métricas atómicas `GetStats()`.
- ❌ Prohibido: Permitir que archivos nucleares superen 350 líneas sin podar primero el código accesorio redundante.

**DEC-117:** Rol K (Arquitecto de Distribución) y Matriz de Compilación Multiplataforma Universal
- Formalización del Rol K en `agentes/ROLES.md` y `agentes/AGENTS.md` para empaquetado y distribución en <30 segundos.
- Matriz universal (`scripts/build_all_platforms.ps1/sh`): generación de 6 binarios estáticos independientes (`windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`) con `CGO_ENABLED=0`, cero dependencias dinámicas y sumas SHA256 en `bin/SHA256SUMS.txt`.
- Sincronización de stubs de plataforma en `src/pkg/core/platform_other.go` y resolución de advertencias de compilación cruzada en macOS (`tun_native_darwin.go`).
- ❌ Prohibido: Reescribir el sistema en Assembly nativo (antítesis de universalidad); prohibido publicar binarios con dependencias de librerías dinámicas del sistema.

**DEC-118:** Rol L (Embajador de Dispositivos) y Arquitectura de Nodos Guardianes con Shadow DIDs
- Formalización del Rol L en `agentes/ROLES.md` y `agentes/AGENTS.md` para convertir computadores en Nodos Guardianes (Edge Ambassadors) de la LAN física.
- Derivación determinista de identidades virtuales (`did:ipvn7:shadow:<sha256(MAC)>`) y asignación de IPs soberanas en `10.7.100.0/24`.
- Reenvío L4 transparente zero-copy con filtrado ZTNA de puertos (impresoras IPP 631 / JetDirect 9100, cámaras RTSP 554, HTTP) e inyección física de Magic Packet Wake-on-LAN.
- Exposición en la API `/api/v1/status` y en la WebUI para visualización y control en 1 clic.
- ❌ Prohibido: Exigir modificación de firmware en dispositivos de hardware cerrado; prohibido exponer dispositivos IoT directamente a Internet sin el escudo ZTNA de la malla.

**DEC-119:** Rol M (Centinela de Versiones) y Auto-Actualización Atómica con Rollback
- Formalización del Rol M en `agentes/ROLES.md` y `agentes/AGENTS.md` para ciclo de vida SemVer y trazabilidad criptográfica.
- Implementación de `VersionManager` en `src/pkg/core/version_manager.go` con verificación forzosa de hash SHA256 antes del reemplazo.
- Reemplazo in-place atómico (`.new -> .old -> .exe`) con capacidad de rollback determinista automático en <5 segundos si el nuevo proceso no supera el health check.
- Flag nativo de CLI `ipvn7 -version` para inspección instantánea de versión, build y wire protocol.
- ❌ Prohibido: Instalar o aceptar actualizaciones sin verificación criptográfica; prohibido dejar un nodo sin internet ante una actualización rota.

**DEC-120:** Rol N (Blindaje Binario y Anti-Ingeniería Inversa con Garble y Stripping DWARF)
- Formalización del Rol N en `agentes/ROLES.md` y `agentes/AGENTS.md` para protección de ejecutables contra ingeniería inversa y extracción estática de cadenas/lógica.
- Integración de pipeline de compilación blindada (`scripts/build_hardened.ps1`): soporte de `garble` con ofuscación de AST (`-tiny`, `-seed=random`) y cifrado estático de literales (`-literals`).
- Invariantes de stripping estricto: erradicación de rutas locales (`-trimpath`), remoción de tabla de símbolos (`-ldflags="-s"`) y depuración DWARF (`-ldflags="-w"`).
- Fallback determinista a build nativo Go con stripping completo si `garble` no está instalado en el host, garantizando cero fallos de CI/CD.
- Registro y aserción incondicional de sumas SHA-256 en `bin/SHA256SUMS.txt`.
- ❌ Prohibido: Distribuir binarios con símbolos de depuración DWARF o rutas locales expuestas; prohibido asumir que la ofuscación reemplaza la seguridad matemática ZTNA/PQC.

**DEC-121:** Panel Visual de 3 Estados (Verde/Amarillo/Rojo), Bypass Nativo de WhatsApp y Desconexión Incondicional
- Implementación de interfaz de usuario de 1 botón con 3 estados en `src/pkg/core/web_ui.go`: 🔴 Rojo (Desconectado/Internet Normal), 🟡 Amarillo (Conectando/Negociando PQC), 🟢 Verde (Protegido/Túnel Activo).
- Arranque seguro desconectado: `ipvn7.exe` inicia por defecto en modo seguro sin secuestrar el proxy de Windows hasta que el usuario pulsa explícitamente el botón.
- Apertura automática de ventana tipo aplicación (`msedge.exe --app=http://127.0.0.1:7070 --window-size=440,680` o navegador por defecto) al iniciar la app.
- Bypass incondicional de WhatsApp y Meta en `ProxyOverride` (`*.whatsapp.net;*.whatsapp.com;*.fbcdn.net;*.facebook.com`) garantizando que WhatsApp Desktop y Web nunca se congelen ni pierdan conexión.
- Botones de acción directa: "🔄 Restaurar Red" y "❌ Salir y Cerrar" que fuerzan `ProxyEnable = 0` y apagan el proceso limpiamente.
- Script de rescate inmediato `scripts/reset_internet.ps1` y accesos directos de 1-clic creados en el Escritorio del usuario ("VPN I7" y "Restaurar Internet").
- ❌ Prohibido: Dejar proxies del sistema huérfanos al cerrar o matar la aplicación; prohibido bloquear el tráfico de aplicaciones de mensajería crítica como WhatsApp.

**DEC-122:** Rol O (Radar de Innovación en IA y Ecosistemas Agénticos) y Transporte Soberano para Agentes
- Formalización del Rol O en `agentes/ROLES.md` y `agentes/AGENTS.md` para vigilancia de IA de frontera (MCP, A2A v1.0, Edge LLMs, inferencia P2P).
- Posicionamiento de IPvN7 como la infraestructura de red soberana, P2P y post-cuántica (PQC Kyber/ML-KEM) para la comunicación segura entre agentes sin depender de servidores centrales.
- Filtro anti-gordura: prohibido incorporar librerías pesadas de ML al núcleo Go. La integración de IA se realiza exclusivamente mediante protocolos ligeros estándar (JSON-RPC, SSE, OpenAPI), inferencia desacoplada (Ollama/llama.cpp en la malla) o heurísticas deterministas embebidas.
- ❌ Prohibido: Inflar el binario con runtimes de Python/PyTorch; prohibido incorporar conceptos de IA teóricos que no generen tracción, usabilidad o velocidad medible en IPvN7.

**DEC-123:** Compatibilidad con Protocolo Agent2Agent (A2A) Este-Oeste y Manifiesto Canónico Agent Card
- Adopción de la especificación A2A v1.0.0 (AAIF / Linux Foundation) para la comunicación horizontal entre agentes autónomos.
- Exposición de manifiesto estándar `/.well-known/agent-card.json` en el servidor web de IPvN7, declarando identidad W3C DID soberana y capacidades de transporte de red PQC.
- Blindaje ZTNA mutuo: la autenticación de pares de agentes se valida mediante firmas Ed25519 y llaves de sesión Kyber/ML-KEM, erradicando los riesgos de suplantación de agentes y envenenamiento de contexto inherentes al estándar A2A comercial.
- ❌ Prohibido: Requerir servidores intermediarios en la nube para conectar agentes; prohibido aceptar peticiones A2A de identidades DID no emparejadas en la malla.

**DEC-124:** Arquitectura de Transporte P2P Zero-Copy para Enjambres de Inferencia Distribuida (BitNet/Exo)
- Posicionamiento del pipeline central zero-copy de IPvN7 (L0-L2) como la canalización física de ultra-baja latencia para la transferencia de activaciones neuronales y tensores entre nodos colaboradores de inferencia en el borde.
- Integración con modelos ternarios ultralivianos de 1-bit (BitNet b1.58 en ~0.4GB RAM) y mallas locales desacopladas (Exo Labs / Petals) mediante túneles L4 autenticados con Shadow DIDs y llaves post-cuánticas ML-KEM-768.
- Filtro anti-gordura: IPvN7 actúa estrictamente como la autopista de transporte rápido de red; prohibido compilar motores de inferencia o tensores dentro del binario del núcleo Go.
- ❌ Prohibido: Forzar el tráfico de inferencia local a través de servidores de nube externos; prohibido violar el invariante zero-copy en la transferencia de datagramas de tensores.

**DEC-125:** Autorización Adaptativa Graduada (IETF `draft-das-agentic-adaptive-authorization-00`) y Cadenas de Delegación Criptográficas PQC-APC
- Adopción del marco de autorización adaptativa para subsanar la brecha de gobernanza identificada en MCP y A2A, sustituyendo el modelo binario permitir/denegar por controles de ejecución graduados (*Ordinary*, *Escalated*, *Quarantine*).
- Implementación de la Cadena de Delegación Agéntica Criptográfica (Agentic Principal Chain - APC) ligada a DIDs soberanos (`did:ipvn7:...`), con atenuación de alcance en cada salto y firma Ed25519 con blindaje post-cuántico ML-KEM-768.
- Verificación matemática local sin consultas a APIs de nube ni consumo de tokens de LLM (latencia <10 µs en el borde).
- ❌ Prohibido: Permitir invocación de acciones de red escaladas (túneles, rutas, proxies) sin verificación criptográfica de la cadena de delegación; prohibido depender de servidores OAuth centralizados para autorizaciones agénticas.

**DEC-126:** Transporte de Ultra-Baja Latencia para Physical AI (VLA 100Hz) y Pasarela de Micropagos Agénticos (AP2/x402)
- Habilitación del pipeline UDP zero-copy (L0-L2) con timestamping de alta resolución (RFC 9221) para sincronización temporal de bucles de control motor de 50–100 Hz en enjambres robóticos (Vision-Language-Action).
- Integración de pasarela de mandatos de gasto agénticos (AP2 / x402) ligados a DIDs para compra/venta autónoma de prioridad en el enrutador Kleinberg, ancho de banda o cómputo en workers Edge AI sin entidades bancarias centralizadas.
- ❌ Prohibido: Compilar librerías de blockchains o pasarelas bancarias pesadas dentro de `ipvn7.exe`; la verificación se limita a firmas criptográficas Ed25519 de los mandatos en los datagramas.

**DEC-127:** Sandboxing de Plugins Agénticos en Pure-Go WebAssembly (Wazero/Reactor) sin Dependencias CGo
- Aprobación de la arquitectura de sandboxing para código y plugins agénticos en el borde basada estrictamente en runtimes WebAssembly en Go puro (`CGO_ENABLED=0`), como Tetratelabs Wazero y el motor nativo `pkg/wasm`.
- Soporte de componentes WASI 0.2 exclusivamente a través del adaptador oficial reactor de la Bytecode Alliance (`wasi_snapshot_preview1.reactor.wasm`), garantizando arranque en <1 ms y consumo <1 MB.
- ❌ Prohibido: Incorporar runtimes de WebAssembly basados en CGo o DLLs nativas de C/Rust (como Wasmtime-Go o Wasmer), para proteger la compilación estática universal y la portabilidad de `ipvn7.exe`.

**DEC-128:** Instalador Gráfico .EXE Autocontenido (0 Consolas, Embed Go) y Orquestador de Distribución Soberana Multiplataforma
- Implementación de instalador gráfico independiente en Go (`cmd/installer`) con subsistema nativo de Windows (`-H=windowsgui`) que erradica consolas negras o comandos terminales.
- Autocontención mediante directivas `//go:embed` integrando `ipvn7.exe` y `wintun.dll` en un único archivo ejecutable portable (`Instalador_VPN_I7.exe`).
- Autodetección de nodos físicos (PC Principal vs Notebook/Satélite) asignando puertos diferenciados (7777/7070 vs 7001/8080) para erradicar colisiones en mallas de red.
- Extensión del sub-rol operativo Rol K y formalización de la skill especializada `ipvn7-distribution-agent` para optimización continua de tamaño, auto-actualización in-place y control SemVer.
- ❌ Prohibido: Exigir comandos de consola al usuario común para instalar o iniciar en Windows; prohibido requerir compiladores o runtimes en el equipo de destino.

**DEC-129:** Plan de Ordenamiento Integral, Poda Preventiva Axioma III en WebUI y Armonización Tríada de Despliegue (Roles K, M, P)
- Poda preventiva del 15% en `src/pkg/core/web_ui.go` (de 343 a 294 líneas) erradicando la alerta preventiva de riesgo y elevando el Health Score global al 100%.
- Saneamiento y eliminación de archivos residuales de test (`data/1.txt`) y sincronización de referencias canónicas en `docs/README.md`.
- Armonización formal de la tríada de despliegue: Rol K (Arquitecto de Compilación Multiplataforma y Distribución Headless), Rol M (Centinela de Versiones y Actualización Atómica) y Rol P (Agente de Instalación Gráfica 1-Clic y Control Evolutivo).
- ❌ Prohibido: Tolerar archivos en zona preventiva (>320 líneas) sin aplicar poda previa; prohibido solapar responsabilidades entre compilación multi-target (Rol K) y empaquetado gráfico de escritorio (Rol P).

**DEC-130:** Sovereign Dynamic Egress Gateway y Salida Óptima Autoconmutable
- Transformación fractal de cada nodo en pasarela de salida a Internet soberana con selección dinámica del mejor gateway de la malla.
- Métrica compuesta de scoring (RTT, tasa BBR y pérdida de paquetes) con histéresis del 25% para evitar oscilaciones (anti-flapping).
- Clamping MSS obligatorio a 1220 bytes en el gateway (MTU rígido 1280B), autorización ZTNA por DID y fallback determinista cero-cortes a conexión directa local.
- ❌ Prohibido: Depender de servidores proxy centrales; prohibido dejar al usuario sin internet si el gateway remoto falla; prohibido violar el límite de 400 líneas.

**DEC-131:** Rol R (Orquestador HIL de Laboratorio Bi-Nodo y Pruebas Físicas) y Expansión del Consejo Agéntico a 18 Sub-Roles
- Formalización del Rol R en `agentes/ROLES.md` y `agentes/AGENTS.md` para automatizar la verificación física obligatoria en el laboratorio de 2 nodos (Nodo A PC Principal `192.168.1.198` $\leftrightarrow$ Nodo B Notebook `192.168.1.106:2201`), despliegue por SFTP/SSH, aserción cruzada de tráfico y simulación de fallas HIL (Reglas 2 y 5).
- Incorporación de directiva de "Economía de Compilación de Release" en Rol P, limitando la generación pesada de `Instalador_VPN_I7.exe` a hitos de release mayor o petición expresa del usuario para optimizar tiempo y consumo de tokens.
- Actualización de directivas en Rol Q para coordinar la agregación multipath (Channel Bonding) junto al Rol R.
- ❌ Prohibido: Dar por finalizada una tarea de red o peering sin aserción cruzada física en ambos nodos del laboratorio; prohibido gastar tokens regenerando instaladores completos ante cambios menores de código.

**DEC-132:** Magna Etapa de Consolidación Mundial y Universal (Ritmo Cardíaco Vital, Matriz 6-Arch y Producción Multi-Dispositivo)
- Adopción del monitor de Ritmo Cardíaco (ECG dinámico a 60 FPS y pulso lub-dub) en la UI minimalista como estándar biomórfico de feedback vital de red en tiempo real.
- Modulación del ritmo cardíaco en función de la telemetría física real (bytes transferidos) con asistolia visual (0 BPM) ante desconexión, erradicando autoengaños de interfaz.
- Generación de la matriz universal de binarios estáticos CGO_ENABLED=0 para 6 arquitecturas clave (Windows amd64/arm64, Linux amd64/arm64, macOS amd64/arm64) con manifiesto inmutable SHA-256.
- Empaquetado de producción en instalador gráfico 1-clic (.EXE con -H=windowsgui y 0 consolas) y bundles ZIP con autodetección de rol anti-colisión para redes multi-dispositivo.
- ❌ Prohibido: Simular ritmos cardíacos activos ante nodos desconectados; prohibido compilar instaladores con consolas negras visibles; prohibido romper el límite de 400 líneas.

**DEC-133:** Orquestación de Rendezvous Cloud Soberano y Señalización Global (Google Firebase RTDB)
- Adopción de Google Firebase Realtime Database (`vpni7-d5a78-default-rtdb`) como directorio de señalización efímero y canal de rendezvous global para descubrimiento de pares y agentes IA a través de CGNAT y redes celulares hostiles.
- Invariante de Desacoplamiento y Circuit Breaker P2P: Firebase actúa únicamente como faro de arranque (bootstrap rendezvous). En cuanto dos pares intercambian endpoints e inician el túnel directo Noise/PQC en UDP, el tráfico de datos fluye 100% P2P sin intermediación en la nube.
- Blindaje de Metadatos y Credenciales: Resguardo estricto del token administrativo en `config/firebase_credentials.json` (ignorado por Git); prohibido registrar claves privadas o tráfico de usuario en la base de datos externa.
- ❌ Prohibido: Crear dependencia permanente de servidores centrales de datos; prohibido exponer credenciales en la raíz o repositorios públicos; prohibido violar el límite de 400 líneas.

**DEC-134:** Poda Preventiva Axioma III en L1 (Firewall y SOCKS5 Gateway) y Restauración del 100% Health Score
- Poda preventiva en `src/pkg/l1/firewall.go` (de 385 a 318 líneas) unificando la lógica de evaluación ZTNA interna (`eval`) y optimizando la serialización de OpenMetrics.
- Poda y compactación preventiva en `src/pkg/l1/socks5_gateway.go` (de 374 a 319 líneas) compactando el pipeline del túnel y la gestión de tramas RFC 1928 / HTTP CONNECT.
- Erradicación total de alertas preventivas en la Magna Multi-Suite (`scripts/verify_ipvn7_standard.ps1`), certificando el **Health Score al 100% (ÓPTIMO / EXCELENCIA)** con 0 violaciones, 0 advertencias, 0 B/op y 100% tests en PASS.
- ❌ Prohibido: Tolerar acumulación de código accesorio en archivos entre 320 y 400 líneas sin aplicar el Principio de Poda Previa (Regla 6).

**DEC-135:** Sistema de Auto-Actualización Online Soberano 1-Clic y Manifiesto de Versiones
- Implementación de motor de actualización OTA online (`CheckOnlineUpdate`, `DownloadAndApplyUpdate`) en `src/pkg/core/version_manager.go`.
- Verificación criptográfica estricta de integridad mediante hash SHA-256 por arquitectura antes de cualquier sustitución de binarios.
- Manifiesto canónico descentralizado en `dist/version_manifest.json` y endpoints `/api/v1/update/check`, `/api/v1/update/apply`, `/api/v1/update/rollback`.
- Notificación e instalación visual 1-clic con banner dark glassmorphism en WebUI y sustitución atómica in-place con soporte de rollback automático.
- ❌ Prohibido: Descargar o ejecutar binarios sin validación previa de hash SHA-256; prohibido auto-actualizar de forma destructiva sin backup criptográfico (.bak) para rollback.

**DEC-136:** Rol U (Evangelizador & Maestro Interactivo de Adopción Soberana) y Sistema de Aprendizaje Progresivo (`guide/`)
- Creación y formalización del Sub-Rol U en `agentes/ROLES.md`, `agentes/AGENTS.md` y `agentes/skills/ipvn7-evangelist-interactive-agent/SKILL.md`.
- Implementación de la suite interactiva de adopción en `guide/` con 5 capas de revelación cognitiva (metáfora de autopista propia, casos de uso en vida/dispositivos/trabajo/comunidad, simulador cuántico y de corte de red Kleinberg, comparativa factual y onboarding 1-clic).
- Autocontención 100% offline con Vanilla HTML5/CSS/JS, estética biomórfica cyberpunk glassmorphism a 60 FPS, servible por el nodo en `/guide`.
- ❌ Prohibido: Presentar al usuario conceptos crípticos sin metáforas cotidianas; prohibido depender de CDNs externas para la guía interactiva; prohibido violar el límite de 400 líneas.

**DEC-137:** Alineación Factual de Auditoría Externa, ML-KEM-768 FIPS 203 Real y Taxonomía de 4 Estados
- Sustitución total de la simulación de intercambio post-cuántico por el estándar oficial NIST FIPS 203 ML-KEM-768 nativo (`crypto/mlkem`) de la biblioteca estándar de Go. Criptograma real de 1088 bytes con decapsulación e Implicit Rejection canónico.
- Corrección de desbordamiento en BufferPool: los paquetes que superen tramas Jumbo (>64 KB) se asignan dinámicamente sin provocar pánico.
- NAT Hole Punching verificado con ACK físico: prohibido declarar éxito si el par no responde en el socket UDP real dentro del timeout.
- Sincronización thread-safe en telemetría de ring buffer y delimitación factual del PacketPacer (Token Bucket estándar, sin simulación BBR).
- Desacoplamiento formal del Núcleo Mínimo I7 (L0 Identidad/PQC, L1 Kleinberg/UDP, L2 Telemetría básica) respecto de los satélites de aplicación (SmartComponents, TV, Impresoras, Chat, Kademlia DHT escalable).
- Adopción estricta de la Taxonomía de 4 Estados: DEMOSTRADO FÍSICAMENTE, IMPLEMENTADO, EXPERIMENTAL y NO DEMOSTRADO en documentación y código.
- ❌ Prohibido: Utilizar HMAC o SHA256 simulando KEM; prohibido declarar éxito de NAT sin recepción de datagramas reales; prohibido pánico por tamaños no estándar en buffers.

**DEC-138:** Cierre Integral de Auditoría Externa: ZTNA en Data Path (Default-Deny), Handshake 1-RTT PQC Wire Framing (1280B MTU), Discovery Silencioso y Cero-Copia Real
- Integración directa del Firewall ZTNA en el bucle principal de recepción (`src/cmd/ipvn7/main.go`): Default-Deny activo tanto en paquetes de datos como en descubrimiento. Eliminada la auto-autorización en balizas de discovery de `src/pkg/l1/autonomous_discovery.go`.
- Negociación e instauración del canal de datos cifrado mediante Handshake 1-RTT PQC (`src/pkg/l1/session_manager.go`): intercambio híbrido ML-KEM-768 FIPS 203 + X25519 empaquetado binario (`EphemeralX25519` + `Salt` + `PQCCiphertext` = 1136 bytes), respetando estrictamente el MTU canónico de 1280 bytes. Cifrado simétrico ChaCha20-Poly1305 para todo payload de datos.
- Erradicación de la copia simulada (`pktBuf = append(...)`) en `main.go`, decodificando directamente el slice `rawBuf[:n]` con validación O(1) de anti-replay por ventana deslizante.
- Saneamiento de ruido de red: desactivado el broadcast periódico ruidoso multi-puerto por defecto ("la red escucha, no grita").
- Armonización técnica: Go 1.24/1.26 unificado en `go.mod` y CI GitHub Actions; eliminadas afirmaciones inexactas de "lock-free" en telemetría; arquitectura Kleinberg normalizada a 12 anillos concéntricos con K-buckets.
- ❌ Prohibido: Admitir tráfico de datos no cifrado o con DIDs no autorizados en el datapath; prohibido rebasar el MTU de 1280B en handshakes KEM; prohibido auto-autorizar DIDs sin intervención explícita.

**DEC-139:** Vinculación Criptográfica Estricta de Identidad en HandshakeInit, Aislamiento Anti-Replay L1 en Datapath y Suite Adversarial Integral
- Vinculación criptográfica obligatoria entre `SourceDID` y la clave pública Ed25519 en `CreateHandshakeInitPacket` mediante firma digital sobre todo el datagrama; `HandleHandshakeInitPacket` valida `PublicKeyFromDID` y `VerifyPacketSignature` ANTES de cualquier decapsulación KEM o autorización en el firewall ZTNA, erradicando la suplantación de identidad (Hallazgo 1).
- Empaquetado canónico X-Wing CFRG de 1120 bytes (`EphemeralX25519` 32B + `ML-KEM-768` 1088B) con derivación determinista de sal, garantizando que el datagrama completo con firma Ed25519 respete estrictamente el MTU determinista de 1280 bytes.
- Sustitución del filtro global L0 por `l1.NewAntiReplayFilter(nil)` en `src/cmd/ipvn7/main.go`, indexando por `originDID:SessionID` con ventana deslizante de 1024 bits y normalización temporal contra interferencias cross-peer y cross-session (Hallazgo 2).
- Manejo fatal de fallos en generación de claves PQC en `main.go`: aborto inmediato (`os.Exit(1)`) sin continuar con claves nulas (Hallazgo 6).
- Validación estricta de respuestas 1-RTT en `HandleHandshakeRespPacket`: comprobación obligatoria de `DestDID == localDID`, correspondencia con `pendingSession` y expiración a 60s (Hallazgo 7).
- Implementación de la suite de pruebas adversariales completas (Tests A hasta H) en `src/pkg/l1/session_adversarial_test.go` (100% PASS).
- ❌ Prohibido: Autorizar DIDs o registrar sesiones ante HandshakeInit sin firma Ed25519 verificada; prohibido mezclar secuencias de distintas sesiones o pares en filtros globales; prohibido silenciar fallos de entropía o generación criptográfica.

**DEC-140:** Desacoplamiento de CI con Build Tags, Aislamiento ZTNA en Roaming, Cierre de Superficie WebUI (127.0.0.1 y RBAC) y Honestidad Criptográfica Reticular
- **Desacoplamiento de CI:** Incorporación del build tag `//go:build windows && installer` en `src/cmd/installer/main.go` y flag `-tags installer` en `scripts/build_installer.ps1`, garantizando que `go vet ./...` y `go test ./...` pasen en limpio en checkout sin depender de binarios precompilados.
- **Aislamiento ZTNA en RoamingUpdate:** Erradicación de `firewall.AuthorizeDID` automático en `src/cmd/ipvn7/main.go` y `src/pkg/core/pipeline_stages.go`. La firma Ed25519 demuestra autenticidad, pero la autorización exige evaluación previa de la política ZTNA local (Default-Deny). Validado con suite `src/pkg/l1/roaming_ztna_test.go`.
- **Cierre de Superficie WebUI:** Bind HTTP restringido a `127.0.0.1` por defecto. Erradicación de CORS `*` en panel administrativo (fijado a loopback). Control de acceso RBAC (`checkAdminAuth`) en operaciones críticas (`/vpn/connect`, `/vpn/disconnect`, `/vpn/exit`, `/update/*`) y mitigación estricta de SSRF en `handleUpdateCheck` limitando manifiestos a fuentes oficiales.
- **Honestidad Criptográfica:** Renombramiento de pseudónimos ML-DSA a identificadores honestos `ExperimentalPQCIdentity` y `ExperimentalSigSize`, documentando que la firma formal de producción descansa en Ed25519 (RFC 8032) y que el vector reticular adjunto es un compromiso experimental no certificado.
- ❌ Prohibido: Auto-autorizar DIDs al recibir paquetes de red firmados; prohibido exponer WebUI administrativa en `0.0.0.0` sin autenticación; prohibido permitir URLs arbitrarias en actualización; prohibido denominar ML-DSA a vectores derivados por HMAC.


**DEC-141:** Consagración de docs/auditoria externa.md como Única Fuente de Verdad Suprema, Modelo de 7 Agentes Especializados y Plan Canónico de 13 Fases
- `docs/auditoria externa.md` se establece como la única fuente de verdad absoluta del proyecto. Cualquier diseño previo queda supeditado a ella.
- Reestructuración del modelo agéntico: erradicada la acumulación de 21 sub-roles, adoptando formalmente los 7 Agentes Especializados de Auditoría Externa (Arquitecto, Seguridad, Core, Testing, CI/Build, Performance, Documentación) con flujo de compuerta estricto (Arquitecto -> Plan -> Programador -> Tester -> Seguridad -> CI -> Merge) y regla de commits unifuncionales.
- Adopción irrevocable de la Regla de Oro: "NO INVENTAR" y de la Taxonomía Factual de 5 Estados (HECHO, TESTEADO, MEDIDO, NO IMPLEMENTADO, EXPERIMENTAL), erradicando auto-declaraciones prematuras de cierre de auditoría.
- Formalización del Baseline Factual en `docs/BASELINE.md` (Fase 0) y los 8 Contratos del Core en `docs/core/*.md` (Fase 7).

**DEC-142:** Consagración de docs/FUENTE_DE_VERDAD.md como Nueva Fuente de Verdad Canónica y Archivo Histórico de la Auditoría Preliminar ee56f89
- `docs/FUENTE_DE_VERDAD.md` se establece como la nueva y definitiva fuente de verdad absoluta del proyecto, superando la naturaleza conversacional transitoria de la auditoría preliminar del commit `ee56f89`.
- Se traslada y archiva formalmente `docs/auditoria externa.md` a `docs/audit/AUDITORIA_EXTERNA_HISTORICA_ee56f89.md` para garantizar estricta trazabilidad de la auditoría externa.
- Se consolidan de forma permanente los 7 Invariantes y Axiomas Inmutables del Núcleo I7, la taxonomía epistemológica de 5 estados, la matriz de resolución de hallazgos (AUD-01 a AUD-07) y el modelo de autogobernanza liderado por FrondaBrick_01.
- ❌ Prohibido: Volver a usar volcados conversacionales de chat como fuentes rectoras; prohibido alterar los 7 axiomas inmutables sin consenso constitucional.

---

**Última actualización:** 2026-10-01  
**Total decisiones activas:** 60 (DEC-079 → DEC-142)  
**Estado:** Todas INVIOLABLE

