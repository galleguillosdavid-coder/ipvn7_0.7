# ipvn7 Network OS — Arquitectura Canónica del Núcleo I7

**Versión:** 0.7.0  
**Paradigma:** Núcleo Mínimo Universal desacoplado (Pure Go, `CGO_ENABLED=0`) con Adaptadores de Transporte y Taxonomía Factual de 4 Estados.

---

## 1. Modelo Arquitectónico Desacoplado

En estricta conformidad con la auditoría externa independiente, la arquitectura de IPvN7 se organiza en tres estratos estrictamente delimitados:

```text
                     APLICACIONES Y PERFILES
   (Chat E2EE · Archivos · Video · Sensores · Agentes · Web UI)
                               │
                       ┌───────▼───────┐
                       │   PROFILES    │
                       └───────┬───────┘
                               │
══════════════════════════════════════════════════════════════════
                    NÚCLEO MÍNIMO I7 (CORE)
══════════════════════════════════════════════════════════════════
 1. Identity   (DID soberano criptográfico did:ipvn7:<pubkey>)
 2. Object     (Carga útil tipada universal con metadatos)
 3. Container  (Trama de alambre compacta MTU <= 1280B)
 4. Session    (Canal cifrado autenticado 1-RTT PQC FIPS 203)
 5. Channel    (Flujos lógicos multiplexados)
 6. Path       (Vector determinista de enrutamiento y saltos)
 7. MTU        (Invariante de 1280 bytes deterministas)
 8. Integrity  (Autenticación Poly1305 / AEAD en tiempo constante)
 9. Routing    (Primitiva de decisión de reenvío por distancia XOR)
 10.Capability (Tokens ZTNA verificables de autorización de recursos)
══════════════════════════════════════════════════════════════════
                               │
                    ADAPTADORES DE TRANSPORTE
                               │
             UDP (I7UDPAdapter)  /  TUN (Wintun / Linux)
                               │
                          RED FÍSICA
```

---

## 2. Taxonomía Factual de 4 Estados

Ninguna funcionalidad se declara "production-ready" ni "certificada" sin evidencia empírica verificable. La clasificación oficial es:

| Estado | Definición | Componentes |
| :--- | :--- | :--- |
| **🟢 DEMOSTRADO FÍSICAMENTE** | Probado en hardware/sockets reales entre equipos o mediante la biblioteca estándar de Go. | • Sockets UDP reales en topología de 2 nodos.<br>• NIST FIPS 203 ML-KEM-768 real (`crypto/mlkem`).<br>• Ed25519 y ChaCha20-Poly1305.<br>• I7UDPAdapter y Container Loopback.<br>• BufferPool multi-tier seguro. |
| **🟡 IMPLEMENTADO** | Código completo con tests unitarios passing, pero pendiente de pruebas a gran escala. | • Enrutador Kleinberg de 12 anillos concéntricos.<br>• Tabla de enrutamiento local basada en K-buckets.<br>• Firewall ZTNA local por DIDs.<br>• Telemetría de red con sincronización `sync.RWMutex`.<br>• Servidor WebUI local. |
| **🟠 EXPERIMENTAL** | Diseños o adaptadores preliminares que requieren calibración en campo. | • Marcapasos Token Bucket con ancho de banda configurable.<br>• Egress Gateway y Nodos Guardianes.<br>• Adaptador TUN / Wintun para interfaz de SO. |
| **🔴 NO DEMOSTRADO** | Hipótesis teóricas o arquitectura futura sin evidencia física. | • Malla planetaria de trillones de nodos.<br>• Perforación NAT simétrica universal en cualquier WAN hostil.<br>• Control de congestión adaptativo BBR dinámico autónomo. |

---

## 3. Invariantes del Núcleo Mínimo

1. **Invariante de MTU (1280B):** Toda trama en el alambre respeta el límite estricto de 1280 bytes para garantizar cero fragmentación en cualquier red física (RFC 8200).
2. **Orquestación Zero-Copy en Pipeline Interno:** La canalización básica de procesamiento interno mantiene 0 B/op y 0 allocs/op (`BenchmarkLinearPipeline_Execute`).
3. **Criptografía Estándar FIPS 203:** Intercambio post-cuántico basado en el estándar canónico oficial (`crypto/mlkem`) con decapsulación e *Implicit Rejection*.
4. **Desacoplamiento Estricto:** Prohibido que componentes periféricos (WASM, MCP, UI, DNS) contaminen o introduzcan dependencias en el Núcleo Mínimo I7.

---

## 4. Orden Canónico del Datapath y Separación de Responsabilidades (Fase 8)

El camino crítico de procesamiento en recepción sigue una secuencia estricta y unidireccional donde cada subsistema responde a una pregunta única y desacoplada:

```text
UDP Socket
   │
   ▼
[1. Frame Validation]      ──► ¿El formato y magic bytes son válidos? (O(1))
   │
   ▼
[2. Identity Verification] ──► ¿Quién es el emisor? (DID autocertificable Ed25519)
   │
   ▼
[3. Anti-Replay L1]        ──► ¿Ya vimos este paquete? (originDID:SessionID:Sequence:Timestamp)
   │
   ▼
[4. Session Lookup]        ──► ¿Existe sesión simétrica activa? (PQC ML-KEM-768 FIPS 203)
   │
   ▼
[5. ZTNA / Authorization]  ──► ¿Tiene permiso para comunicarse? (Default-Deny)
   │
   ▼
[6. AEAD Decrypt]          ──► ¿El contenido es íntegro y auténtico? (ChaCha20-Poly1305)
   │
   ▼
[7. Routing / Next-Hop]    ──► ¿A dónde va? (Distancia XOR Kleinberg / K-buckets)
   │
   ▼
[8. Application / TUN]     ──► Entrega de datos útiles al sistema operativo o aplicación
```

### Reglas Arquitectónicas de Aislamiento
* **El enrutamiento (Routing) NO decide identidad:** Solo determina el siguiente salto topológico hacia el destino.
* **La identidad (Identity) NO decide autorización:** Solo comprueba matemáticamente que el emisor posee la clave privada vinculada al DID.
* **La autorización (ZTNA) NO descifra:** Solo evalúa si la política local permite tráfico del DID autenticado.
* **El cifrado (AEAD) NO decide enrutamiento:** Solo garantiza confidencialidad e integridad del payload.
* **El Anti-Replay NO evalúa permisos:** Solo garantiza que ningún paquete o secuencia sea procesado más de una vez en la misma sesión.

### Mediciones Factuales de Rendimiento
* **Orquestación de Pipeline:** 32.06 ns/op | 0 B/op | 0 allocs/op.
* **Datapath Criptográfico Completo End-to-End (`BenchmarkDatapathEndToEnd`):** 3.68 µs/op | 992 B/op | 18 allocs/op (~270,000 datagramas/segundo por hilo en CPU de 1.1 GHz).

---

## 5. Estructura Real del Repositorio y Desacoplamiento (Auditoría Externa)

El repositorio organiza el código fuente en paquetes atómicos y modulares en estricta correspondencia con los estratos arquitectónicos:

```text
src/
├── cmd/
│   ├── ipvn7/          ──► Daemon principal de nodo de red soberano
│   ├── installer/      ──► Instalador GUI Windows 1-clic (build tag: windows && installer)
│   └── ipvn7-wasm/     ──► Módulo de compilación WebAssembly / navegador
│
└── pkg/
    ├── l0/             ──► Primitivas puras (Identity, Packet CBOR, ML-KEM-768 FIPS 203, Wire)
    ├── l1/             ──► Enrutamiento Kleinberg, Sesiones PQC 1-RTT, ZTNA Firewall, Adaptadores (UDP, TUN, SOCKS5, STUN)
    ├── l2/             ──► Telemetría y métricas vitales protegidas concurrentemente con RWMutex
    ├── core/           ──► Pipeline lineal zero-copy (0 B/op), FSM, WebUI (127.0.0.1) y VersionManager
    └── wasm/           ──► Puente JS / WebAssembly
```

---

## 6. Mecanismos de Descubrimiento de Red

Bajo el axioma *"La red escucha, no grita"*, se formalizan tres mecanismos diferenciados con control estricto de broadcast:

1. **LAN Broadcast (`EnableBroadcast = false` por defecto):**
   - **Estado:** Desactivado de forma predeterminada para evitar contaminación del medio y rastreo pasivo.
   - **Uso:** Solo se activa explícitamente en entornos aislados o laboratorios cerrados.
2. **Discovery Dirigido (Direct Probing / Roaming Updates):**
   - **Estado:** Activado.
   - **Comportamiento:** Emite paquetes autenticados y firmados exclusivamente a los endpoints IP:puerto de pares previamente conocidos o configurados vía CLI (`--peer`).
3. **Rendezvous Ciego y STUN Reflexivo (EBRA / Blind Store):**
   - **Estado:** Activado según configuración.
   - **Comportamiento:** Los nodos publican balizas ciegas cifradas con clave efímera en relays o servidores de almacenamiento ciego sin revelar su topología interna, y utilizan STUN reflexivo estándar (RFC 5389) para conocer su mapeo externo ante NATs.

---

## 7. Matriz Canónica: Función → Archivo → Test → Evidencia

| Estrato / Función | Archivo de Implementación | Suite de Tests | Evidencia Demostrada |
| :--- | :--- | :--- | :--- |
| **Identidad Soberana (Ed25519)** | `src/pkg/l0/identity.go` | `src/pkg/l0/identity_test.go` | DID criptográfico autocertificable, firma y verificación |
| **Criptografía PQC (ML-KEM-768)** | `src/pkg/l0/pqc_kem.go` | `src/pkg/l0/pqc_kem_test.go` | NIST FIPS 203 real (`crypto/mlkem`), decapsulación e implicit rejection |
| **Firewall ZTNA Default-Deny** | `src/pkg/l1/firewall.go` | `src/pkg/l1/roaming_ztna_test.go` | Rechazo estricto de RoamingUpdate no autorizado incluso con firma válida |
| **Sesiones PQC 1-RTT** | `src/pkg/l1/session_manager.go` | `src/pkg/l1/session_adversarial_test.go` | Tests A-H: mitigación total de suplantación, descarte de replay |
| **Enrutador Kleinberg (16 anillos)** | `src/pkg/l1/routing.go` | `src/pkg/l1/l1_test.go` | Enrutamiento voraz por distancia XOR y latencia EWMA |
| **Pipeline Zero-Copy (0 B/op)** | `src/pkg/core/pipeline_stages.go`| `src/pkg/core/pipeline_bench_test.go` | 34.9 ns/op, 0 B/op, 0 allocs/op en `BenchmarkLinearPipeline_Execute` |
| **Web UI Localhost & Control Admin** | `src/pkg/core/web_ui_handlers.go` | `src/pkg/core/web_ui_admin_auth_test.go`| Bind 127.0.0.1, auth token RBAC, CORS sin comodín, SSRF mitigado |
| **Instalador Autocontenido** | `src/cmd/installer/main.go` | `scripts/build_installer.ps1` | Build tag `windows && installer`, zero-admin y auto-elevación UAC |


