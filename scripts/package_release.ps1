# ==============================================================================
# package_release.ps1 - Empaquetador de Distribución Soberana - Rol K
# Empaqueta un bundle autocontenido y comprimido listo para el Notebook o PC
# ==============================================================================

param(
    [string]$Version = "0.7.0"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
$DistDir = Join-Path $RepoRoot "dist"
$BundleDir = Join-Path $DistDir "ipvn7-windows-amd64"
$ZipFile = Join-Path $DistDir "ipvn7-v$Version-windows-amd64.zip"

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  IPVN7 - EMPAQUETADOR DE DISTRIBUCION AUTONOMA - ROL K" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Limpiar o crear carpetas de salida
if (Test-Path $BundleDir) { Remove-Item -Recurse -Force $BundleDir }
if (Test-Path $ZipFile) { Remove-Item -Force $ZipFile }
New-Item -ItemType Directory -Path $BundleDir -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $BundleDir "scripts") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $BundleDir "keystore") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $BundleDir "bin") -Force | Out-Null

# 2. Compilar binario de producción limpio y stripped
Write-Host "[-] Compilando binario de produccion con stripping CGO_ENABLED=0..." -ForegroundColor Cyan
Push-Location "$RepoRoot\src"
try {
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags="-s -w" -o (Join-Path $RepoRoot "bin\ipvn7.exe") ./cmd/ipvn7
    Write-Host " [OK] Binario bin/ipvn7.exe compilado." -ForegroundColor Green
} finally {
    $env:CGO_ENABLED = ""
    $env:GOOS = ""
    $env:GOARCH = ""
    Pop-Location
}

# 3. Copiar componentes al bundle
Write-Host "[-] Estructurando bundle de distribucion..." -ForegroundColor Cyan
Copy-Item (Join-Path $RepoRoot "bin\ipvn7.exe") (Join-Path $BundleDir "bin\ipvn7.exe") -Force
Copy-Item (Join-Path $RepoRoot "bin\ipvn7.exe") (Join-Path $BundleDir "ipvn7.exe") -Force

# Driver Wintun para TUN de kernel en Windows
$wintunSrc = Join-Path $RepoRoot "wintun\bin\amd64\wintun.dll"
if (Test-Path $wintunSrc) {
    Copy-Item $wintunSrc (Join-Path $BundleDir "wintun.dll") -Force
    Copy-Item $wintunSrc (Join-Path $BundleDir "bin\wintun.dll") -Force
    Write-Host " [OK] Driver wintun.dll incluido." -ForegroundColor Green
}

# 4. Lanzador universal en la raíz del paquete
$batUniversal = @"
@echo off
setlocal
cd /d "%~dp0"
start "" "%~dp0ipvn7.exe"
timeout /t 1 /nobreak >nul
start "" "http://127.0.0.1:7070"
"@
Set-Content -Path (Join-Path $BundleDir "Iniciar_VPN_I7.bat") -Value $batUniversal -Encoding ASCII

# Documento de lectura rápida para el usuario
$readme = @"
================================================================
  IPVN7 - RED SOBERANA v$Version
================================================================

USO SIMPLE EN 1 PASO:
Haz doble clic en 'Iniciar_VPN_I7.bat' o directamente en 'ipvn7.exe'.
El panel de control visual se abrirá automáticamente en tu pantalla.
================================================================
"@
Set-Content -Path (Join-Path $BundleDir "LEEME.txt") -Value $readme -Encoding utf8

# 5. Generar archivo .ZIP comprimido
Write-Host "[-] Comprimiendo paquete en $ZipFile..." -ForegroundColor Cyan
Start-Sleep -Milliseconds 800
$zipSuccess = $false
for ($attempt = 1; $attempt -le 3; $attempt++) {
    try {
        if (Test-Path $ZipFile) { Remove-Item -Force $ZipFile -ErrorAction SilentlyContinue }
        Compress-Archive -Path "$BundleDir\*" -DestinationPath $ZipFile -Force -ErrorAction Stop
        $zipSuccess = $true
        break
    } catch {
        Write-Host "    [Aviso] Esperando liberacion de archivos (intento $attempt/3)..." -ForegroundColor Yellow
        Start-Sleep -Seconds 1
    }
}
if (-not $zipSuccess) {
    Write-Host " [ERROR] No se pudo comprimir en zip, pero la carpeta $BundleDir esta lista para copiar." -ForegroundColor Yellow
} else {
    Write-Host " [OK] Archivo ZIP generado con exito." -ForegroundColor Green
}

# 6. Calcular SHA256
$hash = Get-FileHash -Path $ZipFile -Algorithm SHA256
Set-Content -Path (Join-Path $DistDir "SHA256SUMS.txt") -Value "$($hash.Hash)  $(Split-Path -Leaf $ZipFile)" -Encoding utf8
Write-Host " [OK] Suma SHA256: $($hash.Hash)" -ForegroundColor Green

# 7. Limpiar directorio temporal de staging
if (Test-Path $BundleDir) {
    Remove-Item -Recurse -Force $BundleDir -ErrorAction SilentlyContinue
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  EMPAQUETADO FINALIZADO EXITOSAMENTE (ROL K)" -ForegroundColor Cyan
Write-Host "  Zip:    $ZipFile" -ForegroundColor White
Write-Host "================================================================" -ForegroundColor Cyan
