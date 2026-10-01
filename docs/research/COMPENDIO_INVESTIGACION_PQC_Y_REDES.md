# COMPENDIO CANÓNICO DE INVESTIGACIÓN: PQC, TOPOLOGÍA Y TRANSPORTE SOBERANO

> **Estado:** CONSOLIDADO / ESTADO DEL ARTE Y VECTORES DE INVESTIGACIÓN  
> **Fecha de Consolidación:** Octubre 2026  
> **Norma Rectora:** [`docs/FUENTE_DE_VERDAD.md`](../FUENTE_DE_VERDAD.md)  
> **Alcance:** Síntesis técnica de las fichas exógenas RES-001 a RES-018 y Memoria Tecnológica  

---

## 1. CRIPTOGRAFÍA POST-CUÁNTICA Y BLINDAJE DE ENLACE

### RES-001 / RES-003: Estándares KEM Híbridos (X-Wing PQC y ML-KEM-768)
* **X-Wing IETF (`draft-ietf-cfrg-xwing`):** Combinación de X25519 con ML-KEM-768 utilizando SHA3-256 canónico y etiqueta de dominio `\.//^\`. Garantiza seguridad clásica y cuántica simultánea.
* **Limitaciones de WireGuard tradicional:** Carece de PQC nativo acoplado al handshake inicial. IPVN7 integra encapsulación KEM dentro del presupuesto determinista de MTU de 1280 bytes.
* **Firmas Compuestas (RES-012):** Ed25519 para producción determinista inmediata combinable con esquemas reticulares experimentales clasificados transparentemente como `ExperimentalPQCIdentity`.

---

## 2. RESILIENCIA DE RED, NAT Y TOPOLOGÍA SMALL-WORLD

### RES-003 / RES-004: Perforación de NAT Simétrico y Predicción de Puertos
* Análisis de CGNAT celular mediante escaneo de deltas deterministas y asistencia de retransmisión voraz.
* Integración con STUN RFC 5389 para mapeo de endpoints y apertura de agujeros simultáneos en sockets UDP.

### RES-004 / RES-008: Enrutamiento Kleinberg y Defensa Sybil (sKademlia)
* Malla estructurada en 16 anillos concéntricos con métrica de distancia XOR.
* Saltos voraces $O(\log N)$ con retención acotada de pares en memoria (120 pares) para eludir envenenamiento de tablas.

---

## 3. RENDIMIENTO EXTREMO, ZERO-COPY Y KERNEL I/O

### RES-002 / RES-011: Búferes Circulares Lock-Free (Disruptor) y RIO vs IOCP
* Eliminación de contienda mediante ring-buffers atómicos preasignados en tres niveles (64B, 1500B, 64KB).
* Métrica de datapath certificada: **0 B/op** y **0 allocs/op**, alcanzando latencias de procesamiento menores a 20 ns/op en CPU física.

### RES-006 / RES-010: Datagramas para Robótica en Tiempo Real (RFC 9221) y WoL Seguro
* Soporte para datagramas de baja latencia con marcapasos determinista y retransmisión segura de tramas Wake-on-LAN protegidas por HMAC.

---

## 4. INMUNIDAD DE TRÁFICO Y ANONIMATO

### RES-005 / RES-009: Defensa Contra Análisis de Tráfico IA (DAITA) y ECH TLS 1.3
* Camuflaje de datagramas bajo encabezados TLS 1.3 con Encrypted Client Hello (ECH, RFC 9849) para evadir inspección profunda de paquetes (DPI).
* Pacing determinista y padding para mitigar fingerprinting de paquetes mediante aprendizaje automático.

---

## 5. FRONTERAS AGÉNTICAS Y MODELO DE COMPONENTES SATELITALES

### RES-013 / RES-014 / RES-015 / RES-018: Redes Agénticas MCP, A2A y WASI 0.2
* Desacoplamiento total del Core I7 frente a la semántica de agentes: la inteligencia artificial, modelos MCP y contratos de delegación se conectan exclusivamente a través del Smart Component Gateway.
* Aislamiento seguro de micro-servicios mediante WebAssembly WASI 0.2 sin tocar el núcleo de red en Go.
