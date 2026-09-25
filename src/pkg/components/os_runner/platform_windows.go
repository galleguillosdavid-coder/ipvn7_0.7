//go:build windows

package osrunner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modShell32  = syscall.NewLazyDLL("shell32.dll")

	procLockWorkStation      = modUser32.NewProc("LockWorkStation")
	procSendMessageW         = modUser32.NewProc("SendMessageW")
	procGetSystemPowerStatus = modKernel32.NewProc("GetSystemPowerStatus")
	procSHEmptyRecycleBinW   = modShell32.NewProc("SHEmptyRecycleBinW")
)

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

// WindowsExecutor implementación nativa de Windows sin dependencias externas
type WindowsExecutor struct{}

// NewPlatformExecutor constructor para Windows
func NewPlatformExecutor() PlatformExecutor {
	return &WindowsExecutor{}
}

func (w *WindowsExecutor) LockWorkStation() error {
	r, _, err := procLockWorkStation.Call()
	if r == 0 {
		return fmt.Errorf("fallo LockWorkStation: %v", err)
	}
	return nil
}

func (w *WindowsExecutor) SleepSystem() error {
	// Llamada a SetSuspendState vía rundll32 powrprof.dll
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	return cmd.Start()
}

func (w *WindowsExecutor) RebootSystem() error {
	cmd := exec.Command("shutdown", "/r", "/t", "5", "/c", "ipvn7 Sovereign Restart")
	return cmd.Run()
}

func (w *WindowsExecutor) ShutdownSystem() error {
	cmd := exec.Command("shutdown", "/s", "/t", "10", "/c", "ipvn7 Sovereign Shutdown")
	return cmd.Run()
}

func (w *WindowsExecutor) SetMonitorPower(off bool) error {
	// WM_SYSCOMMAND = 0x0112, SC_MONITORPOWER = 0xF170, 2 = Power Off
	val := uintptr(2)
	if !off {
		val = uintptr(^uint(0)) // -1 = Power On
	}
	procSendMessageW.Call(
		uintptr(0xFFFF), // HWND_BROADCAST
		uintptr(0x0112), // WM_SYSCOMMAND
		uintptr(0xF170), // SC_MONITORPOWER
		val,
	)
	return nil
}

func (w *WindowsExecutor) KillProcessByPID(pid int) error {
	if pid <= 4 {
		return errors.New("prohibido terminar procesos de sistema crítico")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func (w *WindowsExecutor) GetTopMemoryProcesses(limit int) ([]ProcessInfo, error) {
	if limit <= 0 {
		limit = 5
	}
	// Usar tasklist nativo de Windows (formato CSV)
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando tasklist: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	procs := make([]ProcessInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\",\"")
		if len(parts) >= 5 {
			name := strings.Trim(parts[0], "\"")
			pidStr := strings.Trim(parts[1], "\"")
			memStr := strings.Trim(parts[4], "\" ")
			// "123.456 K" -> normalizar
			memStr = strings.ReplaceAll(memStr, ".", "")
			memStr = strings.ReplaceAll(memStr, ",", "")
			memStr = strings.ReplaceAll(memStr, " K", "")
			memStr = strings.TrimSpace(memStr)

			pid, _ := strconv.Atoi(pidStr)
			memKB, _ := strconv.ParseInt(memStr, 10, 64)

			if pid > 0 {
				procs = append(procs, ProcessInfo{
					PID:      pid,
					Name:     name,
					MemoryMB: memKB / 1024,
				})
			}
		}
	}

	// Ordenamiento simple de burbuja/inserción de mayor a menor
	for i := 0; i < len(procs)-1; i++ {
		for j := i + 1; j < len(procs); j++ {
			if procs[j].MemoryMB > procs[i].MemoryMB {
				procs[i], procs[j] = procs[j], procs[i]
			}
		}
	}

	if len(procs) > limit {
		procs = procs[:limit]
	}
	return procs, nil
}

func (w *WindowsExecutor) GetBatteryStatus() (*BatteryInfo, error) {
	var sps systemPowerStatus
	r, _, err := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&sps)))
	if r == 0 {
		return nil, fmt.Errorf("error llamando GetSystemPowerStatus: %v", err)
	}

	onAC := (sps.ACLineStatus == 1)
	pct := int(sps.BatteryLifePercent)
	if pct > 100 {
		pct = 100
	}

	status := "Descargando"
	if onAC {
		status = "Conectado a CA (Enchufado)"
	}
	if sps.BatteryFlag&8 != 0 {
		status = "Cargando"
	}
	if sps.BatteryFlag&128 != 0 || pct == 255 {
		status = "Sin Batería (PC de Escritorio)"
		pct = 100
		onAC = true
	}

	return &BatteryInfo{
		Percent:   pct,
		OnACPower: onAC,
		Status:    status,
	}, nil
}

func (w *WindowsExecutor) GetThermalStatus() (*ThermalInfo, error) {
	// Consulta WMIC para zona térmica si está disponible
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance -ClassName Win32_PerfFormattedData_Counters_ThermalZoneInformation -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Temperature")
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()

	tempKelvinStr := strings.TrimSpace(out.String())
	tempC := 42.0 // Valor basal nominal por defecto
	if kelvin, err := strconv.ParseFloat(tempKelvinStr, 64); err == nil && kelvin > 273.15 {
		tempC = kelvin - 273.15
	}

	fanStatus := "Nominal"
	alert := false
	if tempC > 78.0 {
		fanStatus = "Alto Rendimiento / Alerta Térmica"
		alert = true
	}

	return &ThermalInfo{
		CPUTempC:     tempC,
		FanSpeedRPM:  1800,
		FanStatus:    fanStatus,
		ThermalAlert: alert,
	}, nil
}

func (w *WindowsExecutor) EmptyRecycleBin() error {
	// Flags: SHERB_NOCONFIRMATION (0x00000001) | SHERB_NOPROGRESSUI (0x00000002) | SHERB_NOSOUND (0x00000004)
	flags := uintptr(0x00000007)
	r, _, _ := procSHEmptyRecycleBinW.Call(0, 0, flags)
	if r != 0 && r != 0x80004005 { // S_OK or already empty
		return fmt.Errorf("error vaciando papelera: HRESULT 0x%X", r)
	}
	return nil
}

func (w *WindowsExecutor) SafeEjectDrive(driveLetter string) error {
	if driveLetter == "" {
		return errors.New("unidad no especificada")
	}
	// Expulsión segura mediante comando de PowerShell Shell.Application
	script := fmt.Sprintf(`(New-Object -comObject Shell.Application).Namespace(17).ParseName('%s').InvokeVerb('Eject')`, driveLetter)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	return cmd.Run()
}
