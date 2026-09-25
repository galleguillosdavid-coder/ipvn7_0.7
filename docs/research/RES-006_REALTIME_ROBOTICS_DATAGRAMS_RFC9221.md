# RES-006: TELEMETRÍA Y CONTROL DE ACTUADORES EN TIEMPO REAL (RFC 9221 VS IPVN7 DATAGRAM PIPELINE)

## 1. Contexto y Pregunta de Investigación
¿Cómo debe soportar `ipvn7` el control en tiempo real de actuadores, robots, drones, vehículos y telemetría de sensores sin incurrir en bloqueos por pérdida de paquetes (head-of-line blocking) ni añadir capas infladas de middleware (ROS2/DDS pesados)?

## 2. Hallazgos en Estándares Mundiales de Frontera
* **RFC 9221 (IETF Standard):** *"An Unreliable Datagram Extension to QUIC"* define transporte no confiable pero cifrado y con control de congestión dentro de túneles QUIC.
* **Problema de TCP y QUIC Streams:** Si se pierde un datagrama de coordenadas de un dron o joystick, la retransmisión bloquea los datos posteriores, generando picos de latencia fatales.
* **Solución Óptima RFC 9221:** La aplicación descarta datagramas obsoletos; lo que importa es el "estado más reciente" (fresh state over wire).

## 3. Contrastación con Arquitectura IPVN7
* `ipvn7` implementa en su capa L0-L2 un pipeline de datagramas cuantizados a 1280 bytes fijos (`FrameTypeDatagram = 0x01`).
* **Ventaja frente a RFC 9221 / QUIC:**
  1. **Zero-Copy Invariante:** `ipvn7` opera en 41 ns/op con 0 B/op de alocación de memoria, mientras los stacks QUIC de usuario alocan buffers de conexión y tablas de estado.
  2. **Cifrado PQC Híbrido en Banda:** Cifrado con ChaCha20-Poly1305 y llaves derivadas vía X-Wing KEM (`draft-connolly-cfrg-xwing-kem`), inmune a computación cuántica.
  3. **Multi-Hop Autónomo:** Enrutamiento por anillos de Kleinberg sin servidores centrales ni dependencias en brokers ROS2/MQTT.

## 4. Decisión Técnica Adoptada
* Adoptar la semántica de RFC 9221 para el subsistema de telemetría y control de máquinas en `ipvn7`: canal de datagramas sin confirmación (unreliable best-effort) para actuadores y telemetría de alta frecuencia (100-1000 Hz), manteniendo canales de control confiable para órdenes transaccionales (firmware, reinicios).
* Registrado como DEC-098 en el ADR.

## 5. Fuentes Primarias
* IETF RFC 9221: https://datatracker.ietf.org/doc/html/rfc9221
* IETF RFC 9000: QUIC: A UDP-Based Multiplexed and Secure Transport
* quic-go Datagram Implementation: https://quic-go.net/
