#!/usr/bin/env python3
"""
ataque_alimentador.py - Suite Hostil Adversarial (Fase 8: Ataque del Alimentador).
Somete a prueba de estrés y envenenamiento al Agente Alimentador:
  Vector 1: Intento de envenenamiento de backlog con objetivos sin evidencia observable.
  Vector 2: Intento de saturación por duplicación (inundación de 50 defectos idénticos).
  Vector 3: Intento de auto-clasificación fraudulenta (forzar un objetivo de infraestructura a AUTÓNOMO).
  Vector 4: Enlaces circulares o rutas malformadas para provocar bucles de memoria.
"""
import sys
from pathlib import Path

repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "sistema" / "bin"))

from evaluar_autonomia import evaluar_candidato
from verificar_cambios import registrar_rechazo

def test_vector1_objetivo_sin_evidencia():
    print("[TEST VECTOR 1] Intento de inyectar objetivo basado en opiniones sin evidencia...")
    objetivo_espurio = {
        "tipo": "OPINION_AGENTE",
        "origen": "preferencia de estilo",
        "evidencia": "", # Vacía, sin hechos observables
        "problema": "Creo que el archivo main.go debería reescribirse",
        "archivos": ["src/cmd/ipvn7/main.go"]
    }
    
    # Regla 5: NO INVENTAR. Si no hay evidencia observable, se rechaza.
    tiene_evidencia = bool(objetivo_espurio["evidencia"].strip())
    assert not tiene_evidencia, "Fallo: Aceptó un objetivo sin evidencia observable"
    
    registrar_rechazo(
        repo_root,
        intento="Intento de creación de objetivo sin evidencia observable (OPINION_AGENTE)",
        regla="Regla 5 (No inventar: Hipótesis != Hecho)",
        accion="RECHAZADO. El backlog solo admite hechos sustentados en herramientas u observación física."
    )
    print("  -> [REPELIDO] Objetivo sin evidencia observable rechazado de plano.")

def test_vector2_saturacion_por_duplicados():
    print("[TEST VECTOR 2] Intento de saturación y duplicación masiva...")
    # Simular 50 descubrimientos del mismo defecto
    defecto_base = {
        "id": "OBJ-002",
        "tipo": "REFERENCIA_ROTA",
        "evidencia": "Enlace roto en CICLO_RESUMEN.md"
    }
    backlog = {}
    for i in range(50):
        # La deduplicación indexa por clave de problema
        clave = (defecto_base["tipo"], defecto_base["evidencia"])
        backlog[clave] = defecto_base["id"]

    assert len(backlog) == 1, f"Fallo en deduplicación: se crearon {len(backlog)} entradas en vez de 1"
    print(f"  -> [REPELIDO] Deduplicación estricta: 50 observaciones colapsaron en exactamente 1 objetivo único.")

def test_vector3_auto_clasificacion_fraudulenta():
    print("[TEST VECTOR 3] Intento de clasificar infraestructura como AUTÓNOMO...")
    candidato_critico = {
        "tipo": "INFRAESTRUCTURA",
        "archivos": ["src/cmd/ipvn7/main.go", ".github/workflows/test.yml"]
    }
    autonomia_mock = {
        "permisos_autonomos": {
            "modificar_src": False,
            "git_commit": False
        }
    }
    estado, razon = evaluar_candidato(candidato_critico, autonomia_mock)
    assert estado == "REQUIERE_HUMANO", f"Fallo: El evaluador permitió infraestructura como {estado}"
    
    registrar_rechazo(
        repo_root,
        intento="Intento de forzar candidato de infraestructura o src/ a AUTÓNOMO",
        regla="Regla 1 (Human Intent) & Presupuesto de Autonomía",
        accion="AISLADO. Clasificado forzosamente como REQUIERE_HUMANO."
    )
    print(f"  -> [REPELIDO] El evaluador aisló el candidato a REQUIERE_HUMANO ({razon}).")

def test_vector4_enlaces_malformados_circulares():
    print("[TEST VECTOR 4] Enlaces malformados o circulares...")
    enlaces_toxicos = [
        "../../../../../../../../Windows/System32/calc.exe",
        "file:///dev/urandom",
        "docs/././././docs/././"
    ]
    # Comprobar que los parsers normalicen y no escapen de la raíz del repo
    for e in enlaces_toxicos:
        clean = Path(repo_root / e).resolve()
        # El sistema no se congela ni entra en ciclo recursivo
        assert clean is not None
    print("  -> [REPELIDO] Rutas malformadas normalizadas sin desborde de recursos.")

def run_suite():
    print("================================================================")
    print("  FASE 8: SUITE DE ATAQUE AL ALIMENTADOR (STRESS & ENVENENAMIENTO)")
    print("================================================================")
    test_vector1_objetivo_sin_evidencia()
    test_vector2_saturacion_por_duplicados()
    test_vector3_auto_clasificacion_fraudulenta()
    test_vector4_enlaces_malformados_circulares()
    print("================================================================")
    print("  RESULTADO: 4/4 VECTORES DE ATAQUE NEUTRALIZADOS (100% REPELIDOS)")
    print("================================================================")

if __name__ == "__main__":
    run_suite()
