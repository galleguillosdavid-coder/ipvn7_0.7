# ==============================================================================
# build_wasm.ps1 - Compilador reproducible de WebAssembly para IPVN7
# Genera web/ipvn7.wasm y copia wasm_exec.js para compatibilidad universal
# ==============================================================================

$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

Write-Host "--- Compilando IPVN7 WebAssembly (GOOS=js GOARCH=wasm) ---" -ForegroundColor Cyan

[System.Environment]::SetEnvironmentVariable('GOOS', 'js')
[System.Environment]::SetEnvironmentVariable('GOARCH', 'wasm')

go build -o web/ipvn7.wasm ./cmd/ipvn7-wasm
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Fallo en la compilación de ipvn7.wasm" -ForegroundColor Red
    exit 1
}

# Restaurar variables de entorno nativas
[System.Environment]::SetEnvironmentVariable('GOOS', $null)
[System.Environment]::SetEnvironmentVariable('GOARCH', $null)

# Copiar wasm_exec.js si no existe
$goroot = go env GOROOT
$wasmExecSource = Join-Path $goroot 'misc/wasm/wasm_exec.js'
if (-not (Test-Path $wasmExecSource)) {
    $wasmExecSource = Join-Path $goroot 'lib/wasm/wasm_exec.js'
}

if (Test-Path $wasmExecSource) {
    Copy-Item $wasmExecSource -Destination 'web/wasm_exec.js' -Force
    Write-Host "  [OK] wasm_exec.js sincronizado desde Go runtime." -ForegroundColor Green
}

$wasmSize = (Get-Item 'web/ipvn7.wasm').Length / 1MB
Write-Host "  [OK] web/ipvn7.wasm generado exitosamente ($([math]::Round($wasmSize, 2)) MB)." -ForegroundColor Green
Write-Host "--- WebAssembly Listo para uso universal ---" -ForegroundColor Cyan
