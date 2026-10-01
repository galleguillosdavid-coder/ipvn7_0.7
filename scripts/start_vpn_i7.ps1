# ==============================================================================
# start_vpn_i7.ps1 - Lanzador Soberano 1-Clic "VPN I7" con Autodetección Zero-Admin
# ==============================================================================

param(
    [int]$Port = 0,
    [int]$WebPort = 0,
    [string]$Peer = "",
    [switch]$Notebook,
    [switch]$NoBrowser,
    [switch]$NoElevate,
    [switch]$CaptureWeb,
    [switch]$Reset
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
Set-Location $RepoRoot

$localIPs = @(Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty IPAddress)
$isNotebook = $Notebook -or ($env:COMPUTERNAME -eq "Dvd") -or ($localIPs -contains "192.168.1.106")
if ($Port -eq 0) { $Port = if ($isNotebook) { 7001 } else { 7777 } }
if ($WebPort -eq 0) { $WebPort = if ($isNotebook) { 8080 } else { 7070 } }
$socksPort = if ($isNotebook) { 10808 } else { 10807 }
if ($Peer -eq "") { $Peer = if ($isNotebook) { "192.168.1.198:7777" } else { "192.168.1.106:7001" } }

# Función de restauración inmediata de red directa de Windows
function Reset-DirectInternet {
    Write-Host "[*] Asegurando conexion directa a Internet en Windows (Proxy Desactivado)..." -ForegroundColor Green
    reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings" /v ProxyEnable /t REG_DWORD /d 0 /f 2>$null | Out-Null
}

# Siempre asegurar que el proxy esté desactivado al inicio para proteger el internet del usuario
Reset-DirectInternet

if ($Reset) {
    exit 0
}

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

# 2. Detección y Auto-Elevación a Administrador (Mandato AGENTS.md)
$currentPrincipal = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
$isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin -and -not $NoElevate) {
    Write-Host "[*] Intentando auto-ejecución como Administrador (Mandato AGENTS.md: Wintun L3 TUN)..." -ForegroundColor Cyan
    try {
        Start-Process powershell -ArgumentList "-NoExit -ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Verb RunAs
        exit 0
    } catch {
        Write-Host "[!] Permisos de Administrador no concedidos o cancelados." -ForegroundColor Yellow
        Write-Host "[*] Fallback automático: Continuando en Modo Usuario (Zero-Admin)..." -ForegroundColor Green
    }
}

$keystoreFile = Join-Path $RepoRoot "keystore\node_identity.key"
$cmdArgs = @("-port", "$Port", "-web-port", "$WebPort", "-keystore", "$keystoreFile", "-peer", "$Peer")

if ($isAdmin) {
    Write-Host "[MODO] Administrador activo: Interfaz TUN Nativa de Kernel (Wintun L3)..." -ForegroundColor Green
    Write-Host "       (Soporte total para TCP, UDP, Pings, Speedtest y Gaming nativo)" -ForegroundColor DarkGray
    $cmdArgs += @("-tun", "-no-elevate")
} else {
    Write-Host "[MODO] Usuario estándar: Enrutamiento Soberano con Botón 1-Clic..." -ForegroundColor Green
    Write-Host "       (Panel visual de 3 estados: Verde, Amarillo, Rojo con control de WhatsApp)" -ForegroundColor DarkGray
    $cmdArgs += @("-socks5", "$socksPort", "-no-elevate")
    if ($CaptureWeb) {
        $cmdArgs += @("-capture-web")
    }
}

Write-Host "[READY] Ejecutando Núcleo Universal ipvn7 y abriendo Panel..." -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor DarkGray

# 3. Ejecución del binario con Watchdog desacoplado a prueba de fallos
try {
    $proc = Start-Process -FilePath $binPath -WorkingDirectory $RepoRoot -ArgumentList $cmdArgs -PassThru -NoNewWindow
    
    # Abrir ventana tipo aplicación con fallback garantizado (Rol K)
    if (-not $NoBrowser) {
        Start-Sleep -Milliseconds 800
        Write-Host "[UI] Abriendo panel interactivo en: http://127.0.0.1:$WebPort" -ForegroundColor Cyan
        $opened = $false
        try {
            $edgeApp = Start-Process "msedge.exe" -ArgumentList "--app=http://127.0.0.1:$WebPort", "--window-size=440,680" -PassThru -ErrorAction Stop
            if ($edgeApp) { $opened = $true }
        } catch {}

        if (-not $opened) {
            try {
                $chrApp = Start-Process "chrome.exe" -ArgumentList "--app=http://127.0.0.1:$WebPort", "--window-size=440,680" -PassThru -ErrorAction Stop
                if ($chrApp) { $opened = $true }
            } catch {}
        }

        if (-not $opened) {
            # Fallback infalible vía explorer.exe (abre navegador predeterminado del usuario)
            Start-Process "explorer.exe" -ArgumentList "http://127.0.0.1:$WebPort" -ErrorAction SilentlyContinue
        }
    }

    # Watchdog silencioso e independiente: si ipvn7 se cierra o se mata desde el Administrador de Tareas,
    # restaura ProxyEnable = 0 en menos de 1 segundo sin dejar a Windows sin internet.
    $watchdogScript = "Wait-Process -Id $($proc.Id) -ErrorAction SilentlyContinue; reg add 'HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings' /v ProxyEnable /t REG_DWORD /d 0 /f 2>`$null | Out-Null"
    Start-Process powershell -ArgumentList "-NoProfile", "-WindowStyle", "Hidden", "-Command", $watchdogScript -WindowStyle Hidden | Out-Null

    $proc.WaitForExit()
} finally {
    Reset-DirectInternet
}
