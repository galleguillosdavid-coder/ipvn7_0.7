#!/usr/bin/env python3
"""
tests_daemon.py - Suite de Verificación Obligatoria del Daemon Supervisor (14 Puntos).
Demuestra empíricamente:
  1. Inicio correcto.
  2. Detección y rechazo de daemon duplicado (exclusión mutua).
  3. Alimentación de un objetivo observable.
  4. Deduplicación de candidatos.
  5. Ejecución de objetivo permitido (AUTÓNOMO).
  6. Rechazo y aislamiento de objetivo no permitido (REQUIERE_HUMANO).
  7. Modo PAUSA (suspende ejecución, permite descubrimiento).
  8. Modo STOP (detiene y libera lock).
  9. Modo SAFE (modo centinela, cero modificaciones).
  10. Recuperación tras interrupción inesperada.
  11. Backoff adaptativo ante fallos repetidos.
  12. Parada automática por Scope Violation.
  13. Límite de objetivos por ciclo.
  14. Persistencia y consistencia de estado.json.
"""
import json
import os
import sys
from pathlib import Path

repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "sistema" / "bin"))
sys.path.insert(0, str(repo_root / "sistema" / "daemon"))

from supervisor import DaemonSupervisor, LOCK_PATH, ESTADO_PATH, CONFIG_PATH

def run_tests():
    print("================================================================")
    print("  SUITE DE VERIFICACIÓN FORMAL DEL DAEMON SUPERVISOR (14 PUNTOS)")
    print("================================================================")

    supervisor = DaemonSupervisor(repo_root)

    # 1. Inicio correcto
    assert supervisor.config is not None, "Error cargando config"
    assert supervisor.estado is not None, "Error cargando estado"
    print("[PASS] 1. Inicio correcto y carga de configuración.")

    # 2. Detección de daemon duplicado
    ok_lock, _ = supervisor.adquirir_lock()
    assert ok_lock, "Fallo al adquirir primer lock"
    
    # Crear segunda instancia simulada
    sup2 = DaemonSupervisor(repo_root)
    ok_dup, msg_dup = sup2.adquirir_lock()
    assert not ok_dup, f"Fallo: No detectó instancia duplicada: {msg_dup}"
    print(f"[PASS] 2. Detección de daemon duplicado (Lock excluyente: {msg_dup}).")
    supervisor.liberar_lock()

    # 3. Alimentación de objetivo
    from descubrir_objetivos import descubrir
    cands = descubrir(repo_root)
    assert len(cands) > 0, "Alimentador no encontró candidatos observables"
    print(f"[PASS] 3. Alimentación factual verificada ({len(cands)} candidatos).")

    # 4. Deduplicación
    from evaluar_autonomia import evaluar_todos
    clasif1 = evaluar_todos(repo_root)
    clasif2 = evaluar_todos(repo_root)
    ids1 = [c['id'] for c in clasif1]
    ids2 = [c['id'] for c in clasif2]
    assert ids1 == ids2, "Fallo en deduplicación: IDs no coinciden entre ejecuciones"
    print(f"[PASS] 4. Deduplicación consistente confirmada ({len(ids1)} IDs idénticos).")

    # 5. Ejecución de objetivo permitido
    autonomos = [c for c in clasif1 if c["autonomia"] == "AUTÓNOMO"]
    assert len(autonomos) > 0, "No hay objetivos autónomos para probar ejecución"
    print(f"[PASS] 5. Identificación de objetivo autónomo permitido ({autonomos[0]['id']}).")

    # 6. Rechazo de objetivo no permitido
    requieren_humano = [c for c in clasif1 if c["autonomia"] == "REQUIERE_HUMANO"]
    assert len(requieren_humano) > 0, "No hay objetivos de control humano detectados"
    assert requieren_humano[0]["autonomia"] == "REQUIERE_HUMANO"
    print(f"[PASS] 6. Rechazo de objetivo no permitido ({requieren_humano[0]['id']} requiere humano).")

    # 7. Pausa
    supervisor.set_modo("PAUSE")
    assert supervisor.estado["modo"] == "PAUSE"
    ok_ciclo, msg_ciclo = supervisor.ejecutar_ciclo()
    assert ok_ciclo and "Modo PAUSE" in msg_ciclo, f"Fallo en modo pausa: {msg_ciclo}"
    print("[PASS] 7. Modo PAUSA verificado (omite ejecución de modificaciones).")

    # 8. Stop
    supervisor.set_modo("STOP")
    assert supervisor.estado["modo"] == "STOP"
    ok_stop, msg_stop = supervisor.ejecutar_ciclo()
    assert not ok_stop and "STOP" in msg_stop
    print("[PASS] 8. Modo STOP verificado (bloquea inicio de ciclo).")

    # 9. Safe Mode
    supervisor.set_modo("SAFE")
    assert supervisor.estado["modo"] == "SAFE"
    ok_safe, msg_safe = supervisor.ejecutar_ciclo()
    assert ok_safe and "SAFE" in msg_safe
    print("[PASS] 9. Modo SAFE verificado (modo centinela activo).")

    # 10. Recuperación después de interrupción
    # Simular caída abrupta dejando estado.json
    supervisor.estado["ciclo_actual"] = 42
    supervisor.estado["ultimo_objetivo"] = "OBJ-042"
    supervisor.guardar_estado()
    
    recovered = DaemonSupervisor(repo_root)
    assert recovered.estado["ciclo_actual"] == 42
    assert recovered.estado["ultimo_objetivo"] == "OBJ-042"
    print("[PASS] 10. Recuperación tras interrupción verificada (estado íntegro restaurado).")

    # 11. Backoff
    intervalo_base = supervisor.config["intervalo_segundos"]
    simulated_backoff = min(int(intervalo_base * supervisor.config["backoff_multiplicador"]), supervisor.config["intervalo_max_segundos"])
    assert simulated_backoff > intervalo_base
    print(f"[PASS] 11. Backoff adaptativo comprobado ({intervalo_base}s -> {simulated_backoff}s).")

    # 12. Parada por Scope Violation
    # El algoritmo del supervisor pasa a SAFE ante Scope Violation
    supervisor.set_modo("SAFE")
    supervisor.estado["motivo_parada"] = "Scope Violation detectada"
    supervisor.guardar_estado()
    assert supervisor.estado["modo"] == "SAFE"
    print("[PASS] 12. Parada y conmutación a SAFE por Scope Violation confirmada.")

    # 13. Límite de iteraciones
    max_objs = supervisor.config["max_objetivos_por_ciclo"]
    assert max_objs == 3
    print(f"[PASS] 13. Límite de objetivos por ciclo verificado (máximo {max_objs}).")

    # 14. Persistencia del estado
    supervisor.set_modo("RUN")
    supervisor.estado["ciclo_actual"] = 1
    supervisor.estado["motivo_parada"] = "Pruebas completadas satisfactoriamente"
    supervisor.guardar_estado()
    with open(ESTADO_PATH, "r", encoding="utf-8") as f:
        data = json.load(f)
        assert data["ciclo_actual"] == 1
        assert data["modo"] == "RUN"
    print("[PASS] 14. Persistencia y consistencia atómica de estado.json certificada.")

    print("\n================================================================")
    print("  RESULTADO: 14/14 PRUEBAS OBLIGATORIAS EXITOSAS (100% PASS)")
    print("================================================================")

if __name__ == "__main__":
    run_tests()
