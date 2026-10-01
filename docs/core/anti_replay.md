# CONTRATO DEL CORE: ANTI-REPLAY

## 1. ¿Qué representa?
El mecanismo de protección contra ataques de repetición maliciosa de datagramas, asegurando que ningún paquete capturado en el alambre pueda ser reinyectado con éxito.

## 2. ¿Qué datos contiene?
* `WindowSize = 1024`: Tamaño de la ventana deslizante de bits para paquetes fuera de orden.
* `Bitmap`: Vector de 16 enteros uint64 (1024 bits) que representan los números de secuencia aceptados.
* `MaxSeq`: Mayor número de secuencia visto hasta el momento en la sesión.
* `LastSeenTs`: Timestamp del último paquete aceptado.
* `SessionMap`: Tabla de ventanas indexadas por clave compuesta `originDID:sessionID`.

## 3. ¿Quién lo crea?
`src/pkg/l1/anti_replay.go` (`NewAntiReplayFilter`).

## 4. ¿Quién lo modifica?
El pipeline de recepción L1 al evaluar cada paquete con `Accept(originDID, sessionID, sequence, timestamp)`.

## 5. ¿Quién lo destruye?
El recolector de sesiones al cerrarse la sesión asociada o tras expirar por inactividad prolongada.

## 6. ¿Qué invariantes tiene?
* **Aislamiento Cross-Session y Cross-Peer:** Un atacante que repita secuencias de otra sesión o de otro par nunca afecta la ventana de la sesión legítima.
* **Evaluación O(1):** La comprobación y actualización del bitmap de 1024 bits se ejecuta en tiempo constante sin alocaciones en el heap.
* **El Anti-Replay NO evalúa permisos ZTNA:** Solo dictamina frescura y unicidad de secuencia.

## 7. ¿Qué errores puede producir?
* `ErrReplayDetected`: Número de secuencia ya registrado en el bitmap.
* `ErrSequenceTooOld`: Número de secuencia inferior a `MaxSeq - WindowSize`.
* `ErrStaleTimestamp`: Paquete con timestamp desfasado respecto al reloj local.

## 8. ¿Cómo se prueba?
* `src/pkg/l1/session_adversarial_test.go`:
  - Test C (Aislamiento cross-session con números de secuencia idénticos).
  - Test D (Aislamiento cross-peer).
  - Test E (Rechazo inmediato de replay de datos).
  - Test F (Rechazo de replay de HandshakeInit).
