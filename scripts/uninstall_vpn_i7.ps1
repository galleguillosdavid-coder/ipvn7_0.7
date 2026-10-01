# ==============================================================================
# uninstall_vpn_i7.ps1 - Desinstalador Limpio de VPN I7
# ==============================================================================
$ErrorActionPreference = "SilentlyContinue"

Write-Host "[-] Deteniendo procesos ipvn7..." -ForegroundColor Cyan
Stop-Process -Name "ipvn7" -Force

Write-Host "[-] Restaurando proxy de Windows a conexion directa..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyEnable" -Value 0
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name "ProxyServer" -Value ""

Write-Host "[-] Eliminando accesos directos del Escritorio..." -ForegroundColor Cyan
$desktop = [Environment]::GetFolderPath("Desktop")
Get-ChildItem -Path $desktop -Filter "*VPN I7*.lnk" | Remove-Item -Force

Write-Host "[-] Eliminando directorio de instalacion en LocalAppData..." -ForegroundColor Cyan
$localApp = [Environment]::GetFolderPath("LocalApplicationData")
$installDir = Join-Path $localApp "IPVN7"
if (Test-Path $installDir) {
    Remove-Item -Path $installDir -Recurse -Force
}

Write-Host "[OK] Desinstalacion completada con exito." -ForegroundColor Green
Write-Host "    Tu sistema quedo limpio y listo para instalar la nueva version." -ForegroundColor Yellow
