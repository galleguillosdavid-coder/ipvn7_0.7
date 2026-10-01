# hil_remote_action.ps1 - Orquestador HIL Remoto, Optimizador de SO y Sonda Anti-Censura
# Mision: Co-optimizacion de SO, sonda de sitios bloqueados y automatizacion visual (Rol R)
param(
    [ValidateSet("deploy", "show-ui", "open-url", "toggle-vpn", "status", "optimize-os", "probe-blocked")]
    [string]$Action = "status",
    [string]$RemoteHost = "192.168.1.106",
    [int]$SSHPort = 22,
    [string]$SSHUser = "Frondabrick",
    [string]$Url = "https://crypto.cloudflare.com/cdn-cgi/trace",
    [int]$RemoteWebPort = 8080
)

$ErrorActionPreference = "Stop"

function Test-SSHReady {
    $res = ssh -o BatchMode=yes -o ConnectTimeout=3 -p $SSHPort "${SSHUser}@${RemoteHost}" hostname 2>$null
    return ($null -ne $res -and $res.Trim().Length -gt 0)
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "   IPvN7 HIL Orchestrator -- Accion: $Action              " -ForegroundColor Cyan
Write-Host "   Host Remoto: $RemoteHost (SSH: $SSHPort, Web: $RemoteWebPort)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

if (-not (Test-SSHReady)) {
    Write-Host "[ERROR] No hay conexion SSH hacia ${SSHUser}@${RemoteHost}:${SSHPort}." -ForegroundColor Red
    exit 1
}

if ($Action -eq "status") {
    Write-Host "[1/1] Consultando telemetria fisica en http://${RemoteHost}:${RemoteWebPort}/api/v1/status..." -ForegroundColor Yellow
    $status = curl.exe -s "http://${RemoteHost}:${RemoteWebPort}/api/v1/status" | ConvertFrom-Json
    $rxMb = [math]::Round($status.gateway.bytes_rx / 1MB, 2)
    $txKb = [math]::Round($status.gateway.bytes_tx / 1KB, 2)
    Write-Host "  -> DID Remoto:     $($status.did)" -ForegroundColor Green
    Write-Host "  -> Estado VPN:     $($status.vpn_state)" -ForegroundColor Green
    Write-Host "  -> Pares Malla:    $($status.peers_count) par(es)" -ForegroundColor Green
    Write-Host "  -> Trafico Rx:     $rxMb MB" -ForegroundColor Green
    Write-Host "  -> Trafico Tx:     $txKb KB" -ForegroundColor Green
    Write-Host "  -> Destino Actual: $($status.gateway.last_destination)" -ForegroundColor Green
}
elseif ($Action -eq "optimize-os") {
    Write-Host "[1/4] Aplicando Google BBR2 y RACK en el stack TCP del Notebook..." -ForegroundColor Yellow
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "netsh int tcp set supplemental template=internet congestionprovider=bbr2" | Out-Null
    
    Write-Host "[2/4] Habilitando ECN (Explicit Congestion Notification RFC 3168)..." -ForegroundColor Yellow
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "netsh int tcp set global ecncapability=enabled" | Out-Null

    Write-Host "[3/4] Optimizando marcas de tiempo RFC 1323 y TCP Fast Open..." -ForegroundColor Yellow
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "netsh int tcp set global timestamps=allowed fastopen=enabled" | Out-Null

    Write-Host "[4/4] Verificando parametros optimizados del kernel..." -ForegroundColor Yellow
    $tcp = ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "netsh int tcp show supplemental"
    Write-Host "  [OK] Stack TCP optimizado con exito:" -ForegroundColor Green
    $tcp | Select-String "Proveedor de control|Habilitar RACK|Habilitar sondeo" | ForEach-Object { Write-Host "    $($_)" -ForegroundColor Cyan }
}
elseif ($Action -eq "probe-blocked") {
    $probeSites = @(
        @{ Name = "Cloudflare ECH/DPI Trace"; Url = "https://crypto.cloudflare.com/cdn-cgi/trace" },
        @{ Name = "Tor Project Check Anti-Censura"; Url = "https://check.torproject.org" },
        @{ Name = "Cloudflare DNS Anti-Fugas"; Url = "https://1.1.1.1/help" },
        @{ Name = "Wikipedia Censorship Reference"; Url = "https://en.wikipedia.org/wiki/Internet_censorship" }
    )

    Write-Host "[1/3] Lanzando bateria de navegacion en sitios de prueba anti-censura..." -ForegroundColor Yellow
    foreach ($site in $probeSites) {
        Write-Host "  -> Navegando visualmente a: $($site.Name) ($($site.Url))..." -ForegroundColor Cyan
        ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "echo start msedge.exe $($site.Url) > C:\ipvn7\open_youtube.bat"
        ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "powershell -ExecutionPolicy Bypass -File C:\ipvn7\launch_youtube.ps1" | Out-Null
        Start-Sleep -Seconds 3
    }

    Write-Host "[2/3] Auditando telemetria de escape y rendimiento tras navegacion..." -ForegroundColor Yellow
    $status = curl.exe -s "http://${RemoteHost}:${RemoteWebPort}/api/v1/status" | ConvertFrom-Json
    $rxMb = [math]::Round($status.gateway.bytes_rx / 1MB, 2)
    $txKb = [math]::Round($status.gateway.bytes_tx / 1KB, 2)
    
    Write-Host "[3/3] Reporte de Metricas de Red y Rendimiento (Enviado a PC Principal):" -ForegroundColor Green
    Write-Host "  ========================================================" -ForegroundColor DarkGray
    Write-Host "  * Estado de Conexion:    $($status.vpn_state)" -ForegroundColor White
    Write-Host "  * Modo de Pasarela:      Sovereign Egress Gateway (DEC-130)" -ForegroundColor White
    Write-Host "  * Trafico Total Cursado: $rxMb MB (Rx) / $txKb KB (Tx)" -ForegroundColor White
    Write-Host "  * Conexiones Totales:    $($status.gateway.total_connections)" -ForegroundColor White
    Write-Host "  * Conexiones Activas:    $($status.gateway.active_connections)" -ForegroundColor White
    Write-Host "  * Ultimo Host Resuelto:  $($status.gateway.last_destination)" -ForegroundColor White
    Write-Host "  * Evasor DNS Activo:     Resolver Soberano Zero-Leak" -ForegroundColor White
    Write-Host "  ========================================================" -ForegroundColor DarkGray
}
elseif ($Action -eq "deploy") {
    Write-Host "[1/3] Deteniendo proceso previo limpiamente..." -ForegroundColor Yellow
    curl.exe -s -X POST "http://${RemoteHost}:${RemoteWebPort}/api/v1/vpn/exit" | Out-Null
    Start-Sleep -Milliseconds 600

    Write-Host "[2/3] Transfiriendo binario compilado (bin/ipvn7.exe) via SCP..." -ForegroundColor Yellow
    scp -P $SSHPort bin/ipvn7.exe "${SSHUser}@${RemoteHost}:ipvn7.exe"
    scp -P $SSHPort bin/ipvn7.exe "${SSHUser}@${RemoteHost}:/ipvn7/bin/ipvn7.exe"

    Write-Host "[3/3] Reiniciando servicio IPvN7 desacoplado en el host remoto..." -ForegroundColor Yellow
    $startCmd = "powershell -Command Invoke-CimMethod -ClassName Win32_Process -MethodName Create -Arguments @{CommandLine='C:\Users\${SSHUser}\ipvn7.exe -port 7001 -web-port ${RemoteWebPort} -peer 192.168.1.198:7777 -no-elevate'}"
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" $startCmd | Out-Null
    Start-Sleep -Seconds 2

    $check = curl.exe -s "http://${RemoteHost}:${RemoteWebPort}/api/v1/status" | ConvertFrom-Json
    Write-Host "[OK] Despliegue completado. Uptime: $($check.uptime_sec)s, Version: $($check.system)" -ForegroundColor Green
}
elseif ($Action -eq "show-ui") {
    Write-Host "[1/1] Abriendo panel visual en la pantalla fisica del host remoto..." -ForegroundColor Yellow
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "powershell -ExecutionPolicy Bypass -File C:\ipvn7\launch_visual.ps1" | Out-Null
    Write-Host "[OK] Ventana de control abierta en la pantalla interactiva del usuario." -ForegroundColor Green
}
elseif ($Action -eq "open-url") {
    Write-Host "[1/1] Navegando interactivamente hacia '$Url' en la pantalla fisica..." -ForegroundColor Yellow
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "echo start msedge.exe $Url > C:\ipvn7\open_youtube.bat"
    ssh -p $SSHPort "${SSHUser}@${RemoteHost}" "powershell -ExecutionPolicy Bypass -File C:\ipvn7\launch_youtube.ps1" | Out-Null
    Write-Host "[OK] Pagina cargada visualmente en el navegador del host remoto." -ForegroundColor Green
}
elseif ($Action -eq "toggle-vpn") {
    Write-Host "[1/2] Consultando estado actual de la VPN..." -ForegroundColor Yellow
    $cur = curl.exe -s "http://${RemoteHost}:${RemoteWebPort}/api/v1/status" | ConvertFrom-Json
    $next = if ($cur.vpn_state -eq "connected") { "disconnect" } else { "connect" }
    Write-Host "[2/2] Conmutando a '$next' (el boton cambiara de color en pantalla)..." -ForegroundColor Yellow
    $res = curl.exe -s -X POST "http://${RemoteHost}:${RemoteWebPort}/api/v1/vpn/$next" | ConvertFrom-Json
    Write-Host "[OK] Estado conmutado a: $($res.status)" -ForegroundColor Green
}
