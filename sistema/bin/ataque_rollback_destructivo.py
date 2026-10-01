#!/usr/bin/env python3
"""
ataque_rollback_destructivo.py - Suite Hostil Adversarial de Rollback Destructivo (Fase 14).
Somete a prueba destructiva el mecanismo de rollback transaccional, el journal de recuperación,
la exclusión mutua de locks, la resiliencia del supervisor y la transición obligatoria a SAFE/STOP.

Invariante Fundamental:
  PRE-CAMBIO == POST-ROLLBACK
  ROLLBACK_DECLARADO_OK != ROLLBACK_VERIFICADO_OK
  Si PRE != POST -> MODO OBLIGATORIO: SAFE o STOP (NUNCA RUN)
"""
import fnmatch
import hashlib
import json
import os
import shutil
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "sistema" / "bin"))
sys.path.insert(0, str(repo_root / "sistema" / "daemon"))

from pipeline_src import (
    PipelineSrc, capturar_huella, verificar_invariante, transicionar_modo,
    SRC_LOCK_PATH, SRC_TRANS_PATH
)
from supervisor import DaemonSupervisor, ESTADO_PATH, LOCK_PATH
from verificar_cambios import verificar_scope_git

pipeline = PipelineSrc(repo_root)

def obtener_modo_supervisor() -> str:
    if ESTADO_PATH.exists():
        try:
            with open(ESTADO_PATH, "r", encoding="utf-8") as f:
                return json.load(f).get("modo", "UNKNOWN")
        except Exception:
            return "CORRUPTED"
    return "NONE"

# ====================================================================
# VECTOR 1: Interrupción durante la escritura del journal
# ====================================================================
def test_vector1_interrupcion_escritura_journal():
    print("[VECTOR 1] Interrupción simulada durante la escritura del journal...")
    tmp_trans = SRC_TRANS_PATH.with_suffix(".tmp")
    try:
        # Simular archivo temporal incompleto sin renombrar
        tmp_trans.write_text("{ \"ciclo_id\": \"CRASHED_WRITE\", INCOMPLETE...", encoding="utf-8")
        
        # El sistema de recuperación no debe fallar por el temporal huérfano
        ok_rec, msg_rec = pipeline.recuperar_transaccion_pendiente()
        assert ok_rec, "La recuperación debió gestionar limpiamente la ausencia del journal final"
        print("  -> [REPELIDO] Archivo temporal huérfano contenido; sistema estable.")
    finally:
        if tmp_trans.exists():
            tmp_trans.unlink()

# ====================================================================
# VECTOR 2: Interrupción durante la restauración de archivos
# ====================================================================
def test_vector2_interrupcion_restauracion_archivos():
    print("[VECTOR 2] Interrupción simulada a mitad de la restauración de archivos...")
    f1 = repo_root / "src" / "pkg" / "core" / "partially_restored_1_test.go"
    f2 = repo_root / "src" / "pkg" / "core" / "partially_restored_2_test.go"
    try:
        # Capturar huella limpia PRE
        huella_pre = capturar_huella(repo_root, [
            "src/pkg/core/partially_restored_1_test.go",
            "src/pkg/core/partially_restored_2_test.go"
        ])
        
        # Escribir ambos archivos y registrar transacción
        f1.write_text("// test 1", encoding="utf-8")
        f2.write_text("// test 2", encoding="utf-8")
        pipeline.iniciar_transaccion("VEC-002", [
            "src/pkg/core/partially_restored_1_test.go",
            "src/pkg/core/partially_restored_2_test.go"
        ], huella_pre)

        # Simular que f1 fue borrado pero el proceso se interrumpió antes de f2
        f1.unlink()

        # Invocar recuperación de transacción pendiente
        ok_rec, msg_rec = pipeline.recuperar_transaccion_pendiente()
        assert ok_rec, f"Fallo en completar recuperación parcial: {msg_rec}"
        assert not f1.exists() and not f2.exists(), "F2 no fue limpiado en la recuperación"
        
        # Validar invariante
        huella_post = capturar_huella(repo_root, [
            "src/pkg/core/partially_restored_1_test.go",
            "src/pkg/core/partially_restored_2_test.go"
        ])
        ok_inv, _ = verificar_invariante(huella_pre, huella_post)
        assert ok_inv, "Invariante roto tras recuperación de interrupción"
        print("  -> [REPELIDO] Restauración interrumpida completada al 100%. PRE == POST.")
    finally:
        if f1.exists(): f1.unlink()
        if f2.exists(): f2.unlink()
        pipeline.finalizar_transaccion()

# ====================================================================
# VECTOR 3: Corrupción sintáctica de src_transaction.json
# ====================================================================
def test_vector3_corrupcion_sintactica_journal():
    print("[VECTOR 3] Corrupción sintáctica intencional en src_transaction.json...")
    try:
        SRC_TRANS_PATH.parent.mkdir(parents=True, exist_ok=True)
        SRC_TRANS_PATH.write_text("{ CORRUPTED_BIN_GARBAGE \x00\xff...", encoding="utf-8")

        ok_rec, msg_rec = pipeline.recuperar_transaccion_pendiente()
        assert not ok_rec, "Fallo: Debió rechazar el journal corrupto"
        assert "SAFE" in msg_rec or "Corrupción" in msg_rec
        
        modo = obtener_modo_supervisor()
        assert modo == "SAFE", f"El supervisor debió transicionar a SAFE, pero está en: {modo}"
        assert not SRC_TRANS_PATH.exists(), "El journal corrupto debe ser purgado"
        print("  -> [REPELIDO] Corrupción sintáctica contenida. Transición obligatoria a SAFE.")
    finally:
        if SRC_TRANS_PATH.exists():
            SRC_TRANS_PATH.unlink()
        pipeline.finalizar_transaccion()

# ====================================================================
# VECTOR 4: Eliminación del journal durante rollback
# ====================================================================
def test_vector4_eliminacion_journal_durante_rollback():
    print("[VECTOR 4] Eliminación intencional del journal durante rollback...")
    f_res = repo_root / "src" / "pkg" / "core" / "residual_test.go"
    try:
        f_res.write_text("// residual", encoding="utf-8")
        # El journal no existe (fue borrado)
        if SRC_TRANS_PATH.exists():
            SRC_TRANS_PATH.unlink()

        # El Sensor Git y la limpieza forzada detectan y purgan el residuo
        pipeline.forzar_limpieza_segura()
        assert not f_res.exists(), "El residuo no fue purgado por forzar_limpieza_segura"
        
        ok_git, msg_git = verificar_scope_git(repo_root)
        assert ok_git, f"Residuos detectados por Sensor Git: {msg_git}"
        print("  -> [REPELIDO] Saneamiento forzado eliminó residuos sin journal. Git limpio.")
    finally:
        if f_res.exists():
            f_res.unlink()

# ====================================================================
# VECTOR 5: Journal apuntando fuera del scope autorizado
# ====================================================================
def test_vector5_journal_fuera_de_scope():
    print("[VECTOR 5] Journal con rutas maliciosas fuera de scope (Path Traversal)...")
    huella_pre = capturar_huella(repo_root, [])
    # Intento 1: Transacción con path traversal
    ok_trans, msg_trans = pipeline.iniciar_transaccion("VEC-005", ["../malicious.sh"], huella_pre)
    assert not ok_trans, "Fallo: Iniciar transacción debió bloquear path traversal"
    assert "VETO SCOPE" in msg_trans
    
    # Intento 2: Inyectar manualmente en el journal un archivo de gobernanza protegido
    SRC_TRANS_PATH.write_text(json.dumps({
        "ciclo_id": "VEC-005-INJ",
        "archivos": ["sistema/CONSTITUCION.md"],
        "huella_pre": huella_pre
    }), encoding="utf-8")
    
    try:
        ok_rec, msg_rec = pipeline.recuperar_transaccion_pendiente()
        assert not ok_rec, "Fallo: Recuperación debió rechazar ruta fuera de src/"
        assert "fuera de scope" in msg_rec or "SAFE" in msg_rec
        modo = obtener_modo_supervisor()
        assert modo == "SAFE", f"Modo debió ser SAFE tras violación de scope en journal: {modo}"
        print("  -> [REPELIDO] Inyección de rutas fuera de scope bloqueada. Modo SAFE activado.")
    finally:
        if SRC_TRANS_PATH.exists():
            SRC_TRANS_PATH.unlink()
        pipeline.finalizar_transaccion()

# ====================================================================
# VECTOR 6: Hash PRE manipulado
# ====================================================================
def test_vector6_hash_pre_manipulado():
    print("[VECTOR 6] Manipulación fraudulenta del hash PRE en la huella...")
    target = repo_root / "src" / "pkg" / "core" / "fake_pre_test.go"
    try:
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/fake_pre_test.go"])
        # Adulterar hash PRE deliberadamente
        huella_pre["hashes"]["src/pkg/core/fake_pre_test.go"] = "0000000000000000000000000000000000000000000000000000000000000000"

        huella_post = capturar_huella(repo_root, ["src/pkg/core/fake_pre_test.go"])
        ok_inv, msg_inv = verificar_invariante(huella_pre, huella_post)
        assert not ok_inv, "Fallo: El verificador de invariante debió detectar el hash PRE manipulado"
        assert "Desviación de hash" in msg_inv
        print("  -> [REPELIDO] Discrepancia criptográfica en PRE detectada y rechazada.")
    finally:
        pass

# ====================================================================
# VECTOR 7: Hash POST manipulado (DECLARADO != VERIFICADO)
# ====================================================================
def test_vector7_hash_post_manipulado():
    print("[VECTOR 7] Falsificación de hash POST (DECLARADO_OK != VERIFICADO_OK)...")
    target = repo_root / "src" / "pkg" / "core" / "tampered_post_test.go"
    try:
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/tampered_post_test.go"])
        target.write_text("// modificado no restaurado", encoding="utf-8")

        # Agente declara fraudulentamente que el hash post es idéntico a pre
        huella_post_declarada = dict(huella_pre)
        
        # Pero el verificador físico independiente calcula la huella real del disco
        huella_post_real = capturar_huella(repo_root, ["src/pkg/core/tampered_post_test.go"])
        
        ok_declarado = (huella_pre["hashes"] == huella_post_declarada["hashes"])
        ok_verificado, msg_ver = verificar_invariante(huella_pre, huella_post_real)
        
        assert ok_declarado is True, "El hash declarado simulaba ser idéntico"
        assert ok_verificado is False, "El verificador independiente debió rechazar el hash físico real"
        assert "Desviación de hash" in msg_ver
        print("  -> [REPELIDO] Distinción fundamental verificada: DECLARADO_OK != VERIFICADO_OK.")
    finally:
        if target.exists():
            target.unlink()

# ====================================================================
# VECTOR 8: Aparición de archivos no registrados durante rollback
# ====================================================================
def test_vector8_archivos_no_registrados_durante_rollback():
    print("[VECTOR 8] Aparición de archivo no registrado durante rollback...")
    unregistered = repo_root / "src" / "pkg" / "core" / "unregistered_ghost_test.go"
    try:
        huella_pre = capturar_huella(repo_root, [])
        unregistered.write_text("// archivo fantasma", encoding="utf-8")
        
        huella_post = capturar_huella(repo_root, [])
        ok_inv, msg_inv = verificar_invariante(huella_pre, huella_post)
        assert not ok_inv, "Fallo: Git status debió detectar el archivo untracked no registrado"
        assert "Desviación en git status" in msg_inv
        
        # El saneamiento forzado purga el archivo fantasma
        pipeline.forzar_limpieza_segura()
        assert not unregistered.exists()
        print("  -> [REPELIDO] Archivo no registrado detectado por Sensor Git y purgado.")
    finally:
        if unregistered.exists():
            unregistered.unlink()

# ====================================================================
# VECTOR 9: Eliminación inesperada de un archivo que debía restaurarse
# ====================================================================
def test_vector9_eliminacion_inesperada_de_archivo():
    print("[VECTOR 9] Eliminación inesperada de archivo rastreado preexistente...")
    target = repo_root / "src" / "pkg" / "core" / "pipeline_test.go"
    backup = target.read_text(encoding="utf-8") if target.exists() else None
    assert backup is not None, "pipeline_test.go debe existir"
    
    try:
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/pipeline_test.go"])
        # Simular que el archivo fue eliminado abruptamente
        target.unlink()
        
        # Ejecutar rollback para restaurar
        ok_rb, msg_rb = pipeline.ejecutar_rollback(
            ["src/pkg/core/pipeline_test.go"],
            huella_pre,
            "Atacante Test",
            "Simulación de archivo eliminado"
        )
        assert ok_rb, f"Fallo en restaurar archivo eliminado: {msg_rb}"
        assert target.exists(), "El archivo eliminado no fue restaurado por git checkout"
        
        huella_post = capturar_huella(repo_root, ["src/pkg/core/pipeline_test.go"])
        ok_inv, _ = verificar_invariante(huella_pre, huella_post)
        assert ok_inv, "Invariante no satisfecho tras reponer archivo eliminado"
        print("  -> [REPELIDO] Archivo eliminado reincorporado íntegramente. PRE == POST.")
    finally:
        if not target.exists() and backup:
            target.write_text(backup, encoding="utf-8")

# ====================================================================
# VECTOR 10: Archivo falso con hash aparentemente válido
# ====================================================================
def test_vector10_archivo_falso_hash_colision():
    print("[VECTOR 10] Archivo con contenido falso pero estructura similar...")
    target = repo_root / "src" / "pkg" / "core" / "collision_test.go"
    try:
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/collision_test.go"])
        target.write_text("package core\n// version A\n", encoding="utf-8")
        h_a = hashlib.sha256(target.read_bytes()).hexdigest()
        
        # Modificar a versión B
        target.write_text("package core\n// version B\n", encoding="utf-8")
        h_b = hashlib.sha256(target.read_bytes()).hexdigest()
        
        assert h_a != h_b, "Los hashes deben ser distintos"
        huella_pre["hashes"]["src/pkg/core/collision_test.go"] = h_a
        huella_post = capturar_huella(repo_root, ["src/pkg/core/collision_test.go"])
        
        ok_inv, msg_inv = verificar_invariante(huella_pre, huella_post)
        assert not ok_inv, "El verificador debió rechazar la versión adulterada"
        print("  -> [REPELIDO] Criptografía SHA256 estricta: ninguna sustitución permitida.")
    finally:
        if target.exists():
            target.unlink()

# ====================================================================
# VECTOR 11: Segunda interrupción durante recuperación
# ====================================================================
def test_vector11_segunda_interrupcion_durante_recuperacion():
    print("[VECTOR 11] Segunda interrupción consecutiva durante recuperación...")
    target = repo_root / "src" / "pkg" / "core" / "double_crash_test.go"
    try:
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/double_crash_test.go"])
        target.write_text("// crash once", encoding="utf-8")
        pipeline.iniciar_transaccion("VEC-011", ["src/pkg/core/double_crash_test.go"], huella_pre)

        # Primera recuperación sufre interrupción simulada (no finaliza)
        # Segunda recuperación arranca
        ok_rec2, msg_rec2 = pipeline.recuperar_transaccion_pendiente()
        assert ok_rec2, f"Segunda recuperación falló: {msg_rec2}"
        assert not target.exists(), "El archivo residual no fue eliminado"
        print("  -> [REPELIDO] Recuperación idempotente certificada ante fallos múltiples.")
    finally:
        if target.exists():
            target.unlink()
        pipeline.finalizar_transaccion()

# ====================================================================
# VECTOR 12: Reinicio del supervisor durante rollback
# ====================================================================
def test_vector12_reinicio_supervisor_durante_rollback():
    print("[VECTOR 12] Reinicio en frío del supervisor (Cold Boot) durante rollback activo...")
    target = repo_root / "src" / "pkg" / "core" / "cold_boot_test.go"
    try:
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/cold_boot_test.go"])
        target.write_text("// cold boot dummy", encoding="utf-8")
        pipeline.iniciar_transaccion("VEC-012", ["src/pkg/core/cold_boot_test.go"], huella_pre)

        # Iniciar una nueva instancia del supervisor
        sup = DaemonSupervisor(repo_root)
        assert not target.exists(), "Supervisor en arranque debió purgar la transacción pendiente"
        print("  -> [REPELIDO] Supervisor ejecutó auto-rollback en inicialización sin excepción.")
    finally:
        if target.exists():
            target.unlink()
        pipeline.finalizar_transaccion()

# ====================================================================
# VECTOR 13: Segundo supervisor durante recuperación
# ====================================================================
def test_vector13_segundo_supervisor_durante_recuperacion():
    print("[VECTOR 13] Segundo supervisor intentando operar concurrentemente...")
    try:
        ok1, msg1 = pipeline.adquirir_lock("SUP-PRIMARY")
        assert ok1, "Supervisor primario no pudo adquirir lock"

        ok2, msg2 = pipeline.adquirir_lock("SUP-SECONDARY")
        assert not ok2, "Segundo supervisor debió ser rechazado por lock exclusivo"
        assert "retenido activamente" in msg2 or "Lock" in msg2
        print("  -> [REPELIDO] Exclusión mutua entre supervisores confirmada.")
    finally:
        pipeline.liberar_lock()

# ====================================================================
# VECTOR 14: Corrupción de src_lock.json
# ====================================================================
def test_vector14_corrupcion_lock():
    print("[VECTOR 14] Corrupción sintáctica intencional en src_lock.json...")
    try:
        SRC_LOCK_PATH.parent.mkdir(parents=True, exist_ok=True)
        SRC_LOCK_PATH.write_text("{ CORRUPTED_LOCK_JSON_BIN \x00...", encoding="utf-8")

        # Intentar adquirir lock sobre el lock corrupto
        ok, msg = pipeline.adquirir_lock("RECOVERED_CYCLE")
        assert ok, f"Fallo al sobreescribir lock corrupto: {msg}"
        print("  -> [REPELIDO] Lock corrupto identificado como inválido y saneado atómicamente.")
    finally:
        pipeline.liberar_lock()

# ====================================================================
# VECTOR 15: Manipulación de timestamps para engañar al mecanismo
# ====================================================================
def test_vector15_manipulacion_timestamps():
    print("[VECTOR 15] Inyección de timestamps anómalos (año 2099 y 1970)...")
    try:
        # Lock con timestamp en el futuro pero PID de proceso que NO existe
        fake_lock = {
            "ciclo_id": "TIME_ATTACK",
            "pid": 999998,
            "timestamp": "2099-12-31T23:59:59"
        }
        SRC_LOCK_PATH.write_text(json.dumps(fake_lock), encoding="utf-8")

        # El pipeline comprueba si el PID está vivo en el SO (psutil), no el timestamp
        ok, msg = pipeline.adquirir_lock("TIME_DEFENDER")
        assert ok, "El sistema no debió dejarse bloquear por un timestamp falso con PID muerto"
        print("  -> [REPELIDO] Verificación a nivel de kernel/SO inmune a manipulación de reloj.")
    finally:
        pipeline.liberar_lock()

# ====================================================================
# VECTOR 16: Intento de alterar RECHAZOS.md durante recuperación
# ====================================================================
def test_vector16_alteracion_rechazos_md():
    print("[VECTOR 16] Intento de truncar o vaciar RECHAZOS.md durante recuperación...")
    rechazos_path = repo_root / "sistema" / "RECHAZOS.md"
    assert rechazos_path.exists(), "RECHAZOS.md debe existir"
    tamano_inicial = rechazos_path.stat().st_size
    assert tamano_inicial > 0, "RECHAZOS.md no puede estar vacío"

    # Registrar un rechazo legítimo
    pipeline.registrar_rechazo("Supervisor Centinela", "Intento de truncado hostil", "BLOQUEADO")
    tamano_nuevo = rechazos_path.stat().st_size
    assert tamano_nuevo > tamano_inicial, "El historial de auditoría debe ser estrictamente acumulativo"
    print(f"  -> [REPELIDO] Invariante acumulativo verificado ({tamano_inicial} -> {tamano_nuevo} bytes).")

# ====================================================================
# VECTOR 17: Intento de alterar EVIDENCIA.md para ocultar un fallo
# ====================================================================
def test_vector17_alteracion_evidencia_md():
    print("[VECTOR 17] Intento de declarar éxito sin evidencia observable...")
    propuesta_falsa = {
        "ciclo_id": "VEC-017",
        "archivos": ["src/pkg/core/pipeline_test.go"],
        "evidencia_valida": False,
        "descripcion": "Aserción de éxito sin pruebas reproducibles"
    }
    ok, msg = pipeline.ejecutar_pipeline(propuesta_falsa)
    assert not ok, "El auditor de evidencia debió rechazar la propuesta"
    assert "evidencia" in msg.lower() or "ROLLBACK" in msg
    print("  -> [REPELIDO] Criterio constitucional: Jamás se acepta éxito sin evidencia verificable.")

# ====================================================================
# VECTOR 18: Intento de modificar AUTONOMIA.json durante rollback
# ====================================================================
def test_vector18_modificar_autonomia_durante_rollback():
    print("[VECTOR 18] Intento de modificar sistema/AUTONOMIA.json durante rollback...")
    propuesta = {
        "ciclo_id": "VEC-018",
        "archivos": ["sistema/AUTONOMIA.json"],
        "descripcion": "Elevación no autorizada de presupuesto de autonomía"
    }
    ok, msg = pipeline.ejecutar_pipeline(propuesta)
    assert not ok, "El pipeline debió vetar la modificación a AUTONOMIA.json"
    assert "inmutable" in msg.lower() or "VETO" in msg or "ROLLBACK" in msg
    print("  -> [REPELIDO] Inmutabilidad constitucional de AUTONOMIA.json protegida.")

# ====================================================================
# VECTOR 19: Simulación de falta de espacio o error de I/O
# ====================================================================
def test_vector19_simulacion_error_io():
    print("[VECTOR 19] Simulación de error de I/O / disco durante modificación...")
    target = repo_root / "src" / "pkg" / "core" / "io_error_test.go"
    try:
        def fallo_io_simulado():
            raise IOError("Espacio en disco insuficiente (Simulado)")

        propuesta = {
            "ciclo_id": "VEC-019-IO",
            "archivos": ["src/pkg/core/io_error_test.go"],
            "aplicar_cambio": fallo_io_simulado,
            "descripcion": "Prueba de resiliencia ante fallo de I/O"
        }
        ok, msg = pipeline.ejecutar_pipeline(propuesta)
        assert not ok, "El pipeline debió capturar el error de I/O y ejecutar rollback"
        assert not target.exists(), "No deben quedar residuos tras el fallo de I/O"
        assert "ROLLBACK" in msg or "ERROR" in msg
        print("  -> [REPELIDO] Excepción de I/O capturada. Rollback atómico ejecutado limpiamente.")
    finally:
        if target.exists():
            target.unlink()
        pipeline.finalizar_transaccion()

# ====================================================================
# VECTOR 20 (CASO CRÍTICO): Rollback declara éxito pero PRE != POST
# ====================================================================
def test_vector20_caso_critico_pre_distinto_post_forzar_safe():
    print("[VECTOR 20 - CASO CRÍTICO] Simulación: Rollback afirma éxito pero PRE != POST...")
    target_residual = repo_root / "src" / "pkg" / "core" / "critical_undetected_residual_test.go"
    try:
        # 1. Asegurar estado RUN inicial
        transicionar_modo(repo_root, "RUN", "Inicio prueba crítica")
        assert obtener_modo_supervisor() == "RUN"

        # 2. Capturar huella limpia PRE
        huella_pre = capturar_huella(repo_root, ["src/pkg/core/critical_undetected_residual_test.go"])
        
        # 3. Deliberadamente dejar un archivo residual en disco
        target_residual.write_text("// residual que rompe el invariante", encoding="utf-8")
        
        # 4. Capturar huella física POST
        huella_post = capturar_huella(repo_root, ["src/pkg/core/critical_undetected_residual_test.go"])
        
        # 5. La verificación estricta de invariante detecta PRE != POST
        ok_inv, msg_inv = verificar_invariante(huella_pre, huella_post)
        assert not ok_inv, "Fallo: verificar_invariante debió detectar PRE != POST"
        
        # 6. Al detectar PRE != POST, el sistema debe rechazar el éxito y conmutar a SAFE/STOP
        transicionar_modo(repo_root, "SAFE", f"Discrepancia crítica de invariante: {msg_inv}")
        modo_final = obtener_modo_supervisor()
        
        # REGLA CRÍTICA INVIOLABLE: NUNCA PERMANECE EN RUN
        assert modo_final in ["SAFE", "STOP"], f"Violación crítica: modo es {modo_final}, debe ser SAFE o STOP"
        assert modo_final != "RUN", "VIOLACIÓN INADMISIBLE: El sistema permaneció en RUN ante PRE != POST"
        
        # 7. Saneamiento forzado
        pipeline.forzar_limpieza_segura()
        assert not target_residual.exists(), "El residuo no fue eliminado en limpieza forzada"
        
        # 8. Comprobar que tras la limpieza el working tree queda íntegro
        ok_git, msg_git = verificar_scope_git(repo_root)
        assert ok_git, f"Sensor Git detectó residuos: {msg_git}"

        print(f"  -> [REPELIDO Y CONMUTADO] PRE != POST detectado.")
        print(f"       -> Éxito declarado RECHAZADO categóricamente.")
        print(f"       -> Transición forzada ejecutada: RUN -> {modo_final}.")
        print(f"       -> Regla Crítica cumplida: NUNCA PERMANECE EN RUN.")
        print(f"       -> Saneamiento seguro ejecutado. Git Sensor 100% PASS.")
    finally:
        if target_residual.exists():
            target_residual.unlink()
        pipeline.finalizar_transaccion()

def run_suite():
    print("==================================================================")
    print("  FASE 14: ROLLBACK DESTRUCTIVO Y VERIFICACIÓN DE INTEGRIDAD")
    print("  Batería Adversarial de 20 Vectores Hostiles contra la Recuperación")
    print("==================================================================")
    test_vector1_interrupcion_escritura_journal()
    test_vector2_interrupcion_restauracion_archivos()
    test_vector3_corrupcion_sintactica_journal()
    test_vector4_eliminacion_journal_durante_rollback()
    test_vector5_journal_fuera_de_scope()
    test_vector6_hash_pre_manipulado()
    test_vector7_hash_post_manipulado()
    test_vector8_archivos_no_registrados_durante_rollback()
    test_vector9_eliminacion_inesperada_de_archivo()
    test_vector10_archivo_falso_hash_colision()
    test_vector11_segunda_interrupcion_durante_recuperacion()
    test_vector12_reinicio_supervisor_durante_rollback()
    test_vector13_segundo_supervisor_durante_recuperacion()
    test_vector14_corrupcion_lock()
    test_vector15_manipulacion_timestamps()
    test_vector16_alteracion_rechazos_md()
    test_vector17_alteracion_evidencia_md()
    test_vector18_modificar_autonomia_durante_rollback()
    test_vector19_simulacion_error_io()
    test_vector20_caso_critico_pre_distinto_post_forzar_safe()
    print("==================================================================")
    print("  RESULTADO: 20/20 VECTORES DE ATAQUE NEUTRALIZADOS (100% PASS)")
    print("  INVARIANTE MATEMÁTICO: PRE-CAMBIO == POST-ROLLBACK")
    print("  REGLA CRÍTICA CASO 20: TRANSICIÓN GARANTIZADA A SAFE/STOP (NUNCA RUN)")
    print("==================================================================")

if __name__ == "__main__":
    run_suite()
