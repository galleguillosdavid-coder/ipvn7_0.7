# DAEMON SUPERVISOR PERSISTENTE IPVN7
> **Directorio:** `sistema/daemon/`  
> **Control Plane:** [sistema/](../)  
> **Norma:** [sistema/CONSTITUCION.md](../CONSTITUCION.md)

---

## 1. Misión del Supervisor
El Daemon Supervisor orquesta la vigilancia y el ciclo continuo del repositorio sin intervención humana en régimen estándar, respetando el presupuesto de autonomía de [`sistema/AUTONOMIA.json`](../AUTONOMIA.json):

```text
OBSERVAR (Alimentador)
   ↓
CLASIFICAR (Evaluador de Autonomía)
   ↓
PRIORIZAR (Daemon)
   ↓
EJECUTAR (Implementador dentro de Scope)
   ↓
VERIFICAR (Sensor Git & Tests)
   ↓
EVIDENCIAR (sistema/EVIDENCIA.md)
   ↓
CERRAR (historial/ & estado.json)
   ↓
REPOSAR (Intervalo configurable)
   ↓
REPETIR
```

---

## 2. Modos Operativos
- **`RUN`:** Ejecución normal periódica. Ejecuta objetivos autónomos permitidos y descansa el intervalo programado.
- **`PAUSE`:** Suspende la ejecución de acciones; mantiene activo el descubrimiento en modo observación pasiva.
- **`SAFE`:** Modo seguro de emergencia. Solo auditoría y verificación; bloquea cualquier cambio en disco.
- **`STOP`:** Cierre ordenado y liberación de recursos (`daemon.lock`).

---

## 3. Barreras de Seguridad
1. **Exclusión Mutua (`daemon.lock`):** Impide múltiples instancias del daemon en el mismo host.
2. **Backoff Adaptativo:** Ante fallos continuos, duplica el tiempo de espera hasta el máximo configurado; si alcanza `max_fallos_consecutivos`, pasa a `PAUSE` o `SAFE`.
3. **Parada por Scope Violation:** Ante la detección de modificaciones no autorizadas (ej. cambios en `src/`), pasa inmediatamente a `SAFE` y escribe en [`sistema/RECHAZOS.md`](../RECHAZOS.md).
4. **Inmutabilidad:** El daemon jamás modifica `CONSTITUCION.md` ni `AUTONOMIA.json`.
