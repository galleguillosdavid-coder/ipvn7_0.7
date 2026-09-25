# ==============================================================================
# start_vpn_i7.ps1 - Lanzador Soberano 1-Clic "VPN I7" con Autodetección Zero-Admin
# ==============================================================================

param(
    [int]$Port = 7777,
    [int]$WebPort = 7070,
    [switch]$NoBrowser
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir

Write-Host ""
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "         IPVN7 - INICIANDO NODO SOBERANO ZERO-FRICTION" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Asegurar existencia de binario
$binPath = Join-Path $RepoRoot "bin\ipvn7.exe"
if (-not (Test-Path $binPath)) {
    Write-Host "[BUILD] Binario no detectado. Compilando bin/ipvn7.exe..." -ForegroundColor Yellow
    Push-Location "$RepoRoot\src"
    try {
        go build -o "$RepoRoot\bin\ipvn7.exe" ./cmd/ipvn7
        Write-Host "[BUILD] Compilacion exitosa." -ForegroundColor Green
    } finally {
        Pop-Location
    }
}

# 2. Detección de Privilegios de Administrador
$currentPrincipal = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
$isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

$cmdArgs = @("-port", $Port, "-web-port", $WebPort)

if ($isAdmin) {
    Write-Host "[MODO] Administrador detectado: Activando interfaz TUN Nativa de Kernel..." -ForegroundColor Green
    $cmdArgs += "-tun-native"
} else {
    Write-Host "[MODO] Usuario estandar detectado: Activando Userspace FastPath (Zero-Privilegios)..." -ForegroundColor Yellow
    Write-Host "       (SOCKS5 Proxy operativo en localhost:10807 sin requerir permisos root/admin)" -ForegroundColor DarkGray
}

# 3. Lanzamiento del Navegador
if (-not $NoBrowser) {
    $url = "http://localhost:$WebPort"
    Write-Host "[UI] Abriendo panel de control en: $url" -ForegroundColor Cyan
    try {
        Start-Process $url
    } catch {
        Write-Host "[AVISO] No se pudo abrir el navegador automaticamente. Abre manualmente: $url" -ForegroundColor DarkGray
    }
}

Write-Host "[READY] Ejecutando Núcleo Universal ipvn7..." -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor DarkGray

# 4. Ejecución del binario
& $binPath @cmdArgs
