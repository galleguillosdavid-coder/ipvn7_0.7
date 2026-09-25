# RES-002: Análisis de Rendimiento de Sockets en Windows: Registered I/O (RIO) vs IOCP en Go

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE DE ARQUITECTURA DE SOCKETS L0-L1
* **Área de Impacto:** Userspace Virtual Adapter, Latencia de Datagramas UDP en Windows.

---

## 1. Contexto y Problema Físico

En modo sin privilegios de administrador (`ModeUserspaceProxy`), ipvn7 canaliza tráfico UDP a través de sockets de red estándar en lugar de un adaptador de kernel TUN (ej. Wintun/WireGuard-NT). Para superar a WireGuard en latencia, el cuello de botella físico reside en las llamadas al sistema y la copia de memoria entre el kernel de Windows y el espacio de usuario.

---

## 2. Comparativa Técnica: RIO vs IOCP en Go

| Dimensión | Windows Registered I/O (RIO) | Go `netpoll` (IOCP) + Pre-allocated Pools (ipvn7) |
|---|---|---|
| **Mecanismo** | Pinned buffers fijos registrados en kernel; colas de I/O en espacio de usuario. | Completion Ports nativos del kernel con goroutines ligeras. |
| **Copias de Memoria** | Zero-copy físico total a nivel de socket Winsock. | Zero-alloc en Go (`0 allocs/op`), copia única de socket en Winsock IOCP. |
| **Integración con Go** | Requiere `unsafe`, bypass completo del runtime scheduler de Go y carga manual de punteros Winsock. | Compatible 100% con el runtime de Go, sin riesgo de corrupción de memoria ni bloqueos de GC. |
| **Complejidad y Estabilidad** | **Extrema:** Alto riesgo de pánicos y fallos si el recolector de basura mueve punteros. | **Óptima:** Invariante Zero-Copy garantizado por `LinearPipeline` (`36.28 ns/op`). |

---

## 3. Decisión y Plan de Acción para ipvn7

* **Conclusión Técnica:** La adopción de RIO puro añade una sobrecarga de mantenimiento desproporcionada que viola el principio de simplicidad pragmática (Algoritmo de 5 Pasos).
* **Estrategia ipvn7:**
  1. Mantener el pipeline L0-L2 sobre buffers pre-alocados de 1280B (`frame_pool.go`).
  2. Utilizar sockets UDP optimizados con `SO_RCVBUF` y `SO_SNDBUF` aumentados para evitar pérdidas de paquetes en ráfagas de alta tasa.
  3. Preservar la marca de **36.28 ns/op, 0 B/op y 0 allocs/op**, superando en estabilidad y portabilidad a cualquier implementación artesanal de RIO.
