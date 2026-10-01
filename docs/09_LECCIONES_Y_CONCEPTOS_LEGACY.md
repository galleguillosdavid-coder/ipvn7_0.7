# Lecciones Aprendidas y Destilación de Conceptos de Versiones Anteriores (IPv7 v0.1 - v0.6 e ipvn7 vX)

## 1. El Diagnóstico Histórico: La Trampa de la Complejidad
Las versiones preliminares de IPv7 (`0.1` a `0.6` e `ipvn7 vX`) concibieron una arquitectura ambiciosa y visionaria para un Sistema Operativo de Redes Soberano (NOS). Sin embargo, sufrieron del **síndrome de sobre-ingeniería prematura**:
1. **Explosión de Líneas de Código:** Más de 15,000 líneas dispersas en decenas de paquetes concurrentes antes de validar la estabilidad del enlace P2P elemental.
2. **Dependencias CGO y Motores Pesados:** La inclusión de motores de grafos en C++ (KùzuDB), stacks web masivos y dependencias dinámicas rompieron la portabilidad en Windows y Linux, haciendo imposible compilar binarios estáticos universales.
3. **Simulación de Estado y Micro-Features Fragmentadas:** Módulos para robótica (MAVLink), telemetría de vehículos (OBD-II, CAN Bus), agentes LLM embebidos y pasarelas de impresión compitiendo por memoria y CPU en el mismo ejecutable.

La **Revolución del Núcleo Mínimo (DEC-107)** salvó el proyecto reduciéndolo a ~1,700 líneas de Go puro (`CGO_ENABLED=0`), con el Axioma III (límite estricto de $\le 400$ líneas por archivo) y el Invariante Zero-Copy (0 B/op, 0 allocs/op).

---

## 2. Las 10 Joyas Arquitectónicas Rescatadas
A pesar de la sobre-ingeniería de las implementaciones originales, los fundamentos matemáticos y conceptuales desarrollados en las especificaciones históricas (`Plan Maestro NOS`, `Whitepaper Competitivo`, `Especificaciones Técnicas v1-v4`) contienen joyas de ingeniería de vanguardia. A continuación se destilan las 10 ideas maestras para su adopción modular:

### 1. Señalización Ciega Efímera (EBRA — Ephemeral Blind Rendezvous Anchors)
* **Concepto:** Descubrimiento global a través de canales no confiables (Firebase, STUN, HTTP) sin filtrar identidades.
* **Mecanismo:** Derivación de tópicos de Conocimiento Cero por época: $\text{TopicID} = \text{SHA-256}(\text{EpochHour} \parallel \text{RingDegree} \parallel \text{NetworkSeed})$.
* **Aplicación en v0.7:** Conectores en la nube (como Google Firebase RTDB) operan como faros temporales con auto-destrucción (*Consume-and-Burn*) y desconexión inmediata en cuanto se verifica el túnel directo P2P (Circuit Breaker).

### 2. Trama Canónica de 1280B Fijos y Acolchado Anti-DPI (Sphinx / Onion)
* **Concepto:** Erradicar la fragmentación en el perímetro y neutralizar el análisis de tráfico por longitud de paquete (DAITA).
* **Mecanismo:** Todo datagrama se ajusta exactamente al MTU canónico IPv6 (1280 bytes). Si el payload es menor, se aplica relleno determinista; si es mayor, se segmenta en la capa superior sin fragmentación IP.
* **Aplicación en v0.7:** Implementado en L0/L1 mediante serialización CBOR determinista (RFC 8949) y MTU estricto de 1280B.

### 3. Sistema de Nombres Descentralizado dDNS (Petnames)
* **Concepto:** Reemplazo soberano del sistema DNS jerárquico (ICANN) mediante alias locales legibles (`notebook.ipv7` $\to$ `did:ipvn7:...`).
* **Mecanismo:** Relatividad contextual. Cada usuario asigna petnames en su libro de direcciones local, complementado con resolución criptográfica de claves públicas sin servidores autoritativos.
* **Aplicación en v0.7:** El Resolver DNS Soberano (`src/pkg/l1/dns_resolver.go`, DEC-130) ya captura consultas `.ipv7` locales y enruta el resto por la malla cifrada sin fugas hacia el gateway.

### 4. Calidad de Servicio (QoS) y Desafíos PoW Dinámicos Anti-DDoS
* **Concepto:** Prevención de denegación de servicio y amplificación en el espacio de usuario sin firewalls de hardware dedicados.
* **Mecanismo:** Token Bucket clasificado en 3 clases (`Control`, `Interactive`, `Bulk`). Si un par sobrepasa la tasa contratada, debe resolver un micro-acertijo de Prueba de Trabajo (PoW SHA-256 de 16 bits) para que sus paquetes sean procesados.
* **Aplicación en v0.7:** Ya integrado en `src/pkg/l1/qos.go` y la política ZTNA de L1.

### 5. Economía de Tránsito Recíproco (Tit-for-Tat)
* **Concepto:** Incentivar el reenvío de tráfico en la malla de mundo pequeño de Kleinberg sin requerir criptomonedas ni tokens especulativos.
* **Mecanismo:** Contabilidad local de bytes transmitidos vs recibidos. Clasificación en 4 tiers (`Priority`, `Normal`, `Best-Effort`, `Throttled`). Margen de cortesía de 1 MB para bootstrap de nodos nuevos.
* **Aplicación en v0.7:** Métrica incorporable en la selección de rutas greedy de Kleinberg (`src/pkg/l1/routing.go`).

### 6. Almacén Inmutable DAG & Colas Tolerantes a Desconexión (DTN)
* **Concepto:** Entrega garantizada de mensajes y datos en redes intermitentes o aisladas (satélites, minas, zonas remotas).
* **Mecanismo:** Bloques direccionados por contenido (`cid:ipvn7:<sha256>`) con firmas Ed25519 y enlaces causales a padres. Almacenamiento *Store-and-Forward* que sincroniza deltas al reencontrarse los pares.
* **Aplicación en v0.7:** Modelo idóneo para sincronización asíncrona de chat y telemetría histórica como satélite liviano.

### 7. Emparejamiento Fuera-de-Banda (SAS RFC 6189 — Short Authentication String)
* **Concepto:** Vinculación 1-clic a prueba de Hombre-en-el-Medio (MITM) sin intercambiar certificados complejos.
* **Mecanismo:** Derivación canónica de 6 dígitos numéricos o 4 emojis visuales (`🎷 🎨 🔮 🎁`) a partir de la firma compuesta de claves públicas. Los usuarios verifican la coincidencia visual en 3 segundos.
* **Aplicación en v0.7:** Incorporable a la UI minimalista de 1 clic para confirmación instantánea de nuevos pares.

### 8. Red de Confianza Transitiva (Web-of-Trust — WoT)
* **Concepto:** Descubrimiento y admisión de pares sin Autoridades de Certificación (CAs) centralizadas.
* **Mecanismo:** Firmas mutuas de aval. Búsqueda en anchura (BFS) sobre el grafo de contactos con atenuación de reputación por cada salto social (máximo 3 saltos).
* **Aplicación en v0.7:** Permite que dos nodos que no se conocen se comuniquen con confianza si comparten un par amigo verificado.

### 9. Camuflaje Anti-DPI RFC 8446 (TLS 1.3 / Puerto 443) y Modo Proxy Userspace
* **Concepto:** Operar en entornos corporativos o países con cortafuegos hostiles (censura de paquetes VPN estándar como WireGuard o IPsec).
* **Mecanismo:** Encapsulado de datagramas IPvN7 en registros TLS 1.3 sintéticos (`0x17 0x03 0x03 ApplicationData`) sobre TCP/UDP puerto 443. En modo sin privilegios, opera como SOCKS5 (`:10807`) y HTTP CONNECT (`:10808`).
* **Aplicación en v0.7:** Pasarelas de salida soberanas (DEC-130) y transporte resiliente en redes restringidas.

### 10. Resiliencia Post-Apagón con Jitter Descorrelacionado
* **Concepto:** Prevenir el colapso por estampida (*Thundering Herd Problem*) cuando millones de nodos se reconectan tras una caída eléctrica masiva o corte de fibra.
* **Mecanismo:** Backoff estocástico descorrelacionado: $T_{i+1} = \min(T_{\text{max}}, \, \text{Uniforme}(T_{\text{base}}, \, T_{i} \times 3))$.
* **Aplicación en v0.7:** Algoritmo ya incorporado en el bucle de reconexión de pares de L1 (`discovery.go`).

---

## 3. Matriz de Filosofía: Antes vs Ahora

| Dimensión | Versiones Legacy (v0.1 - v0.6 / vX) | Versión Actual (IPvN7 v0.7+) |
|---|---|---|
| **Complejidad de Código** | >15,000 líneas, múltiples crates/paquetes | ~1,700 líneas, monorepo compacto |
| **Límite por Archivo** | Archivos de 1,000 a 3,000 líneas | **Axioma III: $\le 400$ líneas estrictas** |
| **Dependencias Externas** | KùzuDB, CGO, librerías pesadas | **Go puro (`CGO_ENABLED=0`), 0 dependencias runtime** |
| **Alocaciones en Tránsito** | Múltiples copias en memoria por paquete | **Invariante Zero-Copy: 0 B/op, 0 allocs/op** |
| **Canal de Señalización** | Servidores dedicados complejos o nubes dispersas | **P2P nativo + Faro efímero Google Firebase RTDB** |
| **Experiencia de Usuario** | Paneles crípticos con decenas de métricas | **1-clic VPN I7 + Monitor biomórfico de Ritmo Cardíaco** |
| **Verificación** | Mocks en memoria y supuestos teóricos | **Pruebas físicas bi-nodo (HIL) en hardware real** |

---

## 4. Regla de Oro para Futuras Extensiones
Cualquier concepto rescatado de las versiones anteriores debe pasar obligatoriamente por el **Filtro Antihumo de 4 Pasos** antes de escribirse en código:
1. ¿Se puede implementar en Go puro sin dependencias CGO ni bibliotecas pesadas?
2. ¿Mantiene el presupuesto de $\le 400$ líneas y el Invariante Zero-Copy (0 B/op)?
3. ¿Puede verificarse físicamente entre dos nodos reales (PC y Notebook) sin simulaciones?
4. ¿Aporta tracción real al usuario final en menos de 1 clic o 1 comando?

Si la respuesta a alguna de estas preguntas es "No", el concepto debe permanecer como especificación de diseño y jamás contaminar el núcleo de ejecución.
