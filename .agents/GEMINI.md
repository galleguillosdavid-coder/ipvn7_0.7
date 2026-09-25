# DIRECTIVAS DE INGENIERÍA Y MODO DE TRABAJO IPVN7

> [!IMPORTANT]
> **OBJETIVO SUPREMO INVARIABLE:**  
> Red soberana zero-friction más rápida, invisible, intuitiva e indestructible del mundo (1 comando o 1 clic), con una experiencia de usuario (UX) humana, transparente y sin jerga críptica, superando a WireGuard en latencia y a Tor en privacidad, sin servidores centrales ni privilegios de administrador; sustentada en un radar permanente de vigilancia e inteligencia tecnológica exógena que absorbe, contrasta y supera activamente los estándares mundiales de frontera (IETF, NIST, RFCs, papers académicos y proyectos abiertos de vanguardia) mediante rigor físico, pruebas de falsabilidad empíricas y benchmarks factuales, e investigaciones por iniciativa propia de cualquier tema eventualmente necesario o voluntario, para anticipar y superar cualquier amenaza o limitación potencial.
> Todas las reglas canónicas están formalmente definidas en [`AGENTS.md`](./AGENTS.md) y en [`.agents/rules/`](./.agents/rules/).

## Axiomas Esenciales (Resumen Operativo Inviolable)
1. **Algoritmo de 5 Pasos:** Cuestionar requisitos, Eliminar partes innecesarias, Simplificar y optimizar, Acelerar el ciclo, Automatizar.
2. **Cero Simulación:** Sockets UDP/HTTP físicos reales obligatorios; prohibido `time.Sleep` o respuestas fijas.
3. **Realismo Físico 2 Niveles:** Nivel 1 LAN limpia 2 nodos (`.198` <-> `.106`); Nivel 2 Validación Hostil WAN (CGNAT celular, camuflaje TLS 1.3).
4. **Límite Estricto de 400 Líneas (Axioma III):** Ningún archivo >400L. Principio: **"Podar antes de modularizar"** (borrar código innecesario antes de dividir).
5. **Autonomía Operativa Total:** Prohibido preguntar qué hacer al asumir el rol; investigar y actuar por iniciativa propia.
6. **Invariante Zero-Copy:** 0 B/op y 0 allocs/op en L0-L2 (`BenchmarkLinearPipeline_Execute`).
7. **Compuerta Universal:** Obligatorio ejecutar `scripts/verify_ipvn7_standard.ps1` (100% PASS).
8. **Doble Mecanismo Permanente:** Demonio residente PowerShell (`scripts/run_autonomous_daemon.ps1`, `IsDaemon=true`) ejecutándose en background en Windows con 0 tokens de API.
9. **Auto-Interrogación Implacable (Regla 13):** 1) ¿Qué sobra por sobre-ingeniería teórica a podar? 2) ¿Qué falta para 1-clic y adopción masiva? 3) ¿Qué escenario físico hostil falta probar? 4) ¿Qué estándar formal (IETF/NIST) o proyecto abierto líder resuelve esto con mayor elegancia?
10. **Cierre de Bucle Físico Extremo a Extremo (Regla 14):** Prohibido "hecho a medias"; una feature sólo termina cuando el hardware o SO de destino ejecuta físicamente la acción.
11. **Cero Toasts de Éxito Falso (Regla 15):** La UI jamás emite mensajes de éxito sin confirmación física. Si requiere SO (Miracast Win+K, print nativo), se abre el puente nativo con honestidad.
12. **Metodología HIL y Falsabilidad (Regla 16):** Test obligatorio desconectando el dispositivo para probar el camino de error real.
13. **Simplicidad sobre Abstracciones Rotas (Regla 17):** Cero magia inventada; usar puentes físicos soportados por el ecosistema (Miracast, DIAL, /tv receptor web, explorer.exe).
14. **Experiencia de Usuario Radical (Regla 18):** Cero jerga críptica (prohibido "spooler", "soberano", etc. sin traducir a simple); pestañas dedicadas y amplias por función (cero amontonamiento); textos universales multiplataforma y auditoría de usabilidad con agentes persona.
15. **Investigación y Evolución Determinista (Regla 19):** Descubrimiento guiado por 6 Ejes Canónicos; filtro antihumo de 4 pasos (Físico, 5 Pasos, Zero-Copy < 400L, Falsabilidad HIL); ciclo trazable: Web RFC/NIST -> RES-XXX -> DEC-XXX -> TASK-XXX.
16. **Inteligencia Tecnológica Exógena y Estándares Abiertos (Regla 20):** Prohibición de aislamiento en cámara de eco; consulta obligatoria de fuentes primarias (IETF Datatracker, NIST CSRC, repositorios auditados) antes de cada diseño de red.
