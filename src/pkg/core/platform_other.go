//go:build !windows

package core

import (
	"fmt"
)

// BuildSchtasksArgs stub para sistemas no Windows
func BuildSchtasksArgs(taskName, exePath string) []string {
	return nil
}

// RegisterLogonTask stub para sistemas no Windows
func RegisterLogonTask(taskName, exePath string, runner CmdRunner) error {
	return fmt.Errorf("platform: schtasks solo es compatible con Windows")
}

// PurgeOrphanWintunAdapters stub para sistemas no Windows
func PurgeOrphanWintunAdapters(prefix string, runner CmdRunner) (int, error) {
	return 0, nil
}

// SetWindowsUserProxy stub para sistemas no Windows
func SetWindowsUserProxy(proxyServer string, runner CmdRunner) error {
	return nil
}

// ClearWindowsUserProxy stub para sistemas no Windows
func ClearWindowsUserProxy(runner CmdRunner) error {
	return nil
}

// TerminateConflictingProcesses stub para sistemas no Windows
func TerminateConflictingProcesses(runner CmdRunner) {
}

