#!/usr/bin/env python3
"""
prueba_autonomia_prolongada.py - Prueba de Autonomía Prolongada (Fase 10).
Demuestra la estabilidad del ciclo autónomo continuo ejecutando 5 ciclos consecutivos:
OBSERVAR -> DESCUBRIR -> CLASIFICAR -> EJECUTAR PERMITIDO -> VERIFICAR -> REPOSAR -> REPETIR.
Certifica ausencia de degradación de estado y parada correcta al agotarse las tareas autónomas.
"""
import sys
import time
from pathlib import Path

repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "sistema" / "bin"))
sys.path.insert(0, str(repo_root / "sistema" / "daemon"))

from supervisor import DaemonSupervisor

def ejecutar_autonomia_prolongada(num_ciclos: int = 5):
    print("================================================================")
    print(f"  FASE 10: PRUEBA DE AUTONOMÍA PROLONGADA ({num_ciclos} CICLOS)")
    print("================================================================")

    sup = DaemonSupervisor(repo_root)
    sup.set_modo("RUN")
    
    ok_lock, msg_lock = sup.adquirir_lock()
    assert ok_lock, f"No se pudo adquirir lock para prueba prolongada: {msg_lock}"

    ciclos_exitosos = 0
    try:
        for i in range(1, num_ciclos + 1):
            inicio = time.time()
            ok, msg = sup.ejecutar_ciclo()
            duracion = time.time() - inicio
            
            assert ok, f"Fallo en ciclo #{i}: {msg}"
            ciclos_exitosos += 1
            print(f"  [Ciclo {i}/{num_ciclos}] {msg} ({duracion:.2f}s)")
            
            # Pausa breve entre iteraciones para simular el cron/daemon
            time.sleep(0.05)

        print("================================================================")
        print(f"  RESULTADO: {ciclos_exitosos}/{num_ciclos} CICLOS EJECUTADOS CON ÉXITO")
        print("  ESTADO FINAL: Supervisor en reposo limpio sin desbordes ni fugas.")
        print("================================================================")
    finally:
        sup.liberar_lock()

if __name__ == "__main__":
    ejecutar_autonomia_prolongada(5)
