# ==============================================================================
# suite_static_predictive.ps1 - Auditoria Estatica y Proyeccion Preventiva (320-400L)
# ==============================================================================

param(
    [string]$RepoRoot = (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
)

$ErrorActionPreference = "Stop"

Write-Host "--- [1/6] Auditoria Estatica y Proyeccion Preventiva ---" -ForegroundColor Cyan

$violations = @()
$warnings = @()
$extensions = @("*.go", "*.js", "*.css", "*.html", "*.md", "*.ps1", "*.sh")

$trackedFiles = Get-ChildItem -Path $RepoRoot -Recurse -Include $extensions | Where-Object {
    $_.FullName -notmatch "(\.git|vendor|bin|ipvn7\.exe|node_modules|data\\kuzu|wasm_exec\.js)"
}

foreach ($file in $trackedFiles) {
    $lineCount = (Get-Content $file.FullName | Measure-Object -Line).Lines
    $relative = $file.FullName.Replace($RepoRoot, "").TrimStart("\/")

    if ($lineCount -gt 400) {
        $violations += [PSCustomObject]@{ File = $relative; Lines = $lineCount }
    } elseif ($lineCount -ge 320) {
        $warnings += [PSCustomObject]@{ File = $relative; Lines = $lineCount }
    }
}

if ($violations.Count -gt 0) {
    Write-Host "  [FALLA CRITICA] Archivos violando el limite estricto (> 400 lineas):" -ForegroundColor Red
    foreach ($v in $violations) {
        Write-Host "    - $($v.File): $($v.Lines) lineas (Max: 400)" -ForegroundColor Red
    }
} else {
    Write-Host "  [OK] Invariante de 400 lineas cumplido en todos los archivos." -ForegroundColor Green
}

if ($warnings.Count -gt 0) {
    Write-Host "  [PREDICCION/ALERTA] Archivos en zona preventiva (320 - 400 lineas):" -ForegroundColor Yellow
    foreach ($w in $warnings) {
        Write-Host "    - $($w.File): $($w.Lines) lineas (Proximo a modularizar)" -ForegroundColor Yellow
    }
} else {
    Write-Host "  [OK] Ningun archivo se encuentra en la zona preventiva de riesgo." -ForegroundColor Green
}

# 2. Analisis estatico go vet
Write-Host "  -> Ejecutando analisis estatico (go vet)..." -ForegroundColor DarkGray
try {
    $vetOut = go vet ./pkg/... ./cmd/... 2>&1
    Write-Host "  [OK] go vet paso sin observaciones." -ForegroundColor Green
} catch {
    Write-Host "  [ERROR] Fallo en go vet: $_" -ForegroundColor Red
    $violations += [PSCustomObject]@{ File = "go vet"; Lines = 0 }
}

return [PSCustomObject]@{
    Name       = "Static & Predictive"
    Passed     = ($violations.Count -eq 0)
    Violations = $violations
    Warnings   = $warnings
}
