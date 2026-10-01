# PLAN DE CUMPLIMIENTO FACTUAL — AUDITORÍA EXTERNA IPvN7 v0.7

> **Directiva Suprema:** En conformidad con `docs/auditoria externa.md` como única fuente de verdad, cerrar toda brecha entre el diseño arquitectónico y la ejecución física real, conectando el pipeline de seguridad, ZTNA Default-Deny y PQC al camino de datos principal, erradicando falsas rutas de copia y aplicando la taxonomía estricta de 5 estados (HECHO, TESTEADO, MEDIDO, NO IMPLEMENTADO, EXPERIMENTAL).


---

## 1. Diagnóstico de los 5 Bloques Críticos de la Auditoría

| # | Hallazgo de la Auditoría Externa | Diagnóstico Técnico en el Código | Solución Técnica Exacta |
|---|---|---|---|
| **1** | **El Firewall ZTNA no está en el camino de datos principal** | En `main.go`, `l1.NewZTNAFirewall(true)` se crea pero el bucle de recepción UDP despacha paquetes sin llamar a `firewall.EvaluateInbound()`. Además, `autonomous_discovery.go` autoriza DIDs ciegamente desde beacons antes de autenticar posesión criptográfica. | 1. Integrar evaluación ZTNA Default-Deny en cada datagrama de datos recibido en `main.go`.<br>2. Eliminar la auto-autorización ciega en `autonomous_discovery.go`. Solo autorizar DIDs tras verificación criptográfica de firma o handshake PQC. |
| **2** | **El Handshake PQC no gobierna el tráfico de datos principal** | `main.go` solo imprime logs ante `MsgTypeHandshakeInit/Resp`. No existe máquina de estados que establezca claves de sesión y cifre/autentique los paquetes `MsgTypeData` con AEAD ChaCha20-Poly1305. | Implementar `PQCSessionTable` en el camino de datos: intercambio KEM 1-RTT (ML-KEM-768 FIPS 203 + X25519) que derive claves simétricas `TxKey`/`RxKey`, cifrando y autenticando todo `MsgTypeData` con AEAD y anti-replay. |
| **3** | **Contradicción del falso "zero-copy" en `main.go`** | `main.go` adquiere un buffer de `bufferPool`, ejecuta `copy(pktBuf.RawSlice(), buf[:n])` y luego descarta `pktBuf` mientras deserializa `buf[:n]`, incurriendo en una copia inútil. | Erradicar la copia espuria. Deserializar directamente con `l0.DecodePacket(buf[:n])` sin alocaciones intermedias ni copias redundantes en el camino crítico. |
| **4** | **Descubrimiento ruidoso ("la red escucha, no grita")** | `AutonomousDiscoveryEngine` emite broadcasts periódicos cada 12s a 4 puertos (`7777, 7778, 7001, 8080`). | Convertir el broadcast ruidoso en pasivo/opcional. Priorizar aprendizaje local pasivo y consultas dirigidas (`probeTrustedPeers`), suprimiendo el flood periódico no solicitado. |
| **5** | **Contradicciones de configuración y documentación** | • 12 vs 16 anillos Kleinberg (`main.go` imprime 12, código usa 16).<br>• Go 1.26.4 en `go.mod` vs Go 1.24 en CI.<br>• Telemetría declarada "lock-free" pero implementada con `sync.RWMutex`. | • Dinamizar `main.go` con `router.GetConfig().NumRings` y unificar a 16 anillos canónicos.<br>• Sincronizar `src/go.mod` a `go 1.24.0` (estándar FIPS 203).<br>• Corregir la documentación de telemetría a "sincronización concurrente de alta velocidad protegida por sync.RWMutex". |

---

## 2. Fases de Ejecución

### Fase 1: Sincronización de Versión Go y Corrección Documental (Punto 5)
1. Modificar `src/go.mod`: cambiar `go 1.26.4` a `go 1.24.0` (compatible con CI ubuntu-latest y con toolchains modernos).
2. Modificar `src/pkg/l2/telemetry.go`: corregir comentarios eliminando la palabra "lock-free" y sustituyéndola por sincronización concurrente con RWMutex.
3. Unificar enrutador Kleinberg: asegurar que el print en `main.go` y la documentación canónica reflejen `DefaultNumRings = 16`.

### Fase 2: Saneamiento del Motor de Descubrimiento (Punto 4)
1. Modificar `src/pkg/l1/autonomous_discovery.go`:
   - Eliminar `e.Firewall.AuthorizeDID()` de la recepción de beacons no autenticados en `processDiscoveredBeacons`.
   - Añadir control para suprimir el broadcast ruidoso periódico automático, favoreciendo sondeo dirigido y escucha pasiva.
   - Solo autorizar pares en el firewall cuando exista una firma digital verificada (`MsgTypeRoamingUpdate`) o un apretón de manos completado (`MsgTypeHandshakeResp`).

### Fase 3: Integración del Pipeline de Seguridad y Sesión PQC en Datapath (Puntos 1 y 2)
1. Diseñar el gestor de sesiones PQC en `src/pkg/l1/session_manager.go` ($\le 250$ líneas):
   - Mapeo `map[string]*l0.SessionKeys` con exclusión mutua concurrente.
   - Manejo de `MsgTypeHandshakeInit`: desencapsula ML-KEM-768, deriva claves, genera `MsgTypeHandshakeResp`, autoriza par en ZTNA.
   - Manejo de `MsgTypeHandshakeResp`: finaliza 1-RTT, almacena claves de sesión, autoriza par en ZTNA.
   - Cifrado AEAD (`EncryptData`) y descifrado AEAD (`DecryptData`) con ChaCha20-Poly1305.
2. Actualizar `src/cmd/ipvn7/main.go`:
   - Erradicar la copia espuria `pktBuf := bufferPool.Acquire... copy...`.
   - Incorporar `antiReplay := l0.NewAntiReplayFilter()`.
   - Flujo de recepción estricto:
     `UDP -> Decode -> Anti-replay -> Autenticación -> ZTNA -> Session -> AEAD -> DATA`.
   - En `simpleMeshForwarder.ForwardToMesh`: cifrar los datos del túnel/aplicación con la clave de sesión `TxKey` antes de transmitir.

### Fase 4: Batería de Pruebas Físicas y Validación de Seguridad
1. Crear test unitario e integrativo `src/pkg/l1/datapath_security_test.go`:
   - Validar rechazo ZTNA Default-Deny ante DIDs desconocidos.
   - Validar descarte de replay por la ventana de 1024 bits.
   - Validar apretón de manos 1-RTT ML-KEM-768 sobre sockets físicos UDP loopback con cifrado/descifrado AEAD de extremo a extremo.
2. Ejecutar `go test ./pkg/...` asegurando 100% PASS.

### Fase 6: Cierre de Vulnerabilidades Criptográficas y Batería Adversarial (Ronda 2 Auditoría)
1. **Autenticación Obligatoria en HandshakeInit (Hallazgo 1):**
   - Incorporada firma digital Ed25519 obligatoria en `CreateHandshakeInitPacket`.
   - `HandleHandshakeInitPacket` verifica `PublicKeyFromDID(SourceDID)` y `VerifyPacketSignature()` ANTES de decapsular KEM o invocar `firewall.AuthorizeDID()`.
   - Optimización de tamaño de criptograma KEM a 1120B canónico (X-Wing CFRG: 32B X25519 + 1088B ML-KEM-768), garantizando cumplimiento estricto del MTU determinista de 1280B con firma Ed25519.
2. **Conexión de Anti-Replay L1 al Datapath Principal (Hallazgo 2):**
   - Sustituido el filtro global L0 por `l1.NewAntiReplayFilter(nil)` en `src/cmd/ipvn7/main.go`.
   - Aislamiento cross-session estricto mediante evaluación `Accept(packet.SourceDID, sessionID, packet.Sequence, packet.Timestamp)`.
3. **Manejo Fatal de Errores PQC (Hallazgo 6):**
   - Si `GenerateHybridKeyPair` falla en `main.go`, el nodo aborta inmediatamente con `os.Exit(1)`.
4. **Validación Estricta de Respuestas de Handshake (Hallazgo 7):**
   - `HandleHandshakeRespPacket` valida que `pkt.DestDID == identity.DID()`, correspondencia con `pendingSession` y expiración temporal (máx 60s).
5. **Batería de Pruebas Adversariales (Tests A hasta H) (Hallazgo 11):**
   - Creada suite `src/pkg/l1/session_adversarial_test.go`:
     - Test A: Suplantación de SourceDID por atacante -> RECHAZADO (0 sesiones creadas, 0 autorizaciones ZTNA).
     - Test B: Firma inválida/forjada -> RECHAZADO.
     - Test C: Aislamiento Cross-Session (seq idénticos en sesiones distintas) -> AMBOS ACEPTADOS.
     - Test D: Aislamiento Cross-Peer (seq idénticos de pares distintos) -> AMBOS ACEPTADOS.
     - Test E: Replay inmediato de datos -> RECHAZADO.
     - Test F: Replay de HandshakeInit -> RECHAZADO.
     - Test G: Respuesta dirigida a un tercero -> RECHAZADO.
     - Test H: Alteración de SourceDID en tránsito -> RECHAZADO.
6. **Clarificación Factual Zero-Copy (Hallazgo 3):**
   - Separación formal entre el benchmark de orquestación (`BenchmarkLinearPipeline_Execute`, 0 allocs/op) y el datapath UDP con buffers prealocados del pool.

### Fase 7: Resolución de Hallazgos Críticos Ronda 3 (Auditoría Octubre 2026)
1. **Desacoplamiento de CI y Compilación de Instalador (P0):**
   - Incorporado build tag `//go:build windows && installer` en `src/cmd/installer/main.go`.
   - `go vet ./...` y `go test ./...` pasan en limpio en cualquier entorno/CI sin depender de binarios precompilados.
2. **Cierre de Bypass Conceptual ZTNA en RoamingUpdate (P0):**
   - Eliminada la invocación automática `firewall.AuthorizeDID` al recibir `MsgTypeRoamingUpdate` en `main.go` y `pipeline_stages.go`.
   - Autenticación (firma Ed25519) estrictamente separada de Autorización (evaluación Default-Deny).
   - Verificado con suite formal `src/pkg/l1/roaming_ztna_test.go` (Tests de rechazo ante pares no autorizados, aceptación de autorizados, y descarte de firmas inválidas/malformadas).
3. **Cierre de Superficie WebUI y Rutas Administrativas (P0):**
   - Bind HTTP fijado estrictamente a `127.0.0.1` por defecto.
   - Eliminado `Access-Control-Allow-Origin: *` del panel administrativo (reemplazado por loopback restringido).
   - Control de acceso RBAC implementado (`checkAdminAuth`) protegiendo `/vpn/connect`, `/vpn/disconnect`, `/vpn/exit`, `/vpn/cycle` y `/update/*`.
   - Bloqueo estricto de SSRF en `handleUpdateCheck` limitando manifiestos a fuentes de confianza oficiales.
4. **Honestidad Criptográfica en Vectores Reticulares (P1):**
   - Eliminada nomenclatura equívoca de NIST FIPS 204 ML-DSA para componentes derivados por HMAC.
   - Renombrados identificadores a `ExperimentalPQCIdentity` y `ExperimentalSigSize`, declarando explícitamente que la autenticidad en producción descansa en Ed25519 estándar (RFC 8032).

---

## 3. Estado de Evaluación: AUDITADO INTERNAMENTE CON EVIDENCIA LOCAL DEMOSTRADA
La totalidad de los hallazgos P0 y P1 identificados en las rondas de auditoría externa han sido subsanados en el código fuente, respaldados por suites de tests unitarios, adversariales y compuertas de análisis estático en el repositorio local.
