# CONTRATO DEL CORE: PACKET

## 1. ¿Qué representa?
La unidad elemental de datos y control estructurada que viaja serializada en el protocolo IPVN7 a nivel de cable. Transporta metadatos de capa baja y la carga útil.

## 2. ¿Qué datos contiene?
* `Magic`: 4 bytes identificadores de protocolo (`0x49 0x50 0x37 0x31` -> "IP71").
* `Version`: Byte de versión de protocolo (`0x07`).
* `Type`: Tipo de mensaje (`MsgTypeHandshakeInit`, `MsgTypeHandshakeResp`, `MsgTypeData`, `MsgTypeRoamingUpdate`, `MsgTypeBeacon`, etc.).
* `SourceDID`: DID del emisor.
* `DestDID`: DID del destinatario final.
* `Sequence`: Entero uint64 monotónico creciente para control de secuencia y anti-replay.
* `Timestamp`: Timestamp Unix en segundos o nanosegundos para caducidad y frescura.
* `Payload`: Bytes de datos útiles (cifrados o en claro según tipo).
* `Signature`: Firma digital Ed25519 (64 bytes) que cubre todo el datagrama.

## 3. ¿Quién lo crea?
Los módulos de origen del nodo emisor (`l0.CreatePacket`, `l1.CreateHandshakeInitPacket`, etc.) al emitir tráfico.

## 4. ¿Quién lo modifica?
Inmutable en tránsito. Cualquier alteración de un bit invalida el checksum y la firma digital.

## 5. ¿Quién lo destruye?
El receptor descarta y recicla los buffers en el `bufferPool` una vez procesado el datagrama en el pipeline (`l0.RecycleBuffer`).

## 6. ¿Qué invariantes tiene?
* Tamaño total serializado `<= 1280` bytes (Invariante canónico de MTU).
* La firma digital Ed25519 es obligatoria en paquetes de control y señalización.
* Deserialización determinista sin alocaciones intermedias innecesarias (zero-copy).

## 7. ¿Qué errores puede producir?
* `ErrPacketTooLarge`: Longitud superior a 1280 bytes.
* `ErrInvalidMagic`: Cabecera no coincide con "IP71".
* `ErrCorruptedPayload`: Deserialización CBOR truncada o malformada.
* `ErrSignatureMismatch`: Firma inválida al verificar con la clave pública de `SourceDID`.

## 8. ¿Cómo se prueba?
* `src/pkg/l0/l0_test.go`: Serialización, deserialización, tamaño y verificación.
* `src/pkg/l0/wire_fuzz_test.go`: Fuzzing de tramas truncadas, bit-flips y descarte determinista $O(1)$.
