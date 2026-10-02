"""
state_manager.py - Gestor de Estado, Configuración y Cerrojos para el Daemon Supervisor IPVN7.
Módulo desacoplado para mantener compacidad y respetar el umbral de 320 líneas (Axioma III).
"""
import json
import os
from datetime import datetime
from pathlib import Path

DAEMON_DIR = Path(__file__).resolve().parent
CONFIG_PATH = DAEMON_DIR / "config.json"
ESTADO_PATH = DAEMON_DIR / "estado.json"
LOCK_PATH = DAEMON_DIR / "daemon.lock"

def pid_exists(pid: int) -> bool:
    """Verifica si un proceso sigue activo en el sistema operativo."""
    if not pid or pid <= 0:
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

def cargar_config() -> dict:
    """Carga configuración desde config.json con valores seguros por defecto."""
    if CONFIG_PATH.exists():
        try:
            with open(CONFIG_PATH, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception:
            pass
    return {
        "modo_inicial": "RUN",
        "intervalo_segundos": 300,
        "intervalo_min_segundos": 10,
        "intervalo_max_segundos": 3600,
        "backoff_multiplicador": 2.0,
        "max_fallos_consecutivos": 5,
        "max_objetivos_por_ciclo": 3
    }

def guardar_config(config: dict):
    """Persiste configuración en disco."""
    with open(CONFIG_PATH, "w", encoding="utf-8") as f:
        json.dump(config, f, indent=2, ensure_ascii=False)

def cargar_estado(config: dict) -> dict:
    """Carga estado operativo desde estado.json con tolerancia a fallos."""
    if ESTADO_PATH.exists():
        try:
            with open(ESTADO_PATH, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception:
            pass
    return {
        "modo": "STOP",
        "pid": None,
        "ciclo_actual": 0,
        "ultimo_inicio": None,
        "ultimo_fin": None,
        "ultimo_objetivo": None,
        "estado_ultimo_objetivo": None,
        "fallos_consecutivos": 0,
        "intervalo_actual_segundos": config.get("intervalo_segundos", 300),
        "motivo_parada": None,
        "proximo_ciclo": None
    }

def guardar_estado(estado: dict):
    """Persiste el estado operativo en disco."""
    with open(ESTADO_PATH, "w", encoding="utf-8") as f:
        json.dump(estado, f, indent=2, ensure_ascii=False)

def adquirir_lock(estado: dict) -> tuple[bool, str]:
    """Gestiona la exclusión mutua de instancia mediante daemon.lock."""
    if LOCK_PATH.exists():
        try:
            content = LOCK_PATH.read_text(encoding="utf-8").strip()
            lock_data = json.loads(content)
            old_pid = lock_data.get("pid")
            if old_pid and pid_exists(old_pid):
                return False, f"RECHAZADO: Daemon activo detectado en PID {old_pid}."
        except Exception:
            pass

    lock_info = {
        "pid": os.getpid(),
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "modo": estado.get("modo", "RUN")
    }
    LOCK_PATH.write_text(json.dumps(lock_info, indent=2), encoding="utf-8")
    estado["pid"] = os.getpid()
    guardar_estado(estado)
    return True, "Lock adquirido correctamente."

def liberar_lock(estado: dict):
    """Libera el lock de exclusión mutua."""
    if LOCK_PATH.exists():
        try:
            LOCK_PATH.unlink()
        except Exception:
            pass
    estado["pid"] = None
    guardar_estado(estado)
