# RES-010: WAKE-ON-LAN P2P AUTENTICADO Y CONTROL PROFUNDO DE HARDWARE (LOOP CLOSURE)

## 1. Contexto y Problema Físico
¿Cómo encender, suspender o diagnosticar un equipo apagado o en reposo desde una red externa (4G/5G celular o CGNAT) sin abrir puertos en el router, sin servidores en la nube de terceros y sin exponer la red local a ataques de amplificación de broadcast?

## 2. Análisis del Estándar Wake-on-LAN (AMD/HP Magic Packet)
* **Estructura del Magic Packet:** Trama de 102 bytes: prefijo de sincronización `FF:FF:FF:FF:FF:FF` seguido de 16 repeticiones de la dirección MAC de 48 bits del NIC de destino.
* **Limitación FÍSICA de Capa 2:** Los routers y cortafuegos descartan invariablemente los paquetes de difusión dirigida a subred (*Subnet Directed Broadcast*) para prevenir ataques Smurf/DDoS. WoL no es enrutable a través de Internet de forma nativa.
* **Vulnerabilidad de WoL en Internet:** Carece de autenticación y cifrado. Cualquiera con acceso al puerto puede encender equipos arbitrariamente.

## 3. Arquitectura de Solución Soberana en IPVN7
* **Proxy WoL Descentralizado en Malla:**
  1. Si el PC destino está suspendido/apagado, el cliente emisor envía un comando `WoL-Trigger` cifrado extremo a extremo (X-Wing KEM) al nodo `ipvn7` centinela más cercano dentro de la misma LAN (ej. Notebook Dvd o mini-PC).
  2. El centinela valida la firma criptográfica Ed25519 del emisor contra `keystore/trusted_peers.json`.
  3. El centinela inyecta el Magic Packet en el socket físico `255.255.255.255:9`.
* **Cierre de Bucle Físico en el SO (Windows / Linux):**
  - Suspensión/Reposo: Llamada directa a `SetSuspendState` (Powrprof.dll en Windows) o `systemctl suspend` en Linux.
  - Bloqueo de estación de trabajo: `LockWorkStation` (user32.dll).
  - Telemetría de batería: Consulta WMI `Win32_Battery` o `/sys/class/power_supply/BAT*`.

## 4. Decisión Técnica Adoptada
* Formalizar el subsistema WoL-over-P2P autenticado y los puentes nativos de energía del SO sin agentes en la nube.
* Registrado como DEC-102 en el ADR.

## 5. Fuentes Primarias
* AMD Magic Packet Technology: White Paper (Advanced Micro Devices, 1995).
* Microsoft Windows API: Powrprof.h (`SetSuspendState`), Winuser.h (`LockWorkStation`).
