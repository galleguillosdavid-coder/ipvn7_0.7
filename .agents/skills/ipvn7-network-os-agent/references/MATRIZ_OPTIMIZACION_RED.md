# MATRIZ DE OPTIMIZACIÓN DE RED Y PIPELINE ZERO-COPY EN IPVN7

Para que ipvn7 sea el sistema operativo de redes más rápido, seguro y eficiente del mundo, toda mejora técnica debe evaluarse contra esta matriz de optimización.

---

## 1. Reglas de Optimización del Hot-Path de Red

1. **Cero Asignaciones en el Heap (0 B/op):**
   - En el procesamiento de paquetes (`L0`, `L1`, `L2`, `L3`), prohibido usar `new`, slices dinámicos con `append` no acotado o conversión de `[]byte` a `string`.
   - Utilizar buffers pre-asignados reutilizables mediante `sync.Pool` o anillos de memoria (Ring Buffers).
   - Estructuras alineadas a 64 bits para evitar penalizaciones de caché y desalineación de memoria.

2. **Complejidad Algorítmica $O(1)$:**
   - La búsqueda en tablas de enrutamiento y caché de peers debe ser en tiempo constante $O(1)$.
   - La verificación de la ventana anti-replay debe ser un desplazamiento de bits (bit-shift) y operaciones lógicas bit a bit.
   - El descarte de paquetes por exceso de cuota de memoria debe ocurrir antes de cualquier parseo criptográfico.

3. **Concurrencia sin Bloqueos (Lock-Free / Granular):**
   - Evitar mutexes globales (`sync.Mutex`) en el hot-path del pipeline.
   - Preferir primitivas atómicas (`sync/atomic`) y lectura/escritura concurrente con granularidad fina (`sync.RWMutex` particionado por hash).
   - Canales Go con buffers dimensionados adecuadamente para evitar contención de goroutines.

---

## 2. Métricas Clave de Rendimiento (KPIs)

| Indicador | Objetivo Estándar Mundial | Método de Medición |
|---|---|---|
| **Latencia de Tránsito (Pipeline)** | < 25 microsegundos por nodo | `go test -bench=BenchmarkPipeline` |
| **Throughput por Núcleo** | > 1.2 Millones pps (packets/sec) | Pruebas de estrés UDP sintéticas |
| **Consumo de Memoria Base** | < 30 MB en reposo | Monitorización pprof / Sys stats |
| **Asignaciones de Memoria** | 0 allocs/op en hot-path | `go test -benchmem` |
| **Tiempo de Handshake PQC** | < 15 ms (RTT local) | Benchmarking criptográfico NIST L3 |
| **Tiempo de Recuperación de Pérdida** | Inmediato sin colapso de buffer | Pruebas de conmutación de sockets |

---

## 3. Protocolo de Diagnóstico y Perfilado

Antes de aplicar cualquier optimización de red, el Agente debe obtener el perfil base:

```bash
# 1. Ejecutar benchmarks con asignación de memoria
go test -bench=. -benchmem ./pkg/core/...

# 2. Perfil de CPU
go test -bench=. -cpuprofile=cpu.pprof ./pkg/core/...
go tool pprof -top cpu.pprof

# 3. Perfil de Memoria (Heap Allocations)
go test -bench=. -memprofile=mem.pprof ./pkg/core/...
go tool pprof -top -alloc_space mem.pprof
```

Toda optimización debe demostrar una reducción verificable en tiempo de CPU o en bytes asignados, sin alterar la corrección funcional ni violar el límite de 400 líneas por archivo.
