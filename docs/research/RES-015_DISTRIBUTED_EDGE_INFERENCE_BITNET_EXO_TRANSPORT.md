# RES-015: INFERENCIA DISTRIBUIDA EN EL BORDE, MODELOS TERNARIOS BITNET Y TRANSPORTE P2P ZERO-COPY EN IPVN7

## 1. Contexto y Pregunta de Investigación
Hacia finales de 2026, la computación de inteligencia artificial ha experimentado un giro tectónico hacia el borde (*Edge AI*) impulsado por dos avances convergentes:
1. **La Revolución Ternaria de 1-Bit (BitNet b1.58):** Los modelos entrenados con pesos ternarios $\{-1, 0, +1\}$ reemplazan las multiplicaciones en coma flotante por simples adiciones enteras, logrando que modelos de 2B parámetros corran en CPUs convencionales con una huella de tan solo **~0.4 GB de RAM** y latencias locales de ~20 ms por token (frente a 300–500 ms en la nube).
2. **Mallas de Inferencia Colaborativa P2P (Exo Labs, Petals, OpenTela):** Frameworks que enlazan dispositivos heterogéneos de consumo (computadores de escritorio, notebooks, teléfonos) para particionar capas de modelos grandes y ejecutar inferencia distribuida sin depender de servidores centrales.

Sin embargo, el cuello de botella físico de la inferencia distribuida radica en la **latencia de transporte y la fluctuación de paquetes (jitter) al transferir tensores y estados KV entre nodos**. ¿Cómo puede IPvN7 posicionarse como el sustrato de red soberano, zero-copy y P2P idóneo para conectar enjambres de inferencia distribuida?

## 2. Hallazgos en Estándares Mundiales de Frontera (Septiembre 2026)
* **Arquitectura BitNet b1.58:** Estandarizada mediante frameworks nativos en C/C++ (`bitnet.cpp`) con soporte para CPUs ARM/x86 y NPUs. Elimina la necesidad de GPUs de alto consumo para tareas de agente local.
* **Enrutamiento Consciente de Caché de Prefijos (Prefix-Cache-Aware Routing):** Algoritmos emergentes en redes P2P que dirigen las peticiones de inferencia al nodo que ya posee el contexto en memoria (KV-cache caliente), minimizando la re-computación.
* **El Reto de la Latencia Inter-Capa:** En modelos particionados entre dispositivos locales, cada salto de red añade retraso al pipeline secuencial. Si los nodos se comunican mediante HTTP estándar o túneles de nube centralizados, el tiempo por token se dispara de forma inaceptable.

## 3. Oportunidades y Beneficios Estratégicos para IPvN7
1. **Canalización P2P Zero-Copy para Activaciones Neuronales:**
   - La canalización central L0-L2 de IPvN7 certificada en **0 B/op** y latencia de ~30 ns ofrece el medio de transporte más rápido posible sobre UDP para el intercambio de activaciones y tensores entre nodos colaboradores (ej. PC Principal y Notebook), atravesando cortafuegos y CGNAT sin servidores intermediarios.
2. **Enrutamiento Kleinberg con Etiquetas de Capacidad de Cómputo (Inference-Aware Routing):**
   - El enrutador Kleinberg de L1 puede incorporar métricas de capacidad de nodo (ej. nodo con GPU o CPU libre ejecutando BitNet/Ollama). Las solicitudes de inferencia A2A se despachan con algoritmo *greedy* hacia el nodo óptimo más cercano en latencia y recursos.
3. **Escudo de Privacidad Soberana Absoluta (Zero Data Leakage):**
   - Los datos sensibles del usuario y las respuestas del modelo nunca salen de la malla privada hacia centros de datos de terceros. Todo el flujo de inferencia viaja blindado con criptografía post-cuántica **ML-KEM-768** y autenticación ZTNA default-deny.

## 4. Filtro Antihumo Estricto y Presupuesto de Código
* **Aislamiento Total del Núcleo:** IPvN7 **no compilará modelos neuronales ni librerías de tensores** dentro de `ipvn7.exe`. La ejecución de modelos corre en motores externos desacoplados (`bitnet.cpp`, `ollama`, `exo`).
* **Rol de IPvN7:** Proveer la autopista de red física transparente, de ultra-baja latencia y autenticación cuántica que une a estos motores en una malla privada unificada mediante puertos autorizados (Shadow DIDs ZTNA).

## 5. Decisión Técnica Recomendada
* Registrar **DEC-124**: Adopción de arquitectura de transporte P2P para enjambres de inferencia distribuida (BitNet/Exo) con enrutamiento Kleinberg consciente de capacidades en IPvN7.

## 6. Fuentes Primarias
* Microsoft Research: BitNet b1.58 & `bitnet.cpp` Inference Framework (2025–2026).
* Exo Labs: Decentralized Local Mesh AI Architecture (2026).
* Petals: Collaborative Inference of Large Language Models (2026).
* IEEE P3377: Standard for Autonomous AI Agent Communications.
