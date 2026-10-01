# Plan de Implementación: Sovereign Dynamic Egress Gateway (v0.7.0)

**Documento:** `docs/PLAN_SOVEREIGN_GATEWAY_EGRESS.md`  
**Fecha:** 2026-09-29  
**Estado:** LISTO PARA EJECUCIÓN AUTÓNOMA  
**Arquitectura:** Extensión Fractal L1-L4 sobre Núcleo Mínimo Pure Go (`CGO_ENABLED=0`)  
**Decisión Asociada:** ADR DEC-130  

---

## 1. Visión y Fundamentación Fractal

Este plan resuelve la transformación de cada nodo IPvN7 en una **puerta de enlace predeterminada soberana (Sovereign Default Gateway / Exit Node)** con selección automática del nodo óptimo de máxima velocidad.

El diseño sigue el **Principio Fractal de Escala (DEC-108)**:
* **Hoy (1 solo nodo / Localhost):** El nodo actúa como su propia pasarela local blindada (PQC, DNS seguro, métricas base). Si no hay pares con salida, navega transparente por su ISP nativo sin latencia extra.
* **Mañana (2 nodos / Laboratorio A $\leftrightarrow$ B):** Si el PC Principal (`192.168.1.198`) o el Notebook (`192.168.1.106`) sufren degradación o están en redes hostiles (4G/Wi-Fi público), el tráfico conmuta en caliente al otro nodo en <500 ms.
* **Escala Masiva (Billones de Nodos):** Enrutamiento por proximidad en los anillos Kleinberg 0–2 ($O(\log^3 N)$), seleccionando el gateway más rápido de la región sin servidores centrales ni saturar la memoria ($\le 1$ KB por nodo).

---

## 2. Invariantes Técnicos y Restricciones Obligatorias

1. **Axioma III ($\le 400$ Líneas):** Ningún archivo nuevo o modificado superará las 400 líneas. Poda preventiva obligatoria a las 320 líneas.
2. **Invariante Zero-Copy (0 B/op, 0 allocs/op):** El reenvío de datagramas en la canalización central L0-L2 no debe realizar copias en memoria en el hot-path.
3. **MTU Rígido 1280B y TCP MSS Clamping:** Ajustar MSS TCP a 1220 bytes en el gateway para prevenir fragmentación de paquetes en Internet público.
4. **Fallback Determinista Cero-Cortes (DEC-121):** Si el gateway activo se desconecta, el cliente conmuta de inmediato al enlace directo local. El usuario jamás queda sin internet.
5. **Seguridad ZTNA Default-Deny:** Solo DIDs autorizados (anillo de confianza o propios dispositivos) pueden usar un gateway para salir a Internet.

---

## 3. Especificación de Componentes a Desarrollar

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                      CLIENTE (Tu Computador)                           │
│  [Tráfico 0.0.0.0/0] ──► [Wintun L3 / SOCKS5 :10807]                    │
│                                │                                        │
│                 [EgressSelector (src/pkg/l1/)]                          │
│         Evalúa: RTT + Tasa BBR + Packet Loss + Histéresis 25%           │
│                                │                                        │
│     ┌──────────────────────────┴──────────────────────────┐             │
│     ▼ (Si hay Gateway Óptimo)                             ▼ (Fallback)  │
│ [Túnel Datagrama PQC 1280B]                        [Salida Directa ISP] │
└─────────────────┬───────────────────────────────────────────────────────┘
                  │ UDP 7777 (Noise XX + ML-KEM-768)
                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      GATEWAY SOBERANO REMOTO                            │
│  [Socket UDP 7777] ──► [Decapsulación Zero-Copy L0-L1]                  │
│                                │                                        │
│                  [EgressForwarder (src/pkg/l1/)]                        │
│          Verificación ZTNA por DID + Clamping MSS (1220B)               │
│                                │                                        │
│                                ▼                                        │
│                  [Dual-Stack Dialer / SNAT Kernel]                      │
│                                │                                        │
│                                ▼                                        │
│                  [Internet Público (IPv4 / IPv6)]                       │
└─────────────────────────────────────────────────────────────────────────┘
```

### 3.1 Estructuras de Datos Canónicas (`src/pkg/l1/egress_types.go`)
* `EgressCapability`: Anuncio en el latido P2P (`CanExit bool`, `BandwidthMbps uint32`, `ActiveSessions uint16`).
* `GatewayScore`: Puntuación compuesta $S = (\text{RTT} \times 0.4) + (\text{LossRate} \times 0.4) + (\text{Load} \times 0.2)$.
* `Histéresis`: Requisito de mejora $\ge 25\%$ sostenida por 3 segundos para evitar flapping entre pasarelas.

### 3.2 Selector Autónomo en Cliente (`src/pkg/l1/egress_selector.go`)
* Consulta la tabla de pares de Kleinberg (`routing.go`).
* Filtra pares con `CanExit == true` y ZTNA autorizado.
* Conmuta dinámicamente el `TargetExitDID` en `socks5_gateway.go` y `tun_router.go`.
* Si ningún par califica, conmuta a `ModeDirectLocal`.

### 3.3 Reenviador de Salida en Gateway (`src/pkg/l1/egress_forwarder.go`)
* Recibe el datagrama desempaquetado de L1.
* Verifica que el DID emisor tenga permiso de salida en la lista ZTNA.
* Abre la conexión hacia Internet mediante socket pool reutilizable o túnel Wintun.
* Devuelve los paquetes de respuesta encapsulados en datagramas CBOR deterministas hacia el DID origen.

### 3.4 Controles en WebUI (`src/pkg/core/web_ui.go`)
* **Interruptor de Compartir:** `[✓] Compartir mi conexión con mis dispositivos de confianza`.
* **Modo de Salida:** Selector de 3 posiciones:
  1. `⚡ Automática (La más rápida)`
  2. `🛡️ Nodo Específico (Seleccionar de la lista de pares)`
  3. `🏠 Solo Directo Local`
* **Telemetría en Vivo:** Indicador en tiempo real: *"Salida activa: Nodo B (Notebook Dvd) · 1.4 ms · 0% Drop"*.

---

## 4. Fases de Ejecución Paso a Paso

### Fase 1: Anuncio de Capacidades en Wire Protocol (L1)
* **Archivo:** `src/pkg/l1/egress_types.go` (<150 líneas).
* Extender el mensaje de presencia/keepalive CBOR con el campo `egress_flags`.
* Tests unitarios de codificación/decodificación determinista con 0 alocaciones.

### Fase 2: Motor de Selección y Conmutación en Cliente (L1)
* **Archivo:** `src/pkg/l1/egress_selector.go` (<220 líneas).
* Implementar cálculo de score con telemetría de `wan_active_prober.go`.
* Implementar lógica de histéresis anti-flapping y fallback automático.
* Tests unitarios: conmutación ante caída de nodo simulada en <500 ms.

### Fase 3: Forwarder y SNAT en el Gateway (L1/Core)
* **Archivo:** `src/pkg/l1/egress_forwarder.go` (<260 líneas).
* Integrar dialer TCP/UDP dual-stack con clamping MSS de 1220 bytes.
* Validación ZTNA: rechazo inmediato de peticiones de DIDs no autorizados.
* Tests unitarios: transporte bidireccional cliente $\leftrightarrow$ gateway $\leftrightarrow$ servidor web mock.

### Fase 4: Integración en SOCKS5 y Wintun Router
* Actualizar `src/pkg/l1/socks5_gateway.go` y `src/pkg/l1/tun_router.go`.
* Redirigir el tráfico `0.0.0.0/0` al gateway óptimo cuando esté activo.
* Mantener bypass estricto para redes locales y dominios de mensajería (DEC-121).

### Fase 5: Interfaz de Usuario y Experiencia Humana (UX Radical)
* Actualizar `src/pkg/core/web_ui.go` (respetando límite de 400 líneas).
* Añadir controles zen: 1 botón para activar salida compartida y selector de pasarela.
* Validar que ningún término críptico quede sin traducción comprensible (Regla 18).

### Fase 6: Verificación Física en Laboratorio de 2 Nodos (Reglas 1, 2, 5 y 16)
1. Compilar binarios de producción con `scripts/build_all_platforms.ps1`.
2. Desplegar al Notebook (`192.168.1.106`) mediante SFTP.
3. **Test Físico 1 (PC $\rightarrow$ Notebook $\rightarrow$ Internet):** El PC navega saliendo por la IP pública del Notebook.
4. **Test Físico 2 (Falsabilidad ante Caída):** Desconectar Wi-Fi del Notebook mientras se descarga un archivo en el PC; verificar que el PC conmuta a su ISP local en <1s sin congelar la descarga.
5. Ejecutar la compuerta universal `scripts/verify_ipvn7_standard.ps1` (100% PASS, Health Score 100%).

---

## 5. Matriz de Falsabilidad HIL (Casos de Falla Obligatorios)

| Escenario Hostil | Comportamiento Esperado | Falso Éxito Prohibido |
| :--- | :--- | :--- |
| **Gateway Único Apagado** | Fallback a ISP directo en <500 ms, aviso discreto en UI. | Dejar el sistema sin internet o colgar el navegador. |
| **Pérdida de Paquetes 25%** | Reducción de score, reintento BBR o cambio a nodo alterno. | Reportar velocidad máxima fingida en la UI. |
| **DID Desconocido solicita salida** | Paquete descartado inmediatamente en L1 (0 B/op). | Reenviar tráfico no autenticado al Internet público. |
| **Sobrecarga de Ancho de Banda** | Pacing dinámico y limitación suave por cuota token bucket. | Caída del nodo por OOM o bloqueo de sockets locales. |

---

## 6. Registro Arquitectónico Formal: ADR DEC-130

```text
ADR DEC-130: Sovereign Dynamic Egress Gateway y Salida Óptima Autoconmutable
- Fecha: 2026-09-29
- Contexto: Necesidad de permitir que cualquier nodo IPvN7 actúe como salida a Internet
  soberana, acelerando conexiones lentas mediante el mejor gateway de la malla.
- Decisión: Arquitectura híbrida PQC L1/L4 con selección por telemetría RTT/Drop,
  histéresis del 25%, MSS clamping a 1220B, ZTNA Default-Deny y fallback local cero-cortes.
- Anti-Patrón Prohibido: Depender de servidores proxy centrales, usar protocolos no
  encriptados, ignorar el MTU 1280B o dejar al usuario sin internet ante fallos de nodo.
```

---

## 7. Próxima Acción Inmediata

Cuando se inicie la siguiente sesión de trabajo:
1. Abrir este archivo (`docs/PLAN_SOVEREIGN_GATEWAY_EGRESS.md`).
2. Adquirir cerrojo `.agents/task.lock`.
3. Iniciar la codificación de la **Fase 1** (`src/pkg/l1/egress_types.go`).
