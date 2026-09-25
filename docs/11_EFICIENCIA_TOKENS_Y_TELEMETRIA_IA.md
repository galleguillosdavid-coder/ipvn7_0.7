# 11. Eficiencia de Tokens, Telemetría y Optimización Continua de IA

> **Axioma de Observabilidad Zero-Cost:**  
> *"Consultar el gasto de tokens dentro del modelo consume tokens; medirlo fuera de banda en el sistema anfitrión cuesta exactamente 0 tokens."*

---

## 1. Visión y Fundamentos del Consumo de Tokens

En el desarrollo y operación de redes autónomas como IPVN7, el agente de Inteligencia Artificial (IA) y el operador humano deben coordinarse bajo una disciplina estricta de eficiencia de recursos. Un consumo descontrolado de tokens no solo incrementa los costos económicos, sino que degrada la velocidad de inferencia, introduce latencia cognitiva y aproxima prematuramente la ventana de contexto al límite de compactación.

---

## 2. Las Cuatro Dimensiones del Gasto de Tokens

| Dimensión | Definición | Mecanismo de Medición | Costo de Consultar |
| :--- | :--- | :--- | :--- |
| **A. Estimado Previsto** | Lo que calculamos que consumirá una tarea antes de ejecutarla. | Fórmula heurística: $\text{Base} + \text{Historial} + \text{Salida}$. | **0 Tokens** (Cálculo local en cliente). |
| **B. Real Consumido** | El volumen físico de caracteres y pasos generados en la iteración. | Conteo determinista de bytes y caracteres procesados. | **0 Tokens** (Inspección de logs locales). |
| **C. Reportado por la API** | La telemetría inmutable de la plataforma (`usage_metadata`). | `prompt_token_count` + `candidates_token_count`. | **0 Tokens** (Lectura de `transcript.jsonl`). |
| **D. Costo de Consulta** | El impacto de preguntar por el consumo dentro del prompt. | Re-envío de la ventana de contexto completa. | **ALTO (>50k-200k tokens por pregunta)**. |

---

## 3. Modelo Matemático de Estimación Predictiva

El costo de un turno conversacional se rige por:

$$\text{Tokens}_{\text{Turno}} = \text{Tokens}_{\text{Sistema}} + \text{Tokens}_{\text{Historial Acumulado}} + \text{Tokens}_{\text{Mensaje Usuario}} + \text{Tokens}_{\text{Respuesta}}$$

### Constantes y Factores de Conversión
* **Prompt Base Fijo (System Prompt + Tools):** $\approx 18,000 - 25,000 \text{ tokens}$.
* **Factor de Conversión en Español / Código Mixto:**  
  $$\text{Tokens} \approx \frac{\text{Caracteres}}{3.2}$$  
  *(El español utiliza entre un 25% y 35% más tokens que el inglés debido a caracteres acentuados y morfología flexiva).*
* **Costo Referencial (Tarifas Base por Millón de Tokens):**
  * *Entrada (Prompt):* $\$0.075 \text{ USD} / 1\text{M tokens}$.
  * *Salida (Generación):* $\$0.300 \text{ USD} / 1\text{M tokens}$.

---

## 4. De Dónde Provienen las Fugas y Sorpresas de Tokens

El **90% del consumo de tokens** en sesiones de desarrollo asistido proviene de la **Entrada (Prompt Tokens)**, no de lo que escribe la IA. Las tres principales fuentes de fuga son:

1. **Comandos de Terminal Ruidosos:**  
   Ejecutar comandos que devuelven cientos de líneas de trazas, warnings o tests verbose inyecta decenas de miles de caracteres al contexto. Ese texto se reenviará en **todos y cada uno de los turnos siguientes** hasta que la sesión termine o se compacte.
2. **Lectura Completa de Archivos sin Segmentar:**  
   Leer un archivo de 800 líneas cuando solo se requerían 15 líneas añade miles de tokens que se arrastran indefinidamente.
3. **Pings Periódicos por API:**  
   Configurar timers de LLM que despiertan cada 10 minutos para preguntar *"¿Hay tareas?"* re-envía todo el proyecto (100k+ tokens) en cada chequeo.

---

## 5. El Auditor Local Zero-Cost: `scripts/audit_token_usage.ps1`

Para inspeccionar el consumo en cualquier momento con **0 tokens de API**, se provee el script de auditoría en PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\audit_token_usage.ps1
```

### Capacidades del Auditor:
* **Lectura Directa:** Procesa localmente el archivo inmutable `transcript.jsonl` generado por Antigravity en `.system_generated/logs/`.
* **Desglose Métrico:** Pasos de usuario, respuestas de IA, tamaño de la ventana de contexto y tokens de salida.
* **Detección de Picos:** Alerta instantáneamente si un comando o archivo inyectó más de 15,000 caracteres en un solo paso.
* **Cálculo de Costo Financiero:** Estimación acumulada en dólares estadounidenses.
* **Recomendaciones en Tiempo Real:** Avisa si la ventana supera los 150k tokens para sugerir compactación o nuevo chat.

---

## 6. Reglas de Eficiencia para la IA (Auto-Optimización)

Toda IA que opere en IPVN7 debe seguir las siguientes directivas obligatorias:

1. **Preferir Scripts Locales sobre Tareas en Chat:**  
   Utilizar el Demonio Autónomo (`scripts/run_autonomous_daemon.ps1`, `IsDaemon=true`) para tareas repetitivas en segundo plano. Este proceso corre en PowerShell local con **0 tokens de API**.
2. **Lectura Quirúrgica (`StartLine` y `EndLine`):**  
   Al usar `view_file`, especificar siempre el rango exacto de líneas necesarias en lugar de leer el archivo completo.
3. **Filtrar Salidas de Comandos:**  
   Usar comandos de terminal que silencien salidas redundantes (`-q`, `Select-Object -First 10`, etc.).
4. **Respetar el Límite de 400 Líneas (Axioma III):**  
   Archivos compactos significan lecturas rápidas, menor consumo de tokens y diffs quirúrgicos en caliente.

---

## 7. Consejos de Ahorro para el Operador Humano

1. **Prompts Concisos y Específicos:** Instrucciones claras y delimitadas evitan idas y vueltas innecesarias.
2. **Reiniciar Sesión al Cambiar de Dominio:** Si completaste una fase grande y pasas a otra no relacionada, abrir un chat nuevo reinicia la ventana de contexto de 200k a 0 tokens.
3. **Consultar Tokens con el Script Local:** No preguntar en el chat por estadísticas de tokens; ejecutar `scripts/audit_token_usage.ps1` en PowerShell para obtener el dato exacto al instante y gratis.
