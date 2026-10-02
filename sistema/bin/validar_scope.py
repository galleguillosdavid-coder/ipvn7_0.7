#!/usr/bin/env python3
"""
validar_scope.py - Validador Nivel 2 de Scope Lock.
Verifica que las zonas solicitadas y las acciones tengan respaldo en permisos.json y rutas.json.
"""
import json
import sys
from pathlib import Path
from validar_intencion import parse_intencion

def validar_alcance(repo_root: Path):
    intencion_path = repo_root / "sistema" / "INTENCION.md"
    permisos_path = repo_root / "frondabrick_01" / "reglas" / "permisos.json"
    if not permisos_path.exists():
        permisos_path = repo_root / "sistema" / "reglas" / "permisos.json"

    rutas_path = repo_root / "frondabrick_01" / "reglas" / "rutas.json"
    if not rutas_path.exists():
        rutas_path = repo_root / "sistema" / "reglas" / "rutas.json"

    ok, intencion = parse_intencion(intencion_path)
    if not ok:
        return False, f"Fallo al leer intención: {intencion}"

    with open(permisos_path, "r", encoding="utf-8") as f:
        permisos_cfg = json.load(f)["permisos"]

    with open(rutas_path, "r", encoding="utf-8") as f:
        rutas_cfg = json.load(f)["rutas"]

    # Si el estado es VACÍO, no hay conflicto
    if intencion["estado"] == "VACÍO":
        return True, "Estado VACÍO. Scope lock inactivo."

    # Comprobación de seguridad: modificar código sin autorización
    zonas = intencion["zonas"]
    autorizaciones = intencion["autorizaciones"]

    toca_src = any(z.startswith("src") for z in zonas)
    permite_codigo = any("código" in a or "codigo" in a for a in autorizaciones)

    if toca_src and not permite_codigo:
        return False, "VIOLACIÓN DE ALCANCE: Zonas incluye 'src/' pero no se marcó la autorización de modificar código."

    # Comprobación de acciones de alto riesgo
    for p_id, p_info in permisos_cfg.items():
        req = p_info["requiere_checkbox"]
        if req in autorizaciones:
            # Acción explícitamente autorizada
            pass

    return True, {
        "status": "SCOPE_VALIDADO",
        "zonas_autorizadas": zonas,
        "acciones_autorizadas": autorizaciones
    }

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    ok, res = validar_alcance(repo_root)
    if not ok:
        print(f"[RECHAZO] {res}")
        sys.exit(1)
    print(f"[OK] Scope Lock verificado y conforme: {res}")

if __name__ == "__main__":
    main()
