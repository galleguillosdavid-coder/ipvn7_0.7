package l3

import (
	"strings"
	"time"
)

// Constantes de acciones normalizadas
const (
	ActionPowerSleep       = "sleep"
	ActionPowerReboot      = "reboot"
	ActionPowerShutdown    = "shutdown"
	ActionPowerLock        = "lock"
	ActionMonitorOff       = "monitor_off"
	ActionCleanRecycleBin  = "clean_recycle_bin"
	ActionBatteryStatus    = "battery"
	ActionThermalStatus    = "thermal"
	ActionKillHeavyProcess = "kill_heavy"
)

// IntentCategory clasificación de alto nivel de la intención
type IntentCategory string

const (
	CategoryHardwareOS IntentCategory = "hardware_os"
	CategoryRobotDrone IntentCategory = "robot_drone"
	CategoryVehicle    IntentCategory = "vehicle"
	CategoryNetworkP2P IntentCategory = "network_p2p"
	CategoryUnknown    IntentCategory = "unknown"
)

// ParsedIntent intención normalizada derivada del lenguaje natural
type ParsedIntent struct {
	OriginalText string            `json:"original_text"`
	Category     IntentCategory    `json:"category"`
	Action       string            `json:"action"`
	TargetNode   string            `json:"target_node"` // "local", DID o alias (ej. "notebook", "dron")
	Parameters   map[string]string `json:"parameters,omitempty"`
	Confidence   float64           `json:"confidence"`
}

// IntentExecutionResponse resultado de la ejecución física de la intención
type IntentExecutionResponse struct {
	Success      bool                   `json:"success"`
	Category     IntentCategory         `json:"category"`
	Action       string                 `json:"action"`
	TargetNode   string                 `json:"target_node"`
	HumanMessage string                 `json:"human_message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}


// ParseNaturalLanguage interpreta un texto en lenguaje cotidiano (0 tokens / offline)
func ParseNaturalLanguage(text string) ParsedIntent {
	clean := strings.ToLower(strings.TrimSpace(text))
	res := ParsedIntent{
		OriginalText: text,
		Category:     CategoryUnknown,
		Action:       "noop",
		TargetNode:   "local",
		Parameters:   make(map[string]string),
		Confidence:   0.0,
	}

	// Detectar nodo objetivo en el texto
	if strings.Contains(clean, "notebook") || strings.Contains(clean, "laptop") || strings.Contains(clean, "nodo b") {
		res.TargetNode = "notebook"
	} else if strings.Contains(clean, "dron") || strings.Contains(clean, "drone") {
		res.TargetNode = "drone_01"
	} else if strings.Contains(clean, "camion") || strings.Contains(clean, "camión") || strings.Contains(clean, "truck") {
		res.TargetNode = "truck_01"
	} else if strings.Contains(clean, "auto") || strings.Contains(clean, "coche") || strings.Contains(clean, "car") {
		res.TargetNode = "car_01"
	}

	// 1. Energía y Pantalla
	if strings.Contains(clean, "suspender") || strings.Contains(clean, "reposo") || strings.Contains(clean, "dormir") || strings.Contains(clean, "sleep") {
		res.Category = CategoryHardwareOS
		res.Action = ActionPowerSleep
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "bloquear") || strings.Contains(clean, "bloquea") || strings.Contains(clean, "lock") {
		res.Category = CategoryHardwareOS
		res.Action = ActionPowerLock
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "apagar monitor") || strings.Contains(clean, "apaga la pantalla") || strings.Contains(clean, "monitor off") {
		res.Category = CategoryHardwareOS
		res.Action = ActionMonitorOff
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "reiniciar") || strings.Contains(clean, "reboot") {
		res.Category = CategoryHardwareOS
		res.Action = ActionPowerReboot
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "apagar") || strings.Contains(clean, "apaga") || strings.Contains(clean, "shutdown") {
		res.Category = CategoryHardwareOS
		res.Action = ActionPowerShutdown
		res.Confidence = 0.95
		return res
	}

	// 2. Mantenimiento y Hardware
	if strings.Contains(clean, "papelera") || strings.Contains(clean, "reciclaje") || strings.Contains(clean, "limpiar basura") {
		res.Category = CategoryHardwareOS
		res.Action = ActionCleanRecycleBin
		res.Confidence = 0.90
		return res
	}
	if strings.Contains(clean, "bateria") || strings.Contains(clean, "batería") || strings.Contains(clean, "carga") {
		res.Category = CategoryHardwareOS
		res.Action = ActionBatteryStatus
		res.Confidence = 0.90
		return res
	}
	if strings.Contains(clean, "temperatura") || strings.Contains(clean, "calor") || strings.Contains(clean, "ventilador") {
		res.Category = CategoryHardwareOS
		res.Action = ActionThermalStatus
		res.Confidence = 0.90
		return res
	}
	if strings.Contains(clean, "procesos pesados") || strings.Contains(clean, "cerrar tareas lentas") || strings.Contains(clean, "kill") {
		res.Category = CategoryHardwareOS
		res.Action = ActionKillHeavyProcess
		res.Confidence = 0.85
		return res
	}

	// 3. Drones y Robots
	if strings.Contains(clean, "despegar") || strings.Contains(clean, "takeoff") || strings.Contains(clean, "elevar dron") {
		res.Category = CategoryRobotDrone
		res.Action = "takeoff"
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "aterrizar") || strings.Contains(clean, "land") {
		res.Category = CategoryRobotDrone
		res.Action = "land"
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "volver a casa") || strings.Contains(clean, "rtl") || strings.Contains(clean, "retornar") {
		res.Category = CategoryRobotDrone
		res.Action = "rtl"
		res.Confidence = 0.95
		return res
	}
	if strings.Contains(clean, "armar motores") || strings.Contains(clean, "arm") {
		res.Category = CategoryRobotDrone
		res.Action = "arm"
		res.Confidence = 0.90
		return res
	}

	// 4. Vehículos y Máquinas
	if strings.Contains(clean, "borrar fallas") || strings.Contains(clean, "clear dtc") || strings.Contains(clean, "apagar check engine") {
		res.Category = CategoryVehicle
		res.Action = "clear_dtc"
		res.Confidence = 0.90
		return res
	}
	if strings.Contains(clean, "motor") || strings.Contains(clean, "rpm") || strings.Contains(clean, "aceite") {
		res.Category = CategoryVehicle
		res.Action = "status"
		res.Confidence = 0.85
		return res
	}

	return res
}

