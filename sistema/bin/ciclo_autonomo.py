#!/usr/bin/env python3
"""
ciclo_autonomo.py - Orquestador del Ciclo Autónomo Controlado.
1. Ejecuta descubrimiento y evaluación contra AUTONOMIA.json.
2. Actualiza sistema/DESCUBRIMIENTO.md y sistema/OBJETIVOS.md.
3. Respeta límites de ciclo y prohibición estricta de auto-escalamiento.
"""
import json
import sys
from datetime import datetime
from pathlib import Path
from evaluar_autonomia import evaluar_todos

def sincronizar_documentos(repo_root: Path, clasificados: list):
    desc_path = repo_root / "sistema" / "DESCUBRIMIENTO.md"
    obj_path = repo_root / "sistema" / "OBJETIVOS.md"

    # Actualizar DESCUBRIMIENTO.md
    filas = []
    for c in clasificados:
        filas.append(f"| {c['id']} | {c['tipo']} | {c['problema'][:45]}... | {c['autonomia']} | PENDIENTE |")

    tabla = "\n".join(filas) if filas else "| *Sin objetivos descubiertos.* | - | - | - | - |"
    
    desc_content = f"""# REGISTRO DE DESCUBRIMIENTO FACTUAL (sistema/DESCUBRIMIENTO.md)
> Actualizado automáticamente: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}  
> Principio: Todo objetivo está sustentado en evidencia observable.

---

## Índice de Objetivos Descubiertos

| ID | Tipo de Hecho | Problema Resumido | Autonomía | Estado |
| :--- | :--- | :--- | :---: | :---: |
{tabla}

---

## Detalle Factual
"""
    for c in clasificados:
        desc_content += f"""
### [{c['id']}] {c['tipo']}
- **Origen:** {c['problema']}
- **Evidencia Observable:** {c['evidencia']}
- **Archivos:** {', '.join(c['archivos'])}
- **Autonomía:** {c['autonomia']} ({c['razon']})
- **Estado:** PENDIENTE
"""
    desc_path.write_text(desc_content, encoding="utf-8")

    # Actualizar OBJETIVOS.md
    filas_obj = []
    for c in clasificados:
        permiso_str = "AUTORIZADO" if c['autonomia'] == "AUTÓNOMO" else "REQUIERE_HUMANO"
        filas_obj.append(f"| {c['id']} | {c['tipo']} | {c['problema'][:50]}... | {', '.join(c['archivos'])} | {permiso_str} | PENDIENTE |")

    tabla_obj = "\n".join(filas_obj) if filas_obj else "| *Sin objetivos activos.* | - | - | - | - | - |"
    
    obj_content = f"""# BACKLOG DE OBJETIVOS FACTUALES (sistema/OBJETIVOS.md)
> Actualizado automáticamente: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

---

## 1. Objetivos Activos y Priorizados

| ID | Tipo de Hecho | Descripción | Archivos | Presupuesto | Estado |
| :--- | :--- | :--- | :--- | :---: | :---: |
{tabla_obj}

---

## 2. Regla de Autonomía Inmutable
Los objetivos marcados como `REQUIERE_HUMANO` jamás pueden ser ejecutados sin autorización explícita en `sistema/INTENCION.md`.
"""
    obj_path.write_text(obj_content, encoding="utf-8")

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    clasificados = evaluar_todos(repo_root)

    print("[CICLO AUTÓNOMO] Descubriendo y evaluando objetivos...")
    sincronizar_documentos(repo_root, clasificados)
    print(f"[OK] {len(clasificados)} objetivos descubiertos y sincronizados en sistema/OBJETIVOS.md y sistema/DESCUBRIMIENTO.md")

    autonomos = [c for c in clasificados if c["autonomia"] == "AUTÓNOMO"]
    requieren_humano = [c for c in clasificados if c["autonomia"] == "REQUIERE_HUMANO"]

    print(f"     -> Autónomos preautorizados: {len(autonomos)}")
    print(f"     -> Requieren aprobación humana: {len(requieren_humano)}")

if __name__ == "__main__":
    main()
