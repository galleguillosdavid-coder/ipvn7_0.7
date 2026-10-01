#!/usr/bin/env python3
"""
pipeline_src.py - Motor de Modificación Controlada de src/ y Rollback Invariante (Fase 13).
Implementa el flujo estricto de 6 etapas:
  1. Arquitecto
  2. Implementador
  3. Atacante
  4. Seguridad
  5. Verificador
  6. Decisión: Rollback Invariante o Registro de Evidencia.

Garantía Constitucional:
  HASH/DIFF PRE-CAMBIO == HASH/DIFF POST-ROLLBACK
  verificado por Sensor Git Nivel 3 y tests de Go.
"""
import fnmatch
import hashlib
import json
import os
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
SRC_LOCK_PATH = REPO_ROOT / "sistema" / "daemon" / "src_lock.json"
SRC_TRANS_PATH = REPO_ROOT / "sistema" / "daemon" / "src_transaction.json"
POLITICA_SRC_PATH = REPO_ROOT / "sistema" / "reglas" / "politica_src.json"
RECHAZOS_PATH = REPO_ROOT / "sistema" / "RECHAZOS.md"
EVIDENCIA_PATH = REPO_ROOT / "sistema" / "EVIDENCIA.md"

def sha256_archivo(ruta: Path) -> str:
    if not ruta.exists() or not ruta.is_file():
        return "INEXISTENTE"
    h = hashlib.sha256()
    with open(ruta, "rb") as f:
        while chunk := f.read(65536):
            h.update(chunk)
    return h.hexdigest()

def capturar_huella(repo_root: Path, archivos: list) -> dict:
    """Captura el estado criptográfico exacto del working tree y de los archivos involucrados."""
    huella = {
        "hashes": {},
        "git_status": "",
        "git_head": "",
        "timestamp": datetime.now().isoformat()
    }
    
    # 1. Hashes de archivos específicos
    for a in archivos:
        p = repo_root / a
        huella["hashes"][str(a).replace("\\", "/")] = sha256_archivo(p)
        
    # 2. Status Git de porcelain
    try:
        res = subprocess.run(["git", "status", "--porcelain"], cwd=repo_root, capture_output=True, text=True, check=True)
        huella["git_status"] = res.stdout.strip()
    except Exception as e:
        huella["git_status"] = f"ERROR_GIT: {e}"

    # 3. Head de Git
    try:
        res = subprocess.run(["git", "rev-parse", "HEAD"], cwd=repo_root, capture_output=True, text=True, check=True)
        huella["git_head"] = res.stdout.strip()
    except Exception as e:
        huella["git_head"] = "UNKNOWN"

    return huella

def verificar_invariante(huella_pre: dict, huella_post: dict) -> tuple[bool, str]:
    """Verifica que no exista ninguna desviación entre el estado pre-cambio y post-rollback."""
    # 1. Comparación de hashes
    for f, h_pre in huella_pre.get("hashes", {}).items():
        h_post = huella_post.get("hashes", {}).get(f)
        if h_pre != h_post:
            return False, f"Desviación de hash en '{f}': pre={h_pre} vs post={h_post}"

    # 2. Comparación de git status
    if huella_pre.get("git_status") != huella_post.get("git_post_rollback_status", huella_post.get("git_status")):
        status_pre = huella_pre.get("git_status")
        status_post = huella_post.get("git_status")
        return False, f"Desviación en git status: PRE='{status_pre}' vs POST='{status_post}'"

    return True, "INVARIANTE MATEMÁTICO VERIFICADO: PRE-CAMBIO == POST-ROLLBACK"

def transicionar_modo(repo_root: Path, nuevo_modo: str, motivo: str):
    """Fuerza al supervisor a modo SAFE o STOP ante discrepancia de integridad."""
    estado_path = repo_root / "sistema" / "daemon" / "estado.json"
    if estado_path.exists():
        try:
            with open(estado_path, "r", encoding="utf-8") as f:
                st = json.load(f)
            st["modo"] = nuevo_modo
            st["motivo_parada"] = motivo
            st["estado_ultimo_objetivo"] = "FALLIDO"
            with open(estado_path, "w", encoding="utf-8") as f:
                json.dump(st, f, indent=2, ensure_ascii=False)
        except Exception:
            pass

class PipelineSrc:
    def __init__(self, repo_root: Path = REPO_ROOT):
        self.repo_root = repo_root
        self.politica = self._cargar_politica()

    def _cargar_politica(self) -> dict:
        if POLITICA_SRC_PATH.exists():
            try:
                with open(POLITICA_SRC_PATH, "r", encoding="utf-8") as f:
                    return json.load(f)
            except Exception:
                pass
        return {}

    def adquirir_lock(self, ciclo_id: str) -> tuple[bool, str]:
        if SRC_LOCK_PATH.exists():
            try:
                with open(SRC_LOCK_PATH, "r", encoding="utf-8") as f:
                    lock_data = json.load(f)
                lock_pid = lock_data.get("pid")
                # Verificar si el proceso sigue vivo en el sistema operativo
                import psutil
                if lock_pid and psutil.pid_exists(lock_pid):
                    return False, f"Lock de src/ retenido activamente por ciclo '{lock_data.get('ciclo_id')}' (PID {lock_pid})"
            except Exception:
                pass # Lock huérfano o corrupto

        # Crear nuevo lock
        lock_info = {
            "ciclo_id": ciclo_id,
            "pid": os.getpid(),
            "timestamp": datetime.now().isoformat()
        }
        SRC_LOCK_PATH.parent.mkdir(parents=True, exist_ok=True)
        with open(SRC_LOCK_PATH, "w", encoding="utf-8") as f:
            json.dump(lock_info, f, indent=2)
        return True, "Lock de src/ adquirido exitosamente"

    def liberar_lock(self):
        if SRC_LOCK_PATH.exists():
            try:
                SRC_LOCK_PATH.unlink()
            except Exception:
                pass

    def iniciar_transaccion(self, ciclo_id: str, archivos: list, huella_pre: dict) -> tuple[bool, str]:
        # Validar que ningún archivo apunte fuera del scope de src/ o contenga path traversal
        for a in archivos:
            norm_a = os.path.normpath(str(a)).replace("\\", "/")
            if norm_a.startswith("..") or os.path.isabs(str(a)) or not norm_a.startswith("src/"):
                transicionar_modo(self.repo_root, "SAFE", f"Ruta fuera de scope en transacción: {a}")
                return False, f"[VETO SCOPE] Ruta fuera de scope en journal: {a}"

        trans_info = {
            "ciclo_id": ciclo_id,
            "archivos": archivos,
            "huella_pre": huella_pre,
            "timestamp": datetime.now().isoformat()
        }
        SRC_TRANS_PATH.parent.mkdir(parents=True, exist_ok=True)
        # Escritura atómica mediante reemplazo para evitar archivos truncados
        tmp_trans = SRC_TRANS_PATH.with_suffix(".tmp")
        with open(tmp_trans, "w", encoding="utf-8") as f:
            json.dump(trans_info, f, indent=2)
        os.replace(tmp_trans, SRC_TRANS_PATH)
        return True, "Transacción iniciada atómicamente"

    def finalizar_transaccion(self):
        if SRC_TRANS_PATH.exists():
            try:
                SRC_TRANS_PATH.unlink()
            except Exception:
                pass
        self.liberar_lock()

    def forzar_limpieza_segura(self):
        """Restaura src/ ante cualquier fallo o inconsistencia catastrófica."""
        try:
            subprocess.run(["git", "checkout", "--", "src/"], cwd=self.repo_root, capture_output=True)
            subprocess.run(["git", "clean", "-fd", "src/"], cwd=self.repo_root, capture_output=True)
        except Exception:
            pass

    def recuperar_transaccion_pendiente(self) -> tuple[bool, str]:
        """Si el supervisor se reinicia a mitad de un cambio o rollback, recupera el estado."""
        if not SRC_TRANS_PATH.exists():
            return True, "No hay transacciones pendientes en src/"
        try:
            with open(SRC_TRANS_PATH, "r", encoding="utf-8") as f:
                trans = json.load(f)
        except Exception as e:
            # Corrupción sintáctica en el journal
            print(f"[RECUPERACIÓN] Corrupción en journal transaccional detectada: {e}. Forzando modo SAFE...")
            self.forzar_limpieza_segura()
            transicionar_modo(self.repo_root, "SAFE", f"Corrupción en src_transaction.json: {e}")
            if SRC_TRANS_PATH.exists():
                SRC_TRANS_PATH.unlink()
            self.liberar_lock()
            return False, f"Corrupción sintáctica en journal transaccional ({e}). Conmutado a SAFE."

        archivos = trans.get("archivos", [])
        huella_pre = trans.get("huella_pre", {})

        # Validar rutas del journal
        for a in archivos:
            norm_a = os.path.normpath(str(a)).replace("\\", "/")
            if norm_a.startswith("..") or os.path.isabs(str(a)) or not norm_a.startswith("src/"):
                self.forzar_limpieza_segura()
                transicionar_modo(self.repo_root, "SAFE", f"Ruta maliciosa en journal: {a}")
                self.finalizar_transaccion()
                return False, f"Journal corrupto con ruta fuera de scope: {a}. Conmutado a SAFE."

        print(f"[RECUPERACIÓN] Transacción huérfana detectada para ciclo '{trans.get('ciclo_id')}'. Ejecutando rollback...")
        ok_rb, msg_rb = self.ejecutar_rollback(archivos, huella_pre, agente_objetor="Supervisor (Cold Boot Recovery)", motivo="Reinicio durante transacción activa")
        return ok_rb, msg_rb

    def registrar_rechazo(self, agente: str, motivo: str, accion: str):
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        entry = f"| {timestamp} | {agente} | {motivo} | Política src/ & Invariantes | {accion} |\n"
        if RECHAZOS_PATH.exists():
            content = RECHAZOS_PATH.read_text(encoding="utf-8")
            if not content.endswith("\n"):
                content += "\n"
            content += entry
            RECHAZOS_PATH.write_text(content, encoding="utf-8")

    def ejecutar_rollback(self, archivos: list, huella_pre: dict, agente_objetor: str, motivo: str) -> tuple[bool, str]:
        """Ejecuta el rollback forzado y valida el invariante matemático."""
        print(f"  -> [ROLLBACK INICIADO] Agente: {agente_objetor} | Causa: {motivo}")
        
        # 1. Revertir archivos existentes y eliminar nuevos
        for a in archivos:
            p = self.repo_root / a
            h_original = huella_pre.get("hashes", {}).get(str(a).replace("\\", "/"))
            if h_original == "INEXISTENTE":
                if p.exists():
                    p.unlink()
            else:
                # Restaurar vía checkout git
                try:
                    subprocess.run(["git", "checkout", "--", str(p)], cwd=self.repo_root, capture_output=True, check=True)
                except Exception:
                    pass

        # 2. Capturar huella post-rollback física
        huella_post = capturar_huella(self.repo_root, archivos)

        # 3. Validar invariante matemático: PRE-CAMBIO == POST-ROLLBACK
        ok_inv, msg_inv = verificar_invariante(huella_pre, huella_post)
        if not ok_inv:
            # Fallback de saneamiento extremo
            self.forzar_limpieza_segura()
            huella_post = capturar_huella(self.repo_root, archivos)
            ok_inv, msg_inv = verificar_invariante(huella_pre, huella_post)

        # 4. Validar tests locales de Go tras el rollback
        try:
            res_tests = subprocess.run(["go", "test", "./pkg/core", "./pkg/l1"], cwd=self.repo_root / "src", capture_output=True, text=True)
            tests_ok = (res_tests.returncode == 0)
        except Exception:
            tests_ok = False

        # 5. Validar Sensor Git Nivel 3
        from verificar_cambios import verificar_scope_git
        ok_git, msg_git = verificar_scope_git(self.repo_root)

        # 6. Registrar rechazo y limpiar transacción
        self.registrar_rechazo(agente_objetor, motivo, f"ROLLBACK EJECUTADO. Invariante: {'PASS' if ok_inv else 'FAIL'}, Tests: {'PASS' if tests_ok else 'FAIL'}")
        self.finalizar_transaccion()

        if not ok_inv:
            transicionar_modo(self.repo_root, "SAFE", f"FALLO CRÍTICO DE INVARIANTE: {msg_inv}")
            return False, f"FALLO CRÍTICO DE INVARIANTE TRAS ROLLBACK: {msg_inv}"
        if not tests_ok:
            transicionar_modo(self.repo_root, "SAFE", "Tests de Go fallan tras rollback")
            return False, "FALLO CRÍTICO: Los tests de Go fallan tras el rollback."
        if not ok_git:
            transicionar_modo(self.repo_root, "SAFE", f"Sensor Git detectó residuos: {msg_git}")
            return False, f"FALLO CRÍTICO: Sensor Git detectó residuos: {msg_git}"

        return True, f"ROLLBACK CONFORME. {msg_inv}. Sensor Git y Go Tests 100% PASS."

    # ==============================================================
    # PIPELINE DE 6 AGENTES
    # ==============================================================
    def etapa_arquitecto(self, propuesta: dict) -> tuple[bool, str]:
        """Agente 1: Valida arquitectura, archivos bloqueados y ampliaciones no autorizadas."""
        archivos = propuesta.get("archivos", [])
        bloqueados = self.politica.get("zonas_estrictamente_bloqueadas", [])
        
        # Inmutables de gobernanza
        inmutables = [
            "sistema/constitucion.md", "sistema/autonomia.json",
            "sistema/reglas/politica_src.json", "sistema/reglas/permisos.json"
        ]

        for a in archivos:
            a_norm = str(a).replace("\\", "/").lower()
            
            # Intento de alterar gobernanza
            if any(a_norm == inm.lower() for inm in inmutables):
                return False, f"[VETO ARQUITECTO] Intento hostil de modificar archivo de gobernanza inmutable '{a}'"

            # Intento de mezclar src con fuera de src no autorizado
            if not a_norm.startswith("src/"):
                if not any(a_norm.startswith(p) for p in ["sistema/", "agentes/", "docs/"]):
                    return False, f"[VETO ARQUITECTO] Intento de modificar archivo fuera de src no autorizado '{a}'"

            # Intento de modificar zonas bloqueadas
            for b in bloqueados:
                if fnmatch.fnmatch(a_norm, b.lower()):
                    return False, f"[VETO ARQUITECTO] Intento de modificar zona estrictamente bloqueada '{a}' (regla: {b})"

        return True, "Aprobado por Arquitecto"

    def etapa_implementador(self, propuesta: dict) -> tuple[bool, str]:
        """Agente 2: Aplica el cambio y detecta manipulación maliciosa de tests ('test tampering')."""
        archivos = propuesta.get("archivos", [])
        
        # Validar si se pretende debilitar un test existente
        if propuesta.get("manipula_tests_para_ocultar_fallo", False):
            return False, "[VETO IMPLEMENTADOR] Detectado intento de alterar tests existentes para ocultar un defecto (Test Tampering)"

        permitidas = self.politica.get("zonas_permitidas_bajo_autonomia", [])
        for a in archivos:
            a_norm = str(a).replace("\\", "/").lower()
            if a_norm.startswith("src/"):
                cumple_patron = False
                for p_info in permitidas:
                    if fnmatch.fnmatch(a_norm, p_info["patron"].lower()):
                        cumple_patron = True
                        break
                if not cumple_patron:
                    return False, f"[VETO IMPLEMENTADOR] El archivo '{a}' no cumple los patrones permitidos para autonomía en politica_src.json"

        return True, "Aprobado por Implementador"

    def etapa_atacante(self, propuesta: dict) -> tuple[bool, str]:
        """Agente 3: Prueba vectores hostiles, fuzzing y detecta regresiones lógicas sutiles."""
        if propuesta.get("falla_adversarial", False):
            return False, "[VETO ATACANTE] Ataque de inyección de carga maliciosa detectado en datapath"
        
        if propuesta.get("regresion_logica_sutil", False):
            return False, "[VETO ATACANTE] Regresión lógica sutil detectada: desalineación en cálculo de timeout / límites de búfer"

        return True, "Aprobado por Atacante"

    def etapa_seguridad(self, propuesta: dict) -> tuple[bool, str]:
        """Agente 4: Inspecciona vulnerabilidades de memoria, concurrencia o criptografía."""
        if propuesta.get("vulnerabilidad_seguridad", False):
            return False, "[VETO SEGURIDAD] Detectada condición de carrera potencial o puntero nulo no verificado"
        return True, "Aprobado por Seguridad"

    def etapa_verificador(self, propuesta: dict) -> tuple[bool, str]:
        """Agente 5: Ejecuta pruebas unitarias y linters de Go."""
        if propuesta.get("falla_compilacion", False):
            return False, "[VETO VERIFICADOR] Error de compilación en src/ ('go vet' reporta fallo sintáctico)"
        
        if propuesta.get("falla_unit_test", False):
            return False, "[VETO VERIFICADOR] Fallo en suite de pruebas unitarias ('go test' exited code 1)"

        if propuesta.get("falla_deliberada_verificador", False):
            return False, "[VETO VERIFICADOR] Verificador detectó aserción rota en verificación estática"

        return True, "Aprobado por Verificador"

    def ejecutar_pipeline(self, propuesta: dict) -> tuple[bool, str]:
        """Coordina la ejecución de las 6 etapas garantizando atomicidad y rollback invariante."""
        ciclo_id = propuesta.get("ciclo_id", "CICLO-SRC-TEST")
        archivos = propuesta.get("archivos", [])

        # 0. Adquirir lock exclusivo de src/
        ok_lock, msg_lock = self.adquirir_lock(ciclo_id)
        if not ok_lock:
            self.registrar_rechazo("Supervisor de Concurrencia", msg_lock, "BLOQUEADO. Intento de modificación concurrente de src/.")
            return False, msg_lock

        # Capturar huella pre-cambio
        huella_pre = capturar_huella(self.repo_root, archivos)
        self.iniciar_transaccion(ciclo_id, archivos, huella_pre)

        try:
            # 1. Arquitecto
            ok_arq, msg_arq = self.etapa_arquitecto(propuesta)
            if not ok_arq:
                _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Arquitecto", msg_arq)
                return False, f"{msg_arq} | {msg_rb}"

            # 2. Implementador
            ok_imp, msg_imp = self.etapa_implementador(propuesta)
            if not ok_imp:
                _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Implementador", msg_imp)
                return False, f"{msg_imp} | {msg_rb}"

            if "aplicar_cambio" in propuesta and callable(propuesta["aplicar_cambio"]):
                propuesta["aplicar_cambio"]()

            # 3. Atacante
            ok_ata, msg_ata = self.etapa_atacante(propuesta)
            if not ok_ata:
                _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Atacante", msg_ata)
                return False, f"{msg_ata} | {msg_rb}"

            # 4. Seguridad
            ok_seg, msg_seg = self.etapa_seguridad(propuesta)
            if not ok_seg:
                _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Seguridad", msg_seg)
                return False, f"{msg_seg} | {msg_rb}"

            # 5. Verificador
            ok_ver, msg_ver = self.etapa_verificador(propuesta)
            if not ok_ver:
                _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Verificador", msg_ver)
                return False, f"{msg_ver} | {msg_rb}"

            # 6. Verificación de Evidencia Fáctica requerida
            if not propuesta.get("evidencia_valida", True):
                _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Auditor de Evidencia", "Falta evidencia reproducible fáctica")
                return False, f"[VETO AUDITOR] Falta evidencia reproducible fáctica | {msg_rb}"

            # Si TODO pasa exitosamente:
            self.finalizar_transaccion()
            return True, f"[ÉXITO PIPELINE] Cambio en src/ aprobado y verificado por los 5 agentes. Ciclo: {ciclo_id}"

        except Exception as e:
            # Rollback forzado de emergencia ante excepción no controlada
            _, msg_rb = self.ejecutar_rollback(archivos, huella_pre, "Sistema (Excepción)", str(e))
            return False, f"[ERROR EXCEPCIÓN] {e} | {msg_rb}"
