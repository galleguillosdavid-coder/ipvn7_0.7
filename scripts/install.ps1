<#
.SYNOPSIS
    Instalador Universal de IPVN7 Network OS para Windows.
.DESCRIPTION
    Instala los binarios del sistema, driver Wintun (o proxy userspace),
    reglas de Firewall y autoarranque sin requerir obligatoriamente Administrador.
#>
[CmdletBinding()]
param(
    [string]$InstallDir = "",
    [int]$HttpPort = 7070,
    [int]$UdpPort = 7777
)

$ErrorActionPreference = "Stop"

function Write-IPLog {
    param([string]$Message, [string]$Level = "INFO")
    $color = "Green"
    if ($Level -eq "WARN") { $color = "Yellow" }
    if ($Level -eq "ERROR") { $color = "Red" }
    Write-Host "[$Level] $Message" -ForegroundColor $color
}

# 1. Comprobar privilegios de Administrador (Fallback Zero-Admin)
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if ($InstallDir -eq "") {
    if ($isAdmin) {
        $InstallDir = "C:\Program Files\ipvn7"
    } else {
        $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\ipvn7"
    }
}

Write-IPLog "=================================================="
Write-IPLog "  INSTALADOR UNIVERSAL IPVN7 NETWORK OS (WINDOWS) "
Write-IPLog "=================================================="
Write-IPLog "Modo de Privilegios: $(if ($isAdmin) { 'ADMINISTRADOR (Kernel TUN + Wintun)' } else { 'USUARIO ESTÁNDAR (Zero-Admin Userspace)' })"
Write-IPLog "Destino: $InstallDir"

# 2. Crear directorios de destino
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-IPLog "Directorio creado: $InstallDir"
}

$DataDir = Join-Path $InstallDir "data"
if (-not (Test-Path $DataDir)) {
    New-Item -ItemType Directory -Path $DataDir -Force | Out-Null
}

# 3. Copiar binario y Wintun driver
$SourceExe = Join-Path $PSScriptRoot "..\ipvn7.exe"
if (Test-Path $SourceExe) {
    Copy-Item -Path $SourceExe -Destination (Join-Path $InstallDir "ipvn7.exe") -Force
    Write-IPLog "Binario ipvn7.exe instalado en $InstallDir"
} else {
    Write-IPLog "Compilando binario ipvn7.exe..." "INFO"
    go build -trimpath -ldflags="-s -w" -o (Join-Path $InstallDir "ipvn7.exe") ./cmd/ipvn7
    Write-IPLog "Binario compilado e instalado con éxito."
}

# Verificar driver Wintun DLL si dispone de privilegios
$WintunPath = Join-Path $InstallDir "wintun.dll"
if ($isAdmin) {
    $LocalWintun = Join-Path $PSScriptRoot "..\wintun.dll"
    if (Test-Path $LocalWintun) {
        Copy-Item -Path $LocalWintun -Destination $WintunPath -Force
        Write-IPLog "Driver wintun.dll copiado a $InstallDir"
    }
} else {
    Write-IPLog "Instalación en espacio de usuario: operando vía proxy SOCKS5 cuántico (10807) sin privilegios." "INFO"
}

# 4. Configurar reglas de Windows Defender Firewall si es Admin
if ($isAdmin) {
    Write-IPLog "Configurando reglas en Windows Firewall..."
    try {
        netsh advfirewall firewall delete rule name="IPVN7-Mesh-UDP" | Out-Null
        netsh advfirewall firewall add rule name="IPVN7-Mesh-UDP" dir=in action=allow protocol=UDP localport=$UdpPort | Out-Null
        netsh advfirewall firewall delete rule name="IPVN7-Web-HTTP" | Out-Null
        netsh advfirewall firewall add rule name="IPVN7-Web-HTTP" dir=in action=allow protocol=TCP localport=$HttpPort | Out-Null
        Write-IPLog "Reglas de Firewall configuradas: UDP $UdpPort (Mesh), TCP $HttpPort (API/Web)"
    } catch {
        Write-IPLog "Aviso configurando firewall: $_" "WARN"
    }
}

# 5. Registrar autoarranque (Tarea Programada en Admin o Run Key en Usuario)
$ExePath = Join-Path $InstallDir "ipvn7.exe"
$Args = "-port $HttpPort -web-port $HttpPort -udp $UdpPort -data `"$DataDir`" -vpn"

if ($isAdmin) {
    $TaskName = "IPVN7_NetworkOS_Daemon"
    Write-IPLog "Registrando tarea programada de inicio automático con privilegios SYSTEM..."
    try {
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue
        $Action = New-ScheduledTaskAction -Execute $ExePath -Argument $Args -WorkingDirectory $InstallDir
        $Trigger = New-ScheduledTaskTrigger -AtStartup
        $Principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
        $Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
        Register-ScheduledTask -TaskName $TaskName -Action $Action -Trigger $Trigger -Principal $Principal -Settings $Settings -Description "IPVN7 Sovereign Network OS Daemon" | Out-Null
        Write-IPLog "Tarea programada '$TaskName' creada exitosamente."
    } catch {
        Write-IPLog "Aviso en tarea programada: $_" "WARN"
    }
} else {
    Write-IPLog "Registrando autoarranque en perfil de usuario..."
    Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "IPVN7_Daemon" -Value "`"$ExePath`" $Args"
}

# 6. Agregar al PATH del sistema o usuario
$PathScope = if ($isAdmin) { "Machine" } else { "User" }
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", $PathScope)
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", $PathScope)
    Write-IPLog "Agregado $InstallDir al PATH ($PathScope)."
}

Write-IPLog "=================================================="
Write-IPLog "  INSTALACIÓN COMPLETADA EXITOSAMENTE (30 SEGUNDOS)"
Write-IPLog "  Ejecutable: $ExePath                            "
Write-IPLog "  Puerto Mesh UDP: $UdpPort | Dashboard: $HttpPort "
Write-IPLog "=================================================="
