package core

import (
	"context"
	"strings"
	"testing"
)

func TestBuildSchtasksArgs(t *testing.T) {
	taskName := "IPVN7_Daemon"
	exePath := `C:\ipvn7\ipvn7.exe`

	args := BuildSchtasksArgs(taskName, exePath)
	cmdLine := strings.Join(args, " ")

	if !strings.Contains(cmdLine, "/sc onlogon") {
		t.Fatalf("expected /sc onlogon, got: %s", cmdLine)
	}
	if !strings.Contains(cmdLine, "/rl highest") {
		t.Fatalf("expected /rl highest, got: %s", cmdLine)
	}
	if !strings.Contains(cmdLine, "/tn IPVN7_Daemon") {
		t.Fatalf("expected task name in args, got: %s", cmdLine)
	}
}

func TestRegisterLogonTask_Runner(t *testing.T) {
	executed := false
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		executed = true
		if name != "schtasks.exe" {
			t.Fatalf("expected schtasks.exe, got: %s", name)
		}
		return []byte("SUCCESS: The scheduled task was successfully created."), nil
	}

	err := RegisterLogonTask("TestTask", `C:\test\bin.exe`, mockRunner)
	if err != nil {
		t.Fatalf("RegisterLogonTask returned unexpected error: %v", err)
	}
	if !executed {
		t.Fatal("mockRunner was not invoked")
	}
}

func TestPurgeOrphanWintunAdapters_Runner(t *testing.T) {
	netshOutput := `
Admin State    State          Type             Interface Name
-------------------------------------------------------------------------
Enabled        Connected      Dedicated        Ethernet
Enabled        Connected      Dedicated        Wi-Fi
Enabled        Disconnected   Dedicated        ieu0_orphan1
Enabled        Disconnected   Dedicated        ieu0_orphan2
`

	calls := make([]string, 0)
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		fullCmd := name + " " + strings.Join(args, " ")
		calls = append(calls, fullCmd)
		if strings.Contains(fullCmd, "show interface") {
			return []byte(netshOutput), nil
		}
		return []byte("OK"), nil
	}

	purged, err := PurgeOrphanWintunAdapters("ieu0", mockRunner)
	if err != nil {
		t.Fatalf("PurgeOrphanWintunAdapters failed: %v", err)
	}

	if purged != 2 {
		t.Fatalf("expected 2 purged adapters, got: %d", purged)
	}

	if len(calls) != 3 { // 1 show + 2 set disable
		t.Fatalf("expected 3 calls, got %d: %v", len(calls), calls)
	}
}
