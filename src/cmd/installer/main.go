//go:build windows && installer

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/ipvn7.exe
var ipvn7Binary []byte

//go:embed assets/wintun.dll
var wintunDLL []byte

var (
	shell32       = syscall.NewLazyDLL("shell32.dll")
	procShellExec = shell32.NewProc("ShellExecuteW")
)

const swShowNormal = 1

func isElevated() bool {
	cmd := exec.Command("net", "session")
	return cmd.Run() == nil
}

func runAsAdmin() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	args, _ := syscall.UTF16PtrFromString("-elevated")
	dir, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))

	r, _, _ := procShellExec.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(args)),
		uintptr(unsafe.Pointer(dir)),
		uintptr(swShowNormal),
	)
	return r > 32
}

func createDesktopShortcut(exePath, targetDir string) {
	desktop := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
	shortcutPath := filepath.Join(desktop, "VPN I7.lnk")

	psCmd := fmt.Sprintf(
		"$ws=New-Object -ComObject WScript.Shell; $s=$ws.CreateShortcut('%s'); $s.TargetPath='%s'; $s.WorkingDirectory='%s'; $s.Description='VPN I7 - Red Soberana Cuántica'; $s.IconLocation='%s,0'; $s.Save()",
		shortcutPath, exePath, targetDir, exePath,
	)
	_ = exec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", psCmd).Run()
}

func openUI(url string) {
	go func() {
		time.Sleep(1500 * time.Millisecond)
		// 1. Intentar modo App en Edge (ventana limpia nativa sin barras de navegación)
		if err := exec.Command("cmd.exe", "/c", "start", "msedge.exe", fmt.Sprintf("--app=%s", url), "--window-size=440,680").Run(); err == nil {
			return
		}
		// 2. Intentar modo App en Chrome
		if err := exec.Command("cmd.exe", "/c", "start", "chrome.exe", fmt.Sprintf("--app=%s", url), "--window-size=440,680").Run(); err == nil {
			return
		}
		// 3. Fallback en navegador predeterminado
		_ = exec.Command("explorer.exe", url).Start()
	}()
}

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("VPN I7: Instalador exclusivo para Windows.")
		return
	}

	alreadyElevated := isElevated()
	hasElevatedFlag := len(os.Args) > 1 && os.Args[1] == "-elevated"

	// Intentar auto-elevación UAC nativa silenciosa (estilo Chrome / instaladores modernos).
	// Si el usuario cancela o deniega UAC, continúa limpiamente en modo estándar (zero-admin).
	if !alreadyElevated && !hasElevatedFlag {
		if runAsAdmin() {
			os.Exit(0)
		}
	}

	// 1. Directorio de instalación en LocalAppData (%LOCALAPPDATA%\IPVN7)
	localApp := os.Getenv("LOCALAPPDATA")
	if localApp == "" {
		localApp = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	installDir := filepath.Join(localApp, "IPVN7")
	keystoreDir := filepath.Join(installDir, "keystore")
	_ = os.MkdirAll(keystoreDir, 0755)

	// 2. Extraer binario principal y driver Wintun
	exePath := filepath.Join(installDir, "ipvn7.exe")
	wintunPath := filepath.Join(installDir, "wintun.dll")

	_ = os.WriteFile(exePath, ipvn7Binary, 0755)
	if len(wintunDLL) > 0 {
		_ = os.WriteFile(wintunPath, wintunDLL, 0644)
	}

	// 3. Configurar firewall si estamos en modo elevado
	if isElevated() {
		_ = exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=IPVN7-Mesh-UDP", "dir=in", "action=allow", "protocol=UDP", "localport=7777").Run()
		_ = exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=IPVN7-Web-TCP", "dir=in", "action=allow", "protocol=TCP", "localport=7070").Run()
	}

	// 4. Crear acceso directo limpio en el Escritorio: "VPN I7"
	createDesktopShortcut(exePath, installDir)

	// 5. Argumentos de ejecución transparente
	cmdArgs := []string{"-port=7777", "-web-port=7070", "-no-elevate"}
	if isElevated() {
		cmdArgs = append(cmdArgs, "-tun")
	} else {
		cmdArgs = append(cmdArgs, "-socks5=10807")
	}

	// 6. Iniciar VPN I7 inmediatamente en segundo plano
	launchCmd := exec.Command(exePath, cmdArgs...)
	launchCmd.Dir = installDir
	_ = launchCmd.Start()

	// 7. Abrir interfaz visual directamente en pantalla
	openUI("http://127.0.0.1:7070")

	// Pequeña pausa para asegurar que el proceso hijo se desprenda correctamente antes de salir
	time.Sleep(800 * time.Millisecond)
}
