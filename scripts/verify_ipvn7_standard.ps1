# ==============================================================================
# verify_ipvn7_standard.ps1 - Magna Multi-Suite de Verificacion Predictiva IPVN7
# ==============================================================================

$ErrorActionPreference = "Stop"

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  MAGNA MULTI-SUITE PREDICTIVA IPVN7 - AGENTE DE CALIDAD TOTAL" -ForegroundColor Cyan
Write-Host "  Modo: Evaluacion Integral Local y Diagnostico Proactivo (0 Tokens)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

$RepoRoot = Split-Path -Parent $PSScriptRoot
$SrcRoot = "$RepoRoot\src"

# [INVARIANTE DE RAIZ PURA] Prohibicion absoluta de archivos en la raiz del repositorio
$rootFiles = Get-ChildItem -Path $RepoRoot -File
if ($rootFiles.Count -gt 0) {
    Write-Host "`n[ERROR CRITICO] Violacion de Raiz Pura: existen $($rootFiles.Count) archivo(s) en la raiz ($($rootFiles.Name -join ', '))." -ForegroundColor Red
    Write-Host "Todo archivo debe residir en su carpeta correspondiente (config/, docs/, scripts/, src/, etc.)." -ForegroundColor Red
    exit 1
}

Set-Location $SrcRoot
$multiDir = "$RepoRoot\scripts\multisuite"

# 1. Auditoria Estatica y Proyeccion Preventiva (Axioma III + go vet)
$staticRes = & "$multiDir\suite_static_predictive.ps1" -RepoRoot $RepoRoot
if (-not $staticRes.Passed) {
    Write-Host "`n[ERROR CRITICO] Fallo en la auditoria estatica o limite de lineas." -ForegroundColor Red
    exit 1
}

# 2. Deteccion de Carreras y Concurrencia (-race)
$raceRes = & "$multiDir\suite_concurrency_race.ps1" -RepoRoot $RepoRoot
if (-not $raceRes.Passed) {
    Write-Host "`n[ERROR CRITICO] Condiciones de carrera detectadas en modulos concurrentes." -ForegroundColor Red
    exit 1
}

# 3. Invariante Zero-Copy y Proyeccion de Deriva (0 B/op)
$zeroCopyRes = & "$multiDir\suite_zero_copy_drift.ps1" -RepoRoot $RepoRoot
if (-not $zeroCopyRes.Passed) {
    Write-Host "`n[ERROR CRITICO] Regresion en el invariante de memoria zero-copy (0 B/op)." -ForegroundColor Red
    exit 1
}

# 4. Fuzzing Algoritmico y Resiliencia de Frontera (L0/L1)
$fuzzRes = & "$multiDir\suite_algorithmic_fuzzing.ps1" -RepoRoot $RepoRoot
if (-not $fuzzRes.Passed) {
    Write-Host "`n[ERROR CRITICO] Fallo en la suite de fuzzing o tramas malformadas." -ForegroundColor Red
    exit 1
}

# 5. Resiliencia de Red, Jitter RFC 3550 y Failover O(1)
$chaosRes = & "$multiDir\suite_network_chaos.ps1" -RepoRoot $RepoRoot
if (-not $chaosRes.Passed) {
    Write-Host "`n[ERROR CRITICO] Falla en resiliencia de enlace WAN o failover O(1)." -ForegroundColor Red
    exit 1
}

# 6. Validacion Hostil WAN Nivel 2 (STUN RFC 5389, TLS 1.3, BBR, Zero-Admin)
$hostileWanRes = & "$multiDir\suite_network_hostile_wan.ps1" -RepoRoot $RepoRoot
if (-not $hostileWanRes.Passed) {
    Write-Host "`n[ERROR CRITICO] Falla en la validacion hostil WAN (Nivel 2)." -ForegroundColor Red
    exit 1
}

# 7. Suite de Pruebas de Paquetes Completa
Write-Host "`n--- [7/8] Suite Universal de Pruebas Unitarias (go test ./pkg/...) ---" -ForegroundColor Cyan
try {
    go test ./pkg/...
    Write-Host "  [OK] Todos los paquetes pasaron pruebas unitarias (100% PASS)." -ForegroundColor Green
} catch {
    Write-Host "  [ERROR] Fallaron las pruebas unitarias de paquetes." -ForegroundColor Red
    exit 1
}

# 8. Compilacion de Binario Principal
Write-Host "`n--- [8/8] Compilando Binario de Produccion (ipvn7.exe) ---" -ForegroundColor Cyan
$buildPassed = $false
try {
    go build -o "$RepoRoot\bin\ipvn7.exe" ./cmd/ipvn7
    Write-Host "  [OK] Binario bin/ipvn7.exe generado correctamente." -ForegroundColor Green
    $buildPassed = $true
} catch {
    Write-Host "  [ERROR] Fallo al compilar el binario." -ForegroundColor Red
    exit 1
}

# 9. Consolidacion de Informe y Health Score
$reportRes = & "$multiDir\suite_report_generator.ps1" `
    -StaticRes $staticRes `
    -RaceRes $raceRes `
    -ZeroCopyRes $zeroCopyRes `
    -FuzzRes $fuzzRes `
    -ChaosRes $chaosRes `
    -HostileWanRes $hostileWanRes `
    -BuildPassed $buildPassed `
    -RepoRoot $RepoRoot

Write-Host ""
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  VERIFICACION MAGNA COMPLETADA: IPVN7 OPERA BAJO EXCELENCIA" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
