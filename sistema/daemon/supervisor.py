#!/usr/bin/env python3
"""
supervisor.py - Motor del Daemon Supervisor Persistente IPVN7.
Orquesta ciclos continuos, exclusión mutua, modos (RUN, PAUSE, SAFE, STOP),
recuperación de estado, descanso dinámico y reprogramación.
Cumple estrictamente el umbral de compacidad (Axioma III < 320L).
"""
import sys
import time
from datetime import datetime
from pathlib import Path

# Añadir sistema/bin y sistema/daemon al path
repo_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(repo_root / "sistema" / "bin"))
sys.path.insert(0, str(repo_root / "sistema" / "daemon"))

from evaluar_autonomia import evaluar_todos
from verificar_cambios import verificar_scope_git
from ciclo_autonomo import sincronizar_documentos
from state_manager import (
    cargar_config, guardar_config, cargar_estado, guardar_estado,
    adquirir_lock, liberar_lock, pid_exists,
    CONFIG_PATH, ESTADO_PATH, LOCK_PATH
)

class DaemonSupervisor:
    def __init__(self, repo_dir: Path):
        self.repo_dir = repo_dir
        self.config = cargar_config()
        self.estado = cargar_estado(self.config)
        try:
            from pipeline_src import PipelineSrc
            PipelineSrc(self.repo_dir).recuperar_transaccion_pendiente()
        except Exception:
            pass

    def guardar_estado(self):
        guardar_estado(self.estado)

    def cargar_estado(self):
        return cargar_estado(self.config)

    def cargar_config(self):
        return cargar_config()

    def adquirir_lock(self) -> tuple[bool, str]:
        return adquirir_lock(self.estado)

    def liberar_lock(self):
        liberar_lock(self.estado)

    def set_modo(self, nuevo_modo: str) -> tuple[bool, str]:
        modos_validos = ["RUN", "PAUSE", "SAFE", "STOP"]
        if nuevo_modo not in modos_validos:
            return False, f"Modo inválido. Debe ser uno de: {modos_validos}"
        self.estado["modo"] = nuevo_modo
        guardar_estado(self.estado)
        return True, f"Modo cambiado a {nuevo_modo}"

    def ejecutar_ciclo(self) -> tuple[bool, str]:
        modo = self.estado.get("modo", "RUN")
        if modo == "STOP":
            return False, "Daemon detenido (STOP)."

        self.estado["ciclo_actual"] += 1
        self.estado["ultimo_inicio"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        ciclo_num = self.estado["ciclo_actual"]
        print(f"[{datetime.now().strftime('%H:%M:%S')}] Iniciando Ciclo #{ciclo_num} (Modo: {modo})")

        # 1. Descubrimiento y Evaluación de Autonomía
        clasificados = evaluar_todos(self.repo_dir)
        sincronizar_documentos(self.repo_dir, clasificados)

        if modo in ["PAUSE", "SAFE"]:
            msg = f"Modo {modo}: Descubrimiento realizado ({len(clasificados)} candidatos). Ejecución omitida."
            self.estado["ultimo_fin"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            guardar_estado(self.estado)
            return True, msg

        # 2. Selección y Priorización de Objetivos Autónomos
        autonomos = [c for c in clasificados if c.get("autonomia") == "AUTÓNOMO"]
        if not autonomos:
            msg = f"Ciclo #{ciclo_num}: No hay objetivos autónomos pendientes. Sistema en reposo."
            self.estado["ultimo_fin"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            self.estado["fallos_consecutivos"] = 0
            self.estado["intervalo_actual_segundos"] = self.config["intervalo_segundos"]
            guardar_estado(self.estado)
            return True, msg

        obj = autonomos[0]
        self.estado["ultimo_objetivo"] = obj["id"]
        print(f"     -> Ejecutando objetivo autónomo: {obj['id']} ({obj.get('tipo')})")

        # 3. Sensor Git previo
        ok_git, msg_git = verificar_scope_git(self.repo_dir)
        if not ok_git:
            print(f"[ALERTA SEGURIDAD] Scope Violation previa: {msg_git}")
            self.set_modo("SAFE")
            self.estado["motivo_parada"] = f"Scope Violation: {msg_git}"
            guardar_estado(self.estado)
            return False, f"Scope Violation: {msg_git}"

        # 4. Ejecución del objetivo autónomo
        exito_tarea = self.ejecutar_tarea_autonoma(obj)
        if not exito_tarea:
            self.estado["fallos_consecutivos"] += 1
            nuevo_int = min(
                int(self.estado["intervalo_actual_segundos"] * self.config.get("backoff_multiplicador", 2.0)),
                self.config.get("intervalo_max_segundos", 3600)
            )
            self.estado["intervalo_actual_segundos"] = nuevo_int
            if self.estado["fallos_consecutivos"] >= self.config.get("max_fallos_consecutivos", 5):
                self.set_modo("PAUSE")
                self.estado["motivo_parada"] = "Máximo de fallos consecutivos alcanzado."
            guardar_estado(self.estado)
            return False, f"Fallo al ejecutar tarea {obj['id']}."

        # 5. Sensor Git post-ejecución
        ok_git_post, msg_git_post = verificar_scope_git(self.repo_dir)
        if not ok_git_post:
            print(f"[CRÍTICO] Scope Violation producida post-tarea: {msg_git_post}")
            self.set_modo("SAFE")
            self.estado["motivo_parada"] = f"Scope Violation post-tarea: {msg_git_post}"
            guardar_estado(self.estado)
            return False, f"Scope Violation: {msg_git_post}"

        self.estado["fallos_consecutivos"] = 0
        self.estado["intervalo_actual_segundos"] = self.config["intervalo_segundos"]
        self.estado["estado_ultimo_objetivo"] = "COMPLETADO"
        self.estado["ultimo_fin"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        guardar_estado(self.estado)
        return True, f"Ciclo #{ciclo_num} completado exitosamente: {obj['id']} resuelto."

    def ejecutar_tarea_autonoma(self, obj: dict) -> bool:
        """Aplica correcciones para tareas autónomas autorizadas (ej. referencias rotas)."""
        if obj.get("tipo") == "REFERENCIA_ROTA":
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

    def reprogramar_intervalo(self, nuevo_intervalo: int) -> tuple[bool, str]:
        """Reprograma dinámicamente el descanso entre ciclos del daemon."""
        if nuevo_intervalo < 5 or nuevo_intervalo > 86400:
            return False, "El intervalo debe estar entre 5 y 86400 segundos."
        self.config["intervalo_segundos"] = nuevo_intervalo
        guardar_config(self.config)
        self.estado["intervalo_actual_segundos"] = nuevo_intervalo
        guardar_estado(self.estado)
        return True, f"Daemon reprogramado con éxito. Nuevo intervalo de descanso: {nuevo_intervalo}s"

    def bucle_continuo(self, intervalo_override: int = None):
        """Ejecuta el ciclo continuo del daemon con descanso y reprogramación dinámica."""
        if intervalo_override:
            self.reprogramar_intervalo(intervalo_override)

        ok_lock, msg_lock = self.adquirir_lock()
        if not ok_lock:
            print(f"[ERROR] {msg_lock}")
            return False

        print(f"[{datetime.now().strftime('%H:%M:%S')}] Daemon Supervisor IPVN7 iniciado en modo continuo.")
        try:
            while True:
                self.config = cargar_config()
                if self.estado.get("modo", "RUN") == "STOP":
                    print(f"[{datetime.now().strftime('%H:%M:%S')}] Modo STOP detectado. Deteniendo supervisor.")
                    break

                ok, msg = self.ejecutar_ciclo()
                print(f"[{datetime.now().strftime('%H:%M:%S')}] Resultado Ciclo: {'[OK]' if ok else '[AVISO]'} {msg}")

                descanso = self.estado.get("intervalo_actual_segundos", self.config.get("intervalo_segundos", 300))
                proximo_ts = time.time() + descanso
                proximo_str = datetime.fromtimestamp(proximo_ts).strftime("%Y-%m-%d %H:%M:%S")
                self.estado["proximo_ciclo"] = proximo_str
                guardar_estado(self.estado)

                print(f"[{datetime.now().strftime('%H:%M:%S')}] [DESCANSO] Tarea finalizada. Tomando descanso de {descanso}s.")
                print(f"[{datetime.now().strftime('%H:%M:%S')}] [REPROGRAMACIÓN] Próxima ejecución: {proximo_str}")

                while time.time() < proximo_ts:
                    time.sleep(1)
                    try:
                        estado_fresco = cargar_estado(self.config)
                        if estado_fresco.get("modo") == "STOP":
                            print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Interrupción por modo STOP.")
                            return True
                        nuevo_int = estado_fresco.get("intervalo_actual_segundos")
                        if nuevo_int and nuevo_int != descanso:
                            descanso = nuevo_int
                            proximo_ts = min(proximo_ts, time.time() + descanso)
                    except Exception:
                        pass
        except KeyboardInterrupt:
            print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Interrupción recibida (Ctrl+C).")
        finally:
            self.liberar_lock()
            print(f"[{datetime.now().strftime('%H:%M:%S')}] Lock liberado. Daemon en reposo.")
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
        print(f"Próximo ciclo programado: {supervisor.estado.get('proximo_ciclo')}")
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

    elif subcmd == "start":
        intervalo = None
        if len(sys.argv) > 2:
            try:
                intervalo = int(sys.argv[2])
            except ValueError:
                pass
        supervisor.bucle_continuo(intervalo_override=intervalo)

    elif subcmd == "reprogram":
        if len(sys.argv) < 3:
            print("Uso: supervisor.py reprogram <segundos>")
            sys.exit(1)
        try:
            nuevo_int = int(sys.argv[2])
            ok, msg = supervisor.reprogramar_intervalo(nuevo_int)
            print(f"[{'OK' if ok else 'ERROR'}] {msg}")
            sys.exit(0 if ok else 1)
        except ValueError:
            print("[ERROR] El intervalo debe ser un número entero.")
            sys.exit(1)

    elif subcmd == "mode":
        if len(sys.argv) < 3:
            print("Uso: supervisor.py mode <RUN|PAUSE|SAFE|STOP>")
            sys.exit(1)
        ok, msg = supervisor.set_modo(sys.argv[2].upper())
        print(f"[{'OK' if ok else 'ERROR'}] {msg}")
        sys.exit(0 if ok else 1)

    else:
        print(f"Comando desconocido: {subcmd}")
        print("Comandos: status, run-once, start [intervalo_s], reprogram <intervalo_s>, mode <RUN|PAUSE|SAFE|STOP>")
        sys.exit(1)

if __name__ == "__main__":
    main()
