# MANUAL DE OPERACIÓN: CLI Y MODO VPN SOBERANA — IPVN7 v0.7.0

> **Estado:** VIGENTE / GUÍA DE USUARIO  
> **Componentes Cubiertos:** Herramienta de consola `ipvn7-cli` y Modo VPN Soberana 1-Clic  

---

## 1. CONSOLA DE ADMINISTRACIÓN `ipvn7-cli`

La herramienta CLI permite inspeccionar la telemetría, el enrutador Kleinberg y la identidad local sin recurrir al navegador:

```bash
# Estado general del nodo, identidad DID y prefijos soberanos
./bin/ipvn7-cli status

# Inspección de los 16 anillos de Kleinberg y pares conocidos
./bin/ipvn7-cli peers

# Estado de la Máquina de Estados Finitos (FSM) de salud de enlaces
./bin/ipvn7-cli fsm

# Lista de componentes inteligentes satelitales conectados al Gateway
./bin/ipvn7-cli components

# Diagnóstico de rutas y latencia hacia un DID específico
./bin/ipvn7-cli ping did:ipvn7:<peer-public-key>
```

---

## 2. MODO VPN SOBERANA (HERO ZEN 1-CLIC)

El modo VPN Soberana permite al usuario anonimizar su conexión o navegar a través de la malla en un solo clic:

### 2.1 Conexión Rápida
1. Iniciar el nodo principal (`ipvn7.exe`).
2. Abrir el panel de control local en el navegador: `http://127.0.0.1:7070`.
3. Pulsar el botón **"Conectar VPN Soberana"**.
4. El sistema selecciona automáticamente el nodo de salida (Egress Node) de menor latencia y mayor ancho de banda (DEC-130).

### 2.2 Operación Vía API Local
Las aplicaciones locales autorizadas pueden gobernar la VPN mediante llamadas HTTP REST autenticadas:

* `POST http://127.0.0.1:7070/api/v1/vpn/connect`: Conecta el túnel hacia el mejor par disponible.
* `POST http://127.0.0.1:7070/api/v1/vpn/disconnect`: Restaura la conexión local directa en menos de 1 segundo.
* `GET http://127.0.0.1:7070/api/v1/vpn/status`: Devuelve estado (`connected`, `connecting`, `disconnected`), bytes transmitidos y par de salida actual.

### 2.3 Mecanismos de Seguridad y Resiliencia
* **Protección Anti-Cortes (Kill-Switch):** Si el demonio `ipvn7` se detiene inesperadamente, un centinela restaura la configuración de red original de inmediato para evitar pérdida de conectividad a Internet.
* **MSS Clamping a 1220B:** Garantiza que los paquetes TCP no sufran fragmentación en enlaces WAN ni causen bloqueos de carga (black holes).
* **Conmutación Failover O(1):** Si el nodo de salida experimenta pérdidas de paquetes superiores al 5% o desconexión, conmuta automáticamente al siguiente mejor par sin romper las conexiones activas.
