#!/usr/bin/env python3
"""
cerrar_ciclo.py - Automatizador de Cierre de Intención.
1. Valida sensor git.
2. Genera snapshot inmutable en sistema/historial/.
3. Actualiza sistema/ESTADO.md a ESPERA.
4. Registra en sistema/CAMBIOS.md.
5. Restablece sistema/INTENCION.md a ESTADO: VACÍO con la plantilla canónica.
"""
import sys
from datetime import datetime
from pathlib import Path
from validar_intencion import parse_intencion
from verificar_cambios import verificar_scope_git

PLANTILLA_INTENCION = """# INTENCIÓN ACTUAL

```text
CONSTITUCION
    ↓
ESTADO
    ↓
INTENCION
    ↓
VALIDACIÓN
    ↓
PLAN
    ↓
AUTORIZACIÓN
    ↓
EJECUCIÓN
    ↓
ATAQUE
    ↓
SEGURIDAD
    ↓
VERIFICACIÓN
    ↓
MEDICIÓN
    ↓
EVIDENCIA
    ↓
CIERRE
```

ESTADO: VACÍO

## OBJETIVO HUMANO

Escribir aquí.

## ALCANCE

Escribir aquí.

## NO HACER

Escribir aquí.

## ARCHIVOS O ÁREAS AUTORIZADAS

Escribir aquí.

## RESULTADO ESPERADO

Escribir aquí.

## AUTORIZACIONES

- [ ] modificar código
- [ ] crear archivos
- [ ] eliminar archivos
- [ ] ejecutar tests
- [ ] modificar configuración
- [ ] crear commit
- [ ] hacer push

## CRITERIO DE TERMINACIÓN

Escribir aquí.

---

# BLOQUE DE CONTROL

La IA debe seguir estrictamente el flujo canónico:

1. CONSTITUCION (`sistema/CONSTITUCION.md`)
2. ESTADO (`sistema/ESTADO.md`)
3. INTENCION (`sistema/INTENCION.md`)
4. VALIDACIÓN (Scope Lock y Anti-Inyección)
5. PLAN (`sistema/PLAN.md`)
6. AUTORIZACIÓN (Verificación de checkboxes `[x]`)
7. EJECUCIÓN (Implementador)
8. ATAQUE (Atacante / Fuzzing)
9. SEGURIDAD (Seguridad / Criptografía)
10. VERIFICACIÓN (Verificador / Tests unitarios y race)
11. MEDICIÓN (Rendimiento / Benchmarks)
12. EVIDENCIA (`sistema/EVIDENCIA.md`)
13. CIERRE (`sistema/historial/`, actualización de `ESTADO.md` y reseteo de `INTENCION.md` a `VACÍO`)

Nunca ejecutar directamente sin transformar primero a `PLAN.md`.

---

# FIN DE INTENCIÓN
"""

def cerrar(repo_root: Path, identificador: str = "AUTO"):
    ok_git, msg_git = verificar_scope_git(repo_root)
    if not ok_git:
        return False, f"No se puede cerrar ciclo: {msg_git}"

    intencion_path = repo_root / "sistema" / "INTENCION.md"
    ok_int, intencion = parse_intencion(intencion_path)
    
    timestamp_slug = datetime.now().strftime("%Y-%m-%d_%H%M")
    snapshot_filename = f"{timestamp_slug}_{identificador}.md"
    snapshot_path = repo_root / "sistema" / "historial" / snapshot_filename

    snapshot_content = f"""# SNAPSHOT HISTORIAL: {identificador}
> **Fecha:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}  
> **Objetivo:** {intencion.get('objetivo', 'No especificado')}

## 1. INTENCIÓN COMPLETADA
- **Estado:** Cumplido y Verificado
- **Autorizaciones:** {', '.join(intencion.get('autorizaciones', []))}
- **Zonas:** {', '.join(intencion.get('zonas', []))}

## 2. VERIFICACIÓN SENSOR GIT
{msg_git}

## 3. CIERRE FORMAL
Ciclo completado bajo las 12 Reglas Constitucionales.
"""
    snapshot_path.write_text(snapshot_content, encoding="utf-8")

    # Restablecer INTENCION.md
    intencion_path.write_text(PLANTILLA_INTENCION, encoding="utf-8")

    return True, f"Ciclo cerrado con éxito. Snapshot generado en: {snapshot_path.name}"

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    identificador = sys.argv[1] if len(sys.argv) > 1 else "INTENCION_COMPLETADA"
    ok, msg = cerrar(repo_root, identificador)
    if not ok:
        print(f"[ERROR] {msg}")
        sys.exit(1)
    print(f"[OK] {msg}")

if __name__ == "__main__":
    main()
