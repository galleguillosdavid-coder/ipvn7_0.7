# Script para configurar PC principal para conectar al notebook
# EJECUTAR COMO ADMINISTRADOR en la PC principal (192.168.1.198)

Write-Host "=== Configuración PC Principal para conectar al Notebook ===" -ForegroundColor Green
Write-Host ""

# Configurar TrustedHosts
Write-Host "[1/1] Configurando TrustedHosts para 192.168.1.106..." -ForegroundColor Yellow
try {
    Set-Item WSMan:\localhost\Client\TrustedHosts 192.168.1.106 -Force
    Write-Host "✓ 192.168.1.106 agregado a TrustedHosts" -ForegroundColor Green
} catch {
    Write-Host "✗ Error: $_" -ForegroundColor Red
    Write-Host "Debes ejecutar este script como ADMINISTRADOR" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Configuración completada ===" -ForegroundColor Green
Write-Host ""
Write-Host "Ahora puedes conectar al notebook:" -ForegroundColor Cyan
Write-Host "  Enter-PSSession -ComputerName 192.168.1.106 -Credential Frondabrick" -ForegroundColor White
