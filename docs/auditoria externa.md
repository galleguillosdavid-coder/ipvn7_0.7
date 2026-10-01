audita revisa 
galleguillosdavid-coder/ipvn7_0.7

Sí. Hice una auditoría directa del repositorio galleguillosdavid-coder/ipvn7_0.7, incluyendo código, arquitectura, seguridad, CI, tests y documentación.

Repositorio IPVN7 0.7

Resultado ejecutivo

Estado real al 1 de octubre de 2026: NO está cerrado.

Hay bastante trabajo real y varias mejoras importantes, pero encontré fallos críticos que contradicen algunas afirmaciones de “100% PASS / cierre de auditoría”.

Mi clasificación:

Área	Estado
Núcleo criptográfico ML-KEM-768	🟢 Implementación real
Ed25519/X25519/ChaCha20-Poly1305	🟢 Implementado
Anti-replay L1	🟢 Implementado
Tests adversariales	🟢 Existen
MTU 1280	🟢 Bien encaminado
Arquitectura	🟡 Demasiado grande/mezclada
CI	🔴 Falla actualmente
ZTNA	🔴 Hay bypass conceptual en main.go
Web UI/API	🔴 Control remoto sin autenticación
SOCKS5	🟠 Funcional, pero aumenta superficie de ataque
Instalador Windows	🔴 No compila desde el repositorio actual
Claims de auditoría	🔴 Más fuertes que la evidencia actual
1. 🔴 P0 — El CI actual está roto

Esto es objetivo y reproducible.

El último commit:

5c0189b79a24ebb1b90dff76760b53ae57896b5b

ejecutó GitHub Actions y terminó:

FAILURE

El job ni siquiera llegó a los tests.

Falló en:

go vet ./...

con:

cmd/installer/main.go:15:12:
pattern assets/ipvn7.exe:
no matching files found

El instalador contiene:

//go:embed assets/ipvn7.exe
var ipvn7Binary []byte

//go:embed assets/wintun.dll
var wintunDLL []byte

pero esos assets no están presentes en el repositorio.

Consecuencia

La afirmación del documento:

“CI Reproducible”

no está demostrada por el CI actual.

Y tampoco es correcto considerar que la última rama auditada está en 100% PASS cuando la propia ejecución más reciente terminó en failure.

2. 🔴 P0 — Hay un bypass conceptual del ZTNA

Este es más importante que el fallo de CI.

En main.go, al recibir MsgTypeRoamingUpdate, el código hace:

if valid, err := packet.VerifyPacketSignature(); err == nil && valid {
    ...
    firewall.AuthorizeDID(&l1.DIDPolicy{
        DID: packet.SourceDID,
        AllowInbound: true,
        AllowOutbound: true,
        AllowRelay: true,
    })
}

Es decir:

firma válida → autorización automática.

Eso mezcla dos conceptos que la propia arquitectura dice que deben estar separados:

IDENTIDAD
   ↓
¿demostró que posee la clave?

AUTORIZACIÓN
   ↓
¿está autorizado por la política local?

Una firma Ed25519 demuestra posesión de la clave asociada al DID.

No demuestra que ese DID tenga permiso para utilizar el nodo.

Ejemplo del problema

Mallory genera legítimamente:

did:ipvn7:MALLORY

Firma correctamente su RoamingUpdate.

El nodo receptor comprueba:

firma válida = sí

y posteriormente:

AuthorizeDID(Mallory)

Por lo tanto Mallory pasa a estar autorizada.

Eso contradice directamente el modelo:

Default-Deny
+
AuthorizedDIDs

que la documentación afirma utilizar.

Corrección

El flujo debe ser:

RoamingUpdate
       ↓
validar formato
       ↓
validar firma
       ↓
extraer DID
       ↓
¿DID está autorizado por política?
       ├── NO → DROP
       └── SÍ
             ↓
          aceptar

Nunca:

firma válida → AuthorizeDID()

AuthorizeDID() debe ser una operación administrativa/política, no una consecuencia automática de autenticación.

3. 🔴 P0 — Web UI expuesta en 0.0.0.0

Esto es probablemente el problema de seguridad más práctico.

web_ui.go crea:

http.Server{
    Addr: fmt.Sprintf("0.0.0.0:%d", port),
    Handler: mux,
}

Y el main.go utiliza por defecto:

-web-port 7070

Por tanto:

0.0.0.0:7070

no es solamente:

127.0.0.1:7070

Es accesible desde interfaces de red.

Y el instalador incluso agrega una regla de firewall:

IPVN7-Web-TCP
TCP
localport=7070
action=allow
El problema

La API tiene operaciones como:

/vpn/connect
/vpn/disconnect
/vpn/exit
/vpn/cycle
/update/apply
/update/rollback
/mcp
/a2a

Y no veo autenticación fuerte delante de esas rutas.

Además:

Access-Control-Allow-Origin: *

está habilitado.

Esto convierte el WebUI en una superficie de administración remota.

4. 🔴 P0 — /vpn/exit puede apagar el nodo

handleExit() ejecuta:

os.Exit(0)

después de ejecutar el callback de apagado.

Por tanto, si el puerto Web está accesible desde otra máquina y no existe una capa de autenticación que no aparece en estos handlers:

POST /api/v1/vpn/exit

puede convertirse en una operación remota de apagado.

Eso no debería existir así en un daemon de red.

5. 🔴 P0 — Actualización remota demasiado poderosa

También existe:

/api/v1/update/apply
/api/v1/update/rollback

El VersionManager descarga un binario y lo instala como ejecutable.

Tiene una defensa positiva:

SHA-256

Eso está bien.

Pero hash ≠ autenticidad del publicador si el manifiesto que contiene el hash puede ser manipulado.

La cadena actual es:

manifest
   ↓
URL
   ↓
SHA256
   ↓
binario

Para un sistema de red soberano yo exigiría:

manifest
   ↓
firma Ed25519 del fabricante/desarrollador
   ↓
verificación de firma
   ↓
SHA256
   ↓
binario

El hash comprueba:

“este archivo corresponde al hash indicado”.

La firma comprueba:

“este manifiesto fue autorizado por la clave de distribución”.

6. 🟠 CheckOnlineUpdate acepta una URL arbitraria

Esta API:

/api/v1/update/check?url=...

permite suministrar:

manifestURL := r.URL.Query().Get("url")

y luego:

client.Get(manifestURL)

Eso crea una superficie de SSRF.

Aunque no veo inmediatamente una ruta directa a ejecución arbitraria mediante esto, no debería existir en una interfaz de administración expuesta.

Debe existir una política:

URL permitidas:
    github.com/galleguillosdavid-coder/ipvn7_0.7
    o
    servidor de actualización configurado

No:

cualquier URL HTTP
7. 🟢 La criptografía principal sí es mucho más seria que una simulación

Aquí hay una mejora importante respecto de versiones anteriores.

Encontré utilización real de:

crypto/mlkem

y:

mlkem.GenerateKey768()

El código usa los tamaños reales de ML-KEM-768:

Public key:   1184 bytes
Ciphertext:   1088 bytes
Shared secret: 32 bytes

Eso sí es una implementación real del KEM, no simplemente generar bytes con SHA/HMAC y llamarlos ML-KEM.

Además existe:

X25519
+
ML-KEM-768
+
HKDF
+
ChaCha20-Poly1305

La batería de tests también comprueba:

encapsulación;
decapsulación;
ciphertext corrupto;
clave incorrecta;
downgrade;
X-Wing;
UDP loopback;
cifrado/descifrado.

Esto es una parte sólida del trabajo.

8. 🟠 Pero pqc_hybrid.go tiene restos conceptualmente confusos

Encontré esto:

MLDSAPubHex

y:

MLDSA65SeedSize

pero el propio código reconoce que la firma principal es:

Ed25519

y que ML-DSA es experimental.

Eso es correcto como experimento, pero yo eliminaría cualquier representación que parezca una clave ML-DSA real si no existe realmente ML-DSA.

Especialmente esto:

hDSA := sha256.New()
hDSA.Write(...)
hDSA.Write(pqcSignSeed)

y posteriormente:

MLDSAPubHex: hex.EncodeToString(hDSA.Sum(nil))

Eso no constituye una clave pública ML-DSA.

Aunque esté documentado como experimental, el nombre:

MLDSAPubHex

es peligrosamente engañoso.

Mejor:

ExperimentalPQCIdentity

o eliminarlo completamente hasta implementar ML-DSA real.

9. 🟢 Anti-Replay está bastante bien planteado

La nueva implementación L1 utiliza:

originDID
+
sessionID
+
sequence
+
timestamp

y ventana de:

1024 bits

Además existe aislamiento por:

originDID:sessionID

Eso responde directamente a un problema real de replay entre peers/sesiones.

La idea arquitectónica es correcta:

DID
 ↓
Session
 ↓
Sequence Window

y no un contador global para toda la red.

10. 🟠 Hay que revisar profundamente la implementación del anti-replay

Hay una cuestión que todavía no considero cerrada:

if sessionID < lastSessID && tsSec < lastSeenTs

Eso no demuestra por sí solo que una sesión antigua sea ilegítima.

La seguridad real de una sesión debe descansar principalmente en:

handshake autenticado
+
session binding
+
AEAD
+
session lifecycle

El anti-replay debe ser una barrera adicional, no el mecanismo que decide si una sesión es válida.

11. 🟢 MTU 1280 está siendo tratado correctamente como invariante

El proyecto tiene una política explícita:

MAX = 1280

y tests para:

1280 → aceptar
1281 → rechazar

Eso es correcto conceptualmente.

Pero hay que tener cuidado con esta afirmación:

“garantiza cero fragmentación en cualquier red física”.

Eso es demasiado absoluto.

Lo correcto sería:

El protocolo I7 limita su datagrama lógico a 1280 bytes
para evitar depender de fragmentación IP.

La red física subyacente puede tener sus propias características.

12. 🔴 El documento de auditoría está adelantado respecto al código real

Encontré frases como:

100% CUMPLIDO
CERRADO Y CERTIFICADO
100% PASS

pero el CI más reciente:

FAILURE

y además encontré el problema de autorización automática.

Por lo tanto recomiendo cambiar inmediatamente la taxonomía documental.

No:

CERTIFICADO

Sino:

AUDITADO INTERNAMENTE

y:

EVIDENCIA LOCAL

hasta que una ejecución limpia y reproducible pase desde un checkout limpio.

13. 🔴 Hay una inconsistencia muy importante entre documentación y estructura real

ARQUITECTURA.md describe:

src/
├── core/
├── routing/
├── adapters/
└── experimental/

pero el árbol real que audité contiene principalmente:

src/pkg/core
src/pkg/l0
src/pkg/l1
src/pkg/l2
src/pkg/wasm

No es exactamente la arquitectura física que el documento representa.

Esto no es solamente estética.

Para una auditoría futura necesitas poder hacer:

documento → archivo → función → test → evidencia

sin traducciones ambiguas.

14. 🔴 El instalador Windows está incompleto

Además del go:embed roto:

assets/ipvn7.exe
assets/wintun.dll

hay otra cuestión arquitectónica.

El instalador intenta incluir un binario previamente construido:

ipvn7.exe

dentro del propio instalador.

Eso significa que tienes dos pipelines diferentes:

Go source
    ↓
ipvn7.exe
    ↓
installer

pero CI intenta ejecutar:

go vet ./...

antes de tener:

assets/ipvn7.exe
Solución limpia

Separar:

cmd/ipvn7

de:

cmd/installer

y compilar el instalador solamente después de construir el binario.

Ejemplo:

JOB 1
  go test
  go vet
  go build ipvn7

JOB 2
  descargar/compilar Wintun
  copiar ipvn7.exe
  go build installer

JOB 3
  empaquetar release

No meter binarios generados en el source tree.

15. 🟠 El repositorio está creciendo demasiado

Esto me preocupa desde el punto de vista de tu objetivo original de:

núcleo mínimo.

Actualmente conviven:

L0
L1
L2
PQC
X-Wing
Sphinx
SOCKS5
STUN
NAT traversal
UPnP
TUN
Wintun
WASM
MCP
A2A
AP2
x402
Edge AI
Kuzu
Planetary Mesh
Egress
Shadow Devices
Telemetry
QoS
Pacing
...

Eso ya no es un “núcleo mínimo”.

Es prácticamente un:

Network OS + VPN + proxy + AI gateway + agent framework

Eso puede ser una plataforma, pero no debería confundirse con el núcleo I7.

16. Mi separación recomendada

Yo congelaría el núcleo en:

I7 CORE
│
├── Identity
├── Packet
├── Container
├── Session
├── Channel
├── Integrity
├── MTU
├── Anti-Replay
└── Routing

Después:

ADAPTERS
│
├── UDP
├── TUN
├── TCP
├── QUIC
└── WireGuard

Después:

SERVICES
│
├── SOCKS5
├── WebUI
├── STUN
├── NAT traversal
└── Discovery

Y fuera del núcleo:

EXPERIMENTAL
│
├── Sphinx
├── Planetary
├── WASM
├── AI
├── MCP
├── A2A
├── AP2
├── x402
└── Egress intelligence

Esto encaja mucho mejor con tu principio:

Core transporta estructura, no semántica.

17. Orden exacto de corrección

Yo no seguiría agregando funcionalidades todavía.

Haría esto:

P0 — Seguridad
 Eliminar AuthorizeDID() automático desde RoamingUpdate.
 Separar autenticación de autorización.
 Hacer que ZTNA sea realmente Default-Deny.
 Cerrar WebUI a 127.0.0.1 por defecto.
 Autenticación criptográfica para administración.
 Proteger /vpn/exit.
 Proteger /vpn/connect.
 Proteger /update/*.
 Proteger /mcp.
 Proteger /a2a.
 Eliminar CORS * en administración.
P0 — CI
 Resolver assets/ipvn7.exe.
 Resolver assets/wintun.dll.
 Ejecutar go vet ./....
 Ejecutar go test ./....
 Ejecutar go test -race.
 Ejecutar fuzz.
 Compilar Windows.
 Compilar Linux.
 Compilar macOS.
P1 — Criptografía
 Eliminar nomenclatura falsa de ML-DSA.
 Revisar X-Wing contra el estándar exacto.
 Firmar manifiestos de actualización.
 Verificar binding completo del handshake.
 Revisar lifecycle de SessionID.
P1 — Arquitectura
 Congelar L0.
 Congelar L1.
 Sacar experimental del camino crítico.
 Actualizar ARQUITECTURA.md.
 Crear una matriz función → archivo → test → evidencia.
P2 — Rendimiento

Recién después:

 benchmark real;
 allocations;
 pacing;
 congestion control;
 MTU dinámico;
 multipath.
18. El punto más importante

Hay una diferencia entre:

"Tenemos mucho código y muchos tests"

y:

"Tenemos un protocolo pequeño, seguro y demostrado."

IPVN7 0.7 está todavía en el primer estado.

Pero ya tiene piezas técnicamente interesantes y reales, especialmente:

ML-KEM-768
X25519
Ed25519
ChaCha20-Poly1305
Anti-Replay
MTU enforcement
UDP real
TUN
routing

Lo que ahora toca no es añadir más cosas.

Es reducir, separar y demostrar.

Prioridad absoluta

Los tres problemas que yo solucionaría antes de cualquier nueva feature son:

1. 🔴 ZTNA auto-authorize
2. 🔴 WebUI/API sin autenticación
3. 🔴 CI roto por installer/assets

Después de esos tres, volvería a ejecutar una auditoría desde checkout limpio, y recién entonces tendría sentido hablar de “cierre”.

dame una guia exacta separada por fases como corregir y como programar de aqui en adelante cambiando skill y agentes

Sí. Para este repositorio yo cambiaría la forma de programarlo antes de seguir agregando funcionalidades. El problema principal ya no es “falta código”: es que el proyecto necesita una disciplina de arquitectura, seguridad, pruebas y agentes.

Te propongo trabajar desde ahora con fases cerradas, y que cada agente tenga una responsabilidad única.

FASE 0 — Congelar el estado actual

Objetivo: que ningún agente siga modificando cosas mientras no sepamos qué está realmente funcionando.

No agregar funcionalidades.

Checklist:

 Crear rama audit/baseline.
 Registrar commit actual.
 Ejecutar go test ./....
 Ejecutar go vet ./....
 Ejecutar go build ./....
 Registrar exactamente qué falla.
 Ejecutar tests de seguridad existentes.
 Registrar los binarios que realmente se pueden compilar.
 Crear docs/BASELINE.md.
 Eliminar de la documentación afirmaciones como 100% PASS, CERTIFICADO o CERRADO si no están demostradas por CI.

Regla del agente:

No arreglar todavía. Solo medir y documentar.

FASE 1 — Seguridad crítica

Esta es la primera fase de programación real.

1.1 Corregir ZTNA

Actualmente tienes el problema conceptual:

firma válida
      ↓
AuthorizeDID()
      ↓
acceso permitido

Debe quedar:

paquete
   ↓
validar formato
   ↓
verificar firma
   ↓
¿DID está autorizado?
   ├── NO → DROP
   └── SÍ → continuar
Cambiar

En:

src/cmd/ipvn7/main.go

Eliminar la autorización automática producida por MsgTypeRoamingUpdate.

AuthorizeDID() debe ser una operación de política, no una consecuencia de autenticación.

Test obligatorio

Crear pruebas:

TestRoamingValidSignatureUnauthorizedDID
    → DROP

TestRoamingValidSignatureAuthorizedDID
    → ACCEPT

TestRoamingInvalidSignature
    → DROP

TestRoamingUnknownDID
    → DROP
FASE 2 — Cerrar completamente WebUI

Actualmente:

0.0.0.0:7070

es demasiado peligroso para una interfaz administrativa.

Objetivo inicial
127.0.0.1:7070

y solamente posteriormente permitir administración remota mediante un mecanismo autenticado.

Separar:
WebUI pública
        ≠
WebUI administrativa

La administración debe estar detrás de:

Authentication
      ↓
Authorization
      ↓
Operation

No:

HTTP
 ↓
Operation
Rutas críticas

Revisar especialmente:

/vpn/connect
/vpn/disconnect
/vpn/exit
/vpn/cycle
/update/apply
/update/rollback
/mcp
/a2a

Cada una debe tener una política explícita.

Tests

Crear:

TestAdminWithoutAuth
    → 401/403

TestAdminAuthenticatedUnauthorized
    → 403

TestAdminAuthorized
    → operation allowed

Y eliminar:

Access-Control-Allow-Origin: *

del panel administrativo.

FASE 3 — Eliminar SSRF del sistema de actualización

Actualmente:

/api/v1/update/check?url=...

permite introducir una URL arbitraria.

Eso debe desaparecer.

Arquitectura nueva
Configuración
     ↓
Trusted Update Host
     ↓
Manifest
     ↓
Firma digital
     ↓
SHA-256
     ↓
Binary
     ↓
Install

No:

usuario → URL arbitraria → descargar → instalar
Seguridad del update

Usar:

Ed25519 signature
        +
SHA-256

El hash demuestra integridad.

La firma demuestra autenticidad del publicador.

FASE 4 — Arreglar CI antes de seguir

Aquí hay una regla importante:

Nunca vuelvas a programar una funcionalidad nueva sobre un CI roto.

El problema conocido es:

cmd/installer/main.go

requiere:

assets/ipvn7.exe
assets/wintun.dll

pero esos archivos no existen en el checkout normal.

Arquitectura correcta
Job 1 — Core
go test ./...
go vet ./...
go build ./...

Debe funcionar sin instalador.

Job 2 — Windows package
build ipvn7.exe
        ↓
obtener wintun.dll
        ↓
crear assets/
        ↓
compilar installer
Job 3 — Release
core binary
installer
checksums
signed manifest

Así el repositorio fuente no depende de binarios generados.

FASE 5 — Limpiar criptografía

Aquí hay que ser especialmente estricto.

Tienes implementación real de:

ML-KEM-768
X25519
Ed25519
HKDF
ChaCha20-Poly1305

Eso debe permanecer.

Pero hay nombres que sugieren ML-DSA cuando realmente no tienes una implementación ML-DSA equivalente.

Regla

Si no es el algoritmo real:

NO llamarlo ML-DSA

Cambiar nombres experimentales como:

MLDSAPubHex
MLDSA65SeedSize

por nombres honestos.

Por ejemplo:

ExperimentalSignatureSeed
ExperimentalPublicIdentifier

hasta implementar realmente ML-DSA.

FASE 6 — Congelar el CORE IPv7

Aquí haría el cambio arquitectónico más importante.

Tu CORE debería quedar pequeño.

CORE
Identity
Packet
Container
Object
Session
Channel
Integrity
AntiReplay
MTU
Routing

Nada más.

La regla debe ser:

El CORE transporta estructura. No transporta semántica de aplicaciones.

Por tanto:

CORE
 │
 ├── Identity
 ├── Packet
 ├── Container
 ├── Session
 ├── Channel
 ├── Integrity
 ├── AntiReplay
 ├── MTU
 └── Routing

Después:

ADAPTERS
 │
 ├── UDP
 ├── TCP
 ├── QUIC
 ├── TUN
 └── WireGuard

Después:

SERVICES
 │
 ├── Discovery
 ├── STUN
 ├── NAT
 ├── SOCKS5
 └── WebUI

Y fuera del camino crítico:

EXPERIMENTAL
 │
 ├── Sphinx
 ├── Planetary
 ├── WASM
 ├── AI
 ├── MCP
 ├── A2A
 ├── AP2
 ├── x402
 └── Egress intelligence
FASE 7 — Crear contratos del CORE

Antes de añadir código nuevo, cada componente debe tener una interfaz clara.

Por ejemplo:

Packet
Container
Session
Channel
Identity
Path

Cada uno debe responder:

¿Qué representa?
¿Qué datos contiene?
¿Quién lo crea?
¿Quién lo modifica?
¿Quién lo destruye?
¿Qué invariantes tiene?
¿Qué errores puede producir?
¿Cómo se prueba?

Crear:

docs/core/
    identity.md
    packet.md
    container.md
    session.md
    channel.md
    routing.md
    mtu.md
    anti_replay.md
FASE 8 — Contratos matemáticos

Aquí aprovecharía tu filosofía de matemática sobre algoritmos innecesariamente complejos.

Cada elemento crítico debe tener invariantes.

Ejemplo:

MAX_DATAGRAM = 1280

Entonces:

payload + header <= 1280

Debe existir una prueba que lo demuestre.

Para sesiones:

SessionID != 0
SessionID pertenece a una sesión válida
SessionID no puede reutilizarse incorrectamente

Para anti-replay:

(originDID, sessionID, sequence)

debe determinar correctamente la ventana.

Esto debe convertirse en tests, no solamente documentación.

FASE 9 — Pruebas destructivas

Después de seguridad y CORE.

Crear una batería permanente:

tests/
 ├── unit/
 ├── integration/
 ├── security/
 ├── adversarial/
 ├── interoperability/
 ├── performance/
 └── fuzz/
Seguridad

Probar:

packet corrupto
firma inválida
DID desconocido
DID válido pero no autorizado
replay
sequence inválido
session inválida
container corrupto
MTU > 1280
fragmentación incorrecta
downgrade
handshake incompleto
Fuzzing

Especialmente:

Packet parser
Container parser
TLV parser
Session parser
Routing parser
FASE 10 — Recién aquí rendimiento

No optimizar antes.

Medir:

throughput
latency
jitter
CPU
RAM
allocations
packet rate
pacing
MTU

Y especialmente tu idea de:

velocidad constante

en lugar de:

máximo → congestión → pérdida → recuperación

Pero primero medirla.

No declararla superior antes de tener resultados.

FASE 11 — Interoperabilidad

Después:

IPv7 node A
       ↓
IPv7 node B
       ↓
IPv7 node C

Probar:

A → B
A → C
A → B → C

y después:

UDP
TUN
WireGuard
QUIC

La regla:

El CORE no debería saber qué transporte físico hay debajo.

FASE 12 — Servicios

Solo después de que el CORE esté estable:

Chat
Files
VPN
Remote Support
IoT
Gateway

Estos deben consumir el CORE.

No modificarlo para acomodar cada aplicación.

CAMBIO DE SKILLS Y AGENTES

Aquí está probablemente el cambio más importante para tu flujo con IA.

No uses un único agente para todo.

Crea agentes especializados.

AGENTE 1 — Arquitecto

Responsabilidad:

arquitectura
interfaces
dependencias
límites

No programa.

Su trabajo:

¿Dónde debería vivir esto?
AGENTE 2 — Seguridad

Responsabilidad exclusiva:

ZTNA
auth
authorization
crypto
replay
SSRF
WebUI
updates

No agrega funcionalidades.

Pregunta permanente:

¿Cómo puede abusarse de esto?

AGENTE 3 — CORE

Solo:

Identity
Packet
Container
Session
Channel
MTU
AntiReplay
Routing

No tocar:

MCP
AI
WASM
Sphinx
A2A
x402
AGENTE 4 — Testing

No desarrolla funcionalidades.

Su función:

buscar fallos
crear tests
fuzzing
regresiones

Debe intentar romper lo construido.

Este agente debería ser hostil al código.

AGENTE 5 — CI/Build

Solo:

GitHub Actions
Go versions
Windows
Linux
macOS
installer
release
artifacts

No modificar protocolos.

AGENTE 6 — Performance

Solo entra después de que:

security = OK
CI = OK
tests = OK

Trabaja con mediciones reales.

AGENTE 7 — Documentación

No inventa capacidades.

Solo puede escribir:

HECHO
TESTEADO
MEDIDO
NO IMPLEMENTADO
EXPERIMENTAL

Esto evita que la IA convierta una hipótesis en una característica “certificada”.

REGLA DE TRABAJO ENTRE AGENTES

Nunca:

Agente A modifica
      ↓
Agente B modifica
      ↓
Agente C modifica
      ↓
todo mezclado

Usaría:

ARQUITECTO
     ↓
PLAN
     ↓
PROGRAMADOR
     ↓
TESTER
     ↓
SEGURIDAD
     ↓
CI
     ↓
MERGE

Cada etapa debe poder rechazar el trabajo anterior.

REGLA DE LOS COMMITS

Cada commit debe responder una sola pregunta.

Bien:

fix: prevent unauthorized roaming authorization

Bien:

test: reject replayed session packets

Bien:

fix: bind admin webui to localhost

Mal:

improve ipv7 security performance architecture installer
REGLA DE ORO PARA LOS AGENTES

Yo pondría este principio en todos tus skills:

NO INVENTAR.

Si algo no está demostrado por código, test,
benchmark o documentación verificable,
debe declararse como hipótesis o pendiente.

Nunca convertir:
HIPÓTESIS → HECHO
INTENCIÓN → IMPLEMENTACIÓN
TEST UNITARIO → SEGURIDAD COMPLETA
BUILD LOCAL → CI REPRODUCIBLE
HASH → AUTENTICIDAD
FIRMA VÁLIDA → AUTORIZACIÓN

Y otra regla:

ANTES DE PROGRAMAR:

1. leer arquitectura existente
2. localizar código afectado
3. identificar invariantes
4. identificar tests existentes
5. explicar cambio mínimo
6. programar
7. probar
8. revisar seguridad
9. ejecutar CI
10. documentar evidencia
ORDEN EXACTO QUE YO USARÍA

No intentaría arreglar todo simultáneamente.

FASE 0
BASELINE
   ↓
FASE 1
ZTNA
   ↓
FASE 2
WEBUI / ADMIN
   ↓
FASE 3
UPDATE / SSRF
   ↓
FASE 4
CI / BUILD
   ↓
FASE 5
CRYPTO
   ↓
FASE 6
CORE
   ↓
FASE 7
CONTRATOS
   ↓
FASE 8
INVARIANTES
   ↓
FASE 9
FUZZ / ADVERSARIAL
   ↓
FASE 10
PERFORMANCE
   ↓
FASE 11
INTEROPERABILIDAD
   ↓
FASE 12
SERVICIOS

Y recién después volvería a evaluar si realmente necesitas:

Sphinx
Planetary Mesh
WASM
AI
MCP
A2A
AP2
x402
Edge AI
Kuzu
Egress

No necesariamente hay que eliminarlos. Hay que impedir que contaminen el núcleo.

La meta no debería ser tener un repositorio enorme.

La meta debería ser poder decir:

“Este pequeño núcleo hace exactamente X, puedo demostrarlo con estas pruebas, y todo lo demás se conecta a él sin modificarlo.”

Ese sería, para mí, el cambio fundamental de tu IPv7 0.7: pasar de “seguir construyendo funcionalidades” a construir un núcleo que pueda sobrevivir a los agentes que lo programan.