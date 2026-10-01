audita revisa 
galleguillosdavid-coder/ipvn7_0.7

Sí. Audité el HEAD actual de main del repositorio `galleguillosdavid-coder/ipvn7_0.7`, commit f1a1308 del 1 de octubre de 2026.

Mi conclusión es bastante clara:

El proyecto tiene una base técnica real y bastante trabajo implementado, pero yo NO consideraría todavía que el núcleo de seguridad IPvN7 v0.7 esté cerrado.

Encontré al menos 2 problemas de seguridad críticos, además de varias pruebas que certifican componentes aislados pero no el camino de datos completo.

1. Hallazgo crítico: el HandshakeInit permite suplantar un DID

Este es el problema más importante.

En session_manager.go, HandleHandshakeInitPacket() recibe:

SourceDID

pero no verifica una firma de ese paquete antes de usarlo.

Después hace:

Respond1RTT(...)

y luego:

m.sessions[pkt.SourceDID] = sessionKeys

y finalmente:

firewall.AuthorizeDID(pkt.SourceDID)

Es decir:

Atacante
   │
   │ declara SourceDID = DID de otra persona
   │
   │ ML-KEM válido para Bob
   ▼
Bob
   │
   ├─ crea sesión
   ├─ acepta el KEM
   └─ AUTORIZA el DID declarado

El problema es que el KEM demuestra que quien hizo el encapsulado conoce el secreto resultante, pero no demuestra que esa persona posea la clave Ed25519 correspondiente al SourceDID.

Y el propio protocolo ya tiene una solución para esto: el DID es autocertificable porque se deriva directamente de la clave pública Ed25519:

did:ipvn7:<hex de clave pública Ed25519>

Eso significa que el HandshakeInit debería quedar criptográficamente vinculado al DID.

Ataque concreto

Mallory podría construir:

SourceDID = DID_Alice
RecipientDID = DID_Bob

pero encapsular usando las claves públicas de Bob.

Bob podría terminar con:

sessions[DID_Alice] = sesión_de_Mallory

y además:

AuthorizeDID(DID_Alice)

Eso rompe la correspondencia:

DID ↔ posesión de clave
Corrección

El HandshakeInit debe incluir autenticación de identidad.

Mínimo:

HandshakeInit
    SourceDID
    DestinationDID
    ephemeral X25519
    ML-KEM ciphertext
    timestamp
    sequence
    firma Ed25519

Y Bob debe verificar:

DID → PublicKeyFromDID()
PublicKey → Verify(signature)

antes de autorizar el DID.

La regla debería ser:

KEM válido
      ↓
firma válida del SourceDID
      ↓
DID coincide con clave pública
      ↓
crear sesión
      ↓
autorizar DID

No:

KEM válido
      ↓
autorizar DID
2. Segundo problema crítico: el Anti-Replay correcto existe, pero main.go no lo utiliza

Esto me parece especialmente importante porque el repositorio afirma haber solucionado precisamente el problema de cross-session replay.

Existe:

src/pkg/l1/anti_replay_session.go

y ahí sí hay una arquitectura mucho mejor:

originDID
sessionID
sequence
timestamp

con una ventana independiente por sesión.

Pero el main.go utiliza:

antiReplay := l0.NewAntiReplayFilter()

y luego:

antiReplay.ValidateAndUpdate(packet.Sequence)

El filtro L0 solamente mantiene:

lastSeq
bitmap

globalmente.

No mantiene:

DID + SessionID
Consecuencia

Supongamos:

Alice → seq 1
Bob   → seq 1

El filtro global puede considerar el seq=1 de Bob como repetido aunque sea completamente legítimo.

Más grave todavía:

Alice/session-A/seq=100
Bob/session-B/seq=100

las sesiones no están aisladas.

Y existe código L1 que aparentemente fue diseñado precisamente para resolverlo:

AntiReplayFilter.Accept(
    originDID,
    sessionID,
    seq,
    timestamp
)

pero no está conectado al datapath principal.

Esto es una contradicción arquitectónica

Actualmente tienes dos sistemas:

L0 AntiReplay
    ↓
main.go
    ↓
datapath real

y:

L1 AntiReplay por sesión
    ↓
tests
    ↓
NO llega al datapath principal

Para mí esto debe corregirse antes de considerar cerrado el sistema.

3. El supuesto "Zero-Copy 0 B/op" no demuestra Zero-Copy del datapath

Este hallazgo es menos peligroso pero importante para la documentación.

La suite ejecuta:

go test -run=^$ -bench=BenchmarkLinearPipeline_Execute -benchmem ./pkg/core

Y el benchmark es:

pipeline := NewLinearPipeline().
    AddStage(&noopStage{}).
    AddStage(&noopStage{}).
    AddStage(&noopStage{})

Es decir:

no decodifica un paquete real.

No hace:

UDP
 ↓
CBOR
 ↓
Packet
 ↓
AntiReplay
 ↓
ZTNA
 ↓
AEAD
 ↓
routing

Por tanto:

0 B/op
0 allocs/op

demuestra que ese benchmark artificial del pipeline vacío no genera allocations.

No demuestra:

main.go UDP datapath = zero-copy

De hecho DecodePacket() utiliza CBOR y construye:

var p Packet
cborDecMode.Unmarshal(data, &p)

por lo que la afirmación de zero-copy absoluto del camino de red requiere una medición diferente.

Yo cambiaría la afirmación

En lugar de:

Zero-Copy certificado.

usar:

Eliminada la copia intermedia explícita del buffer UDP antes de la decodificación. El pipeline de procesamiento sigue pendiente de caracterización de allocations end-to-end.

Eso sería técnicamente defendible.

4. El sistema PQC sí es real en una parte importante

Aquí hay que reconocer lo que está bien.

pqc_hybrid.go utiliza realmente:

crypto/mlkem
ML-KEM-768

y las dimensiones corresponden al esquema estándar:

public key = 1184 bytes
ciphertext = 1088 bytes
shared secret = 32 bytes

Además, hay pruebas que comprueban:

Encapsulate
     ↓
Decapsulate
     ↓
shared secret coincide

y también pruebas UDP reales sobre 127.0.0.1.

Eso sí constituye evidencia de una implementación funcional de:

X25519
+
ML-KEM-768
+
HKDF
+
ChaCha20-Poly1305

en el componente criptográfico.

5. Pero cuidado con llamar a la firma "ML-DSA"

Aquí el propio código es bastante honesto y eso está bien.

pqc_signatures.go dice explícitamente que el componente reticular:

NO constituye una implementación formal completa de NIST FIPS 204.

Y efectivamente:

MLDSA65SigSize = 128

no corresponde al tamaño de una firma ML-DSA-65 real.

Lo que hay es:

Ed25519 real
+
HMAC/SHA-256 experimental

Por tanto:

HECHO

Ed25519:

real.

ML-KEM-768:

real.

ML-DSA:

experimental / no implementado como ML-DSA FIPS 204.

Esto está correctamente advertido en el código, pero yo evitaría cualquier documentación que diga simplemente:

IPvN7 tiene ML-DSA

porque sería una descripción engañosa.

6. Otro problema: fallo de generación PQC no detiene el nodo

En main.go:

hybridKeys, err := l1.GenerateHybridKeyPair(identity)

if err != nil {
    core.LogError(...)
}

y después:

sessionMgr := l1.NewPQCSessionManager(
    identity,
    hybridKeys,
    firewall,
)

Es decir, si la generación de claves falla:

hybridKeys = nil

pero el proceso continúa.

Después determinadas rutas pueden intentar utilizar:

m.localKeys

y terminar en comportamiento inválido o panic.

Para el modo seguro debería ser:

GenerateHybridKeyPair
       │
       ├── OK → continuar
       │
       └── ERROR → abortar nodo

No:

ERROR → registrar → continuar como si nada
7. El handshake tampoco está completamente unido a la identidad

Hay otro detalle relacionado con el primero.

HandleHandshakeRespPacket() verifica:

pkt.VerifyPacketSignature()

Eso está bien.

Pero no verifica explícitamente que:

pkt.DestDID == identidad local

ni que el SourceDID corresponda exactamente al peer esperado por el pendingSession.

Actualmente la asociación principal es:

pendingSessions[pkt.SourceDID]

Yo endurecería esto a algo equivalente a:

SessionID pendiente
+
expectedPeerDID
+
expectedDestinationDID
+
timestamp/expiry

Así una respuesta no puede simplemente presentarse ante cualquier sesión pendiente.

8. El descubrimiento está bastante mejor de lo que parece

Aquí encontré algo positivo respecto de las versiones anteriores.

EnableBroadcast existe y por defecto queda:

false

por lo que el broadcast:

7777
7778
7001
8080

no se ejecuta automáticamente.

Eso coincide bastante bien con tu principio:

"la red escucha, no grita".

Pero el sistema todavía hace periódicamente:

STUN keepalive
+
rendezvous
+
probeTrustedPeers
+
DiscoverAllPeers

por lo que "silencioso" no significa realmente pasivo.

Eso no es necesariamente un bug; es una decisión de diseño que debe quedar claramente diferenciada:

broadcast LAN = opcional/desactivado
discovery unicast/rendezvous = activo
STUN = activo
9. El repositorio tiene una cantidad importante de arquitectura periférica

Veo:

L0
L1
L2
Core
interfaces
TUN
SOCKS5
WebUI
STUN
Rendezvous
Kleinberg
XDP/eBPF
WASM
SDK Go
SDK Python
SDK Rust
SDK TypeScript
Docker
scripts
agents
skills
etc.

Eso es muchísimo para un núcleo que conceptualmente quieres mantener pequeño.

Y aquí veo una tensión con tu propia filosofía anterior:

"Core transports structure, not semantics."

El proyecto está empezando a parecer un Network OS completo, no solamente un núcleo de transporte IPvN7.

No digo que haya que borrar esos componentes.

Yo los separaría físicamente:

CORE
├── wire
├── identity
├── session
├── channel
├── routing
├── integrity
└── fragmentation

ADAPTERS
├── TUN
├── SOCKS5
├── STUN
├── DERP
├── WebUI
├── XDP
└── etc.

EXPERIMENTAL
├── AI
├── X-Wing
├── ML-DSA
└── autonomous

Eso haría mucho más fácil auditar el verdadero IPvN7.

10. El router no debería formar parte de la prueba de seguridad

Hay una mezcla que conviene evitar:

PQC
ZTNA
AntiReplay
Kleinberg
Discovery
TUN

en el mismo camino lógico.

El camino mínimo debería poder demostrarse sin router sofisticado:

UDP
 ↓
Wire Decode
 ↓
Identity
 ↓
AntiReplay
 ↓
Session
 ↓
AEAD
 ↓
Application

Después:

Routing

debería decidir únicamente:

¿por dónde envío?

no:

¿quién eres?
¿puedo confiar en ti?
¿estás cifrado?

Eso mantiene limpio el diseño.

11. Las pruebas actuales tienen una laguna importante

Las pruebas comprueban muy bien:

Alice ↔ Bob

pero faltan pruebas adversariales del tipo:

Test A — DID falso
Mallory
SourceDID = Alice
KEM → Bob

Debe fallar.

Test B — DID correcto + clave incorrecta
SourceDID = Alice
firma = Mallory

Debe fallar.

Test C — Cross-session
Alice/session1/seq10
Alice/session2/seq10

Debe aceptar ambos.

Test D — Cross-peer
Alice/seq10
Bob/seq10

Debe aceptar ambos.

Test E — Replay
Alice/session1/seq10
Alice/session1/seq10

Segundo debe fallar.

Test F — Handshake replay

Reenviar exactamente el mismo HandshakeInit.

Debe ser rechazado.

Test G — respuesta para otro destinatario
Bob → Carol

presentada ante Alice.

Debe fallar.

Test H — modificación de SourceDID

Modificar solamente:

SourceDID

después del handshake.

Debe fallar.

12. La afirmación "100% PASS" necesita una corrección conceptual

El repositorio documenta:

Health Score 100%
100% PASS
DEMOSTRADO FÍSICAMENTE

El problema no es que los tests sean falsos.

El problema es qué están demostrando.

Por ejemplo:

Test PQC

demuestra que:

ML-KEM → secreto compartido → AEAD

funciona.

No demuestra:

DID autenticado correctamente

porque precisamente el handshake tiene la vulnerabilidad que encontré.

Y:

AntiReplay L1

demuestra que su algoritmo por sesión funciona.

Pero main.go utiliza otro AntiReplay.

Por eso yo clasificaría actualmente el estado así:

Componente	Estado que puedo verificar
ML-KEM-768	IMPLEMENTADO / probado
X25519	IMPLEMENTADO / probado
ChaCha20-Poly1305	IMPLEMENTADO / probado
Ed25519/DID	IMPLEMENTADO / probado
ZTNA Default-Deny	IMPLEMENTADO / integrado
Discovery sin broadcast por defecto	IMPLEMENTADO
Anti-Replay por sesión	IMPLEMENTADO pero no conectado al datapath principal
Anti-Replay utilizado por main.go	IMPLEMENTADO, pero global
Zero-copy end-to-end	NO DEMOSTRADO
ML-DSA FIPS 204	NO IMPLEMENTADO; experimental
Identidad autenticada durante HandshakeInit	DEFECTUOSO
Seguridad completa del datapath	NO CERRADA
13. Mi diagnóstico de la arquitectura actual

La veo aproximadamente así:

                 ┌──────────────────────┐
                 │       Discovery      │
                 └──────────┬───────────┘
                            │
                            ▼
UDP ──► Decode ──► AntiReplay L0 ──► Dispatch
                            │
             ┌──────────────┼──────────────┐
             │              │              │
          Handshake       Data         Roaming
             │              │              │
             ▼              ▼              ▼
          PQC/KEM         ZTNA          Ed25519
             │              │
             ▼              ▼
          Session         AEAD
                            │
                            ▼
                           TUN

El diseño que yo perseguiría para el núcleo sería:

UDP
 │
 ▼
FRAME VALIDATION
 │
 ├── size
 ├── version
 ├── destination
 └── timestamp
 │
 ▼
IDENTITY
 │
 └── DID ↔ Ed25519
 │
 ▼
ANTI-REPLAY
 │
 └── DID + SessionID + Sequence
 │
 ▼
SESSION
 │
 ├── X25519
 ├── ML-KEM-768
 └── key derivation
 │
 ▼
ZTNA
 │
 └── policy
 │
 ▼
AEAD
 │
 └── ChaCha20-Poly1305
 │
 ▼
ROUTING
 │
 ▼
APPLICATION / TUN

Ese orden es mucho más defendible.

14. Qué arreglaría primero

No añadiría nuevas funciones todavía.

Haría exactamente estas 5 correcciones:

1. Autenticar HandshakeInit

Firma Ed25519 obligatoria y DID autocertificable.

2. Sustituir AntiReplay L0 por AntiReplay L1

El datapath debe usar:

DID + SessionID + Sequence + Timestamp

y eliminar el filtro global del camino principal.

3. Añadir prueba de suplantación

Un test debe demostrar:

Mallory ≠ Alice

aunque Mallory pueda crear un KEM válido.

4. Rehacer la prueba Zero-Copy

Medir:

UDP Read
→ DecodePacket
→ validation
→ ZTNA
→ decrypt

con:

-benchmem
-race

y separar claramente:

pipeline orchestration benchmark

de:

network datapath benchmark
5. Si falla PQC → abortar

No continuar con:

hybridKeys == nil
Veredicto técnico

La versión 0.7 ya dejó de ser solamente una maqueta. Hay código real, sockets UDP reales, ML-KEM-768 real, Ed25519 real, AEAD real, TUN, routing y pruebas bastante amplias.

Pero encontré una contradicción importante entre lo que el repositorio declara haber cerrado y lo que realmente está conectado al datapath.

El problema principal es este:

           IDENTIDAD
               │
               X       ← falta vínculo criptográfico aquí
               │
             KEM
               │
             SESIÓN
               │
          AUTORIZACIÓN

Debe ser:

           IDENTIDAD
               │
          Ed25519 OK
               │
             KEM
               │
       SessionID creado
               │
      AntiReplay por sesión
               │
          ZTNA authorize
               │
             AEAD
               │
              DATA

Yo no haría más funcionalidades de IPvN7 hasta corregir esos puntos. En particular, no añadiría todavía DHT, más perfiles, más agentes, más SDK ni más "galactic scale". Primero cerraría el núcleo criptográfico y demostraría mediante tests adversariales que la identidad, la sesión y el paquete son inseparables.

Además, el estado de CI que pude consultar para el commit f1a1308 no muestra ejecuciones de GitHub Actions asociadas; por tanto, el "100% PASS" que aparece en la documentación no lo puedo considerar una ejecución CI independiente verificable desde GitHub en este momento.