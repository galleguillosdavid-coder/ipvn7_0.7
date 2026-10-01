# PLAN DE MEJORAS PRIORITARIAS - IPVN7 v0.7.0

**Fecha:** 2026-09-30 | **Versión:** 0.7.0 | **Estado:** 10/10 Completadas (100%)

---

## 📊 RESUMEN EJECUTIVO

| # | Mejora | Componente | Prioridad | Estado |
|:---|---|---|:---:|:---:|
| 1 | Configurar Sleep en TUN Router | `pkg/l1/tun_router.go` | 🔥 CRÍTICA | ✅ Completado |
| 2 | Exportar OpenMetrics L1 | Múltiples L1 | 🔥 CRÍTICA | ✅ Completado |
| 3 | Cache de Políticas Firewall | `pkg/l1/firewall.go` | 🔥 ALTA | ✅ Completado |
| 4 | Métricas de Anti-Replay | `pkg/l0/anti_replay.go` | 🔥 ALTA | ✅ Completado |
| 5 | Health Check Endpoint `/health` | `pkg/core/web_ui.go` | 🔥 ALTA | ✅ Completado |
| 6 | Configuración Timeouts LinkHealing | `pkg/l1/link_healing.go` | 🔥 ALTA | ✅ Completado |
| 7 | Graceful Shutdown Coordinado | `pkg/core/shutdown.go` | 🔥 MEDIA | ✅ Completado |
| 8 | Bulk Operations en Firewall | `pkg/l1/firewall.go` | 🔥 MEDIA | ✅ Completado |
| 9 | Prioridad QoS en Gateway | `pkg/core/gateway.go` | 🔥 ALTA | ✅ Completado |
| 10 | Validación de MTU en Wire Format | `pkg/l0/wire.go` | 🔥 MEDIA | ✅ Completado |

---

## 1. Configurar Sleep TUN Router
- **Archivo:** `src/pkg/l1/tun_router.go`
- **Solución:** Variable `tunRouterSleepDuration` configurable vía `ConfigureTunRouterSleep()` y variable de entorno `IPVN7_TUN_ROUTER_SLEEP_MS` (0ms prod, 10ms pruebas).

## 2. Exportar OpenMetrics L1
- **Archivos:** `firewall.go`, `memory_arbiter.go`, `qos.go`, `nat_traversal.go`, `multipath.go`.
- **Solución:** Implementación de `ToOpenMetrics() string` compatible con Prometheus en componentes L1.

## 3. Cache de Políticas Firewall
- **Archivo:** `src/pkg/l1/firewall.go`
- **Solución:** Cache `policyCache` con TTL de 30s e invalidación automática al modificar políticas, optimizando evaluación de `EvaluateInbound/Outbound`.

## 4. Métricas Anti-Replay
- **Archivo:** `src/pkg/l0/anti_replay.go`
- **Solución:** Contadores atómicos `accepted` y `rejected` en `AntiReplayFilter` con método `Stats()`.

## 5. Health Check Endpoint
- **Archivo:** `src/pkg/core/web_ui.go`
- **Solución:** Endpoint HTTP GET `/health` reportando estado del nodo, uptime, pares y versión JSON.

## 6. Configuración de Timeouts LinkHealing
- **Archivo:** `src/pkg/l1/link_healing.go`
- **Solución:** Variables configurables `MaxConsecutiveLosses`, `DegradedLatencyMs`, `LinkHealingCooldown` y función `ConfigureLinkHealing()`.

## 7. Graceful Shutdown Coordinado
- **Archivo:** `src/pkg/core/shutdown.go`, `web_ui.go`
- **Solución:** Función `GracefulShutdown(fn, timeout)` con `context.WithTimeout` para terminación determinista.

## 8. Bulk Operations en Firewall
- **Archivo:** `src/pkg/l1/firewall.go`
- **Solución:** Métodos `AuthorizeDIDsBatch([]*DIDPolicy)` y `RevokeDIDsBatch([]string)` bajo un único cerrojo.

## 9. Prioridad de Tráfico QoS en Gateway
- **Archivo:** `src/pkg/core/gateway.go`
- **Solución:** `SendDatagram(env, tc)` integrando `TrafficClass` y evaluación con `QoSManager`.

## 10. Validación de MTU en Wire Format
- **Archivo:** `src/pkg/l0/wire.go`
- **Solución:** Límite estricto de MTU canónico de 1280 bytes en serialización de datagramas IPv6/IPvN7.
