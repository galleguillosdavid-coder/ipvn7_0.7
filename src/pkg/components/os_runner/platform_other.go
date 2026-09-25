//go:build !windows

package osrunner

import (
	"errors"
	"os"
	"os/exec"
)

// GenericExecutor implementación para Linux / macOS
type GenericExecutor struct{}

// NewPlatformExecutor constructor para sistemas no Windows
func NewPlatformExecutor() PlatformExecutor {
	return &GenericExecutor{}
}

func (g *GenericExecutor) LockWorkStation() error {
	// Intentar loginctl o xdg-screensaver
	cmd := exec.Command("loginctl", "lock-session")
	return cmd.Run()
}

func (g *GenericExecutor) SleepSystem() error {
	cmd := exec.Command("systemctl", "suspend")
	return cmd.Run()
}

func (g *GenericExecutor) RebootSystem() error {
	cmd := exec.Command("systemctl", "reboot")
	return cmd.Run()
}

func (g *GenericExecutor) ShutdownSystem() error {
	cmd := exec.Command("systemctl", "poweroff")
	return cmd.Run()
}

func (g *GenericExecutor) SetMonitorPower(off bool) error {
	state := "on"
	if off {
		state = "off"
	}
	cmd := exec.Command("xset", "dpms", "force", state)
	return cmd.Run()
}

func (g *GenericExecutor) KillProcessByPID(pid int) error {
	if pid <= 1 {
		return errors.New("prohibido terminar init/systemd")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func (g *GenericExecutor) GetTopMemoryProcesses(limit int) ([]ProcessInfo, error) {
	return []ProcessInfo{
		{PID: 100, Name: "ipvn7", MemoryMB: 28},
	}, nil
}

func (g *GenericExecutor) GetBatteryStatus() (*BatteryInfo, error) {
	return &BatteryInfo{
		Percent:   100,
		OnACPower: true,
		Status:    "Conectado a CA",
	}, nil
}

func (g *GenericExecutor) GetThermalStatus() (*ThermalInfo, error) {
	return &ThermalInfo{
		CPUTempC:     40.0,
		FanSpeedRPM:  1500,
		FanStatus:    "Nominal",
		ThermalAlert: false,
	}, nil
}

func (g *GenericExecutor) EmptyRecycleBin() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	trashDir := home + "/.local/share/Trash/files"
	return os.RemoveAll(trashDir)
}

func (g *GenericExecutor) SafeEjectDrive(driveLetter string) error {
	cmd := exec.Command("eject", driveLetter)
	return cmd.Run()
}
