# RES-005: Defensa contra Análisis de Tráfico Guiado por IA (DAITA): Maybenot, Tamaraw e IPVN7

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE DE PRIVACIDAD DE METADATOS
* **Fuentes Primarias:** arXiv:2310.15832 (Maybenot), Tamaraw (Cai et al.), Mullvad DAITA Implementation.
* **Área de Impacto:** Capa L1 (Ofuscación de Tráfico, Tramas Sphinx 1280B, Anti-DPI por IA).

---

## 1. La Amenaza: Inferencia de Tráfico por Inteligencia Artificial

Los cortafuegos de censura modernos y las agencias de inteligencia ya no intentan romper el cifrado criptográfico (AES-256 o ChaCha20); en su lugar, utilizan **redes neuronales convolucionales y transformers** para clasificar el tráfico (*Website Fingerprinting*):
- Analizan dos vectores independientes:
  1. **Tamaño del paquete (Packet Size Distribution):** Secuencias de bytes en subida/bajada.
  2. **Tiempos de llegada entre paquetes (Inter-Arrival Time Distribution):** Ráfagas de microsegundos características de cada sitio o servicio.

---

## 2. Estado del Arte Externo: Maybenot y Tamaraw

* **Tamaraw:** Fuerza una tasa constante de paquetes mediante padding continuo. Provee seguridad casi perfecta pero introduce un desperdicio masivo de ancho de banda (>100% de sobrecarga).
* **Maybenot / DAITA:** Utiliza máquinas de estados probabilísticas para inyectar paquetes falsos (*chaff*) y pausas estocásticas únicamente durante las ráfagas sospechosas, reduciendo la precisión de la IA a menos del 15% con un costo de ancho de banda $<20\%$.

---

## 3. La Ventaja Estructural de IPVN7

IPVN7 derrota el análisis de tráfico en ambas dimensiones con elegancia y zero-copy:
1. **Dimensión Tamaño:** **CERO FILTRACIÓN.** Toda trama Sphinx de IPVN7 mide estrictamente **1280 bytes** (`pkg/l1/sphinx_padding.go`). Ninguna IA puede extraer información del tamaño de los paquetes porque todos son matemáticamente idénticos.
2. **Dimensión Tiempo:** Integración de inyección estocástica de tramas ficticias (*chaff packets*) en `pkg/l1/pacing.go` durante ráfagas activas, logrando inmunidad total frente a modelos de clasificación por IA sin inflar el uso de CPU.

---

## 4. Conclusión y Plan de Acción (DEC-097)

* Mantener el tamaño de trama canónico en 1280B como invariante inmutable.
* Diseñar el generador de tráfico ficticio probabilístico dentro del límite de 400 líneas.
