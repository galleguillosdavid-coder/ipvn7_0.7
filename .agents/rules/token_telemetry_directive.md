# DIRECTIVA DE TELEMETRÍA DE TOKENS POR ITERACIÓN EN ANTIGRAVITY

> **Regla Operativa Obligatoria para el Agente:**  
> Al finalizar cada turno o iteración en la interfaz de Antigravity, el agente DEBE incluir un pie de telemetría de consumo de tokens basado en la fórmula determinista de `docs/11_EFICIENCIA_TOKENS_Y_TELEMETRIA_IA.md` y los datos del transcript local.

---

## Formato del Bloque de Telemetría (Pie de Mensaje)

El bloque debe presentarse con la siguiente estructura:

```markdown
---
### 📊 Telemetría de Tokens de la Iteración (Antigravity)
* **Tokens de Entrada Estimados (Prompt Context Window):** ~[N] tokens
* **Tokens de Salida Generados (Esta Iteración):** ~[N] tokens
* **Pasos Registrados en la Sesión:** [Total] pasos | [U] turnos de usuario
* **Costo Acumulado Estimado de la Sesión:** $[X.XX] USD
* **Estado de Eficiencia:** [🟢 ÓPTIMO / 🟡 MODERADO / 🔴 ALTO]
* **Recomendación Proactiva:** [Consejo contextual de ahorro y optimización]
---
```

---

## Factores de Cálculo Determinista
1. **Factor de Conversión:** $\text{Tokens} \approx \text{Caracteres} / 3.2$ (Español y código fuente).
2. **Tarifas Base de Referencia:**
   * Entrada: $0.075 USD / 1M tokens.
   * Salida: $0.300 USD / 1M tokens.
3. **Umbrales de Alerta:**
   * Ventana $< 80,000$ tokens: 🟢 Óptimo.
   * Ventana entre $80,000$ y $150,000$ tokens: 🟡 Moderado.
   * Ventana $> 150,000$ tokens: 🔴 Alto (Sugerir compactación o inicio de nueva sesión limpia).
