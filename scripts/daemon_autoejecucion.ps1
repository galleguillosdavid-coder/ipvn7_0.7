# ==============================================================================
# daemon_autoejecucion.ps1 - Daemon Supervisor y Autodisparador Autonomo IPVN7
# ==============================================================================
# Uso:
#   .\scripts\daemon_autoejecucion.ps1 start [-IntervalSeconds 300] [-Background]
#   .\scripts\daemon_autoejecucion.ps1 stop
#   .\scripts\daemon_autoejecucion.ps1 status
#   .\scripts\daemon_autoejecucion.ps1 restart [-IntervalSeconds 300]
#   .\scripts\daemon_autoejecucion.ps1 run-once
#   .\scripts\daemon_autoejecucion.ps1 reprogram -IntervalSeconds 600
#   .\scripts\daemon_autoejecucion.ps1 logs [-Follow]
# ==============================================================================

param(
    [Parameter(Position=0)]
    [ValidateSet("start", "stop", "status", "restart", "run-once", "reprogram", "logs", "help")]
    [string]$Action = "status",

    [Parameter(Position=1)]
    [int]$IntervalSeconds = 300,

    [switch]$Background,
    [switch]$Follow
)

$ErrorActionPreference = "Continue"
$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

$DataDir = Join-Path $RepoRoot "data"
if (-not (Test-Path $DataDir)) { New-Item -ItemType Directory -Path $DataDir -Force | Out-Null }

$LogFile = Join-Path $DataDir "daemon_autoejecucion.log"
$PidFile = Join-Path $RepoRoot "sistema\daemon\autoejecucion.pid"
$PythonExe = "python"

function Write-Info($msg) {
    $ts = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    Write-Host "[$ts] [INFO] $msg" -ForegroundColor Cyan
}

function Write-Ok($msg) {
    $ts = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    Write-Host "[$ts] [OK]   $msg" -ForegroundColor Green
}

function Write-Warn($msg) {
    $ts = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    Write-Host "[$ts] [WARN] $msg" -ForegroundColor Yellow
}

function Write-Err($msg) {
    $ts = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    Write-Host "[$ts] [ERR]  $msg" -ForegroundColor Red
}

function Test-NodeHealth() {
    try {
        $res = Invoke-RestMethod -Uri "http://127.0.0.1:7070/api/v1/status" -TimeoutSec 2 -ErrorAction Stop
        return @{ Alive = $true; State = $res.vpn_state; Uptime = $res.uptime_sec; Peers = $res.peers_count }
    } catch {
        return @{ Alive = $false; State = "Down"; Uptime = 0; Peers = 0 }
    }
}

function Start-NodeAutoHeal() {
    $health = Test-NodeHealth
    if (-not $health.Alive) {
        Write-Warn "Nodo IPVN7 en 127.0.0.1:7070 no responde. Auto-curando..."
        $bin = Join-Path $RepoRoot "bin\ipvn7.exe"
        if (Test-Path $bin) {
            Start-Process -FilePath $bin -ArgumentList "-port 7777 -web-port 7070 -socks5 10807 -no-elevate" -WindowStyle Hidden
            Start-Sleep -Seconds 2
            $postHealth = Test-NodeHealth
            if ($postHealth.Alive) {
                Write-Ok "Auto-curacion exitosa: Nodo restablecido (Uptime: $($postHealth.Uptime)s)."
            } else {
                Write-Err "Fallo al restablecer nodo IPVN7 local."
            }
        } else {
            Write-Warn "Binario $bin no encontrado. Omitiendo auto-curacion de nodo."
        }
    }
}

function Execute-AutoCycle() {
    Write-Info "Ejecutando ciclo integral de autoejecucion..."
    Start-NodeAutoHeal

    # 1. Ciclo de supervision y autogobernanza
    try {
        & $PythonExe "$RepoRoot\sistema\bin\daemon.py" run-once
    } catch {
        Write-Err "Error al ejecutar ciclo de gobernanza: $_"
    }

    # 2. Rotacion preventiva de logs si supera 10MB
    if (Test-Path $LogFile) {
        $size = (Get-Item $LogFile).Length
        if ($size -gt 10MB) {
            Write-Warn "Rotando log de autoejecucion ($([math]::Round($size/1MB, 2)) MB)..."
            Move-Item -Path $LogFile -Destination (Join-Path $DataDir "daemon_autoejecucion.old.log") -Force
        }
    }
}

function Run-Loop([int]$interval) {
    Write-Info "Iniciando bucle de autoejecucion (Intervalo base: ${interval}s)..."
    while ($true) {
        Execute-AutoCycle

        # Consultar si el supervisor reprogramo dinamicamente el descanso
        $estadoPath = Join-Path $RepoRoot "sistema\daemon\estado.json"
        $descanso = $interval
        if (Test-Path $estadoPath) {
            try {
                $st = Get-Content $estadoPath -Raw | ConvertFrom-Json
                if ($st.intervalo_actual_segundos -and $st.intervalo_actual_segundos -gt 0) {
                    $descanso = $st.intervalo_actual_segundos
                }
                if ($st.modo -eq "STOP") {
                    Write-Warn "Modo STOP detectado en estado.json. Finalizando autoejecutor."
                    break
                }
            } catch {}
        }

        $nextTime = (Get-Date).AddSeconds($descanso).ToString("HH:mm:ss")
        Write-Info "[DESCANSO] Tarea completada. Reposo de ${descanso}s. Proxima ronda: $nextTime"

        # Reposo dinamico verificando interrupciones cada 2 segundos
        $end = (Get-Date).AddSeconds($descanso)
        while ((Get-Date) -lt $end) {
            Start-Sleep -Seconds 2
            if (Test-Path $estadoPath) {
                try {
                    $fresh = Get-Content $estadoPath -Raw | ConvertFrom-Json
                    if ($fresh.modo -eq "STOP") {
                        Write-Warn "Detencion solicitada durante descanso."
                        return
                    }
                } catch {}
            }
        }
    }
}

# --- DESPACHADOR DE ACCIONES ---
switch ($Action.ToLower()) {
    "start" {
        # Restablecer modo a RUN en caso de haber quedado en STOP
        try { & $PythonExe "$RepoRoot\sistema\bin\daemon.py" mode RUN | Out-Null } catch {}

        if (Test-Path $PidFile) {
            $existingPid = Get-Content $PidFile -Raw -ErrorAction SilentlyContinue
            if ($existingPid) {
                $proc = Get-Process -Id ([int]$existingPid.Trim()) -ErrorAction SilentlyContinue
                if ($proc -and $proc.Id -ne $PID) {
                    Write-Warn "El Daemon de Autoejecucion YA esta activo (PID: $($proc.Id)). Usa 'stop' o 'restart'."
                    exit 0
                }
            }
        }

        if ($Background) {
            Write-Info "Lanzando Daemon de Autoejecucion en SEGUNDO PLANO..."
            $errLog = Join-Path $DataDir "daemon_autoejecucion_err.log"
            $thisScript = if ($PSCommandPath) { $PSCommandPath } else { Join-Path $RepoRoot "scripts\daemon_autoejecucion.ps1" }
            $argList = @(
                "-NoProfile",
                "-ExecutionPolicy", "Bypass",
                "-File", $thisScript,
                "start",
                "-IntervalSeconds", "$IntervalSeconds"
            )
            $p = Start-Process -FilePath "powershell.exe" -ArgumentList $argList -RedirectStandardOutput $LogFile -RedirectStandardError $errLog -WindowStyle Hidden -PassThru
            Write-Ok "Daemon iniciado en segundo plano (PID: $($p.Id)). Logs en: $LogFile"
        } else {
            $PID | Set-Content -Path $PidFile -Force
            try {
                Run-Loop -interval $IntervalSeconds
            } finally {
                if (Test-Path $PidFile) { Remove-Item $PidFile -Force }
            }
        }
    }

    "stop" {
        Write-Info "Deteniendo Daemon de Autoejecucion..."
        $stopped = $false

        # Notificar modo STOP al supervisor
        try { & $PythonExe "$RepoRoot\sistema\bin\daemon.py" mode STOP | Out-Null } catch {}

        if (Test-Path $PidFile) {
            $pidVal = Get-Content $PidFile -Raw -ErrorAction SilentlyContinue
            if ($pidVal) {
                $proc = Get-Process -Id ([int]$pidVal.Trim()) -ErrorAction SilentlyContinue
                if ($proc) {
                    Stop-Process -Id $proc.Id -Force
                    Write-Ok "Proceso daemon en segundo plano (PID: $($proc.Id)) terminado."
                    $stopped = $true
                }
            }
            Remove-Item $PidFile -Force -ErrorAction SilentlyContinue
        }

        # Detener posibles instancias huerfanas de supervisor.py
        $pyProcs = Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -match "supervisor\.py" }
        foreach ($py in $pyProcs) {
            Stop-Process -Id $py.ProcessId -Force -ErrorAction SilentlyContinue
            Write-Ok "Supervisor Python terminado (PID: $($py.ProcessId))."
            $stopped = $true
        }

        # Limpiar locks
        $lockFile = Join-Path $RepoRoot "sistema\daemon\daemon.lock"
        if (Test-Path $lockFile) { Remove-Item $lockFile -Force }

        if (-not $stopped) {
            Write-Info "No se detectaron procesos activos de autoejecucion."
        }
    }

    "status" {
        Write-Host "================================================================" -ForegroundColor Cyan
        Write-Host "         ESTADO INTEGRAL DEL DAEMON DE AUTOEJECUCION           " -ForegroundColor Cyan
        Write-Host "================================================================" -ForegroundColor Cyan

        # 1. Proceso de Autoejecucion
        $daemonAlive = $false
        $daemonPid = "INACTIVO"
        if (Test-Path $PidFile) {
            $savedPid = (Get-Content $PidFile -Raw).Trim()
            $proc = Get-Process -Id ([int]$savedPid) -ErrorAction SilentlyContinue
            if ($proc) {
                $daemonAlive = $true
                $memMb = [math]::Round($proc.WorkingSet64 / 1048576, 1)
                $daemonPid = "$($proc.Id) (Activo - Memoria $memMb MB)"
            }
        }
        $fgColor = if ($daemonAlive) { "Green" } else { "DarkGray" }
        Write-Host "Proceso Autoejecutor:     $daemonPid" -ForegroundColor $fgColor

        # 2. Estado Supervisor Python
        try {
            & $PythonExe "$RepoRoot\sistema\bin\daemon.py" status
        } catch {
            Write-Warn "No se pudo consultar el estado del supervisor Python."
        }

        # 3. Estado Nodo IPVN7
        $node = Test-NodeHealth
        Write-Host ""
        Write-Host "--- SALUD DEL NODO IPVN7 (Puerto 7070) ---" -ForegroundColor Cyan
        if ($node.Alive) {
            Write-Host "Estado del Nodo:          EN LINEA (Uptime: $($node.Uptime)s, Red: $($node.State), Pares: $($node.Peers))" -ForegroundColor Green
        } else {
            Write-Host "Estado del Nodo:          DESCONECTADO / DETENIDO" -ForegroundColor Red
        }

        # 4. Archivos de Bloqueo
        $dLock = Test-Path "$RepoRoot\sistema\daemon\daemon.lock"
        $tLock = Test-Path "$RepoRoot\agentes\task.lock"
        $dLockStr = if ($dLock) { "BLOQUEADO" } else { "LIBRE" }
        $dLockCol = if ($dLock) { "Yellow" } else { "Green" }
        $tLockStr = if ($tLock) { "PRESENTE" } else { "LIBRE" }
        $tLockCol = if ($tLock) { "Yellow" } else { "Green" }
        Write-Host "Lock Supervisor:          $dLockStr" -ForegroundColor $dLockCol
        Write-Host "Lock Tareas Agentes:      $tLockStr" -ForegroundColor $tLockCol
        Write-Host "Archivo de Registro:      $LogFile" -ForegroundColor DarkGray
    }

    "restart" {
        & $PSCommandPath stop
        Start-Sleep -Seconds 1
        & $PSCommandPath start -IntervalSeconds $IntervalSeconds -Background
    }

    "run-once" {
        Execute-AutoCycle
    }

    "reprogram" {
        if ($IntervalSeconds -lt 5) {
            Write-Err "El intervalo debe ser de al menos 5 segundos."
            exit 1
        }
        & $PythonExe "$RepoRoot\sistema\bin\daemon.py" reprogram $IntervalSeconds
    }

    "logs" {
        if (-not (Test-Path $LogFile)) {
            Write-Warn "No existe archivo de log en $LogFile"
            exit 0
        }
        if ($Follow) {
            Get-Content -Path $LogFile -Wait -Tail 30
        } else {
            Get-Content -Path $LogFile -Tail 50
        }
    }

    default {
        Write-Host 'Uso: .\scripts\daemon_autoejecucion.ps1 {start|stop|status|restart|run-once|reprogram|logs} [opciones]'
        Write-Host '  start [-IntervalSeconds 300] [-Background] : Inicia el daemon'
        Write-Host '  stop                                       : Detiene el daemon'
        Write-Host '  status                                     : Muestra dashboard'
        Write-Host '  restart                                    : Reinicia el daemon'
        Write-Host '  run-once                                   : Ejecuta 1 ciclo'
        Write-Host '  reprogram -IntervalSeconds [seg]           : Ajusta descanso'
        Write-Host '  logs [-Follow]                             : Muestra logs'
    }
}
