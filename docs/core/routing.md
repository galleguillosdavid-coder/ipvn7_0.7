# CONTRATO DEL CORE: ROUTING & PATH

## 1. ¿Qué representa?
El motor y vector de decisión topológica para el reenvío de datagramas entre nodos en la malla IPVN7, utilizando la métrica de distancia XOR de Kleinberg sobre identificadores de 256 bits y anillos concéntricos con K-buckets.

## 2. ¿Qué datos contiene?
* `RoutingTable`: Tabla de pares organizados en anillos concéntricos / K-buckets.
* `PeerEntry`: DID del par, dirección IP:puerto física, distancia métrica XOR, latencia EWMA estimada y último contacto.
* `NextHop`: Función determinista que dado un `DestDID` retorna el par más cercano en espacio métrico.

## 3. ¿Quién lo crea?
El enrutador Kleinberg (`src/pkg/l1/routing.go`) al iniciar el nodo de red.

## 4. ¿Quién lo modifica?
El motor de descubrimiento y sondeo pasivo (`autonomous_discovery.go`, `wan_active_prober.go`) al actualizar latencias o registrar nuevos pares autorizados.

## 5. ¿Quién lo destruye?
El recolector de vecinos inactivos que purga entradas con fallas consecutivas de heartbeat.

## 6. ¿Qué invariantes tiene?
* **El enrutamiento NO decide identidad ni autorización:** Solo responde a la pregunta *¿A dónde va el paquete y cuál es el siguiente salto topológico?*.
* Búsqueda determinista $O(\log N)$ saltos teóricos.
* Prohibido el auto-emparejamiento (un nodo nunca se incluye a sí mismo como par externo).

## 7. ¿Qué errores puede producir?
* `ErrNoRouteToHost`: No existen pares disponibles en la tabla que acerquen el paquete al destino.
* `ErrRoutingLoopDetected`: Detección de bucle por conteo de saltos (Hop Limit).

## 8. ¿Cómo se prueba?
* `src/pkg/l1/l1_test.go`: Inserción de pares, cálculo de distancia XOR, selección de next-hop y saturación de anillos.
