# RES-003: Perforación de NAT Simétrico, Paradoja del Cumpleaños (RFC 5128) y Relay Kleinberg

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE DE ARQUITECTURA DE CONECTIVIDAD P2P
* **Área de Impacto:** Capa L1 (STUN RFC 5389), Capa L2 (Kleinberg Sovereign Relay / MASQUE Fallback).

---

## 1. El Reto del CGNAT Simétrico (Endpoint-Dependent Mapping)

En redes móviles celulares (4G/5G) y conexiones corporativas, los proveedores asignan NATs simétricos donde cada nuevo destino recibe un puerto externo aleatorio. Esto inutiliza el mapeo estático de STUN tradicional.

---

## 2. Evaluación de Técnicas del Estado del Arte

| Técnica | Mecanismo | Viabilidad en ipvn7 | Decisión |
|---|---|---|---|
| **Birthday Paradox Port Prediction (RFC 5128)** | Envía ráfagas masivas de paquetes a puertos consecutivos esperando una colisión estadística. | **Pésima:** Satura la red móvil, drena la batería y es bloqueado como escaneo de puertos por DPI. | ❌ **DESCARTADO** (Anti-patrón frágil). |
| **Relays Centrales Propietarios (Tailscale DERP / TURN tradicional)** | Todo el tráfico pasa por servidores centrales operados por una sola empresa. | **Inadmisible:** Viola el mandato soberano de cero servidores centrales y crea puntos de censura. | ❌ **RECHAZADO** (Dependencia central). |
| **Kleinberg Sovereign Relay (ipvn7)** | Retransmisión cifrada de tramas 1280B a través del nodo con menor latencia en los 12 anillos de Kleinberg. | **Óptima:** Cero servidores centrales, preserva E2EE total y no requiere privilegios especiales. | ✅ **ADOPTADO COMO BASELINE**. |
| **MASQUE CONNECT-UDP (RFC 9298)** | Encapsulamiento de tramas UDP sobre TLS 1.3 en puerto 443 cuando UDP directo está bloqueado. | **Excelente:** Supera bloqueos extremos sin permisos de administrador. | ✅ **ADOPTADO COMO FALLBACK HOSTIL**. |

---

## 3. Algoritmo Determinista de Conexión en ipvn7 (Cascada 4-Niveles)

```
[ Iniciar Conexión P2P ]
           │
           ▼
 [ 1. Intento Directo STUN RFC 5389 ] ──(Éxito)──► Conexión Directa (<2ms)
           │ (Falla por CGNAT Simétrico Doble)
           ▼
 [ 2. Detección Asimétrica ] ────────────(Éxito)──► Perforación Unilateral
           │ (Ambos extremos en Hard-NAT)
           ▼
 [ 3. Kleinberg Sovereign Relay ] ───────(Éxito)──► Retransmisión E2EE en Malla
           │ (UDP Bloqueado por DPI)
           ▼
 [ 4. MASQUE RFC 9298 sobre TLS 1.3:443 ] ───────► Túnel Invisible Zero-Admin
```
