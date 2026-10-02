@echo off
rem ==============================================================================
rem start_autodaemon.bat - Lanzador 1-Clic del Daemon de Autoejecución IPVN7
rem ==============================================================================
setlocal
cd /d "%~dp0\.."

echo Iniciando Daemon de Autoejecucion IPVN7 en segundo plano...
powershell -NoProfile -ExecutionPolicy Bypass -File "scripts\daemon_autoejecucion.ps1" start -IntervalSeconds 300 -Background

echo.
powershell -NoProfile -ExecutionPolicy Bypass -File "scripts\daemon_autoejecucion.ps1" status

pause
