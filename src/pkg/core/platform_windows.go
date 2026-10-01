// Package core provee utilidades de endurecimiento de plataforma Windows
// y limpieza de adaptadores Wintun huérfanos, rescatado de Ipv7-4 (core/security_windows.go)
// y adaptado para ipvn7 v0.7.
package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// BuildSchtasksArgs genera la línea de argumentos canónica para elevación en inicio de sesión
func BuildSchtasksArgs(taskName, exePath string) []string {
	return []string{
		"/create",
		"/f",
		"/sc", "onlogon",
		"/rl", "highest",
		"/tn", taskName,
		"/tr", fmt.Sprintf("\"%s\"", exePath),
	}
}

// RegisterLogonTask registra el binario en el Programador de Tareas de Windows sin UAC prompt
func RegisterLogonTask(taskName, exePath string, runner CmdRunner) error {
	if runtime.GOOS != "windows" && runner == nil {
		return fmt.Errorf("platform: schtasks solo es compatible con Windows (actual: %s)", runtime.GOOS)
	}
	if runner == nil {
		runner = DefaultCmdRunner
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := BuildSchtasksArgs(taskName, exePath)
	out, err := runner(ctx, "schtasks.exe", args...)
	if err != nil {
		return fmt.Errorf("schtasks error: %w (salida: %s)", err, string(out))
	}
	return nil
}

// PurgeOrphanWintunAdapters busca adaptadores de red huérfanos que coincidan con el prefijo y los purga
func PurgeOrphanWintunAdapters(prefix string, runner CmdRunner) (int, error) {
	if runner == nil {
		runner = DefaultCmdRunner
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Obtener lista de adaptadores via netsh interface show interface
	out, err := runner(ctx, "netsh", "interface", "show", "interface")
	if err != nil {
		return 0, fmt.Errorf("platform: fallo al consultar interfaces de red: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	purged := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Si la línea contiene el prefijo del adaptador huérfano (ej: ieu0, ipvn7_tun)
		if strings.Contains(trimmed, prefix) {
			// Extraer nombre de la interfaz (última columna de netsh)
			fields := strings.Fields(trimmed)
			if len(fields) >= 4 {
				ifaceName := fields[len(fields)-1]
				// Intentar deshabilitar / eliminar la interfaz huérfana
				_, _ = runner(ctx, "netsh", "interface", "set", "interface", ifaceName, "admin=disable")
				purged++
			}
		}
	}

	return purged, nil
}

// SetWindowsUserProxy activa el proxy de usuario de Windows para capturar tráfico sin privilegios admin
func SetWindowsUserProxy(proxyServer string, runner CmdRunner) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	if runner == nil {
		runner = DefaultCmdRunner
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	_, err := runner(ctx, "reg.exe", "add", regKey, "/v", "ProxyServer", "/t", "REG_SZ", "/d", proxyServer, "/f")
	if err != nil {
		return fmt.Errorf("error configurando ProxyServer: %w", err)
	}
	_, _ = runner(ctx, "reg.exe", "add", regKey, "/v", "ProxyOverride", "/t", "REG_SZ", "/d", "<local>;localhost;127.*;10.*;192.168.*;*.whatsapp.net;*.whatsapp.com;*.facebook.com;*.fbcdn.net", "/f")
	_, err = runner(ctx, "reg.exe", "add", regKey, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "1", "/f")
	if err != nil {
		return fmt.Errorf("error activando ProxyEnable: %w", err)
	}

	// Watchdog inmune a cierres forzosos: si el proceso es terminado por Task Manager o cierre de consola,
	// restaura la conexión directa de Windows de forma automática e inmediata.
	watchdogCmd := fmt.Sprintf("Wait-Process -Id %d -ErrorAction SilentlyContinue; reg add '%s' /v ProxyEnable /t REG_DWORD /d 0 /f", os.Getpid(), regKey)
	watchdog := exec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", watchdogCmd)
	_ = watchdog.Start()

	return nil
}

// ClearWindowsUserProxy desactiva el proxy de usuario de Windows y restaura la conexión directa
func ClearWindowsUserProxy(runner CmdRunner) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	if runner == nil {
		runner = DefaultCmdRunner
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	_, err := runner(ctx, "reg.exe", "add", regKey, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f")
	return err
}

// TerminateConflictingProcesses busca y termina procesos residuales o zombies (vpi7, ipvn7 previo)
func TerminateConflictingProcesses(runner CmdRunner) {
	if runtime.GOOS != "windows" {
		return
	}
	if runner == nil {
		runner = DefaultCmdRunner
	}
	myPID := os.Getpid()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Terminar cualquier zombie residual histórico de vpi7
	_, _ = runner(ctx, "taskkill.exe", "/F", "/IM", "vpi7.exe")

	// Terminar cualquier otra instancia de ipvn7 con PID distinto al actual
	filter := fmt.Sprintf("PID ne %d", myPID)
	_, _ = runner(ctx, "taskkill.exe", "/F", "/IM", "ipvn7.exe", "/FI", filter)
}

