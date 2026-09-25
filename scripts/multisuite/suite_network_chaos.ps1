# ==============================================================================
# suite_network_chaos.ps1 - Resiliencia de Enlace WAN, Jitter y Caos UDP
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Stop"

Write-Host "--- [5/6] Resiliencia de Red, Jitter RFC 3550 y Failover O(1) ---" -ForegroundColor Cyan

$chaosTests = @(
    "TestWANActiveProber_LifecycleAndEcho",
    "TestLinkHealingEngine",
    "TestKleinbergRouter_SaturationAndEviction",
    "TestKleinbergRouter_ConcurrentAccessStress"
)

Write-Host "  -> Verificando sockets fisicos loopback, fluctuacion de jitter y failover..." -ForegroundColor DarkGray
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$chaosFailures = @()
$passedCount = 0

try {
    $out = go test -v -run="TestWANActiveProber|TestLinkHealing|TestKleinbergRouter" ./pkg/l1 2>&1
    $sw.Stop()

    foreach ($ct in $chaosTests) {
        if ($out -match "--- PASS: $ct") {
            Write-Host "  [OK] $($ct): Resiliencia certificada (sockets UDP reales)." -ForegroundColor Green
            $passedCount++
        } else {
            Write-Host "  [FALLA] $($ct): Falla en prueba de resiliencia de red." -ForegroundColor Red
            $chaosFailures += $ct
        }
    }
} catch {
    $sw.Stop()
    Write-Host "  [ERROR CRITICO] Excepcion en suite de resiliencia de red: $_" -ForegroundColor Red
    $chaosFailures += "NetworkChaosExecutionError"
}

return [PSCustomObject]@{
    Name        = "Network Chaos & Failover Resilience"
    Passed      = ($chaosFailures.Count -eq 0)
    PassedCount = $passedCount
    TotalTests  = $chaosTests.Count
    ElapsedMs   = $sw.ElapsedMilliseconds
    Failures    = $chaosFailures
}
