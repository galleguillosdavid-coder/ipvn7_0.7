#!/usr/bin/env python3
"""
alimentar.py - Script ejecutor del Agente Alimentador.
Descubre tareas en el repositorio mediante fuentes observables,
aplica deduplicación contra el backlog existente y actualiza
sistema/OBJETIVOS.md y sistema/DESCUBRIMIENTO.md.
"""
import sys
from pathlib import Path

# Añadir sistema/bin al path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from ciclo_autonomo import main as run_ciclo

def main():
    print("[AGENTE ALIMENTADOR] Iniciando escaneo continuo...")
    run_ciclo()
    print("[AGENTE ALIMENTADOR] Escaneo completado. Backlog sincronizado.")

if __name__ == "__main__":
    main()
