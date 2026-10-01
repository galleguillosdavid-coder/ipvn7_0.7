#!/usr/bin/env python3
"""
daemon.py - CLI Wrapper para el Daemon Supervisor IPVN7.
Uso:
  python sistema/bin/daemon.py status
  python sistema/bin/daemon.py run-once
  python sistema/bin/daemon.py start [intervalo_segundos]
  python sistema/bin/daemon.py reprogram <intervalo_segundos>
  python sistema/bin/daemon.py mode <RUN|PAUSE|SAFE|STOP>
"""
import sys
from pathlib import Path

supervisor_path = Path(__file__).resolve().parent.parent / "daemon" / "supervisor.py"

if __name__ == "__main__":
    import subprocess
    args = [sys.executable, str(supervisor_path)] + sys.argv[1:]
    res = subprocess.run(args)
    sys.exit(res.returncode)
