# Script de configuración del notebook para acceso remoto completo
# Ejecutar como ADMINISTRADOR en el notebook (192.168.1.106)

Write-Host "=== Configuración del Notebook para IPVN7 ===" -ForegroundColor Green
Write-Host ""

# 1. Habilitar PowerShell Remoting
Write-Host "[1/5] Habilitando PowerShell Remoting..." -ForegroundColor Yellow
try {
    Enable-PSRemoting -Force -ErrorAction Stop
    Write-Host "✓ PowerShell Remoting habilitado" -ForegroundColor Green
} catch {
    Write-Host "✗ Error habilitando PowerShell Remoting: $_" -ForegroundColor Red
}

# 2. Configurar TrustedHosts para permitir conexión desde PC principal
Write-Host "[2/5] Configurando TrustedHosts..." -ForegroundColor Yellow
try {
    Set-Item WSMan:\localhost\Client\TrustedHosts "192.168.1.198" -Force -Concatenate
    Write-Host "✓ 192.168.1.198 agregado a TrustedHosts" -ForegroundColor Green
} catch {
    Write-Host "✗ Error configurando TrustedHosts: $_" -ForegroundColor Red
}

# 3. Configurar Firewall para permitir WinRM
Write-Host "[3/5] Configurando Firewall para WinRM..." -ForegroundColor Yellow
try {
    New-NetFirewallRule -DisplayName "Windows Remote Management" -Direction Inbound -LocalPort 5985 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue
    Write-Host "✓ Regla de firewall creada para puerto 5985" -ForegroundColor Green
} catch {
    Write-Host "✗ Error configurando firewall: $_" -ForegroundColor Red
}

# 4. Configurar Firewall para permitir IPVN7 puertos
Write-Host "[4/5] Configurando Firewall para IPVN7..." -ForegroundColor Yellow
try {
    New-NetFirewallRule -DisplayName "IPVN7 UDP 7001" -Direction Inbound -LocalPort 7001 -Protocol UDP -Action Allow -ErrorAction SilentlyContinue
    New-NetFirewallRule -DisplayName "IPVN7 UDP 7777" -Direction Inbound -LocalPort 7777 -Protocol UDP -Action Allow -ErrorAction SilentlyContinue
    New-NetFirewallRule -DisplayName "IPVN7 Web TCP 8080" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue
    New-NetFirewallRule -DisplayName "IPVN7 Web TCP 7070" -Direction Inbound -LocalPort 7070 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue
    Write-Host "✓ Reglas de firewall creadas para puertos IPVN7 (UDP 7001, 7777 / TCP 8080, 7070)" -ForegroundColor Green
} catch {
    Write-Host "✗ Error configurando firewall IPVN7: $_" -ForegroundColor Red
}

# 5. Crear directorio de trabajo
Write-Host "[5/5] Creando directorio de trabajo..." -ForegroundColor Yellow
$testDir = "C:\Users\Frondabrick\Desktop\ipvn7_test"
if (-not (Test-Path $testDir)) {
    New-Item -ItemType Directory -Path $testDir -Force | Out-Null
    New-Item -ItemType Directory -Path "$testDir\keystore" -Force | Out-Null
    Write-Host "✓ Directorio creado: $testDir" -ForegroundColor Green
} else {
    Write-Host "✓ Directorio ya existe: $testDir" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Configuración completada ===" -ForegroundColor Green
Write-Host ""
Write-Host "Ahora puedes copiar el binario ipvn7.exe al notebook:" -ForegroundColor Cyan
Write-Host "  Copy-Item bin\ipvn7.exe \\192.168.1.106\c$\Users\Frondabrick\Desktop\ipvn7_test\" -ForegroundColor White
Write-Host ""
Write-Host "Para ejecutar IPVN7 en el notebook desde PC principal:" -ForegroundColor Cyan
Write-Host "  Enter-PSSession -ComputerName 192.168.1.106 -Credential Frondabrick" -ForegroundColor White
Write-Host "  cd C:\Users\Frondabrick\Desktop\ipvn7_test" -ForegroundColor White
Write-Host "  .\ipvn7.exe --port 7001 --web-port 8080 --peer 192.168.1.198:7777 --keystore keystore\notebook_identity.key" -ForegroundColor White
