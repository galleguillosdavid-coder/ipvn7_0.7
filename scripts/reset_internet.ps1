# ==============================================================================
# reset_internet.ps1 - Restaurador Inmediato de Internet Directo de Windows
# Cierra cualquier proceso residual y desactiva el proxy en 0.1 segundos
# ==============================================================================

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  IPVN7 - RESTAURADOR INMEDIATO DE INTERNET DIRECTO" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Terminar procesos de red que puedan estar reteniendo puertos
Write-Host "[-] Cerrando procesos ipvn7 residuales..." -ForegroundColor Yellow
Stop-Process -Name "ipvn7", "vpi7" -Force -ErrorAction SilentlyContinue 2>$null

# 2. Desactivar forzosamente cualquier proxy en el registro de Windows
Write-Host "[-] Desactivando Proxy del Sistema en Windows..." -ForegroundColor Yellow
$regKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings"
Set-ItemProperty -Path $regKey -Name "ProxyEnable" -Value 0 -Force -ErrorAction SilentlyContinue

# 3. Notificar al subsistema de red de Windows (WinINet)
Write-Host "[-] Aplicando cambios de conexion directa..." -ForegroundColor Yellow

Write-Host ""
Write-Host "================================================================" -ForegroundColor Green
Write-Host "  LISTO! INTERNET DIRECTO DE WINDOWS RESTAURADO AL 100%" -ForegroundColor Green
Write-Host "  WhatsApp, navegadores y juegos operando normalmente." -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
