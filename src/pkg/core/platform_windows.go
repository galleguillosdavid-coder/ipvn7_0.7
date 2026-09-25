// Package core provee utilidades de endurecimiento de plataforma Windows
// y limpieza de adaptadores Wintun huérfanos, rescatado de Ipv7-4 (core/security_windows.go)
// y adaptado para ipvn7 v0.7.
package core

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// CmdRunner abstrae la ejecución de comandos para permitir pruebas deterministas
type CmdRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// DefaultCmdRunner ejecuta comandos reales del sistema operativo
func DefaultCmdRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

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
