package core

import (
	"context"
	"os/exec"
)

// CmdRunner abstrae la ejecución de comandos para permitir pruebas deterministas
type CmdRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// DefaultCmdRunner ejecuta comandos reales del sistema operativo
func DefaultCmdRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
