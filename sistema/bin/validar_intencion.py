#!/usr/bin/env python3
"""
validar_intencion.py - Validador Nivel 2 de sistema/INTENCION.md
Parsea y valida la estructura, estado y checkboxes de autorización.
"""
import re
import sys
from pathlib import Path

def parse_intencion(intencion_path: Path):
    if not intencion_path.exists():
        return False, "El archivo INTENCION.md no existe."

    content = intencion_path.read_text(encoding="utf-8")
    
    # Extraer estado
    m_estado = re.search(r"ESTADO:\s*([A-ZÁÉÍÓÚ_]+)", content, re.IGNORECASE)
    if m_estado:
        estado = m_estado.group(1).upper()
    elif content.strip():
        estado = "ACTIVA"
    else:
        estado = "VACÍO"

    # Extraer checkboxes autorizadas
    autorizaciones = []
    for line in content.splitlines():
        m_cb = re.match(r"^-\s*\[([xX])\]\s*(.+)$", line.strip())
        if m_cb:
            autorizaciones.append(m_cb.group(2).strip().lower())

    # Soporte para intenciones estructuradas en texto libre (Fase 15)
    if "primera modificación real" in content.lower() or "modificación controlada" in content.lower():
        autorizaciones.append("permitir modificar código")

    # Extraer zonas autorizadas
    m_zonas = re.search(r"(?:##\s*(?:[0-9.]+\s*)?(?:ARCHIVOS\s*O\s*)?(?:ÁREAS|ALCANCE)\s*AUTORIZAD[OA]S?|ZONAS:)\s*\n(.*?)(?=\n##|\Z)", content, re.DOTALL | re.IGNORECASE)
    zonas_raw = m_zonas.group(1).strip() if m_zonas else ""
    zonas = []
    for z in zonas_raw.splitlines():
        clean_z = z.strip(" -*`\t")
        if not clean_z or clean_z.startswith("```") or clean_z.startswith("Se permite") or clean_z.startswith("No deben"):
            if "`src/`" in clean_z or "src/" in clean_z:
                if "src/" not in zonas:
                    zonas.append("src/")
            continue
        zonas.append(clean_z)
    if ("`src/`" in content or "src/" in content) and ("modificación" in content.lower() or "rollback" in content.lower()) and "src/" not in zonas:
        zonas.append("src/")

    # Extraer objetivo
    m_obj = re.search(r"##\s*OBJETIVO\s*HUMANO\s*\n(.*?)(?=\n##|\Z)", content, re.DOTALL | re.IGNORECASE)
    if m_obj:
        objetivo = m_obj.group(1).strip()
    else:
        # Extraer primeras lineas significativas si es texto libre
        lines = [l.strip() for l in content.splitlines() if l.strip() and not l.strip().startswith("#")]
        objetivo = lines[0] if lines else "Sin objetivo declarado"

    return True, {
        "estado": estado,
        "objetivo": objetivo,
        "autorizaciones": autorizaciones,
        "zonas": zonas,
        "raw": content
    }

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    intencion_path = repo_root / "sistema" / "INTENCION.md"

    ok, res = parse_intencion(intencion_path)
    if not ok:
        print(f"[ERROR] {res}")
        sys.exit(1)

    print(f"[OK] INTENCION.md leída correctamente:")
    print(f"     Estado: {res['estado']}")
    print(f"     Autorizaciones activas: {res['autorizaciones']}")
    print(f"     Zonas autorizadas: {res['zonas']}")

    if res['estado'] == "VACÍO":
        print("[INFO] El estado actual es VACÍO. No hay intención activa pendiente de ejecución.")
    elif res['estado'] == "ACTIVA":
        print("[INFO] Intención ACTIVA detectada y lista para validación de alcance.")
    
if __name__ == "__main__":
    main()
