# CONTRATO DEL CORE: SESSION

## 1. ¿Qué representa?
El estado criptográfico efímero y seguro establecido entre dos identidades DID soberanas tras completar con éxito un apretón de manos 1-RTT post-cuántico (ML-KEM-768 FIPS 203 + X25519).

## 2. ¿Qué datos contiene?
* `SessionID`: Identificador único de 64 bits de la sesión.
* `PeerDID`: DID del par remoto.
* `TxKey`: Clave simétrica de transmisión (32 bytes ChaCha20-Poly1305).
* `RxKey`: Clave simétrica de recepción (32 bytes ChaCha20-Poly1305).
* `CreatedAt`: Timestamp de inicio de la sesión.
* `LastSeen`: Timestamp de última actividad verificada.
* `AntiReplay`: Instancia de ventana anti-replay dedicada para esta sesión específica.
* `TxSequence`: Contador monotónico de transmisión.

## 3. ¿Quién lo crea?
El `SessionManager` (`src/pkg/l1/session_manager.go`) al completar el handshake 1-RTT (`MsgTypeHandshakeInit` y `MsgTypeHandshakeResp`).

## 4. ¿Quién lo modifica?
* El transmisor incrementa `TxSequence` en cada datagrama.
* El receptor actualiza `LastSeen` y la ventana anti-replay.
* Las claves se rotan periódicamente según la política de Forward Secrecy.

## 5. ¿Quién lo destruye?
El `SessionManager` por expiración de inactividad (TTL) o cierre explícito de conexión.

## 6. ¿Qué invariantes tiene?
* `SessionID != 0`.
* Aislamiento estricto: Las secuencias y anti-replay de una sesión nunca interfieren con otra sesión del mismo o distinto par.
* Las claves simétricas `TxKey` y `RxKey` derivan de un secreto compartido con decapsulación real ML-KEM-768 vía HKDF-SHA256.

## 7. ¿Qué errores puede producir?
* `ErrSessionNotFound`: Datagrama de datos recibido sin sesión establecida previa.
* `ErrSessionExpired`: Sesión caducada por inactividad.
* `ErrDecryptionFailed`: Fallo en AEAD ChaCha20-Poly1305 (contenido corrupto o clave incorrecta).
* `ErrReplayDetected`: Datagrama duplicado o fuera de ventana.

## 8. ¿Cómo se prueba?
* `src/pkg/l1/session_adversarial_test.go`: Tests A a H (suplantación, forjado, aislamiento cross-session y cross-peer, descarte de replay).
* `src/pkg/l0/pqc_kem_test.go`: Tests de decapsulación e implicit rejection de ML-KEM-768.
