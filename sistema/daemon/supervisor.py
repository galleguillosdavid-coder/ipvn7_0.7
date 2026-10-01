#!/usr/bin/env python3
"""
supervisor.py - Motor del Daemon Supervisor Persistente IPVN7.
Orquesta ciclos continuos, exclusión mutua por lock, gestión de modos
(RUN, PAUSE, SAFE, STOP), recuperación de estado y backoff ante fallos.
"""
import json
import os
import sys
import time
from datetime import datetime
from pathlib import Path

# Añadir sistema/bin al path para acceder a los módulos de gobierno
repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "sistema" / "bin"))

from evaluar_autonomia import evaluar_todos
from verificar_cambios import verificar_scope_git, registrar_rechazo
from ciclo_autonomo import sincronizar_documentos
from cerrar_ciclo import cerrar

CONFIG_PATH = Path(__file__).resolve().parent / "config.json"
ESTADO_PATH = Path(__file__).resolve().parent / "estado.json"
LOCK_PATH = Path(__file__).resolve().parent / "daemon.lock"

def pid_exists(pid: int) -> bool:
    """Verifica si un proceso sigue activo en el sistema operativo."""
    if pid <= 0:
        return False
    try:
        if os.name == "nt":
            import ctypes
            kernel32 = ctypes.windll.kernel32
            SYNCHRONIZE = 0x00100000
            process = kernel32.OpenProcess(SYNCHRONIZE, False, pid)
            if process != 0:
                kernel32.CloseHandle(process)
                return True
            return False
        else:
            os.kill(pid, 0)
            return True
    except Exception:
        return False

class DaemonSupervisor:
    def __init__(self, repo_dir: Path):
        self.repo_dir = repo_dir
        self.config = self.cargar_config()
        self.estado = self.cargar_estado()
        try:
            from pipeline_src import PipelineSrc
            PipelineSrc(self.repo_dir).recuperar_transaccion_pendiente()
        except Exception:
            pass

    def cargar_config(self) -> dict:
        if CONFIG_PATH.exists():
            with open(CONFIG_PATH, "r", encoding="utf-8") as f:
                return json.load(f)
        return {
            "modo_inicial": "RUN",
            "intervalo_segundos": 600,
            "intervalo_min_segundos": 60,
            "intervalo_max_segundos": 3600,
            "backoff_multiplicador": 2.0,
            "max_fallos_consecutivos": 5,
            "max_objetivos_por_ciclo": 3
        }

    def cargar_estado(self) -> dict:
        if ESTADO_PATH.exists():
            try:
                with open(ESTADO_PATH, "r", encoding="utf-8") as f:
                    return json.load(f)
            except Exception:
                pass # Archivo corrupto: fallback seguro
        return {
            "modo": "STOP",
            "pid": None,
            "ciclo_actual": 0,
            "ultimo_inicio": None,
            "ultimo_fin": None,
            "ultimo_objetivo": None,
            "estado_ultimo_objetivo": None,
            "fallos_consecutivos": 0,
            "intervalo_actual_segundos": self.config.get("intervalo_segundos", 600),
            "motivo_parada": None,
            "proximo_ciclo": None
        }

    def guardar_estado(self):
        with open(ESTADO_PATH, "w", encoding="utf-8") as f:
            json.dump(self.estado, f, indent=2, ensure_ascii=False)

    def adquirir_lock(self) -> tuple[bool, str]:
        if LOCK_PATH.exists():
            try:
                content = LOCK_PATH.read_text(encoding="utf-8").strip()
                lock_data = json.loads(content)
                old_pid = lock_data.get("pid")
                if old_pid and pid_exists(old_pid):
                    return False, f"RECHAZADO: Daemon activo detectado en PID {old_pid}."
            except Exception:
                pass # Lock inválido o corrupto, se sobreescribe

        # Adquirir lock para el proceso actual
        lock_info = {
            "pid": os.getpid(),
            "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "modo": self.estado.get("modo", "RUN")
        }
        LOCK_PATH.write_text(json.dumps(lock_info, indent=2), encoding="utf-8")
        self.estado["pid"] = os.getpid()
        self.guardar_estado()
        return True, "Lock adquirido correctamente."

    def liberar_lock(self):
        if LOCK_PATH.exists():
            try:
                LOCK_PATH.unlink()
            except Exception:
                pass
        self.estado["pid"] = None
        self.guardar_estado()

    def set_modo(self, nuevo_modo: str):
        modos_validos = ["RUN", "PAUSE", "SAFE", "STOP"]
        if nuevo_modo not in modos_validos:
            return False, f"Modo inválido. Debe ser uno de: {modos_validos}"
        self.estado["modo"] = nuevo_modo
        self.guardar_estado()
        return True, f"Modo cambiado a {nuevo_modo}"

    def ejecutar_ciclo(self) -> tuple[bool, str]:
        modo = self.estado.get("modo", "RUN")
        if modo == "STOP":
            return False, "Daemon detenido (STOP)."
        
        self.estado["ciclo_actual"] += 1
        self.estado["ultimo_inicio"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        ciclo_num = self.estado["ciclo_actual"]

        print(f"[{datetime.now().strftime('%H:%M:%S')}] Iniciando Ciclo #{ciclo_num} (Modo: {modo})")

        # 1. Alimentación y Descubrimiento
        clasificados = evaluar_todos(self.repo_dir)
        sincronizar_documentos(self.repo_dir, clasificados)

        if modo in ["PAUSE", "SAFE"]:
            msg = f"Modo {modo}: Descubrimiento realizado ({len(clasificados)} candidatos). Ejecución de modificaciones omitida."
            self.estado["ultimo_fin"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            self.guardar_estado()
            return True, msg

        # 2. Selección y Priorización de Objetivos
        autonomos = [c for c in clasificados if c["autonomia"] == "AUTÓNOMO"]
        if not autonomos:
            msg = f"Ciclo #{ciclo_num}: No hay objetivos autónomos pendientes. Sistema en reposo."
            self.estado["ultimo_fin"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            self.estado["fallos_consecutivos"] = 0
            self.estado["intervalo_actual_segundos"] = self.config["intervalo_segundos"]
            self.guardar_estado()
            return True, msg

        # Tomar el primer objetivo autónomo priorizado
        obj = autonomos[0]
        self.estado["ultimo_objetivo"] = obj["id"]
        print(f"     -> Ejecutando objetivo autónomo: {obj['id']} ({obj['tipo']})")

        # 3. Sensor Git previo y ejecución controlada
        ok_git, msg_git = verificar_scope_git(self.repo_dir)
        if not ok_git:
            print(f"[ALERTA DE SEGURIDAD] Scope Violation detectada antes de ejecutar: {msg_git}")
            self.set_modo("SAFE")
            self.estado["motivo_parada"] = f"Scope Violation: {msg_git}"
            self.guardar_estado()
            return False, f"Scope Violation: {msg_git}"

        # 4. Ejecución del objetivo (Corrección autónoma permitida en docs)
        exito_tarea = self.ejecutar_tarea_autonoma(obj)

        if not exito_tarea:
            self.estado["fallos_consecutivos"] += 1
            # Aplicar backoff
            nuevo_intervalo = min(
                int(self.estado["intervalo_actual_segundos"] * self.config["backoff_multiplicador"]),
                self.config["intervalo_max_segundos"]
            )
            self.estado["intervalo_actual_segundos"] = nuevo_intervalo
            if self.estado["fallos_consecutivos"] >= self.config["max_fallos_consecutivos"]:
                self.set_modo("PAUSE")
                self.estado["motivo_parada"] = "Máximo de fallos consecutivos alcanzado."
            self.guardar_estado()
            return False, f"Fallo al ejecutar tarea {obj['id']}."

        # 5. Verificación Git post-ejecución
        ok_git_post, msg_git_post = verificar_scope_git(self.repo_dir)
        if not ok_git_post:
            print(f"[CRÍTICO] Scope Violation producida tras la ejecución: {msg_git_post}")
            self.set_modo("SAFE")
            self.estado["motivo_parada"] = f"Scope Violation post-tarea: {msg_git_post}"
            self.guardar_estado()
            return False, f"Scope Violation: {msg_git_post}"

        # 6. Éxito: restablecer backoff y persistir
        self.estado["fallos_consecutivos"] = 0
        self.estado["intervalo_actual_segundos"] = self.config["intervalo_segundos"]
        self.estado["estado_ultimo_objetivo"] = "COMPLETADO"
        self.estado["ultimo_fin"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        self.guardar_estado()
        return True, f"Ciclo #{ciclo_num} completado exitosamente: {obj['id']} resuelto."

    def ejecutar_tarea_autonoma(self, obj: dict) -> bool:
        """Aplica la corrección para tareas autónomas preautorizadas (ej. referencias rotas)."""
        tipo = obj.get("tipo")
        if tipo == "REFERENCIA_ROTA":
            # Corregir enlaces rotos conocidos hacia .agents -> agentes o core -> src/pkg/core
            for f_rel in obj.get("archivos", []):
                p = self.repo_dir / f_rel
                if p.exists() and p.suffix == ".md":
                    try:
                        content = p.read_text(encoding="utf-8")
                        content_fixed = content.replace("../.agents/", "../agentes/")
                        content_fixed = content_fixed.replace("../pkg/core/", "../src/pkg/core/")
                        p.write_text(content_fixed, encoding="utf-8")
                        return True
                    except Exception as e:
                        print(f"Error modificando doc: {e}")
                        return False
        return True

def main():
    supervisor = DaemonSupervisor(repo_root)
    subcmd = sys.argv[1].lower() if len(sys.argv) > 1 else "status"

    if subcmd == "status":
        print("==================================================")
        print("  ESTADO DEL DAEMON SUPERVISOR IPVN7")
        print("==================================================")
        print(f"Modo actual:              {supervisor.estado.get('modo')}")
        print(f"PID activo:               {supervisor.estado.get('pid')}")
        print(f"Ciclo actual:             {supervisor.estado.get('ciclo_actual')}")
        print(f"Último inicio:            {supervisor.estado.get('ultimo_inicio')}")
        print(f"Último objetivo:          {supervisor.estado.get('ultimo_objetivo')} ({supervisor.estado.get('estado_ultimo_objetivo')})")
        print(f"Fallos consecutivos:      {supervisor.estado.get('fallos_consecutivos')}")
        print(f"Intervalo programado:     {supervisor.estado.get('intervalo_actual_segundos')}s")
        print(f"Motivo parada/pausa:      {supervisor.estado.get('motivo_parada')}")
        sys.exit(0)

    elif subcmd == "run-once":
        ok_lock, msg_lock = supervisor.adquirir_lock()
        if not ok_lock:
            print(f"[ERROR] {msg_lock}")
            sys.exit(1)
        try:
            ok, msg = supervisor.ejecutar_ciclo()
            print(f"[RESULTADO] {'[OK]' if ok else '[FALLO]'} {msg}")
            sys.exit(0 if ok else 1)
        finally:
            supervisor.liberar_lock()

    elif subcmd == "mode":
        if len(sys.argv) < 3:
            print("Uso: supervisor.py mode <RUN|PAUSE|SAFE|STOP>")
            sys.exit(1)
        ok, msg = supervisor.set_modo(sys.argv[2].upper())
        print(f"[{'OK' if ok else 'ERROR'}] {msg}")
        sys.exit(0 if ok else 1)

    else:
        print(f"Comando desconocido: {subcmd}")
        print("Comandos disponibles: status, run-once, mode <RUN|PAUSE|SAFE|STOP>")
        sys.exit(1)

if __name__ == "__main__":
    main()
