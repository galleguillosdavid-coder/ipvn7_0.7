# ==============================================================================
# build_installer.ps1 - Compilador del Instalador Gráfico Autocontenido (Rol K)
# Genera un único .EXE que se instala con doble clic sin línea de comandos
# ==============================================================================

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
$SrcDir = Join-Path $RepoRoot "src"
$DistDir = Join-Path $RepoRoot "dist"
$AssetsDir = Join-Path $SrcDir "cmd\installer\assets"
$OutputFile = Join-Path $DistDir "Instalador_VPN_I7.exe"

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  IPVN7 - COMPILANDO INSTALADOR GRAFICO EXE AUTOCONTENIDO" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Asegurar directorios
if (-not (Test-Path $DistDir)) { New-Item -ItemType Directory -Path $DistDir -Force | Out-Null }
if (-not (Test-Path $AssetsDir)) { New-Item -ItemType Directory -Path $AssetsDir -Force | Out-Null }

# 2. Asegurar binario actualizado
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

# 3. Copiar assets para embed
Copy-Item $binExe (Join-Path $AssetsDir "ipvn7.exe") -Force
$wintunDll = Join-Path $RepoRoot "wintun\bin\amd64\wintun.dll"
if (Test-Path $wintunDll) {
    Copy-Item $wintunDll (Join-Path $AssetsDir "wintun.dll") -Force
}

# 4. Compilar instalador con subsistema GUI de Windows (-H=windowsgui para 0 consolas)
Write-Host "[-] Compilando $OutputFile (Subsistema GUI)..." -ForegroundColor Cyan
Push-Location $SrcDir
try {
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -tags installer -ldflags="-H=windowsgui -s -w" -o $OutputFile ./cmd/installer
    Write-Host " [OK] Instalador compilado con exito." -ForegroundColor Green
} finally {
    $env:CGO_ENABLED = ""
    $env:GOOS = ""
    $env:GOARCH = ""
    Pop-Location
}

# 5. Resumen
$size = (Get-Item $OutputFile).Length / 1MB
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  INSTALADOR AUTOCONTENIDO LISTO: $OutputFile" -ForegroundColor Green
Write-Host "  Tamano: $($size.ToString('F2')) MB (100% independiente, 0 dependencias)" -ForegroundColor White
Write-Host "  Uso: Copiar al Notebook o cualquier PC y hacer doble clic." -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Cyan
