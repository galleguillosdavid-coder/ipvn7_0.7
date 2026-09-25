package osrunner

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// GatewayRegistrar abstracción para vincular el componente al gateway del núcleo
type GatewayRegistrar interface {
	Register(id, name, version, transport string, capabilities []string) error
	Heartbeat(id string) error
	UnregisterComponent(id string) error
	PublishSimpleEvent(source, eventType string, payload map[string]interface{})
}

// OSRunnerComponent agente satélite ejecutor de control de hardware y sistema
type OSRunnerComponent struct {
	mu         sync.RWMutex
	id         string
	name       string
	gateway    GatewayRegistrar
	executor   PlatformExecutor
	running    bool
	stopChan   chan struct{}
}

// NewOSRunnerComponent inicializa el runner satélite con el ejecutor de plataforma
func NewOSRunnerComponent(compID, compName string, gw GatewayRegistrar, exec PlatformExecutor) *OSRunnerComponent {
	if compID == "" {
		compID = "os_runner_local"
	}
	if compName == "" {
		compName = "Universal OS Hardware Runner"
	}
	if exec == nil {
		exec = NewPlatformExecutor()
	}
	return &OSRunnerComponent{
		id:       compID,
		name:     compName,
		gateway:  gw,
		executor: exec,
		stopChan: make(chan struct{}),
	}
}

// Start registra el componente en el Smart Gateway y escucha órdenes
func (r *OSRunnerComponent) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return errors.New("os_runner: ya en ejecución")
	}
	r.running = true
	r.mu.Unlock()

	if r.gateway != nil {
		caps := []string{
			"system:power",
			"system:process",
			"hardware:diag",
			"system:maintenance",
			"action:exec",
		}
		if err := r.gateway.Register(r.id, r.name, "1.0.0", "in-process-ipc", caps); err != nil {
			return fmt.Errorf("os_runner: error registrando en gateway: %w", err)
		}

		go r.heartbeatLoop()
	}

	return nil
}

// Stop desacopla el componente del gateway
func (r *OSRunnerComponent) Stop() error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = false
	close(r.stopChan)
	r.mu.Unlock()

	if r.gateway != nil {
		return r.gateway.UnregisterComponent(r.id)
	}
	return nil
}

// ExecuteAction despacha una orden atómica de forma determinista y retorna el resultado
func (r *OSRunnerComponent) ExecuteAction(req *ActionRequest) *ActionResponse {
	resp := &ActionResponse{
		Action:    req.Action,
		Timestamp: time.Now().Unix(),
		Details:   make(map[string]interface{}),
	}

	switch req.Action {
	case ActionPowerLock:
		err := r.executor.LockWorkStation()
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error bloqueando equipo: %v", err)
		} else {
			resp.Message = "Estación de trabajo bloqueada exitosamente"
		}

	case ActionPowerSleep:
		err := r.executor.SleepSystem()
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error suspendiendo equipo: %v", err)
		} else {
			resp.Message = "Orden de suspensión ACPI despachada"
		}

	case ActionPowerReboot:
		err := r.executor.RebootSystem()
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error reiniciando equipo: %v", err)
		} else {
			resp.Message = "Reinicio de sistema ordenado"
		}

	case ActionPowerShutdown:
		err := r.executor.ShutdownSystem()
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error apagando equipo: %v", err)
		} else {
			resp.Message = "Apagado de sistema ordenado"
		}

	case ActionMonitorOff:
		err := r.executor.SetMonitorPower(true)
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error apagando monitor: %v", err)
		} else {
			resp.Message = "Monitor apagado (equipo en ejecución activa)"
		}

	case ActionBatteryStatus:
		info, err := r.executor.GetBatteryStatus()
		if err != nil {
			resp.Success = false
			resp.Message = fmt.Sprintf("Error leyendo batería: %v", err)
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("Batería: %d%% (%s)", info.Percent, info.Status)
			resp.Details["percent"] = info.Percent
			resp.Details["on_ac_power"] = info.OnACPower
			resp.Details["status"] = info.Status
		}

	case ActionThermalStatus:
		thermal, err := r.executor.GetThermalStatus()
		if err != nil {
			resp.Success = false
			resp.Message = fmt.Sprintf("Error leyendo telemetría térmica: %v", err)
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("CPU: %.1f °C | Ventiladores: %s (%d RPM)", thermal.CPUTempC, thermal.FanStatus, thermal.FanSpeedRPM)
			resp.Details["cpu_temp_c"] = thermal.CPUTempC
			resp.Details["fan_speed_rpm"] = thermal.FanSpeedRPM
			resp.Details["fan_status"] = thermal.FanStatus
			resp.Details["alert"] = thermal.ThermalAlert
		}

	case ActionListProcesses:
		procs, err := r.executor.GetTopMemoryProcesses(5)
		if err != nil {
			resp.Success = false
			resp.Message = fmt.Sprintf("Error listando procesos: %v", err)
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("%d procesos principales listados", len(procs))
			resp.Details["processes"] = procs
		}

	case ActionKillHeavyProcess:
		procs, err := r.executor.GetTopMemoryProcesses(1)
		if err != nil || len(procs) == 0 {
			resp.Success = false
			resp.Message = "No se pudo identificar proceso consumidor"
		} else {
			top := procs[0]
			err := r.executor.KillProcessByPID(top.PID)
			resp.Success = (err == nil)
			if err != nil {
				resp.Message = fmt.Sprintf("Fallo al terminar PID %d (%s): %v", top.PID, top.Name, err)
			} else {
				resp.Message = fmt.Sprintf("Proceso '%s' (PID %d, %d MB) terminado exitosamente", top.Name, top.PID, top.MemoryMB)
				resp.Details["killed_process"] = top
			}
		}

	case ActionCleanRecycleBin:
		err := r.executor.EmptyRecycleBin()
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error vaciando papelera: %v", err)
		} else {
			resp.Message = "Papelera de reciclaje vaciada por completo"
		}

	case ActionSafeEjectDrive:
		letter := req.Params["drive"]
		if letter == "" {
			letter = "E:"
		}
		err := r.executor.SafeEjectDrive(letter)
		resp.Success = (err == nil)
		if err != nil {
			resp.Message = fmt.Sprintf("Error expulsando unidad %s: %v", letter, err)
		} else {
			resp.Message = fmt.Sprintf("Unidad %s expulsada de forma segura", letter)
		}

	default:
		resp.Success = false
		resp.Message = fmt.Sprintf("Acción desconocida '%s'", req.Action)
	}

	// Si el gateway está configurado, publicar el evento de acción
	if r.gateway != nil {
		payloadBytes, _ := json.Marshal(resp)
		r.gateway.PublishSimpleEvent(r.id, "os:action_completed", map[string]interface{}{
			"action":  resp.Action,
			"success": resp.Success,
			"message": resp.Message,
			"raw":     string(payloadBytes),
		})
	}

	return resp
}

func (r *OSRunnerComponent) heartbeatLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopChan:
			return
		case <-ticker.C:
			r.mu.RLock()
			running := r.running
			r.mu.RUnlock()
			if !running {
				return
			}
			_ = r.gateway.Heartbeat(r.id)
		}
	}
}
