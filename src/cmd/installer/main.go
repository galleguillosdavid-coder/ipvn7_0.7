//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

var (
	shell32       = syscall.NewLazyDLL("shell32.dll")
	procShellExec = shell32.NewProc("ShellExecuteW")
	user32        = syscall.NewLazyDLL("user32.dll")
	procMsgBox    = user32.NewProc("MessageBoxW")
)

const (
	swShowNormal = 1
	mbIconError  = 0x00000010
	mbOk         = 0x00000000
)

func showMessage(title, text string, flags uintptr) {
	tPtr, _ := syscall.UTF16PtrFromString(title)
	mPtr, _ := syscall.UTF16PtrFromString(text)
	_, _, _ = procMsgBox.Call(0, uintptr(unsafe.Pointer(mPtr)), uintptr(unsafe.Pointer(tPtr)), flags)
}

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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func locateBinary(exeDir, fileName string, searchPaths []string) string {
	for _, p := range searchPaths {
		candidate := filepath.Join(p, fileName)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return candidate
		}
	}
	return ""
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
		// 1. Intentar modo App en Edge
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

	if !alreadyElevated && !hasElevatedFlag {
		if runAsAdmin() {
			os.Exit(0)
		}
	}

	selfExe, err := os.Executable()
	if err != nil {
		selfExe = os.Args[0]
	}
	selfDir := filepath.Dir(selfExe)

	// Arquitectura B: Localización de artefactos sin go:embed en el árbol Git
	searchRoots := []string{
		selfDir,
		filepath.Join(selfDir, "bin"),
		filepath.Join(selfDir, "..", "bin"),
		filepath.Join(selfDir, "dist"),
		".",
		"bin",
	}

	srcExe := locateBinary(selfDir, "ipvn7.exe", searchRoots)
	if srcExe == "" {
		showMessage("VPN I7 — Instalador",
			"No se encontró el binario ipvn7.exe adyacente para la instalación.\nAsegúrese de ubicar ipvn7.exe junto al instalador o en el directorio bin/.",
			uintptr(mbIconError|mbOk))
		os.Exit(1)
	}

	wintunRoots := append(searchRoots,
		filepath.Join(selfDir, "wintun", "bin", "amd64"),
		filepath.Join(selfDir, "..", "wintun", "bin", "amd64"),
		"wintun/bin/amd64",
	)
	srcWintun := locateBinary(selfDir, "wintun.dll", wintunRoots)

	// 1. Directorio de instalación en LocalAppData (%LOCALAPPDATA%\IPVN7)
	localApp := os.Getenv("LOCALAPPDATA")
	if localApp == "" {
		localApp = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	installDir := filepath.Join(localApp, "IPVN7")
	keystoreDir := filepath.Join(installDir, "keystore")
	_ = os.MkdirAll(keystoreDir, 0755)

	// 2. Copiar binario principal y driver Wintun (si está disponible)
	exePath := filepath.Join(installDir, "ipvn7.exe")
	wintunPath := filepath.Join(installDir, "wintun.dll")

	if err := copyFile(srcExe, exePath); err != nil {
		showMessage("VPN I7 — Error de Instalación",
			fmt.Sprintf("No se pudo copiar ipvn7.exe a destino: %v", err),
			uintptr(mbIconError|mbOk))
		os.Exit(1)
	}

	if srcWintun != "" {
		_ = copyFile(srcWintun, wintunPath)
	}

	// 3. Configurar firewall si estamos en modo elevado: ÚNICAMENTE transporte UDP 7777
	// (Se erradica la apertura innecesaria de TCP 7070 para la WebUI localhost)
	if isElevated() {
		_ = exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=IPVN7-Mesh-UDP", "dir=in", "action=allow", "protocol=UDP", "localport=7777").Run()
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

	time.Sleep(800 * time.Millisecond)
}
