# ==============================================================================
# view_logs.ps1 - Visor en tiempo real de logs de ejecución de IPVN7
# ==============================================================================

param(
    [switch]$Follow,
    [int]$Lines = 50,
    [string]$LogFile = ""
)

$RepoRoot = Split-Path -Parent $PSScriptRoot

if (-not $LogFile) {
    $LogFile = Join-Path $RepoRoot "data\ipvn7.log"
}

if (-not (Test-Path $LogFile)) {
    Write-Host "[!] Archivo de log no encontrado en: $LogFile" -ForegroundColor Yellow
    Write-Host "    Inicia ipvn7 para generar el registro de eventos." -ForegroundColor DarkGray
    exit 0
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "           REGISTRO DE EJECUCION IPVN7 ($LogFile)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

if ($Follow) {
    Write-Host "[*] Siguiendo eventos en vivo (Ctrl+C para salir)...`n" -ForegroundColor Green
    Get-Content $LogFile -Tail $Lines -Wait
} else {
    Get-Content $LogFile -Tail $Lines
    Write-Host "`n[*] Para seguir eventos en vivo, ejecuta: .\scripts\view_logs.ps1 -Follow" -ForegroundColor DarkGray
}
