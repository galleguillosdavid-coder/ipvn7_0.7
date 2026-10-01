# ==============================================================================
# suite_zero_copy_drift.ps1 - Invariante Zero-Copy y Deriva de Rendimiento
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Continue"

Write-Host "--- [3/6] Invariante Zero-Copy y Proyeccion de Deriva ---" -ForegroundColor Cyan

$benchOut = go test -run=^$ -bench=BenchmarkLinearPipeline_Execute -benchmem ./pkg/core 2>&1 | Out-String
$matched = ([regex]::Matches($benchOut, 'BenchmarkLinearPipeline_Execute\S*\s+(\d+)\s+([\d\.]+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op')).Groups

$passed = $false
$bytesPerOp = -1
$allocsPerOp = -1
$nsPerOp = -1.0
$warnings = @()

if ($matched.Count -ge 5) {
    $nsPerOp = [double]$matched[2].Value
    $bytesPerOp = [int]$matched[3].Value
    $allocsPerOp = [int]$matched[4].Value

    Write-Host "  -> Metricas registradas: $nsPerOp ns/op | $bytesPerOp B/op | $allocsPerOp allocs/op" -ForegroundColor DarkGray

    # Invariante XI: 0 B/op y 0 allocs/op
    if ($bytesPerOp -eq 0 -and $allocsPerOp -eq 0) {
        Write-Host "  [OK] Invariante Zero-Copy CERTIFICADO (0 B/op, 0 allocs/op)." -ForegroundColor Green
        $passed = $true
    } else {
        Write-Host "  [FALLA CRITICA] Regresion de Memoria detectada: $bytesPerOp B/op, $allocsPerOp allocs/op (Esperado: 0/0)" -ForegroundColor Red
        $passed = $false
    }

    # Deriva de rendimiento preventiva (Umbral de alerta: 55 ns/op)
    if ($nsPerOp -gt 55.0) {
        $warnings += "Deriva de latencia elevada: $nsPerOp ns/op (> 55.0 ns umbral de degradacion)."
        Write-Host "  [PREDICCION/ALERTA] $($warnings[-1])" -ForegroundColor Yellow
    } else {
        Write-Host "  [OK] Rendimiento nominal optimo ($nsPerOp ns/op <= 55.0 ns)." -ForegroundColor Green
    }
} else {
    Write-Host "  [FALLA] No se pudo analizar la salida del benchmark." -ForegroundColor Red
}

return [PSCustomObject]@{
    Name         = "Zero-Copy & Memory Drift"
    Passed       = $passed
    NsPerOp      = $nsPerOp
    BytesPerOp   = $bytesPerOp
    AllocsPerOp  = $allocsPerOp
    Warnings     = $warnings
}
