# ==============================================================================
# suite_concurrency_race.ps1 - Deteccion de Carreras y Deadlocks Concurrente
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Stop"

Write-Host "--- [2/6] Verificacion Concurrente y Deteccion de Carreras (-race) ---" -ForegroundColor Cyan

$targets = @("./pkg/core", "./pkg/l1", "./pkg/l2")
$raceFailures = @()
$timings = @{}

$hasGCC = $false
try {
    $gccCheck = Get-Command gcc -ErrorAction SilentlyContinue
    if ($gccCheck) { $hasGCC = $true }
} catch {}

$modeLabel = if ($hasGCC) { "CGO ThreadSanitizer (-race)" } else { "Multi-Core Stress (-cpu=4,8 Contention)" }
Write-Host "  Modo Concurrente: $modeLabel" -ForegroundColor DarkGray

foreach ($pkg in $targets) {
    Write-Host "  -> Verificando contienda y concurrencia en $pkg..." -ForegroundColor DarkGray
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        if ($hasGCC) {
            $testOut = go test -race -timeout 30s $pkg 2>&1
        } else {
            $testOut = go test "-cpu=4,8" -count=1 -timeout 30s $pkg 2>&1
        }
        $sw.Stop()
        $timings[$pkg] = "$($sw.ElapsedMilliseconds) ms"

        if ($LASTEXITCODE -ne 0 -or $testOut -match "DATA RACE" -or $testOut -match "FAIL") {
            Write-Host "  [FALLA - CONCURRENCIA / RACE] en $pkg" -ForegroundColor Red
            $raceFailures += [PSCustomObject]@{ Package = $pkg; Output = ($testOut -join "`n") }
        } else {
            Write-Host "  [OK] $pkg validado bajo alta concurrencia ($($sw.ElapsedMilliseconds) ms)." -ForegroundColor Green
        }
    } catch {
        $sw.Stop()
        Write-Host "  [ERROR CRITICO] Timeout o excepcion en $($pkg): $_" -ForegroundColor Red
        $raceFailures += [PSCustomObject]@{ Package = $pkg; Output = $_.ToString() }
    }
}

return [PSCustomObject]@{
    Name       = "Concurrency & Race Detector"
    Passed     = ($raceFailures.Count -eq 0)
    Failures   = $raceFailures
    Timings    = $timings
}
