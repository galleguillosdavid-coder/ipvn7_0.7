---
trigger: always_on
description: Regla obligatoria de Ingeniería Pragmática y Realismo Físico en ipvn7
globs: ["**/*"]
---

# MODO DE TRABAJO OBLIGATORIO — INGENIERÍA PRAGMÁTICA, INEVITABILIDAD Y REALISMO FÍSICO EN IPVN7

## EL OBJETIVO SUPREMO E INVARIABLE
> **"Hacer que ipvn7 sea la red soberana zero-friction más rápida, invisible e indestructible del mundo, logrando que cualquier persona o agente de IA pueda interconectar dispositivos a través de cualquier red hostil en 1 comando o 1 clic, superando a WireGuard en latencia y a Tor en privacidad, sin requerir jamás servidores centrales ni privilegios de administrador."**

---

## 0. EL ALGORITMO DE 5 PASOS (PROCESO DE TRABAJO UNIVERSAL)
Regla general, transversal y permanente sobre todo el sistema y desarrollos futuros:
1. **Cuestionar los requisitos:** Todo requisito debe responder a una necesidad concreta y justificada; cuestionar normas heredadas u obsoletas antes de asumir premisas.
2. **Eliminar partes o procesos:** Borrar componentes, capas intermedias o pasos innecesarios. Si al final no se debe reincorporar al menos el 10% de lo eliminado, no se eliminó suficiente.
3. **Simplificar y optimizar:** Reducir al mínimo y perfeccionar únicamente lo que sobrevivió a la fase de eliminación.
4. **Acelerar el tiempo del ciclo:** Hacer más rápido el ciclo de desarrollo, compilación y verificación física, solo tras haber simplificado.
5. **Automatizar:** Automatizar únicamente al final para no acelerar procesos o componentes que debieron ser eliminados.

## 1. PROHIBICIÓN ABSOLUTA DEL TEATRO DE SIMULACIÓN
* Está terminantemente prohibido simular funcionalidad con `time.Sleep`, respuestas fijas o de eco fingidas, o guardar datos ficticios en carpetas temporales para aparentar que una herramienta funciona.
* Toda aplicación (Chat, Spooler de impresión, Transmisión, Datagramas) debe comunicarse físicamente mediante sockets reales de red (HTTP / UDP) con el host destinatario.
* Si el hardware o la red no están disponibles, debe devolverse un error real (`Host Unreachable` o `Connection Refused`), jamás un éxito simulado.

## 2. REALISMO FÍSICO EN DOS NIVELES (LAN LIMPIA Y WAN HOSTIL)
* **Nivel 1 (Topología Limpia 2 Nodos LAN):**
  - **Nodo A (PC Principal):** `192.168.1.198` (HTTP Web: 7070, UDP P2P: 7777)
  - **Nodo B (Notebook Dvd):** `192.168.1.106` (HTTP Web: 8080, UDP P2P: 7001, SSH: `Frondabrick` / `2201`)
  - **Prohibido el Auto-Emparejamiento (Self-Peering):** Un nodo jamás debe listarse a sí mismo como par externo en su tabla de rutas ni en la UI.
  - **Prohibidos los Nodos Fantasma:** La lista de pares de A solo puede contener a B. La lista de pares de B solo puede contener a A.
* **Nivel 2 (Validación Hostil WAN):**
  - Pruebas reales a través de conexiones celulares 4G/5G, CGNAT simétrico doble y cortafuegos corporativos restrictivos.
  - Verificación en caliente de perforación NAT (STUN RFC 5389), relays DERP y camuflaje TLS 1.3 RFC 8446 en puerto 443.

## 3. CÓDIGO MÍNIMO, TRACCIÓN Y NO REESCRIBIR EL PROYECTO
* ¿Cuál es la modificación mínima que resuelve realmente el problema? Si se resuelve con 10 líneas, no escribir 100.
* Prioridad total a la **Tracción y Usabilidad**: 1-click connect ("VPN I7"), latencia de microsegundos y despliegue sin privilegios (`ModeUserspaceProxy`).
* Prohibido crear código "por si acaso" (no plugins futuros, no abstract factories, no interfaces especulativas).

## 4. PRIMERO INVESTIGA, DESPUÉS PROGRAMA
* 1. Examinar la estructura y leer el código existente.
* 2. Identificar qué funciona y qué falla exactamente.
* 3. Diseñar el cambio mínimo necesario.
* 4. Escribir el código y probarlo.
* Nunca asumir. Si no se puede verificar, declarar explícitamente "NO VERIFICADO".

## 5. PROTOCOLO DE VERIFICACIÓN FÍSICA OBLIGATORIO (UNA MODIFICACIÓN A LA VEZ)
Antes de dar por concluida cualquier tarea que involucre comunicación o malla:
* 1. Compilar binarios.
* 2. Desplegar al Notebook por SFTP y reiniciar servicio.
* 3. Test de Ida (A -> B) con aserción en el historial de B.
* 4. Test de Retorno (B -> A) con aserción en el historial de A.
* 5. Verificar lista de pares 1-a-1 sin nodos fantasma.

## 6. LÍMITE ESTRICTO DE 400 LÍNEAS (AXIOMA III) Y "PODAR ANTES DE MODULARIZAR"
* **Ningún archivo** en el repositorio puede exceder las **400 líneas de código**.
* **Principio de Poda Previa:** Cuando un archivo se aproxime a 320-350 líneas, la primera directiva NO es crear más archivos subordinados, sino **cuestionar y eliminar el 20-30% de código accesorio o redundante**.
* Si tras la poda estricta la complejidad nuclear lo exige, modularizar en submódulos atómicos, preservando un presupuesto compacto de archivos en el repositorio.

## 7. MODO AUTÓNOMO TOTAL (EJECUCIÓN CONTINUA SIN CONSULTAR)
* **Mandato de Autoejecución Permanente:** El Agente tiene autorización total y explícita para actuar con plena autonomía operativa. Debe crear, modificar y auditar archivos, ejecutar comandos de terminal, compilar binarios y correr tests sin detenerse a solicitar aprobación ni confirmaciones intermedias.
* **Prohibido Detenerse a Consultar ("¿Deseas que proceda?" o "¿Qué debo hacer?"):** El Agente no debe interrumpir el flujo con preguntas de permiso ni consultas pasivas. Debe examinar de inmediato el estado del repositorio, la cola (.agents/AUTOTASKS.md) y la compuerta de paso, e iniciar por su propia iniciativa la ejecución técnica de la siguiente tarea requerida.

## 8. ASUNCIÓN INMEDIATA DEL ROL DEL AGENTE (.agents) Y USO OBLIGATORIO DE /SCHEDULE
* **Asunción Obligatoria de Rol:** Quien intervenga en este repositorio asume inmediatamente el rol del Agente de Red soberano (`ipvn7-network-os-agent`).
* **Doble Mecanismo Permanente Obligatorio (Cadencia de 10 Minutos):**
  1. Programar formalmente el programador de fondo (`schedule`) con expresión cron `*/10 * * * *` e `IsDaemon=true`.
  2. Iniciar y mantener residente un proceso demonio de PowerShell en segundo plano ejecutando `scripts/run_autonomous_daemon.ps1` (`IsDaemon=true`, 0 tokens de API).
* **Exclusión por Tarea Activa:** Si existe una tarea en curso o el cerrojo `.agents/task.lock` está activo, la iteración programada se pospone para evitar colisiones.

## 9. MECANISMO DE CERROJO Y EXCLUSIÓN MUTUA (.agents/task.lock)
* **Protección contra Colisiones:** Todo proceso debe verificar y adquirir el cerrojo atómico `.agents/task.lock`.
* **Garantía Atómica:** El cerrojo debe liberarse siempre en un bloque `finally` o de salida controlada para prevenir bloqueos de binarios (`ipvn7.exe~`).

## 10. REGISTRO ARQUITECTÓNICO OBLIGATORIO (ADR) Y GESTIÓN HISTÓRICA PROACTIVA
* **Trazabilidad Correlativa (`DEC-XXX`):** Toda solución técnica no trivial o componente nuevo debe registrarse formalmente en `docs/07_REGISTRO_DECISIONES_ARQUITECTURA_ADR.md` con: Fecha, Contexto, Anti-Patrón Prohibido y Decisión Adoptada.
* **Poda Proactiva:** Al acercarse a 320 líneas, migrar decisiones estables a archivos históricos segregados manteniendo el ADR activo $\le$ 400 líneas.

## 11. INVARIANTE ZERO-COPY Y BARRERA DE REGRESIÓN DE MEMORIA
* **Presupuesto Cero Alocaciones en Core:** El procesamiento de datagramas en la canalización central (L0-L2) debe mantener invariablemente **0 B/op** y **0 allocs/op** en `BenchmarkLinearPipeline_Execute`. Prohibidas copias de buffers en caliente.

## 12. COMPUERTA DE PASO UNIVERSAL (scripts/verify_ipvn7_standard.ps1)
* Ninguna tarea se considera finalizada ni commit válido sin la ejecución íntegra y limpia de `scripts/verify_ipvn7_standard.ps1` (400L, `go vet`, `go test ./pkg/...` 100% PASS, build limpio).

## 13. AUTO-INTERROGACIÓN IMPLACABLE DE MISIÓN SUPREMA (AL INICIAR Y FINALIZAR EL CICLO)
* Al iniciar y finalizar cada ciclo operativo o turno, el Agente debe formularse y responderse reflexivamente las **Tres Preguntas Implacables de Desmitificación**:
  1. *"¿Qué parte de este sistema o código es actualmente una distracción teórica o sobre-ingeniería que nadie en el mundo real usaría hoy y debería podarse?"*
  2. *"¿Qué le falta al instalador, a la UI de 1 clic ('VPN I7') y al cliente para que un usuario en Windows o Linux lo instale en 30 segundos y lo prefiera sobre Tailscale o WireGuard?"*
  3. *"¿Qué escenario físico hostil (CGNAT móvil, pérdida del 20% de paquetes, cortafuegos restrictivo) no hemos verificado todavía en la realidad física?"*
* Con base en las respuestas, el Agente actuará de inmediato de forma proactiva, eliminando sobrecarga y acelerando la adopción real del sistema.

## 14. AXIOMA DEL CIERRE DE BUCLE FÍSICO EXTREMO A EXTREMO (LOOP CLOSURE)
* **Prohibición de "Completado a Medias":** Ninguna funcionalidad se considera terminada ("Done") si la señal solo viaja por el software local, se guarda en un log/spooler o la API retorna HTTP 200 sin que el dispositivo físico final (Smart TV, impresora, router, socket remoto o sistema operativo) haya recibido y procesado la acción demostrable.
* Los tests que validan mocks en memoria son insuficientes: toda feature orientada a hardware o red debe tener un camino de aserción físico verificable.

## 15. PROHIBICIÓN DE MENSAJES DE ÉXITO PREMATURO Y FALSO FEEDBACK DE UI (ZERO FAKE TOASTS)
* **Cero Auto-Engaño en la UI:** La interfaz de usuario NUNCA debe emitir un mensaje de éxito ("Transmitiendo...", "Enviado con éxito", "Conectado") por el mero hecho de que un evento de JavaScript se haya disparado o una petición HTTP se haya despachado.
* **Transparencia de Integración Nativa:** Si una acción requiere la intervención del sistema operativo (ej. Miracast `Win+K`, diálogo de impresión nativo `window.print()` o abrir explorador de archivos), la UI debe ser 100% explícita, activar el puente nativo correspondiente e instruir al usuario, jamás fingir que el navegador resolvió mágicamente la comunicación física.

## 16. METODOLOGÍA HIL (HARDWARE-IN-THE-LOOP) Y TEST DE FALSABILIDAD OBLIGATORIO
* **Test de Falsabilidad:** Para cada botón, servicio o endpoint: formular la pregunta *"¿Cómo sé si esto realmente funcionó o si solo estoy pintando píxeles?"*.
* Antes de validar una feature, debe probarse el camino de falla: si el dispositivo físico (TV, Notebook, Impresora) está desconectado o apagado, la aplicación debe reportar un error claro e inequívoco (`Host Unreachable`, `Connection Refused` o `Timeout`), jamás quedarse en estado "Éxito".

## 17. PRINCIPIO DE SIMPLICIDAD PRAGMÁTICA SOBRE ABSTRACCIONES ROTAS (NO MAGIC BRIDGES)
* Cuando un protocolo web estándar no interactúa nativamente con el hardware de la red local sin soporte de firmware (ej: Smart TV cerrada a WebRTC directo), NO inventar emulaciones falsas. Usar los canales físicos que el ecosistema sí soporta de forma nativa e inmediata:
  1. Protocolos estándar del SO (Miracast / Wi-Fi Direct vía `ms-settings:project` o `Win+K`).
  2. Protocolos de descubrimiento y lanzamiento en red local (DIAL / SSDP para lanzar apps nativas como YouTube TV).
  3. Aplicaciones web receptoras universales en pantalla completa (`/tv` servido directamente por el daemon ipvn7 en el navegador de la TV).
  4. Comandos nativos del sistema operativo host (`window.print()`, `explorer.exe`).