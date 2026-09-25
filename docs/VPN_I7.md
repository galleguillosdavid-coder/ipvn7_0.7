# Especificación de Arquitectura y Hoja de Ruta: VPN I7

**Estado:** PRESERVADO PARA IMPLEMENTACIÓN FUTURA  
**Fecha:** 2026-09-21  
**Filosofía:** El Algoritmo de 5 Pasos (Ingeniería Lean / Elon Musk)  
**Objetivo:** Transformación del sistema en un cliente y nodo VPN real, soberano y ultra-ligero para Windows ("VPN I7"), con UI simplificada, reducción drástica de código y purga de historial git.

---

## 1. Fundamentación y Algoritmo de 5 Pasos

### Paso 1: Cuestionar cada Requisito
* **Cuestionamiento del Frontend:** El sistema actual acumula motores de física de partículas de grafos en canvas (`app_kuzu_physics.js`, `app_kuzu_canvas.js`, `style_kuzu_controls.css`) que consumen más de 1.100 líneas de código y recursos de CPU sin aportar al propósito nuclear: ser una VPN segura, determinista y de alto rendimiento.
* **Cuestionamiento de Git:** El historial de `.git` acumula 50 MB debido a binarios históricos y commits pesados. Para despliegues ágiles, debe ser purgado a un commit raíz limpio (<2 MB).
* **Propósito Primario:** La red debe presentarse al usuario final como **"VPN I7"**, una VPN soberana nativa para Windows con control de conexión en un clic, micro-segmentación ZTNA y acceso transparente a herramientas locales.

### Paso 2: Eliminar Partes o Procesos Innecesarios
1. **Poda Frontend:**
   * Retirar `web/app_kuzu_physics.js` (~341 líneas).
   * Retirar `web/app_kuzu_canvas.js` (~180 líneas).
   * Retirar `web/style_kuzu_controls.css` (~324 líneas).
   * Retirar `web/style_kuzu_workspace.css` (~250 líneas).
   * *Ahorro neto:* > 1.090 líneas de código frontend eliminadas.
2. **Purga de Git:**
   * Crear rama huérfana (`orphan`) para reiniciar el árbol sin historial binario obsoleto.
   * Reducción estimada del repositorio: de 50 MB a < 2 MB.

### Paso 3: Simplificar y Optimizar
1. **Identidad y Diseño de UI "VPN I7":**
   * Panel unificado tipo "Dark Glassmorphism" (obsidian `#090d16`, acento `#00e5ff` cian neón y `#00ff9d` verde cuántico de túnel activo).
   * **Interruptor Maestro Central:** Conectar / Desconectar VPN con animación de pulso criptográfico.
   * **Tarjeta de Identidad Soberana:** Muestra IP Virtual (`10.7.0.x / fd07::...`), DID criptográfico y estado de cifrado post-cuántico.
   * **Selector de Nodo Físico:** Selección 1-clic de pasarela con latencia ping en vivo (Nodo A `192.168.1.198:7777`, Nodo B `192.168.1.106:7001`).
   * **Telemetría Instantánea:** Rendimiento TX/RX en tiempo real, pérdida de paquetes y estado del Cortafuegos ZTNA.
   * **Herramientas Soberanas Compactas:** Pestañas desplegables o modales para Chat E2EE P2P, Spooler de impresión física y IoT Shield sin saturar la pantalla principal.
2. **Límite Estricto $\le$ 400 Líneas (Axioma III):**
   * Reducción de `web/index.html` de 383 líneas a ~220 líneas.
   * `web/app_core.js` enfocado exclusivamente en estado de túnel, métricas y cambio de nodos (<250 líneas).

### Paso 4: Acelerar el Tiempo de Ciclo
* Pruebas automáticas completas en < 3s (`go test ./...`).
* Carga de la aplicación web en navegador < 100ms gracias a la eliminación del motor de física.

### Paso 5: Automatizar
* Binarios finales limpios y ligeros en `bin/windows_amd64/` (`ipvn7.exe`, `ipvn7-cli.exe`, `ipvn7-bridge.exe`).
* Script de compilación reproducible `scripts/build_dual.ps1`.

---

## 2. Especificación Técnica de la UI "VPN I7"

### Estructura de Componentes Visuales:
```text
┌─────────────────────────────────────────────────────────────┐
│  [🛡️ VPN I7]  Sovereign Network OS         [🟢 PROTEGIDO]    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    ┌───────────────────┐                    │
│                    │   ((  🔘  ))      │                    │
│                    │  DESCONECTAR VPN  │                    │
│                    │   10.7.0.198/16   │                    │
│                    └───────────────────┘                    │
│                                                             │
│   IP SOBERANA: 10.7.0.198        ENCRIPTACIÓN: Dilithium3   │
│   PASARELA: Nodo B (192.168.1.106)  LATENCIA: 1.4 ms RTT    │
│                                                             │
├───────────────────────────────┬─────────────────────────────┤
│ 📊 TELEMETRÍA Y TRÁFICO       │ 🌐 PASARELAS DISPONIBLES    │
│   TX: 1.4 MB | RX: 3.8 MB     │   * Nodo B (Notebook Dvd)   │
│   Drop: 0 | ZTNA: Estricto    │   * Modo Local Directo      │
├───────────────────────────────┴─────────────────────────────┤
│ 🛠️ HERRAMIENTAS SOBERANAS: [💬 Chat E2EE] [🖨️ Spooler] [🛡️ IoT] │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Guía de Ejecución Futura (Paso a Paso)

Cuando el usuario instruya la implementación de esta especificación:
1. **Fase 1 (Poda de Archivos):** Eliminar los archivos JS/CSS de física de Kùzu detallados en el Paso 2.
2. **Fase 2 (Reestructuración UI):** Actualizar `web/index.html`, `web/style_base.css`, `web/style_dashboard.css` y `web/app_core.js` con los componentes de "VPN I7".
3. **Fase 3 (Conectores API):** Conectar endpoints `/api/v1/status` y `/api/v1/peers` para reflejar el estado del túnel VPN en tiempo real.
4. **Fase 4 (Purga Git):** Ejecutar rama huérfana limpia y reindexación de objetos para comprimir el espacio del repositorio.
5. **Fase 5 (Certificación):** Validar `go test ./...` (<3s), auditoría de $\le 400$ líneas y compilación cruzada.
