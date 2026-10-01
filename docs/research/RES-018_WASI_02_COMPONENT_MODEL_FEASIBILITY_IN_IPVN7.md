# RES-018: FACTIBILIDAD Y ANÁLISIS DE IMPLEMENTACIÓN DEL ESTÁNDAR WASI 0.2 EN IPVN7

## 1. Pregunta de Investigación y Contexto
¿Es viable y conveniente implementar el estándar **WASI 0.2 (WebAssembly System Interface Preview 2 / Component Model)** en el software soberano de IPvN7 para la ejecución de plugins y lógica agéntica en el borde, sin violar las directivas de cero dependencias CGo, compacidad ($\le 400$ líneas) e invariante zero-copy?

---

## 2. Radiografía del Estándar WASI 0.2 (Septiembre 2026)
WASI 0.2 representa un cambio de paradigma impulsado por la Bytecode Alliance:
* **Abandono del modelo POSIX crudo (`wasi_snapshot_preview1`):** WASI 0.1 operaba como una traducción casi directa de syscalls de Unix (open, read, write).
* **Adopción del WebAssembly Component Model:** Los programas ya no son meros ejecutables monolíticos, sino "Componentes" que exponen y consumen interfaces de alto nivel tipadas con **WIT (WebAssembly Interface Types)** como `wasi:cli`, `wasi:http` y `wasi:filesystem`.
* **Aislamiento Granular ("Shared-Nothing"):** Cada componente posee su propio espacio de memoria lineal independiente. La comunicación entre componentes se realiza mediante el paso de mensajes estructurados (canonical ABI), eliminando el riesgo de que un plugin corrompa la memoria del host.

---

## 3. Evaluación Comparativa de Runtimes para Go en IPvN7

| Criterio | Opción A: Wasmtime-Go (CGo / Rust) | Opción B: Tetratelabs Wazero (Pure Go) | Opción C: Motor WASM Nativo IPvN7 (`pkg/wasm`) |
| :--- | :---: | :---: | :---: |
| **Soporte Nativo WASI 0.2** | Nativo 100% | Vía Adaptador Reactor Bytecode Alliance | Primitivas Criptográficas DIDs |
| **Dependencia de CGo** | **SÍ (Requiere CGo y `.dll`/`.so`)** | **NO (100% Go puro)** | **NO (100% Go puro)** |
| **Compilación Cruzada** | Muy compleja y frágil | Trivial (`CGO_ENABLED=0`) | Trivial (`CGO_ENABLED=0`) |
| **Peso Agregado al Binario** | +35 MB a +60 MB | +4 MB a +6 MB | +0 KB (ya integrado, 126 líneas) |
| **Tiempo de Arranque** | 10–25 ms | <1 ms (JIT nativo en memoria) | <0.01 ms |
| **Veredicto IPvN7** | ❌ **PROHIBIDO TERMINANTEMENTE** | ✅ **VIABLE Y RECOMENDADO** | ✅ **NÚCLEO BASE OPERATIVO** |

---

## 4. Dictamen Técnico y Arquitectura Recomendada para IPvN7

### 4.1 La Regla de Oro: Prohibición Absoluta de CGo
La adopción de `wasmtime` o `wasmer` en IPvN7 introduciría bibliotecas dinámicas compartidas de C/C++, destruyendo la portabilidad de `ipvn7.exe` (que hoy corre en cualquier máquina Windows, Linux o Docker scratch sin prerequisitos). Por tanto, **cualquier integración de WASI en IPvN7 debe ser incondicionalmente en Go puro (`CGO_ENABLED=0`)**.

### 4.2 El Camino Pragmático: Arquitectura de 2 Fases
1. **Fase 1 (Inmediata / Estado Actual):**
   - El motor [`src/pkg/wasm`](../../src/pkg/wasm) provee las primitivas de identidad soberana, firma Ed25519 y verificación para entornos WebAssembly/JS con 0 sobrecarga.
2. **Fase 2 (Habilitación de Plugins WASI 0.2 en el Borde):**
   - Cuando un agente requiera ejecutar código de filtrado de paquetes o cálculo dinámico, se usará el motor puro Go **Wazero** junto con el adaptador oficial `wasi_snapshot_preview1.reactor.wasm` de la Bytecode Alliance. Esto permite ejecutar componentes WASI 0.2 sin contaminar el núcleo con CGo.

---

## 5. Beneficios Estratégicos para IPvN7
* **Sandboxing Total para Agentes:** Si un sub-agente genera código malicioso o defectuoso, este queda confinado en el sandbox de memoria de Wasm sin poder acceder a sockets locales no autorizados ni alterar el stack del sistema operativo.
* **Cold Boot en Microsegundos:** Los plugins agénticos se instancian en $<1$ ms bajo demanda y se destruyen al finalizar la tarea, manteniendo el consumo en reposo en 0 MB.

---

## 6. Decisión Técnica Recomendada
* Registrar **DEC-127**: Aprobación de la arquitectura de sandboxing agéntico basada en WebAssembly Pure-Go (Wazero/Reactor) sin dependencias CGo, descartando Wasmtime por violación del principio de binario autocontenido.

---

## 7. Fuentes Primarias
* Bytecode Alliance: WebAssembly Component Model & WASI 0.2 Specification (2025–2026).
* Tetratelabs Wazero: Zero-Dependency WebAssembly Runtime for Go.
* Wasmtime: Component Model Reference Implementation and Limitations in Static Go Builds.
* W3C WebAssembly Community Group: Interface Types & Canonical ABI.
