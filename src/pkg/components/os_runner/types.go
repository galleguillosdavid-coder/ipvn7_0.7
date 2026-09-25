// Package osrunner implementa el Agente Satélite Universal de Sistema Operativo
// para control de hardware, energía, procesos y diagnóstico en ipvn7 v0.7.
package osrunner

// Tipos de acciones del sistema operativo
const (
	ActionPowerSleep       = "power:sleep"
	ActionPowerReboot      = "power:reboot"
	ActionPowerShutdown    = "power:shutdown"
	ActionPowerLock        = "power:lock"
	ActionMonitorOff       = "power:monitor_off"
	ActionKillHeavyProcess = "process:kill_heavy"
	ActionListProcesses    = "process:list"
	ActionBatteryStatus    = "hardware:battery"
	ActionThermalStatus    = "hardware:thermal"
	ActionCleanRecycleBin  = "system:clean_recycle_bin"
	ActionSafeEjectDrive   = "storage:safe_eject"
)

// ActionRequest solicitud de comando recibida a través de la malla ipvn7
type ActionRequest struct {
	Action    string            `json:"action"`
	Params    map[string]string `json:"params,omitempty"`
	CallerDID string            `json:"caller_did"`
	Timestamp int64             `json:"timestamp"`
}

// ActionResponse resultado determinista retornado tras la ejecución
type ActionResponse struct {
	Action    string                 `json:"action"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// BatteryInfo telemetría física de batería
type BatteryInfo struct {
	Percent   int    `json:"percent"`
	OnACPower bool   `json:"on_ac_power"`
	Status    string `json:"status"`
}

// ThermalInfo telemetría térmica y de ventiladores
type ThermalInfo struct {
	CPUTempC      float64 `json:"cpu_temp_c"`
	FanSpeedRPM   int     `json:"fan_speed_rpm"`
	FanStatus     string  `json:"fan_status"`
	ThermalAlert  bool    `json:"thermal_alert"`
}

// ProcessInfo ficha de proceso consumidor de recursos
type ProcessInfo struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	MemoryMB   int64   `json:"memory_mb"`
	MemoryPct  float64 `json:"memory_pct"`
	CPUPercent float64 `json:"cpu_pct"`
}

// PlatformExecutor interfaz abstracta para operaciones nativas del sistema operativo
type PlatformExecutor interface {
	LockWorkStation() error
	SleepSystem() error
	RebootSystem() error
	ShutdownSystem() error
	SetMonitorPower(off bool) error
	KillProcessByPID(pid int) error
	GetTopMemoryProcesses(limit int) ([]ProcessInfo, error)
	GetBatteryStatus() (*BatteryInfo, error)
	GetThermalStatus() (*ThermalInfo, error)
	EmptyRecycleBin() error
	SafeEjectDrive(driveLetter string) error
}
