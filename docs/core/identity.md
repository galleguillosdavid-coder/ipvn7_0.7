# CONTRATO DEL CORE: IDENTITY

## 1. ¿Qué representa?
Representa la identidad criptográfica soberana e inmutable de un nodo o actor en la red IPVN7, desacoplada por completo de su dirección física o de red efímera (IP:puerto).

## 2. ¿Qué datos contiene?
* `DID`: Cadena canónica con formato `did:ipvn7:<Base58/Hex(PublicKeyEd25519)>`.
* `PublicKey`: Clave pública Ed25519 (32 bytes).
* `PrivateKey`: Clave privada Ed25519 (64 bytes, custodiada en keystore local).
* `PQCIdentity`: Identificador o clave pública post-cuántica (ML-KEM-768 / experimental).

## 3. ¿Quién lo crea?
El keystore local del nodo (`l0.GenerateIdentity` o `l0.LoadIdentity`) al arrancar el daemon o al inicializar un nuevo perfil de identidad.

## 4. ¿Quién lo modifica?
Inmutable. Una identidad una vez generada no cambia de clave pública ni de DID. La rotación de identidades genera un nuevo DID soberano independiente.

## 5. ¿Quién lo destruye?
Se mantiene en memoria mientras el proceso del nodo esté en ejecución. La eliminación del archivo en disco destruye la identidad física del nodo.

## 6. ¿Qué invariantes tiene?
* Longitud de clave Ed25519: exactamente 32 bytes para pública, 64 bytes para privada.
* El DID deriva deterministamente de la clave pública: `DID == "did:ipvn7:" + encode(PublicKey)`.
* Cero simulación: firmas reales generadas con `crypto/ed25519`.

## 7. ¿Qué errores puede producir?
* `ErrInvalidPrivateKey`: Clave privada corrupta o longitud no conforme a RFC 8032.
* `ErrDIDMismatch`: El DID recibido no se corresponde matemáticamente con la clave pública adjunta.
* `ErrSignatureInvalid`: Fallo en la verificación criptográfica de la firma del datagrama.

## 8. ¿Cómo se prueba?
* `src/pkg/l0/identity_test.go`: Generación, derivación determinista de DID, firma y verificación.
* `src/pkg/l1/session_adversarial_test.go`: Test A (rechazo estricto ante suplantación de DID) y Test B (rechazo de firmas forjadas).
