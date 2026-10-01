# CONTRATO DEL CORE: CHANNEL

## 1. ¿Qué representa?
Un conducto lógico multiplexado dentro de una sesión cifrada activa, permitiendo múltiples flujos concurrentes independientes (ej. túnel VPN, control, telemetría, archivos) sin colisión de orden.

## 2. ¿Qué datos contiene?
* `ChannelID`: Identificador de 16 bits del sub-canal (`SubPort`).
* `SessionID`: Identificador de la sesión padre a la que pertenece el canal.
* `Priority`: Nivel de prioridad de despacho (control, tiempo real, best-effort).
* `State`: Estado del canal (`Open`, `Draining`, `Closed`).

## 3. ¿Quién lo crea?
La capa de multiplexación L1 al instanciar un servicio sobre una sesión establecida.

## 4. ¿Quién lo modifica?
El planificador de datagramas al encolar o despachar paquetes en ese canal.

## 5. ¿Quién lo destruye?
El servicio que lo abrió al terminar su flujo o el gestor de sesiones al cerrarse la sesión padre.

## 6. ¿Qué invariantes tiene?
* Un canal solo transmite datos si su sesión padre está autenticada y en estado activo.
* Canales independientes no bloquean mutuamente el despacho en caso de pérdida de paquetes fuera de orden (transporte orientado a datagramas).

## 7. ¿Qué errores puede producir?
* `ErrChannelClosed`: Intento de emitir en un canal cerrado.
* `ErrUnknownChannel`: Recepción de datos para un `ChannelID` no registrado.

## 8. ¿Cómo se prueba?
* `src/pkg/l1/l1_test.go`: Multiplexación de subpuertos y canales lógicos.
