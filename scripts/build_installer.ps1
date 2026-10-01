# ==============================================================================
# build_installer.ps1 - Compilador del Instalador Windows Desacoplado (Arquitectura B)
# Genera el ejecutable instalador sin contaminar el árbol fuente con binarios embed
# ==============================================================================

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
$SrcDir = Join-Path $RepoRoot "src"
$DistDir = Join-Path $RepoRoot "dist"
$OutputFile = Join-Path $DistDir "Instalador_VPN_I7.exe"

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  IPVN7 - COMPILANDO INSTALADOR WINDOWS (ARQUITECTURA B)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Asegurar directorios de distribución
if (-not (Test-Path $DistDir)) { New-Item -ItemType Directory -Path $DistDir -Force | Out-Null }

# 2. Asegurar binario principal de producción
$binExe = Join-Path $RepoRoot "bin\ipvn7.exe"
if (-not (Test-Path $binExe)) {
    Write-Host "[-] Compilando bin/ipvn7.exe..." -ForegroundColor Cyan
    Push-Location $SrcDir
    try {
        $env:CGO_ENABLED = "0"
        go build -trimpath -ldflags="-s -w" -o $binExe ./cmd/ipvn7
    } finally {
        $env:CGO_ENABLED = ""
        Pop-Location
    }
}

# 3. Compilar instalador con subsistema GUI de Windows (-H=windowsgui)
Write-Host "[-] Compilando $OutputFile (Subsistema GUI desacoplado)..." -ForegroundColor Cyan
Push-Location $SrcDir
try {
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags="-H=windowsgui -s -w" -o $OutputFile ./cmd/installer
    Write-Host " [OK] Instalador compilado con éxito." -ForegroundColor Green
} finally {
    $env:CGO_ENABLED = ""
    $env:GOOS = ""
    $env:GOARCH = ""
    Pop-Location
}

# 4. Copiar artefactos adyacentes a dist/ para distribución lista
Copy-Item $binExe (Join-Path $DistDir "ipvn7.exe") -Force
$wintunDll = Join-Path $RepoRoot "wintun\bin\amd64\wintun.dll"
if (Test-Path $wintunDll) {
    Copy-Item $wintunDll (Join-Path $DistDir "wintun.dll") -Force
}

# 5. Resumen
$size = (Get-Item $OutputFile).Length / 1MB
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  INSTALADOR DESACOPLADO LISTO: $OutputFile" -ForegroundColor Green
Write-Host "  Tamaño Instalador: $($size.ToString('F2')) MB" -ForegroundColor White
Write-Host "  Carpeta de Distribución: $DistDir (contiene instalador + binario + wintun)" -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Cyan
