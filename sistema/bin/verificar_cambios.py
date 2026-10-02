#!/usr/bin/env python3
"""
verificar_cambios.py - Sensor Git Nivel 3.
Compara los archivos modificados en Git contra las zonas autorizadas en INTENCION.md.
Ante cualquier archivo fuera de scope: aborta, escribe en RECHAZOS.md y sale con código 1.
"""
import subprocess
import sys
from datetime import datetime
from pathlib import Path
from validar_intencion import parse_intencion

def obtener_archivos_modificados(repo_root: Path):
    cmd = ["git", "status", "--porcelain"]
    try:
        res = subprocess.run(cmd, cwd=repo_root, capture_output=True, text=True, check=True)
    except Exception as e:
        return []

    archivos = []
    for line in res.stdout.splitlines():
        if not line.strip():
            continue
        # formato: XY ruta o R  origen -> destino
        parts = line[3:].strip()
        if "->" in parts:
            parts = parts.split("->")[-1].strip()
        parts = parts.strip('"\'')
        archivos.append(parts.replace("\\", "/"))
    return archivos

def registrar_rechazo(repo_root: Path, intento: str, regla: str, accion: str):
    rechazos_path = repo_root / "sistema" / "RECHAZOS.md"
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    entry = f"| {timestamp} | Sensor Git (Nivel 3) | {intento} | {regla} | {accion} |\n"
    
    if rechazos_path.exists():
        content = rechazos_path.read_text(encoding="utf-8")
        if not content.endswith("\n"):
            content += "\n"
        content += entry
        rechazos_path.write_text(content, encoding="utf-8")

def verificar_scope_git(repo_root: Path):
    intencion_path = repo_root / "sistema" / "INTENCION.md"
    ok, intencion = parse_intencion(intencion_path)
    if not ok:
        return False, f"Error leyendo intención: {intencion}"

    zonas = [z.strip("/").lower() for z in intencion.get("zonas", [])]
    modificados = obtener_archivos_modificados(repo_root)

    # Zonas de control e infraestructura base del repositorio
    permitidos_base = [
        "sistema/", "agentes/", "docs/",
        ".github/", ".vscode/", "scripts/", "frondabrick_01/", "config/",
        ".gitignore", "readme.md"
    ]

    import fnmatch
    import json

    # Cargar política de src si existe
    politica_src_path = repo_root / "frondabrick_01" / "reglas" / "politica_src.json"
    if not politica_src_path.exists():
        politica_src_path = repo_root / "sistema" / "reglas" / "politica_src.json"

    bloqueados_src = []
    if politica_src_path.exists():
        try:
            with open(politica_src_path, "r", encoding="utf-8") as pf:
                pol_data = json.load(pf)
                bloqueados_src = pol_data.get("zonas_estrictamente_bloqueadas", [])
        except Exception:
            pass

    # Cargar archivos aprobados formalmente en CAMBIOS.md
    archivos_aprobados_cambios = []
    cambios_path = repo_root / "sistema" / "CAMBIOS.md"
    if cambios_path.exists():
        try:
            content_c = cambios_path.read_text(encoding="utf-8")
            import re
            for m in re.finditer(r"(src/[\w./\-]+)", content_c):
                archivos_aprobados_cambios.append(m.group(1).lower())
        except Exception:
            pass

    # Archivos de gobernanza estrictamente inmutables para el loop autónomo
    inmutables_gobierno = [
        "sistema/constitucion.md",
        "sistema/autonomia.json",
        "frondabrick_01/reglas/politica_src.json",
        "frondabrick_01/reglas/permisos.json",
        "frondabrick_01/reglas/rutas.json"
    ]

    violaciones = []
    for f in modificados:
        f_norm = f.lower()

        # 1. Comprobar inmutabilidad constitucional
        if any(f_norm == inm.lower() for inm in inmutables_gobierno):
            violaciones.append(f"{f} [ARCHIVO DE GOBERNANZA INMUTABLE]")
            continue

        # 2. Comprobar zonas estrictamente bloqueadas en src/
        if f_norm.startswith("src/"):
            bloqueado = False
            for pat in bloqueados_src:
                if fnmatch.fnmatch(f_norm, pat.lower()):
                    violaciones.append(f"{f} [ZONA ESTRICTAMENTE BLOQUEADA EN POLITICA_SRC]")
                    bloqueado = True
                    break
            if bloqueado:
                continue

        # 3. Comprobar si pertenece a alguna zona autorizada o cambio formal aprobado en CAMBIOS.md
        autorizado = (
            any(f_norm.startswith(z) for z in zonas)
            or any(f_norm.startswith(p) for p in permitidos_base)
            or any(f_norm == apr for apr in archivos_aprobados_cambios)
        )
        if not autorizado:
            violaciones.append(f"{f} [FUERA DE ALCANCE]")

    if violaciones:
        detalle = f"Modificaciones no autorizadas detectadas en: {', '.join(violaciones)}"
        registrar_rechazo(
            repo_root,
            intento=detalle,
            regla="Regla 4 (Scope Lock) & Regla 1 (Constitución) & Política src/",
            accion="ABORTADO. Sensor Git Nivel 3 bloqueó la ejecución."
        )
        return False, detalle

    return True, f"Verificación Git limpia. {len(modificados)} archivos dentro de scope."

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    ok, msg = verificar_scope_git(repo_root)
    if not ok:
        print(f"[ERROR CRÍTICO - SCOPE VIOLATION] {msg}")
        sys.exit(1)
    print(f"[OK - SENSOR GIT] {msg}")

if __name__ == "__main__":
    main()
