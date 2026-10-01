# PLAN DE EJECUCIÓN: FASE 17 — INVESTIGACIÓN AUTÓNOMA

## 1. INTENCIÓN Y OBJETIVO
- **Referencia:** [sistema/INTENCION.md](INTENCION.md)
- **Objetivo:** Encontrar el siguiente cuello de botella real de IPVN7 mediante medición empírica y profiling de datapath.
- **Restricciones Absolutas:**
  - ❌ NO modificar código en `src/`.
  - ❌ NO crear CHG nuevo.
  - ❌ NO hacer commit/push.
  - ❌ No optimizar nada todavía.
  - ❌ No asumir a priori que el cuello de botella es criptografía, red o pipeline.
  - ❌ Medir antes de concluir.

---

## 2. ETAPAS DE INVESTIGACIÓN Y MEDICIÓN

1. **Revisión Arquitectónica y Tests Actuales:**
   - Confirmar que todo el repositorio pasa `go test -count=1 ./...` y `go vet ./...`.
   - Inspeccionar los componentes de datapath en `src/pkg/core`, `src/pkg/l0`, `src/pkg/l1`, `src/pkg/l2`.

2. **Identificación de Rutas Críticas del Datapath:**
   - Ruta A: Deserialización y Serialización CBOR RFC 8949 (`Packet.Encode()` / `DecodePacket()`) en `pkg/l0`.
   - Ruta B: Verificación criptográfica y firma Ed25519 (`SignPacket()` / `VerifyPacket()`) en `pkg/l0`.
   - Ruta C: Pool y Gestión de Contextos de Paquete (`AcquirePacketContext()` / `ReleasePacketContext()`) en `pkg/core`.
   - Ruta D: Filtrado ZTNA (`ZTNAFirewall.Evaluate()`) en `pkg/l1`.
   - Ruta E: Enrutamiento Kleinberg y Búsqueda de Anillos (`KleinbergRouter.Route()`) en `pkg/l1`.
   - Ruta F: Pipeline general con etapas completas (`DecodeCBORStage` + `ZTNAFilterStage` + `QoSFilterStage` + `TopologyFSMStage`).

3. **Ejecución de Benchmarks de Profiling:**
   - Construir arnés de perfilado factual en `sistema/bin/investigar_cuellos_botella.go`.
   - Medir para cada componente:
     - Latencia media ($\mu\text{s}$ o $\text{ns/op}$)
     - Asignaciones de memoria (Bytes/op y Allocs/op)
     - Consumo relativo porcentual en el datapath completo.

4. **Análisis de Trabajo Redundante y Asignaciones Evitables:**
   - Comparar el peso relativo de cada operación en el ciclo de vida por paquete.
   - Identificar asignaciones de memoria que presionan el Garbage Collector de Go en alta concurrencia.

5. **Formulación de Candidatos (Máximo 3):**
   - Candidato 1, 2 y 3 detallando:
     - Evidencia fáctica
     - Medición empírica
     - Impacto potencial
     - Dificultad técnica
     - Riesgo de regresión
     - Experimento mínimo demostrativo.

6. **Selección Formal:**
   - Seleccionar un único candidato como `OBJETIVO-017`.
   - Clasificar según criterio: `CUELLO DE BOTELLA DEMOSTRADO`, `SOSPECHA` o `INCONCLUSO`.

7. **Cierre de Ciclo:**
   - Actualizar `sistema/EVIDENCIA.md` y `sistema/ESTADO.md`.
   - Crear snapshot inmutable en `sistema/historial/2026-10-01_1900_FASE17_INVESTIGACION_CUELLOS_BOTELLA.md`.
   - Restablecer `sistema/INTENCION.md` a `ESTADO: VACÍO`.

---

## 3. ARCHIVOS AUTORIZADOS
- `sistema/bin/investigar_cuellos_botella.go`
- `sistema/PLAN.md`
- `sistema/ESTADO.md`
- `sistema/EVIDENCIA.md`
- `sistema/historial/*`
- `sistema/INTENCION.md`

## 4. ESTADO DEL PLAN
**APROBADO PARA EJECUCIÓN (RENDIMIENTO + ARQUITECTO + AUDITOR)**
