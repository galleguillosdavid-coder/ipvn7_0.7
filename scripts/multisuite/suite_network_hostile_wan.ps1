# ==============================================================================
# suite_network_hostile_wan.ps1 - Validación Hostil WAN Nivel 2 IPVN7 (DEC-064)
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Stop"

Write-Host "--- [NIVEL 2] Validación Hostil WAN: CGNAT, Camuflaje TLS 1.3 & Zero-Admin ---" -ForegroundColor Cyan

$hostileTests = @(
    "TestSTUNHeaderGenerationAndParsing",
    "TestSTUNRFC5389ResponseParsing",
    "TestNATTraversal_HolePunchingAndRelay",
    "TestNATHolePunchEngine_SimultaneousPunch",
    "TestTLSMasquerade_WrapUnwrap",
    "TestTLSOptionEngine_ClientHelloGeneration",
    "TestTLSOptionEngine_ApplicationDataWrapUnwrap",
    "TestXWingKEM_ConformityAndRoundtrip",
    "TestSOCKS5GatewayEchoTunnel",
    "TestNativeTunWindows_FallbackWhenMissingDLL",
    "TestPQC_PhysicalUDPLoopback_Bidirectional",
    "TestI7UDPAdapter_PhysicalTransmissionLoopback"
)

Write-Host "  -> Verificando perforacion NAT STUN RFC 5389, CGNAT, Camuflaje TLS 1.3, X-Wing PQC, Zero-Admin y Loopback Fisico UDP..." -ForegroundColor DarkGray
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$hostileFailures = @()
$passedCount = 0

try {
    Push-Location "$RepoRoot\src"
    $out = go test -v -run="TestSTUN|TestNATTraversal_HolePunchingAndRelay|TestNATHolePunchEngine_SimultaneousPunch|TestTLSMasquerade|TestTLSOptionEngine|TestXWing|TestSOCKS5GatewayEchoTunnel|TestNativeTunWindows_FallbackWhenMissingDLL|TestPQC_PhysicalUDPLoopback_Bidirectional|TestI7UDPAdapter_PhysicalTransmissionLoopback" ./pkg/l1 ./pkg/core 2>&1
    Pop-Location
    $sw.Stop()

    foreach ($ht in $hostileTests) {
        if ($out -match "--- PASS: $ht") {
            Write-Host "  [OK] $($ht): Certificado contra condiciones hostiles WAN." -ForegroundColor Green
            $passedCount++
        } else {
            Write-Host "  [FALLA] $($ht): Falla en prueba hostil WAN." -ForegroundColor Red
            $hostileFailures += $ht
        }
    }
} catch {
    $sw.Stop()
    Write-Host "  [ERROR CRITICO] Excepción en suite hostil WAN: $_" -ForegroundColor Red
    $hostileFailures += "HostileWANExecutionError"
}

return [PSCustomObject]@{
    Name        = "Hostile WAN Resilience (Level 2)"
    Passed      = ($hostileFailures.Count -eq 0)
    PassedCount = $passedCount
    TotalTests  = $hostileTests.Count
    ElapsedMs   = $sw.ElapsedMilliseconds
    Failures    = $hostileFailures
}
