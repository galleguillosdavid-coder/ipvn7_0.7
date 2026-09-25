@echo off
cd /d "%~dp0"
echo [*] Iniciando IPVN7 Master Node...
start /b bin\ipvn7.exe -port 7777 -web-port 7070 -vpn
echo [OK] Nodo Maestro iniciado en puerto UDP 7777 (Web: http://localhost:7070)
