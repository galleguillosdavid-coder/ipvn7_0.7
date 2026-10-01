#!/usr/bin/env python3
"""
gobierno.py - Controlador Unificado CLI del Sistema IPVN7 (Nivel 2 y Nivel 3)
Uso:
  python sistema/bin/gobierno.py status
  python sistema/bin/gobierno.py validate
  python sistema/bin/gobierno.py verify
  python sistema/bin/gobierno.py close [ID]
"""
import sys
from pathlib import Path
from validar_intencion import parse_intencion
from validar_scope import validar_alcance
from verificar_cambios import verificar_scope_git
from cerrar_ciclo import cerrar

def cmd_status(repo_root: Path):
    print("==================================================")
    print("  CONTROLADOR DE GOBIERNO IPVN7 (Nivel 2 & 3)")
    print("==================================================")
    intencion_path = repo_root / "sistema" / "INTENCION.md"
    ok, data = parse_intencion(intencion_path)
    if not ok:
        print(f"[ERROR] {data}")
        return 1

    print(f"Estado Intención:  {data['estado']}")
    print(f"Objetivo:          {data['objetivo'][:60]}..." if len(data['objetivo']) > 60 else f"Objetivo:          {data['objetivo']}")
    print(f"Zonas:             {', '.join(data['zonas']) if data['zonas'] else 'Ninguna'}")
    print(f"Autorizaciones:    {', '.join(data['autorizaciones']) if data['autorizaciones'] else 'Ninguna'}")
    ok_git, msg_git = verificar_scope_git(repo_root)
    estado_git = "[CONFORME]" if ok_git else "[SCOPE VIOLATION]"
    print(f"Sensor Git (L3):   {estado_git}")
    if not ok_git:
        print(f"                   {msg_git}")
    return 0 if ok_git else 1

def cmd_validate(repo_root: Path):
    print("[1/2] Validando estructura de intención...")
    ok_int, int_data = parse_intencion(repo_root / "sistema" / "INTENCION.md")
    if not ok_int:
        print(f"[RECHAZO] {int_data}")
        return 1
    print(f"[OK] Intención parseada. Estado: {int_data['estado']}")

    print("[2/2] Validando alcance contra matriz de permisos y rutas...")
    ok_scope, scope_data = validar_alcance(repo_root)
    if not ok_scope:
        print(f"[RECHAZO] {scope_data}")
        return 1
    print(f"[OK] Scope Lock verificado y aprobado.")
    return 0

def cmd_verify(repo_root: Path):
    print("[Sensor Git Nivel 3] Inspeccionando cambios en el working tree...")
    ok_git, msg_git = verificar_scope_git(repo_root)
    if not ok_git:
        print(f"[ERROR - VIOLACIÓN DE ALCANCE] {msg_git}")
        return 1
    print(f"[OK] {msg_git}")
    return 0

def cmd_close(repo_root: Path, ident: str):
    print(f"Cerrando ciclo con identificador: {ident}...")
    ok, msg = cerrar(repo_root, ident)
    if not ok:
        print(f"[ERROR] {msg}")
        return 1
    print(f"[OK] {msg}")
    return 0

def main():
    repo_root = Path(__file__).resolve().parent.parent.parent
    subcmd = sys.argv[1].lower() if len(sys.argv) > 1 else "status"

    if subcmd == "status":
        sys.exit(cmd_status(repo_root))
    elif subcmd == "validate":
        sys.exit(cmd_validate(repo_root))
    elif subcmd == "verify":
        sys.exit(cmd_verify(repo_root))
    elif subcmd == "discover":
        from ciclo_autonomo import main as run_ciclo
        run_ciclo()
        sys.exit(0)
    elif subcmd == "daemon":
        import subprocess
        daemon_script = repo_root / "sistema" / "daemon" / "supervisor.py"
        test_script = repo_root / "sistema" / "bin" / "tests_daemon.py"
        daemon_args = sys.argv[2:]
        if daemon_args and daemon_args[0] == "test":
            res = subprocess.run([sys.executable, str(test_script)])
        else:
            res = subprocess.run([sys.executable, str(daemon_script)] + daemon_args)
        sys.exit(res.returncode)
    else:
        print(f"Comando desconocido: {subcmd}")
        print("Comandos disponibles: status, validate, verify, discover, daemon <status|run-once|mode|test>, close [ID]")
        sys.exit(1)

if __name__ == "__main__":
    main()
