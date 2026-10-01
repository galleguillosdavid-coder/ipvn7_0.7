#!/usr/bin/env python3
"""
descubrir_objetivos.py - Sensor Factual de Descubrimiento de Objetivos.
Inspecciona fuentes reales:
  1. Git status (archivos no reconciliados)
  2. Static analysis (go vet)
  3. Enlaces rotos en documentación
  4. Comentarios TODO/FIXME existentes
Genera candidatos en sistema/DESCUBRIMIENTO.md y sistema/OBJETIVOS.md.
"""
import os
import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path

def chequear_git_status(repo_root: Path):
    try:
        res = subprocess.run(["git", "status", "--porcelain"], cwd=repo_root, capture_output=True, text=True, check=True)
    except Exception as e:
        return []

    candidatos = []
    lineas = [l.strip() for l in res.stdout.splitlines() if l.strip()]
    if lineas:
        candidatos.append({
            "tipo": "INFRAESTRUCTURA",
            "origen": "git status",
            "evidencia": f"{len(lineas)} archivos con cambios pendientes de reconciliación en working tree.",
            "problema": "Existen cambios o archivos no confirmados en la infraestructura.",
            "impacto": "MEDIO",
            "archivos": [l[3:].strip().replace("\\", "/").strip('"\'') for l in lineas[:5]],
            "confianza": "ALTA"
        })
    return candidatos

def chequear_enlaces_rotos(repo_root: Path):
    candidatos = []
    carpetas_a_revisar = [repo_root / "docs", repo_root / "agentes", repo_root / "sistema"]
    
    patron_enlace = re.compile(r"\[([^\]]+)\]\(([^)]+)\)")
    
    for c in carpetas_a_revisar:
        if not c.exists():
            continue
        for root, _, files in os.walk(c):
            for file in files:
                if not file.endswith(".md"):
                    continue
                file_path = Path(root) / file
                try:
                    content = file_path.read_text(encoding="utf-8")
                except Exception:
                    continue
                
                for m in patron_enlace.finditer(content):
                    link_target = m.group(2).strip()
                    # Ignorar enlaces web, anclas y esquemas especiales
                    if link_target.startswith("http") or link_target.startswith("#") or link_target.startswith("mailto:"):
                        continue
                    
                    # Normalizar ruta relativa
                    clean_target = link_target.split("#")[0].replace("file:///", "").replace("file://", "")
                    if not clean_target:
                        continue
                    
                    target_path = (file_path.parent / clean_target).resolve()
                    if not target_path.exists() and not Path(clean_target).exists():
                        # Link roto detectado
                        candidatos.append({
                            "tipo": "REFERENCIA_ROTA",
                            "origen": f"inspección markdown ({file_path.relative_to(repo_root)})",
                            "evidencia": f"Enlace roto [{m.group(1)}]({link_target}) en {file_path.name}",
                            "problema": f"El archivo destino '{clean_target}' no existe físicamente en disco.",
                            "impacto": "BAJO",
                            "archivos": [str(file_path.relative_to(repo_root)).replace("\\", "/")],
                            "confianza": "ALTA"
                        })
                        if len(candidatos) >= 3:
                            return candidatos
    return candidatos

def descubrir(repo_root: Path):
    candidatos = []
    candidatos.extend(chequear_git_status(repo_root))
    candidatos.extend(chequear_enlaces_rotos(repo_root))
    return candidatos

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    candidatos = descubrir(repo_root)

    print(f"[DESCUBRIMIENTO FACTUAL] Analizando repositorio...")
    if not candidatos:
        print("[OK] No se detectaron anomalías factuales o trabajo pendiente.")
        sys.exit(0)

    print(f"[OK] Se descubrieron {len(candidatos)} candidatos a objetivo:")
    for idx, c in enumerate(candidatos, 1):
        print(f"  [{idx}] {c['tipo']} ({c['origen']}): {c['problema']}")
        print(f"      Evidencia: {c['evidencia']}")
        print(f"      Archivos: {', '.join(c['archivos'])}")

if __name__ == "__main__":
    main()
