@echo off
setlocal
cd /d "%~dp0.."
powershell -ExecutionPolicy Bypass -NoProfile -File "%~dp0start_vpn_i7.ps1" %*
pause
