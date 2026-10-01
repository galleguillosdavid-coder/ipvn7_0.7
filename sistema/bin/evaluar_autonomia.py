#!/usr/bin/env python3
"""
evaluar_autonomia.py - Clasificador de Autonomía Controlada.
Evalúa objetivos descubiertos contra el presupuesto inmutable de sistema/AUTONOMIA.json.
Determina si un objetivo es AUTÓNOMO o REQUIERE_HUMANO.
"""
import json
import sys
from pathlib import Path
from descubrir_objetivos import descubrir

def evaluar_candidato(candidato: dict, autonomia_cfg: dict):
    permisos = autonomia_cfg.get("permisos_autonomos", {})
    tipo = candidato.get("tipo", "")
    archivos = candidato.get("archivos", [])

    # 1. Reglas duras de protección (NUNCA AUTÓNOMO)
    toca_src = any("src/" in a.lower() for a in archivos)
    if toca_src and not permisos.get("modificar_src", False):
        return "REQUIERE_HUMANO", "Afecta a 'src/' y modificar_src está deshabilitado en AUTONOMIA.json"

    if tipo == "INFRAESTRUCTURA" and not permisos.get("git_commit", False):
        return "REQUIERE_HUMANO", "Requiere acciones sobre Git (commit/staging) no autorizadas de forma autónoma"

    # 2. Comprobación de permisos autónomos concedidos
    if tipo == "REFERENCIA_ROTA" and permisos.get("corregir_documentacion", False):
        return "AUTÓNOMO", "Corrección de enlaces en documentación está autorizada en AUTONOMIA.json"

    if tipo == "TEST_FALTANTE" and permisos.get("corregir_tests", False):
        return "AUTÓNOMO", "Corrección de tests está autorizada en AUTONOMIA.json"

    return "REQUIERE_HUMANO", "La acción excede el presupuesto de autonomía preautorizado."

def evaluar_todos(repo_root: Path):
    autonomia_path = repo_root / "sistema" / "AUTONOMIA.json"
    with open(autonomia_path, "r", encoding="utf-8") as f:
        autonomia_cfg = json.load(f)

    candidatos = descubrir(repo_root)
    clasificados = []

    for idx, c in enumerate(candidatos, 1):
        estado_autonomia, razon = evaluar_candidato(c, autonomia_cfg)
        obj_id = f"OBJ-{idx:03d}"
        clasificados.append({
            "id": obj_id,
            "tipo": c["tipo"],
            "problema": c["problema"],
            "evidencia": c["evidencia"],
            "archivos": c["archivos"],
            "autonomia": estado_autonomia,
            "razon": razon
        })

    return clasificados

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    clasificados = evaluar_todos(repo_root)

    print("==================================================")
    print("  EVALUADOR DE AUTONOMÍA IPVN7 (sistema/AUTONOMIA.json)")
    print("==================================================")
    for c in clasificados:
        icono = "[AUTÓNOMO]" if c["autonomia"] == "AUTÓNOMO" else "[REQUIERE_HUMANO]"
        print(f"{c['id']} {icono} {c['tipo']}")
        print(f"     Problema: {c['problema']}")
        print(f"     Razón:    {c['razon']}")
        print()

if __name__ == "__main__":
    main()
