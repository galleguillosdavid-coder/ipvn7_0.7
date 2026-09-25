# RES-007: VINCULACIÓN ZERO-FRICTION 1-CLIC (SHORT AUTHENTICATION STRING - RFC 6189 / BLUETOOTH NUMERIC COMPARISON)

## 1. Contexto y Desafío UX
¿Cómo permitir que dos dispositivos se vinculen de forma 100% segura en 1 solo clic o 5 segundos sin obligar al usuario a copiar y pegar llaves públicas hexadecimales de 64 caracteres ni depender de un servidor central de registro (como Tailscale / Auth0)?

## 2. Hallazgos en Estándares y Criptografía de Frontera
* **RFC 6189 (ZRTP SAS Model):** Define el uso de *Short Authentication String* (SAS). Un hash criptográfico truncado (normalmente 4 dígitos o 2 palabras cortas) derivado del intercambio de claves Diffie-Hellman / KEM efímero.
* **Bluetooth Core Spec (Numeric Comparison):** Ambos dispositivos muestran el mismo código numérico de 6 dígitos; si un atacante MITM intercepta la conexión, los códigos divergen con probabilidad $1 - 10^{-6}$.
* **Formalización de Vaudenay (SAS Protocols):** Garantiza seguridad demostrable contra ataques activos sin infraestructura de clave pública (PKI) centralizada.

## 3. Contrastación con Arquitectura IPVN7
* En `ipvn7`, el handshake híbrido post-cuántico (X-Wing KEM: X25519 + ML-KEM-768) genera un secreto compartido $SS$.
* **Mecanismo SAS en IPVN7 (Eje 5 UX Radical):**
  1. Descubrimiento local instantáneo por mDNS/SSDP o invitación numérica de 4 dígitos.
  2. Al recibir solicitud de vinculación, ambos nodos calculan $SAS = \text{Truncate}_{16}(\text{HMAC-SHA256}(SS, \text{PubA} \parallel \text{PubB})) \pmod{10000}$.
  3. La UI muestra una notificación limpia: *"Notebook Dvd solicita vincularse. Código: 4819. ¿Coincide?" [Vincular] [Rechazar]*.
  4. Al pulsar "Vincular", la clave se guarda automáticamente en `keystore/trusted_peers.json`.

## 4. Decisión Técnica Adoptada
* Adoptar el modelo SAS de RFC 6189 en la capa de emparejamiento de `ipvn7` para eliminar toda fricción y jerga técnica de claves públicas.
* Registrado como DEC-099 en el ADR.

## 5. Fuentes Primarias
* IETF RFC 6189: ZRTP: Media Path Key Agreement for Unicast Secure RTP (Section 5.5: Short Authentication String)
* Vaudenay, S. (2005): Secure Communications over Insecure Channels Based on Short Authenticated Strings.
