# ==============================================================================
# suite_algorithmic_fuzzing.ps1 - Fuzzing Algoritmico y Resiliencia de Frontera
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Stop"

Write-Host "--- [4/6] Fuzzing Algoritmico y Resiliencia de Protocolo (L0/L1) ---" -ForegroundColor Cyan

$fuzzFailures = @()
$fuzzTests = @(
    "TestFuzz_PacketBoundaryAndTruncation",
    "TestFuzz_ChecksumTamperingBitFlips",
    "TestFuzz_MalformedPQCEncapsulation",
    "TestFuzz_TamperedSignatureRejection"
)

Write-Host "  -> Ejecutando bateria de fuzzing determinista en caliente..." -ForegroundColor DarkGray
$sw = [System.Diagnostics.Stopwatch]::StartNew()
try {
    $out = go test -v -run="TestFuzz" ./pkg/l0 2>&1
    $sw.Stop()

    $passedCount = 0
    foreach ($ft in $fuzzTests) {
        if ($out -match "--- PASS: $ft") {
            Write-Host "  [OK] $($ft): Validado (0 panics, descarte O(1) determinista)." -ForegroundColor Green
            $passedCount++
        } else {
            Write-Host "  [FALLA] $($ft): No paso la validacion de frontera." -ForegroundColor Red
            $fuzzFailures += $ft
        }
    }
} catch {
    $sw.Stop()
    Write-Host "  [ERROR CRITICO] Excepcion en ejecucion de fuzzing: $_" -ForegroundColor Red
    $fuzzFailures += "FuzzingExecutionError"
}

return [PSCustomObject]@{
    Name         = "Algorithmic Fuzzing & Wire Resiliency"
    Passed       = ($fuzzFailures.Count -eq 0)
    PassedCount  = $passedCount
    TotalTests   = $fuzzTests.Count
    ElapsedMs    = $sw.ElapsedMilliseconds
    Failures     = $fuzzFailures
}
