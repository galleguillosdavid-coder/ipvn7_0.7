# CONTRATO DEL CORE: MTU (MAXIMUM TRANSMISSION UNIT)

## 1. ¿Qué representa?
El invariante dimensional que limita el tamaño máximo del datagrama lógico IPVN7 a nivel de cable para evitar la fragmentación en la capa de red IP (RFC 8200).

## 2. ¿Qué datos contiene?
* `MaxDatagramSize = 1280`: Límite superior constante en bytes para cualquier datagrama serializado.
* `MaxPayloadSize = 1280 - HeadersSize`: Espacio útil disponible para datos cifrados.

## 3. ¿Quién lo crea?
Constante canónica definida en `src/pkg/l0/wire.go`.

## 4. ¿Quién lo modifica?
Inmutable. Ninguna capa, protocolo ni adaptador puede generar o aceptar un datagrama lógico que supere los 1280 bytes.

## 5. ¿Quién lo destruye?
N/A (invariante constante).

## 6. ¿Qué invariantes tiene?
* `Len(datagrama) <= 1280` bytes.
* Cualquier paquete recibido con longitud superior a 1280 bytes es descartado de inmediato en tiempo $O(1)$ sin alocar memoria en el heap.
* El apretón de manos post-cuántico (HandshakeInit / HandshakeResp con ML-KEM-768 FIPS 203 + X25519 + firma Ed25519) debe caber íntegramente en un único datagrama `<= 1280` bytes sin fragmentación.

## 7. ¿Qué errores puede producir?
* `ErrDatagramTooLarge`: Intento de serializar un paquete con carga útil que exceda los 1280 bytes.

## 8. ¿Cómo se prueba?
* `src/pkg/l0/l0_test.go`: Test explícito de frontera: `1280` -> ACEPTAR, `1281` -> RECHAZAR.
* `src/pkg/l1/session_adversarial_test.go`: Verificación de tamaño de HandshakeInit con firma (1216B <= 1280B).
