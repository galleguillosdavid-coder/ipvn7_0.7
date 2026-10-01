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

---

## 4. Comandos CLI, Descanso y Reprogramación Dinámica

El daemon cuenta con un wrapper CLI unificado en `sistema/bin/daemon.py`:

```bash
# Ver estado actual del supervisor y próximo ciclo programado
python sistema/bin/daemon.py status

# Ejecutar un único ciclo síncrono
python sistema/bin/daemon.py run-once

# Iniciar bucle continuo autónomo (con descanso e intervalo opcional en segundos)
python sistema/bin/daemon.py start [intervalo_segundos]

# Reprogramar dinámicamente el periodo de descanso entre ciclos
python sistema/bin/daemon.py reprogram <intervalo_segundos>

# Cambiar modo de operación en caliente
python sistema/bin/daemon.py mode <RUN|PAUSE|SAFE|STOP>
```

### Ciclo de Descanso y Reprogramación
Al culminar cada ciclo o tarea, el supervisor:
1. Lee la configuración activa de descanso (`intervalo_actual_segundos`).
2. Persiste la marca temporal del próximo despertar en `estado.json` (`proximo_ciclo`).
3. Entra en reposo / descanso adaptativo monitoreando señales de terminación o reprogramaciones en caliente.
4. Al expirar el descanso, despierta de forma autónoma y ejecuta el siguiente ciclo.

