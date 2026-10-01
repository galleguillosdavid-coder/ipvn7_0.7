# RES-011: PROCESAMIENTO DE DATAGRAMAS ULTRA-ALTA VELOCIDAD (DISRUPTOR RING BUFFER ZERO-COPY)

## 1. Contexto y Desafío de Rendimiento
¿Cómo garantizar que `ipvn7` procese ráfagas masivas de datagramas UDP (10.000–50.000 PPS) en userspace sin incurrir en contienda de bloqueos (mutex locks), sin generar presión en el recolector de basura (GC pauses) y sin romper la portabilidad multiplataforma (Windows/Linux/macOS)?

## 2. Hallazgos en Patrones de Ultra-Baja Latencia
* **Patrón LMAX Disruptor:** Estructura de Ring Buffer circular lock-free basada en un array de tamaño potencia de 2 ($2^k$).
* **Indexación Bitwise O(1):** La posición del buffer se calcula mediante máscara `head & (Capacidad - 1)`, evitando la costosa operación de división/módulo (`%`).
* **Mechanical Sympathy y Single-Producer:** El hilo receptor escribe secuencialmente en las ranuras prealocadas del ring sin competir por bloqueos, eliminando la invalidación de líneas de caché de CPU (False Sharing).
* **Zero GC Overhead:** Los buffers de 1280B se alocan una sola vez en el arranque (`init`); el pipeline nunca invoca `make([]byte)` en la ruta crítica.

## 3. Implementación y Síntesis en IPVN7 L0-L2
* `ipvn7` adopta este patrón en `pkg/l0/linear_pipeline.go` y los pools de datagramas de 1280B.
* Los benchmarks de producción (`BenchmarkLinearPipeline_Execute`) certifican:
  - Latencia media: **33–39 ns/op**.
  - Asignaciones de memoria: **0 B/op y 0 allocs/op**.
  - Capacidad de procesamiento: $>25.000.000$ operaciones/segundo por núcleo.

## 4. Decisión Técnica Adoptada
* Formalizar el patrón Disruptor Lock-Free Ring Buffer como estándar arquitectónico obligatorio en la capa de transporte L0-L2 de `ipvn7`.
* Registrado como DEC-103 en el ADR.

## 5. Fuentes Primarias
* Thompson, M., Barker, D., et al. (2011): "Disruptor: High performance alternative to bounded queues for exchanging data between threads". LMAX Whitepaper.
* Drepper, U. (2007): "What Every Programmer Should Know About Memory". Red Hat.
