package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// AutoElevationManager maneja la elevación automática silenciosa tipo Chrome
type AutoElevationManager struct {
	serviceName string
	servicePath string
}

// NewAutoElevationManager crea el gestor de elevación automática
func NewAutoElevationManager() *AutoElevationManager {
	exePath, _ := os.Executable()
	return &AutoElevationManager{
		serviceName: "IPVN7TunnelService",
		servicePath: exePath,
	}
}

// IsElevated verifica si el proceso corre con permisos de administrador (agnóstico al idioma)
func IsElevated() bool {
	if runtime.GOOS != "windows" {
		return os.Geteuid() == 0
	}
	cmd := exec.Command("net", "session")
	return cmd.Run() == nil
}

// RequestSelfElevation intenta relanzar el proceso actual como Administrador mediante UAC
// Garantiza una ventana de consola visible, interactiva y permanente para depuración en tiempo real.
func RequestSelfElevation(args []string) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	cwd, _ := os.Getwd()
	argStr := strings.Join(args, " ")

	// Ejecutar a través de cmd.exe /k con título para que la ventana sea 100% visible e interactiva,
	// y no se cierre abruptamente ante cualquier salida, facilitando la depuración total al usuario.
	psCmd := fmt.Sprintf("Start-Process -FilePath 'cmd.exe' -ArgumentList '/k title IPVN7 Sovereign Network OS (Kernel L3) & \"%s\" %s' -WorkingDirectory '%s' -Verb RunAs", exe, argStr, cwd)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Normal", "-Command", psCmd)
	return cmd.Run() == nil
}

// HasAdminPrivileges verifica si la app corre con admin
func (ae *AutoElevationManager) HasAdminPrivileges() bool {
	return IsElevated()
}

// InstallTUNService instala el servicio TUN silenciosamente (solo una vez)
func (ae *AutoElevationManager) InstallTUNService() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-elevation solo soportado en Windows")
	}

	if ae.HasAdminPrivileges() {
		return ae.installWindowsService()
	}
	
	// Si no tiene admin, intentar auto-elevar silenciosamente
	return ae.autoElevateAndInstall()
}

// installWindowsService instala el servicio Windows con SYSTEM
func (ae *AutoElevationManager) installWindowsService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Usar sc.exe para crear servicio
	args := []string{
		"create", ae.serviceName,
		"binPath=", fmt.Sprintf("\"%s\" -service-mode", ae.servicePath),
		"start=", "auto",
		"DisplayName=", "IPVN7 Tunnel Service",
	}
	
	cmd := exec.CommandContext(ctx, "sc.exe", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creando servicio: %w (salida: %s)", err, string(output))
	}

	// Configurar el servicio para correr como SYSTEM
	args = []string{
		"config", ae.serviceName,
		"obj=", "LocalSystem",
	}
	
	cmd = exec.CommandContext(ctx, "sc.exe", args...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error configurando servicio: %w (salida: %s)", err, string(output))
	}

	// Iniciar el servicio
	args = []string{"start", ae.serviceName}
	cmd = exec.CommandContext(ctx, "sc.exe", args...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error iniciando servicio: %w (salida: %s)", err, string(output))
	}

	return nil
}

// autoElevateAndInstall eleva silenciosamente e instala el servicio
func (ae *AutoElevationManager) autoElevateAndInstall() error {
	// Crear script temporal PowerShell
	scriptPath := filepath.Join(os.TempDir(), "ipvn7_install_service.ps1")
	scriptContent := fmt.Sprintf(`
# Script de instalación silenciosa de servicio TUN
$ErrorActionPreference = "Stop"

try {
    # Crear servicio
    sc.exe create "%s" binPath= "%s -service-mode" start= auto DisplayName= "IPVN7 Tunnel Service"
    if ($LASTEXITCODE -ne 0) { throw "Error creando servicio" }
    
    # Configurar como SYSTEM
    sc.exe config "%s" obj= LocalSystem
    if ($LASTEXITCODE -ne 0) { throw "Error configurando servicio" }
    
    # Iniciar servicio
    sc.exe start "%s"
    if ($LASTEXITCODE -ne 0) { throw "Error iniciando servicio" }
    
    Write-Output "SUCCESS"
} catch {
    Write-Output "ERROR: $_"
    exit 1
}
`, ae.serviceName, ae.servicePath, ae.serviceName, ae.serviceName)

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0600)
	if err != nil {
		return fmt.Errorf("error creando script temporal: %w", err)
	}
	defer os.Remove(scriptPath)

	// Ejecutar PowerShell con elevación silenciosa
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error elevando e instalando: %w (salida: %s)", err, string(output))
	}

	if !strings.Contains(string(output), "SUCCESS") {
		return fmt.Errorf("instalación falló: %s", string(output))
	}

	return nil
}

// IsTUNServiceRunning verifica si el servicio TUN está corriendo
func (ae *AutoElevationManager) IsTUNServiceRunning() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sc.exe", "query", ae.serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "RUNNING")
}

// EnsureTUNAvailable asegura que el TUN esté disponible (automático)
func (ae *AutoElevationManager) EnsureTUNAvailable() error {
	// Si el servicio ya corre, perfecto
	if ae.IsTUNServiceRunning() {
		return nil
	}

	// Si tiene admin, instalar directamente
	if ae.HasAdminPrivileges() {
		return ae.InstallTUNService()
	}

	// Si no tiene admin, intentar auto-elevar
	return ae.autoElevateAndInstall()
}
