# REGLAS Y POLÍTICAS OPERATIVAS DE FRONDABRICK_01

> **Custodio Soberano:** [FrondaBrick_01](../../agentes/Frondabrick01/AGENTE.md)  
> **Directorio Rector:** `frondabrick_01/reglas/`  
> **Norma Suprema:** [`sistema/CONSTITUCION.md`](../../sistema/CONSTITUCION.md) y [`docs/FUENTE_DE_VERDAD.md`](../../docs/FUENTE_DE_VERDAD.md)  

---

## 1. MISIÓN Y CENTRALIZACIÓN
Por mandato de autogobernanza y resolución de ambigüedad arquitectónica, la totalidad de los marcos normativos, restricciones de scope, políticas de modificación de código fuente (`src/`), permisos de acciones y catálogos de criticidad de rutas han sido absorbidos y centralizados bajo la supervisión directa del **Agente Principal FrondaBrick_01**.

## 2. CATÁLOGO DE REGLAS ABSORBIDAS
| Archivo | Función |
| :--- | :--- |
| [`acciones.json`](acciones.json) | Definición y estados de acciones permitidas (lectura, escritura, commit, push). |
| [`permisos.json`](permisos.json) | Mapeo de autorizaciones y zonas requeridas para cada capacidad técnica. |
| [`politica_src.json`](politica_src.json) | Zonas bloqueadas del Core (`wire.go`, `crypto.go`, `pqc_*`, `interfaces/`) y pipeline de 6 agentes. |
| [`rutas.json`](rutas.json) | Niveles de criticidad y clasificación taxonómica de los directorios del repositorio. |
| [`fuentes_descubrimiento.json`](fuentes_descubrimiento.json) | Sensores autorizados para el descubrimiento factual de anomalías y trabajo pendiente. |

## 3. UNIFICACIÓN DE GOBERNANZA
Cualquier inspección, validación o ejecución de compuertas (Nivel 2 y Nivel 3) opera consultando estas definiciones centralizadas para evitar divergencias o desincronizaciones entre subsistemas.
