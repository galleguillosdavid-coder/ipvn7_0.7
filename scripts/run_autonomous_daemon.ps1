# ==============================================================================
# run_autonomous_daemon.ps1 - Demonio Autodisparador de Autotareas (10 min)
# ==============================================================================

$ErrorActionPreference = "Continue"

$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  AUTODISPARADOR PERMANENTE IPVN7 - CADENCIA 10 MINUTOS" -ForegroundColor Cyan
Write-Host "  Modo: Daemon Autonomo con Exclusion Mutua (.agents/task.lock)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

$intervalSeconds = 600

while ($true) {
    $now = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    $lockFile = "$RepoRoot\.agents\task.lock"

    $isLocked = $false
    if (Test-Path $lockFile) {
        $lockContent = Get-Content $lockFile -Raw
        if ($lockContent -notmatch "STATUS=FREE" -and $lockContent -match "PID=") {
            $isLocked = $true
        }
    }

    if ($isLocked) {
        Write-Host "[$now] [AUTODISPARADOR] Tarea en curso detectada ($lockFile). Omitiendo ciclo." -ForegroundColor DarkYellow
    } else {
        Write-Host "[$now] [AUTODISPARADOR] Disparando ciclo autonomo programado..." -ForegroundColor Cyan
        try {
            & powershell -ExecutionPolicy Bypass -File "$RepoRoot\scripts\autonomous_cycle.ps1"
        } catch {
            Write-Host "[$now] [AUTODISPARADOR] Error ejecutando ciclo: $_" -ForegroundColor Red
        }
    }

    Write-Host "[$now] [AUTODISPARADOR] En reposo durante $intervalSeconds segundos (proximo disparo a las $((Get-Date).AddSeconds($intervalSeconds).ToString('HH:mm:ss')))...`n" -ForegroundColor DarkGray
    Start-Sleep -Seconds $intervalSeconds
}
