package osrunner

import (
	"errors"
	"testing"
)

type mockGateway struct {
	registered   bool
	unregistered bool
	events       []string
}

func (m *mockGateway) Register(id, name, version, transport string, capabilities []string) error {
	m.registered = true
	return nil
}

func (m *mockGateway) Heartbeat(id string) error {
	return nil
}

func (m *mockGateway) UnregisterComponent(id string) error {
	m.unregistered = true
	return nil
}

func (m *mockGateway) PublishSimpleEvent(source, eventType string, payload map[string]interface{}) {
	m.events = append(m.events, eventType)
}

type mockExecutor struct {
	locked      bool
	slept       bool
	rebooted    bool
	shutdown    bool
	monitorOff  bool
	killedPID   int
	emptiedBin  bool
	ejectedDisk string
}

func (m *mockExecutor) LockWorkStation() error {
	m.locked = true
	return nil
}

func (m *mockExecutor) SleepSystem() error {
	m.slept = true
	return nil
}

func (m *mockExecutor) RebootSystem() error {
	m.rebooted = true
	return nil
}

func (m *mockExecutor) ShutdownSystem() error {
	m.shutdown = true
	return nil
}

func (m *mockExecutor) SetMonitorPower(off bool) error {
	m.monitorOff = off
	return nil
}

func (m *mockExecutor) KillProcessByPID(pid int) error {
	if pid <= 0 {
		return errors.New("pid inválido")
	}
	m.killedPID = pid
	return nil
}

func (m *mockExecutor) GetTopMemoryProcesses(limit int) ([]ProcessInfo, error) {
	return []ProcessInfo{
		{PID: 1234, Name: "heavy_task.exe", MemoryMB: 1024, MemoryPct: 85.0},
		{PID: 5678, Name: "chrome.exe", MemoryMB: 512, MemoryPct: 15.0},
	}, nil
}

func (m *mockExecutor) GetBatteryStatus() (*BatteryInfo, error) {
	return &BatteryInfo{
		Percent:   88,
		OnACPower: true,
		Status:    "Cargando",
	}, nil
}

func (m *mockExecutor) GetThermalStatus() (*ThermalInfo, error) {
	return &ThermalInfo{
		CPUTempC:     48.5,
		FanSpeedRPM:  2100,
		FanStatus:    "Nominal",
		ThermalAlert: false,
	}, nil
}

func (m *mockExecutor) EmptyRecycleBin() error {
	m.emptiedBin = true
	return nil
}

func (m *mockExecutor) SafeEjectDrive(driveLetter string) error {
	m.ejectedDisk = driveLetter
	return nil
}

func TestOSRunnerComponent_LifecycleAndActions(t *testing.T) {
	gw := &mockGateway{}
	mock := &mockExecutor{}
	runner := NewOSRunnerComponent("os_runner_test", "Test OS Hardware Runner", gw, mock)

	// 1. Registro
	if err := runner.Start(); err != nil {
		t.Fatalf("error iniciando runner: %v", err)
	}

	if !gw.registered {
		t.Fatal("componente no registrado en el gateway")
	}

	// 2. Ejecutar Acciones
	// Batería
	respBat := runner.ExecuteAction(&ActionRequest{Action: ActionBatteryStatus})
	if !respBat.Success || respBat.Details["percent"] != 88 {
		t.Fatalf("fallo en ActionBatteryStatus: %+v", respBat)
	}

	// Térmico
	respTherm := runner.ExecuteAction(&ActionRequest{Action: ActionThermalStatus})
	if !respTherm.Success || respTherm.Details["cpu_temp_c"] != 48.5 {
		t.Fatalf("fallo en ActionThermalStatus: %+v", respTherm)
	}

	// Matar proceso pesado
	respKill := runner.ExecuteAction(&ActionRequest{Action: ActionKillHeavyProcess})
	if !respKill.Success || mock.killedPID != 1234 {
		t.Fatalf("fallo en ActionKillHeavyProcess: %+v", respKill)
	}

	// Bloquear equipo
	respLock := runner.ExecuteAction(&ActionRequest{Action: ActionPowerLock})
	if !respLock.Success || !mock.locked {
		t.Fatalf("fallo en ActionPowerLock: %+v", respLock)
	}

	// Monitor off
	respMon := runner.ExecuteAction(&ActionRequest{Action: ActionMonitorOff})
	if !respMon.Success || !mock.monitorOff {
		t.Fatalf("fallo en ActionMonitorOff: %+v", respMon)
	}

	// Expulsar disco
	respEject := runner.ExecuteAction(&ActionRequest{
		Action: ActionSafeEjectDrive,
		Params: map[string]string{"drive": "F:"},
	})
	if !respEject.Success || mock.ejectedDisk != "F:" {
		t.Fatalf("fallo en ActionSafeEjectDrive: %+v", respEject)
	}

	// Acción desconocida
	respUnknown := runner.ExecuteAction(&ActionRequest{Action: "invalid:action"})
	if respUnknown.Success {
		t.Fatal("esperaba fallo para acción desconocida")
	}

	// 3. Desregistro
	if err := runner.Stop(); err != nil {
		t.Fatalf("error deteniendo runner: %v", err)
	}
}
