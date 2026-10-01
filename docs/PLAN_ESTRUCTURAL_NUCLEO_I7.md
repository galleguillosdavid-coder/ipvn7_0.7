# Plan Estructural Maestro: Retorno al Núcleo Mínimo I7 y Purga Factual

> **"Unificar la red alrededor de un Núcleo Mínimo de 10 primitivas universales, desacoplado de transportes físicos y satélites periféricos, bajo una rigurosa taxonomía de 4 estados."**
> Formulado a partir de las recomendaciones estructurales de [`auditoria externa.md`](../auditoria%20externa.md).

---

## Estado Global de Ejecución

- [x] **Fase 1: Formalización Canónica del Núcleo Mínimo I7 (Las 10 Primitivas)**
  - [x] **Check 1.1:** Definir las 10 primitivas nucleares puras en `src/pkg/core/i7_primitives.go` (Identity, Object, Container, Session, Channel, Path, MTU, Integrity, Routing, Capability) sin dependencias externas.
  - [x] **Check 1.2:** Crear suite unitaria de validación contractual `src/pkg/core/i7_primitives_test.go` verificando la instanciación e invariante de 1280B de cada primitiva.

- [x] **Fase 2: Conexión Desacoplada con Transportes Físicos (UDP Transport Adapter)**
  - [x] **Check 2.1:** Implementar `I7UDPAdapter` en `src/pkg/core/i7_adapter.go` vinculando el pipeline del Núcleo Mínimo con sockets físicos de red UDP (emisión y recepción real).
  - [x] **Check 2.2:** Validar en `src/pkg/core/i7_adapter_test.go` la transmisión bidireccional socket a socket de un I7 Container con aserción estricta de entrega y verificación de integridad.

- [x] **Fase 3: Saneamiento Factual de la Documentación Central**
  - [x] **Check 3.1:** Actualizar `docs/ARQUITECTURA.md` estableciendo formalmente la separación tripartita: Aplicaciones / Perfiles $\rightarrow$ Núcleo Mínimo I7 $\rightarrow$ Adaptadores de Transporte.
  - [x] **Check 3.2:** Erradicar de `README.md` y documentos principales las afirmaciones de "production-ready" o "100% certificado", aplicando la Taxonomía Factual de 4 Estados.

- [x] **Fase 4: Benchmark Factual con Entorno Explícito y Compuerta Universal**
  - [x] **Check 4.1:** Ejecutar microbenchmark registrando métricas con especificación de entorno (CPU, SO, versión Go, commit, GOMAXPROCS): 37.57 ns/op, 0 B/op, 0 allocs/op en Intel i5-1030NG7.
  - [x] **Check 4.2:** Ejecutar la compuerta de paso universal `scripts/verify_ipvn7_standard.ps1` y confirmar 100% PASS, Invariante Zero-Copy y compilación limpia.
