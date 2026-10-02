# REGISTRO DE CAMBIOS (sistema/CAMBIOS.md)

| Fecha | Identificador | Tipo | Componente | Descripción |
| :--- | :--- | :--- | :--- | :--- |
| 2026-10-01 | CHG-001 | FIX | Infraestructura | Restauración de .github/ y .vscode/ desde github/ y vscode/. |
| 2026-10-01 | CHG-002 | FEAT | Gobernanza | Creación del control plane sistema/ (CONSTITUCION, INTENCION, ESTADO, etc.). |
| 2026-10-01 | CHG-003 | DOCS | Documentación | Creación de docs/AUDITORIA_SISTEMA_ACTUAL.md y corrección de enlace en docs/README.md. |
| 2026-10-01 | CHG-004 | REFACTOR | Agentes | Especialización modular de agentes en agentes/. |
| 2026-10-01 | CHG-005 | FEAT | Gobernanza | Formalización del ciclo cerrado operativo en sistema/PROTOCOLO_GOBIERNO.md y vinculación agéntica. |
| 2026-10-01 | CHG-006 | FEAT | Automatización | Implementación del controlador CLI (gobierno.ps1/gobierno.py) con Validador Nivel 2 y Sensor Git Nivel 3. |
| 2026-10-01 | CHG-007 | FEAT | Autonomía | Implementación del motor de descubrimiento factual (OBJETIVOS, DESCUBRIMIENTO, AUTONOMIA.json y comando discover). |
| 2026-10-01 | CHG-008 | FEAT | Supervisor | Implementación del Daemon Supervisor persistente (sistema/daemon/), Agente Alimentador (agentes/alimentador/), modos de seguridad y suite 14/14 PASS. |
| 2026-10-01 | CHG-009 | TEST | Hardening | Ejecución de suites hostiles Fases 7, 8, 9 (13/13 vectores repelidos), prueba de autonomía prolongada Fase 10 (5/5 ciclos) y política de control de src/ (Fase 11 y 12). |
| 2026-10-01 | CHG-010 | FEAT/TEST | Seguridad src/ | Implementación del motor de modificación controlada de src/ (pipeline_src.py), verificación del Invariante de Rollback (PRE == POST) y neutralización de 19/19 vectores adversariales (Fase 13). |
| 2026-10-01 | CHG-011 | TEST/HARDENING | Integridad Transaccional | Implementación de la suite de Rollback Destructivo (ataque_rollback_destructivo.py) con 20 vectores de ataque, verificación formal DECLARADO_OK != VERIFICADO_OK, transición obligatoria a SAFE ante PRE != POST (Regla Caso 20) y robustecimiento transaccional de pipeline_src.py (Fase 14). |
| 2026-10-01 | CHG-012 | PERF/REFACTOR | Datapath src/ | Optimización lock-free de LinearPipeline en src/pkg/core/pipeline.go mediante atomic.Pointer[pipelineSnapshot] y evaluación directa de ctx.Err(). Reducción de latencia del 61.01% (38.80 ns -> 15.13 ns/op, 0 allocs), 100% tests PASS y dictamen MEJORA DEMOSTRADA (Fase 15). |
| 2026-10-01 | CHG-013 | GOBERNANZA | Control Plane | Cierre formal y archivado de Fase 20 en historial/. Reseteo canónico de sistema/INTENCION.md a ESTADO: VACÍO y sincronización de ESTADO.md. |
| 2026-10-01 | CHG-014 | GOBERNANZA | FrondaBrick_01 | Absorción y centralización soberana de reglas (acciones, fuentes, permisos, política src, rutas) en frondabrick_01/reglas/ y actualización de enrutamiento en sistema/bin/. |
