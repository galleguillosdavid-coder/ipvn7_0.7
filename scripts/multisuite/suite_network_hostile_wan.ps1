# ==============================================================================
# suite_network_hostile_wan.ps1 - Validación Hostil WAN Nivel 2 IPVN7 (DEC-064)
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Stop"

Write-Host "--- [NIVEL 2] Validación Hostil WAN: CGNAT, Camuflaje TLS 1.3 & Pacing BBR ---" -ForegroundColor Cyan

$hostileTests = @(
    "TestSTUNHeaderGenerationAndParsing",
    "TestSTUNRFC5389ResponseParsing",
    "TestTLSMasquerade_WrapUnwrap",
    "TestTLSOptionEngine_ClientHelloGeneration",
    "TestBBRController_StartupAndConvergence",
    "TestCorporateVPNZeroAdminFallback",
    "TestCorporateVPNTLSDisguise",
    "TestMASQUE_WrapAndUnwrapDatagramCapsule",
    "TestXWingKEM_ConformityAndRoundtrip"
)

Write-Host "  -> Verificando perforación NAT STUN RFC 5389, MASQUE RFC 9298, X-Wing PQC y Zero-Admin..." -ForegroundColor DarkGray
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$hostileFailures = @()
$passedCount = 0

try {
    Push-Location "$RepoRoot\src"
    $out = go test -v -run="TestSTUN|TestTLSMasquerade|TestTLSOptionEngine|TestBBRController|TestCorporateVPN|TestMASQUE|TestXWing" ./pkg/l1 2>&1
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
