# METODOLOGÍA DE INVESTIGACIÓN, VIGILANCIA Y EVOLUCIÓN DETERMINISTA EN IPVN7

> **"Más inteligente que documentar tareas es documentar cómo descubrirlas, sobre qué fuentes basarse y cómo absorber el conocimiento del mundo exterior con control, rigor físico y determinismo matemático."**

---

## 1. El Propósito del Marco Metodológico
El objetivo supremo de ipvn7 es ser la red soberana zero-friction más rápida, invisible, intuitiva e indestructible del mundo, sustentada en un radar permanente de vigilancia e inteligencia tecnológica exógena e investigaciones por iniciativa propia de cualquier tema necesario o voluntario, para anticipar y superar cualquier amenaza o limitación potencial. Para liderar frente a corporaciones y estándares consolidados (WireGuard, Tailscale, Tor, Cloudflare, MASQUE), ipvn7 no inventa ocurrencias en aislamiento; opera mediante un **sistema determinista de descubrimiento, vigilancia tecnológica activa, indagación voluntaria proactiva y absorción de fuentes externas primarias**.

---

## 2. Los 6 Ejes Temáticos Fundacionales (Sobre qué investigar)

Toda nueva investigación, búsqueda externa o propuesta de desarrollo debe encuadrarse estrictamente en uno de estos 6 pilares:

```
                  ┌─────────────────────────────────────────┐
                  │      IPVN7 SOVEREIGN NETWORK OS         │
                  └────────────────────┬────────────────────┘
          ┌──────────────┬─────────────┼─────────────┬──────────────┐
          ▼              ▼             ▼             ▼              ▼
   [1. Criptografía [2. Rendimiento [3. Resiliencia [4. Enrutamiento [5. UX Radical [6. Hardware &
     PQC & FIPS]    Zero-Copy]      WAN Hostil]     Sin Servidor]   Zero-Friction]   Vehículos]
```

1. **Criptografía Post-Cuántica e Identidad Soberana (L0/L1):**
   * *Fuentes Primarias:* NIST CSRC (FIPS 203 ML-KEM, FIPS 204 ML-DSA, SP 800-208), IETF CFRG drafts (X-Wing KEM, RFC 10024), criptografía de curva elíptica Ed25519/X25519, pruebas ZKP compactas y Micro-PoW.
   * *Pregunta Guía:* *"¿Cómo garantizamos confidencialidad perfecta post-cuántica sin superar el MTU de 1280B?"*

2. **Rendimiento Físico y Asignación Cero (Zero-Copy Sockets L0/L1):**
   * *Fuentes Primarias:* Windows IOCP / RIO, Linux `io_uring` y `eBPF/XDP`, buffers en anillo pre-alocados (`sync.Pool`), algoritmos de pacing (BBRv3) y control de contienda multi-core.
   * *Pregunta Guía:* *"¿Cómo conmutamos paquetes a velocidad de cable con 0 B/op y menos de 40 ns/op en userspace?"*

3. **Resiliencia en Redes Hostiles y Traspaso NAT (L1/L2):**
   * *Fuentes Primarias:* IETF RFCs (STUN RFC 5389, ICE RFC 8445, MASQUE RFC 9298 CONNECT-UDP, TLS 1.3 RFC 8446), camuflaje anti-DPI y relaying distribuido.
   * *Pregunta Guía:* *"¿Cómo atravesamos un CGNAT simétrico móvil o un firewall restrictivo sin servidores centrales?"*

4. **Enrutamiento Descentralizado y Malla Auto-Reparable (L1/L2):**
   * *Fuentes Primarias:* Grafos navegables de Kleinberg (12 anillos concéntricos), S/Kademlia resistente a Sybil, CRDTs de convergencia libre de conflictos y gossip anti-amplificación.
   * *Pregunta Guía:* *"¿Cómo encontramos la ruta óptima en $O(\log^2 N)$ saltos sin que ningún nodo conozca la topología entera?"*

5. **Experiencia de Usuario Radical (UX Humana y Cero Fricción):**
   * *Fuentes Primarias:* Heurísticas de Nielsen Norman Group, estándares W3C/WCAG, psicología de interfaces sin jerga técnica, pruebas con Agentes Persona (Doña Carmen, Carlos, Elena) y conexión en 1 clic ("VPN I7").
   * *Pregunta Guía:* *"¿Puede cualquier persona conectar sus dispositivos o transmitir a su Smart TV en 5 segundos sin ver términos técnicos?"*

6. **Hardware-in-the-Loop, Robótica y Sistemas Operativos:**
   * *Fuentes Primarias:* MAVLink v2 (drones y robótica), CAN Bus ISO 11898 / OBD-II / SAE J1939 (vehículos), llamadas nativas Win32/POSIX para control físico de hardware.
   * *Pregunta Guía:* *"¿La acción física se ejecutó demostrablemente sobre el hardware de destino con un camino de error falsable?"*

---

## 3. Protocolo Permanente de Inteligencia y Fuentes Externas (Radar de Frontera)

Para asegurar que ipvn7 incorpore constantemente los avances más recientes de la ingeniería global:

### A. Matriz de Fuentes Primarias Obligatorias
* **Estándares y Borradores IETF / CFRG:** IETF Datatracker (`datatracker.ietf.org`), RFCs canónicos y drafts activos de los grupos de trabajo TLS, MASQUE, QUIC, CBOR y CFRG.
* **Estándares Criptográficos NIST:** Portal CSRC NIST (`csrc.nist.gov`), reportes de estandarización PQC y parámetros oficiales de matrices reticulares.
* **Repositorios de Código de Vanguardia:** Implementaciones de referencia auditadas (`golang/go`, `wireguard-go`, `tailscale/tailscale`, `quic-go`, `torproject`, `tor-spec`, `cloudflare/boring`).
* **Investigación Abierta y Vulnerabilidades:** Repositorios de pre-prints IACR Cryptology ePrint Archive (`eprint.iacr.org`) y bases de datos CVE/NVD para evasión y resiliencia proactiva.

### B. Procedimiento de Indagación Activa
1. **Disparador Exógeno:** Antes de diseñar o refactorizar cualquier protocolo, KDF, transporte o algoritmo, el Agente realiza una consulta activa a fuentes externas (`search_web`).
2. **Cotejo Doble:** Ninguna afirmación técnica o benchmark se asume válida sin contrastarla contra al menos dos fuentes técnicas independientes o código fuente abierto verificable.
3. **Filtro Anti-Vaporware:** Descartar anuncios de marketing corporativo que no incluyan especificación formal, RFC o código reproducible.

---

## 4. El Embudo de Filtrado Determinista (Filtro Antihumo en 4 Pasos)

Toda tecnología descubierta en el exterior debe superar 4 compuertas antes de admitirse en el código:

```
[ Idea / Paper / RFC / Repo ]
          │
          ▼
 1. ¿Es Físicamente Verificable? ──► [NO] ──► DESCARTAR (Prohibido Teatro de Simulación)
          │ [SÍ]
          ▼
 2. Algoritmo de 5 Pasos ──────────► [SOBRE-INGENIERÍA] ──► PODAR 30-50%
          │ [SIMPLIFICADO]
          ▼
 3. Invariante Zero-Copy & 400L ───► [ALOCA MEMORIA / >400L] ──► REDISEÑAR A ZERO-ALLOC
          │ [0 B/op & <= 400L]
          ▼
 4. Test de Falsabilidad HIL ──────► [NO DETECTA FALLA FÍSICA] ──► RECHAZAR
          │ [PASA CAMINO DE ERROR]
          ▼
[ APROBADO: Pasa a Ficha Técnica RES-XXX y Tarea TASK-XXX ]
```

1. **Filtro 1 — Realismo Físico (Regla 1 & 2):** Si la tecnología depende de servicios cerrados en la nube, simulación con `time.Sleep` o mocks fingidos, queda descartada.
2. **Filtro 2 — Algoritmo de 5 Pasos (Regla 0):** Cuestionar el requisito. Podar capas superfluas. Si se resuelve con 40 líneas nativas de Go sin dependencias pesadas, ése es el camino.
3. **Filtro 3 — Presupuesto Zero-Copy y Axioma III (Regla 6 & 11):** Debe encajar en archivos $\le 400$ líneas y garantizar **0 B/op** en la ruta crítica.
4. **Filtro 4 — Falsabilidad Demostrable (Regla 16):** Debe probarse el camino de error físico antes del de éxito (`Host Unreachable` / `Connection Refused`).

---

## 5. Ciclo Operativo de Investigación a Código (El Método de 4 Pasos)

1. **Paso A — Sondeo Focalizado (`search_web`):** Búsqueda de borradores IETF, NIST o código en repositorios de frontera.
2. **Paso B — Ficha Técnica Factual (`docs/research/RES-XXX.md`):**
   - Fuente Externa Primaria (URLs, RFCs, papers).
   - Análisis comparativo frente a ipvn7 y competidores (WireGuard, Tailscale, Tor).
   - Decisión de adopción, adaptación o rechazo.
3. **Paso C — Registro Arquitectónico (`docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md`):**
   - Entrada formal `DEC-XXX` con Contexto, Anti-Patrón Prohibido y Decisión Invariable.
4. **Paso D — Tarea Atómica en Cola (`.agents/AUTOTASKS.md`):**
   - Registro de `TASK-XXX` validable mediante `scripts/verify_ipvn7_standard.ps1`.
