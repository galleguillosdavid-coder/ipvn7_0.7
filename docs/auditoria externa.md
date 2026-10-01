revisa 
galleguillosdavid-coder/ipvn7_0.7

Sí. Revisé directamente el repositorio galleguillosdavid-coder/ipvn7_0.7, incluyendo arquitectura, L0/L1/L2, PQC, routing, pacing, DHT, malla planetaria, tests y documentación de auditoría.

Mi conclusión inicial es importante:

Hay bastante código real y varias piezas están efectivamente implementadas, pero la auditoría interna del propio repositorio sobreestima el nivel de seguridad/producción en algunos puntos. Encontré al menos un problema de seguridad crítico que debe corregirse antes de seguir agregando funcionalidades.

1. Estado general que veo

El repositorio ya es grande: el árbol contiene una implementación Go bastante extensa, con L0/L1/L2, TUN/Wintun, UDP, routing, PQC, firewall, DHT, pacing, WebUI, etc.

La arquitectura documentada es coherente con la dirección que hemos venido buscando:

APPLICATIONS / PROFILES
          ↓
      I7 CORE
 Identity
 Object
 Container
 Session
 Channel
 Path
 MTU
 Integrity
 Routing
 Capability
          ↓
 TRANSPORT ADAPTERS
 UDP / TUN
          ↓
      RED FÍSICA

Eso está bien encaminado conceptualmente.

Además, el repositorio tiene una característica que considero positiva: sí distingue DEMOSTRADO / IMPLEMENTADO / EXPERIMENTAL / NO DEMOSTRADO. Eso es mucho mejor que declarar todo como "production-ready".

2. 🔴 Problema crítico: hay una clave privada real dentro del repositorio

Encontré:

keystore/node_identity.key

y contiene una estructura con:

public_key
private_key

Es decir, una clave privada Ed25519 está actualmente versionada en GitHub.

Esto contradice directamente la afirmación de la auditoría:

"No hay hardcoded secrets"

y también hace que la frase:

"Keystore seguro"

sea engañosa en el contexto del repositorio público.

El problema no es que el archivo tenga permisos 0600 en una máquina local. Una vez que la clave privada está en Git, esos permisos ya no protegen el secreto histórico del repositorio.

Esto es prioridad 0.

Hay que:

Revocar/abandonar esa identidad.
Generar una identidad nueva.
Eliminar keystore/node_identity.key del árbol actual.
Eliminarla también del historial Git.
Añadir keystore/*.key y equivalentes al .gitignore.
Hacer que el programa genere el keystore localmente.
Nunca volver a distribuir una identidad privada dentro del repositorio.

Y no basta con borrar el archivo en el próximo commit, porque la clave continúa en el historial.

No voy a reproducir aquí la clave privada encontrada.

3. 🔴 Segundo problema: pqc_hybrid.go tiene un fallback que no es ML-KEM

Este punto es mucho más importante de lo que parece.

El código declara:

HybridKEMAlgorithm = "X25519+ML-KEM-768"

pero Encapsulate() hace esto:

if len(targetKEMBytes) == mlkem.EncapsulationKeySize768 {
    ...
    pqcSecret, fullCipher = ek.Encapsulate()
}

if pqcSecret == nil {
    pqcSecret, err = deriveFallbackPQC(...)
}

El fallback calcula un secreto mediante SHA-256.

Por tanto:

clave ML-KEM válida
        ↓
ML-KEM real

pero:

clave ML-KEM inválida/truncada
        ↓
SHA-256 personalizado
        ↓
se sigue presentando como HybridKEMAlgorithm

Eso no es ML-KEM.

La función incluso se llama:

deriveFallbackPQC

Esto es peligroso en un protocolo criptográfico porque puede crear una falsa sensación de que el canal continúa siendo post-cuántico cuando en realidad no lo es.

Mi recomendación

En el núcleo de protocolo:

ML-KEM inválido
       ↓
ERROR
       ↓
NO ESTABLECER SESIÓN

No:

ML-KEM inválido
       ↓
"inventemos otro secreto"

Si quieres compatibilidad degradada, debe ser explícita:

HYBRID_PQC
CLASSICAL_ONLY
EXPERIMENTAL_FALLBACK

y jamás compartir el mismo identificador criptográfico.

4. 🟠 Hay una confusión de tamaños en el híbrido

En pqc_hybrid.go aparece:

MLKEM768CipherSize = 128

pero ML-KEM-768 tiene un ciphertext de 1088 bytes.

El propio proyecto lo sabe: l0/pqc_kem.go define correctamente:

MLKEM768CiphertextBytes = 1088

El problema es que L1 crea:

pqcCipher := make([]byte, MLKEM768CipherSize)

y copia solamente los primeros 128 bytes:

copy(pqcCipher, fullCipher[:MLKEM768CipherSize])

Después guarda simultáneamente:

PQCCiphertext
FullPQCCiphertext

Esto produce dos representaciones diferentes del mismo KEM.

La ruta completa funciona porque Decapsulate() utiliza:

FullPQCCiphertext

cuando tiene los 1088 bytes.

Pero conceptualmente está mal diseñado.

Yo lo simplificaría

Para el protocolo real:

X25519 ephemeral      32 B
ML-KEM-768 ciphertext 1088 B
--------------------------------
KEM ciphertext        1120 B

y nada de:

"ciphertext compacto" = primeros 128 bytes

Los primeros 128 bytes no son un ciphertext ML-KEM válido independiente.

5. 🟢 X-Wing está mucho mejor planteado

La parte:

XWingEncapsulate()
XWingDecapsulate()
DeriveXWingSharedSecret()

es conceptualmente bastante más limpia.

El código utiliza:

ML-KEM-768
+
X25519
+
SHA-256 KDF

y construye:

1088 + 32 = 1120 bytes

También hay test de round-trip:

Alice
 ↓
X-Wing encapsulate
 ↓
UDP / wire
 ↓
Bob
 ↓
X-Wing decapsulate
 ↓
same shared secret

Eso sí es una base útil para el protocolo.

Pero yo separaría claramente:

FIPS 203 ML-KEM

de:

X-Wing

en la API. Ahora están demasiado mezclados dentro de pqc_hybrid.go.

6. 🟠 ML-DSA: la implementación declarada no coincide completamente con lo que hace el código

La documentación afirma:

Ed25519 + ML-DSA-65

pero GenerateHybridKeyPair() hace:

pqcSignSeed := make([]byte, 32)

y luego:

SHA256("ML-DSA-65-PUBLIC-MATRIX-DERIVATION" + seed)

para producir algo llamado:

MLDSAPubHex

Eso no constituye una implementación de ML-DSA-65.

Hay que distinguir:

ML-DSA real

de:

identificador/hash experimental relacionado con una semilla

Los tests pueden comprobar que ese mecanismo funciona internamente, pero eso no lo convierte en FIPS 204.

Esto es exactamente el tipo de cosa que nuestro principio de creatividad fundamentada en la realidad debería detectar.

7. 🟢 El routing sí tiene bastante sustancia

routing.go no es simplemente pseudocódigo.

Hay:

DID real.
extracción de clave pública.
distancia XOR.
rings.
límites de peers.
health state.
latency.
jitter.
pérdida.
cuarentena.
eviction.
roaming.
locks concurrentes.

La FSM:

HEALTHY
   ↓
DEGRADED
   ↓
UNSTABLE
   ↓
UNREACHABLE
   ↓
QUARANTINED

es interesante para el concepto que veníamos desarrollando de nodos buenos que reciben más recursos y nodos problemáticos que pierden privilegios.

Pero no llamaría todavía a esto "routing Kleinberg probado a escala".

Es:

un algoritmo de selección/routing local inspirado en Kleinberg con métricas adicionales.

Eso es mucho más preciso.

8. 🟠 El DHT no es todavía una DHT completa

El propio comentario de dht_kademlia.go lo reconoce:

"No constituye una DHT distribuida completa con RPCs de red"

Y estoy de acuerdo.

Actualmente tienes:

NodeID
XOR distance
K-buckets
replacement cache
FindClosest

Eso es una tabla de routing Kademlia local.

Faltan las operaciones distribuidas propiamente tales:

PING
FIND_NODE
FIND_VALUE
STORE
iterative lookup
alpha concurrency
timeouts
republishing
expiration
network discovery

Por tanto:

HECHO

Kademlia-style routing table local.

NO DEMOSTRADO

DHT distribuida funcional.

Eso debería mantenerse explícito.

9. 🔴 La "malla planetaria" todavía es simulación

Aquí encontré algo muy importante.

planetary_mesh.go crea:

node-000
node-001
...

en una retícula 2D artificial.

Luego genera:

ShortLinks
LongLinks

con probabilidad:

1 / distance²

Eso sirve como simulación matemática de Kleinberg.

Pero no demuestra una malla planetaria real.

Más importante todavía:

SimulateFailure()

utiliza:

rand.Float64()

mientras la construcción usa un RNG local determinista.

Y el test:

SimulateFailure(0.10)

no comprueba realmente que la red continúe funcionando después de la caída.

Solo registra cuántos nodos cayeron.

Por tanto el nombre:

Scale50NodesAndResilience

es demasiado fuerte.

Realmente prueba:

Scale50NodesAndFailureSimulation

No resiliencia.

10. 🟠 El pacing tiene una buena idea, pero no es todavía control de congestión

PacketPacer utiliza correctamente un token bucket:

golang.org/x/time/rate

y además:

MaxBurstPackets = 1

Esto encaja muy bien con tu idea anterior:

velocidad estable en lugar de ráfagas + pérdida.

Pero hay un detalle fundamental:

BottleneckRateBytesPerSec

es un valor configurado.

No se está estimando automáticamente el verdadero:

BtlBw
RTprop

Por tanto:

IMPLEMENTADO

Pacing token bucket.

NO IMPLEMENTADO

BBR.

La propia documentación lo reconoce, y eso está bien.

Yo incluso eliminaría temporalmente cualquier referencia a BBR del núcleo hasta tener una implementación real.

11. 🔴 Hay una inconsistencia muy importante entre la auditoría y el código

La auditoría dice:

Health Score Global: 99%

y:

APROBADO CON EXCELENCIA

y:

production-ready

Pero el propio código todavía contiene:

fallback criptográfico no-Ml-KEM;
ML-DSA que no es ML-DSA;
DHT que no es DHT distribuida;
planetary mesh simulada;
resiliencia no realmente verificada;
keystore privado dentro del repositorio.

Por lo tanto yo no usaría ese 99%.

No porque el proyecto sea malo, sino porque ese número está mezclando:

"tests pasan"

con:

"protocolo seguro y demostrado"

Son cosas diferentes.

12. Lo más interesante del proyecto

Después de revisar el código, creo que el núcleo que vale la pena conservar es bastante más pequeño que todo el repositorio.

Yo lo reduciría conceptualmente a:

                    I7
                     │
        ┌────────────┼────────────┐
        │            │            │
     IDENTITY     CONTAINER     SESSION
        │            │            │
        └────────────┼────────────┘
                     │
                 PATH / MTU
                     │
                  ROUTING
                     │
             ┌───────┴───────┐
             │               │
            UDP             TUN

Y alrededor:

PQC
Pacing
Discovery
NAT traversal
DHT
Gateway
VPN
Applications

como módulos.

Esto se acerca mucho más a la filosofía que hemos venido buscando:

el núcleo transporta estructura, no semántica.

13. Mi clasificación actual
Componente	Mi evaluación
Identidad Ed25519	🟢 Real
DID derivado de clave	🟢 Real
IPv4/IPv6 virtual derivada	🟢 Implementado
UDP	🟢 Real
ChaCha20-Poly1305	🟢 Real
ML-KEM-768 L0	🟢 Real
X25519 + ML-KEM	🟡 Real con diseño híbrido que debe limpiarse
X-Wing	🟡 Implementación funcional que requiere validación externa/interoperabilidad
ML-DSA-65	🔴 La implementación mostrada no es ML-DSA-65
Token Bucket	🟢 Real
BBR autónomo	🔴 No implementado
Routing XOR/Kleinberg	🟡 Implementado localmente
Kademlia routing table	🟢 Implementado localmente
DHT distribuida	🔴 No implementada
Planetary Mesh	🟠 Simulación
Resiliencia planetaria	🔴 No demostrada
VPN/TUN	🟠 Hay implementación, falta validar extremo a extremo en escenarios reales
Seguridad del repositorio	🔴 Problema crítico por clave privada versionada
14. Qué haría ahora

Yo no agregaría ninguna funcionalidad nueva todavía.

Haría una fase que llamaría:

HARDENING-0.7.1

En este orden:

1. 🔴 Eliminar identidad privada del Git
2. 🔴 Rotar identidad
3. 🔴 Eliminar fallback PQC
4. 🔴 Separar ML-KEM real de mecanismos experimentales
5. 🔴 Corregir ML-DSA: real o renombrarlo
6. 🟠 Eliminar "99% / production-ready"
7. 🟠 Separar:
       DEMOSTRADO
       IMPLEMENTADO
       EXPERIMENTAL
       HIPÓTESIS
8. 🟠 Tests físicos 2 nodos
9. 🟠 Tests físicos 3 nodos
10. 🟠 MTU real
11. 🟠 pérdida real
12. 🟠 cambio de ruta
13. 🟠 pacing real
14. 🟡 recién después DHT
15. 🟡 recién después WAN/NAT
Y hay una regla que pondría en el proyecto:

Ninguna función criptográfica puede degradar silenciosamente a otra función y seguir declarando el mismo algoritmo.

Esa regla es particularmente importante para I7.

Mi impresión después de esta revisión

El proyecto ya dejó de ser simplemente una idea. Hay bastante implementación real.

Pero también veo que el repositorio está entrando en una fase peligrosa: la arquitectura está creciendo más rápido que la demostración física de cada componente.

Y eso es exactamente lo que conviene detener ahora.

La siguiente etapa no debería ser "hacer IPv7 más grande".

Debería ser:

hacer que cada pieza que ya existe sea inequívocamente real, segura y demostrable.

Repositorio ipvn7_0.7