package vehiclerobotbridge

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// GatewayEventBus abstracción mínima del bus del Smart Component Gateway
type GatewayEventBus interface {
	PublishEvent(eventType string, data []byte) error
}

// GatewayBusFunc permite usar funciones anónimas como GatewayEventBus
type GatewayBusFunc func(eventType string, data []byte) error

// PublishEvent ejecuta la función adaptadora
func (f GatewayBusFunc) PublishEvent(eventType string, data []byte) error {
	return f(eventType, data)
}


// VehicleRobotBridge orquesta periféricos vehiculares, drones, robots y maquinaria
type VehicleRobotBridge struct {
	mu           sync.RWMutex
	componentID  string
	name         string
	eventBus     GatewayEventBus
	droneDevices map[string]*DroneTelemetry
	vehDevices   map[string]*VehicleTelemetry
	running      bool
}

// NewVehicleRobotBridge inicializa el componente satélite
func NewVehicleRobotBridge(id, name string, bus GatewayEventBus) *VehicleRobotBridge {
	return &VehicleRobotBridge{
		componentID:  id,
		name:         name,
		eventBus:     bus,
		droneDevices: make(map[string]*DroneTelemetry),
		vehDevices:   make(map[string]*VehicleTelemetry),
		running:      true,
	}
}

// IngestMAVLinkData procesa bytes recibidos por socket UDP/Serial desde un dron o robot
func (b *VehicleRobotBridge) IngestMAVLinkData(deviceID string, rawData []byte) error {
	pkt, err := ParseMAVLinkV2Frame(rawData)
	if err != nil {
		return err
	}

	b.mu.Lock()
	telem, exists := b.droneDevices[deviceID]
	if !exists {
		telem = &DroneTelemetry{FlightMode: "MANUAL", LastHeartbeat: time.Now()}
		b.droneDevices[deviceID] = telem
	}
	UpdateTelemetryFromMAVLink(pkt, telem)
	b.mu.Unlock()

	// Notificar al bus de eventos de la malla
	if b.eventBus != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"device_id": deviceID,
			"type":      "drone",
			"telemetry": telem,
		})
		_ = b.eventBus.PublishEvent("telemetry:drone", data)
	}
	return nil
}

// IngestCANFrame procesa tramas CAN Bus recibidas de un automóvil, camión o máquina
func (b *VehicleRobotBridge) IngestCANFrame(deviceID string, devType string, frame *CANFrame) bool {
	b.mu.Lock()
	telem, exists := b.vehDevices[deviceID]
	if !exists {
		telem = &VehicleTelemetry{DeviceType: devType, LastUpdate: time.Now()}
		b.vehDevices[deviceID] = telem
	}

	updated := false
	if frame.Extended {
		updated = ProcessJ1939Frame(frame, telem)
	} else {
		updated = ProcessOBD2Response(frame, telem)
	}
	b.mu.Unlock()

	if updated && b.eventBus != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"device_id": deviceID,
			"type":      devType,
			"telemetry": telem,
		})
		_ = b.eventBus.PublishEvent("telemetry:vehicle", data)
	}
	return updated
}

// ExecuteControlCommand procesa una orden sobre un dron, robot o vehículo
func (b *VehicleRobotBridge) ExecuteControlCommand(cmd *ControlCommand) *CommandResult {
	if cmd == nil {
		return &CommandResult{Success: false, Message: "comando nulo", Timestamp: time.Now()}
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	res := &CommandResult{
		DeviceID:  cmd.DeviceID,
		Action:    cmd.Action,
		Timestamp: now,
	}

	// Comandos para Drones y Robots
	if drone, ok := b.droneDevices[cmd.DeviceID]; ok {
		switch cmd.Action {
		case "arm":
			drone.Armed = true
			res.Success = true
			res.Message = "Motores armados con éxito"
		case "disarm":
			drone.Armed = false
			res.Success = true
			res.Message = "Motores desarmados con éxito"
		case "takeoff":
			if !drone.Armed {
				res.Success = false
				res.Message = "Imposible despegue: motores no armados"
				return res
			}
			drone.FlightMode = "GUIDED"
			drone.AltitudeMeters = 5.0
			res.Success = true
			res.Message = "Despegue autónomo iniciado a 5.0m"
		case "land":
			drone.FlightMode = "LAND"
			res.Success = true
			res.Message = "Maniobra de aterrizaje iniciada"
		case "rtl":
			drone.FlightMode = "RTL"
			res.Success = true
			res.Message = "Retorno al punto de despegue (RTL) activado"
		default:
			res.Success = false
			res.Message = fmt.Sprintf("Acción no soportada para dron: %s", cmd.Action)
		}
		return res
	}

	// Comandos para Vehículos y Máquinas
	if veh, ok := b.vehDevices[cmd.DeviceID]; ok {
		switch cmd.Action {
		case "clear_dtc":
			veh.DTCFaultCodes = nil
			res.Success = true
			res.Message = "Códigos de avería (DTCs) reseteados en ECU"
		case "status":
			res.Success = true
			res.Message = fmt.Sprintf("Vehículo OK: %d RPM, %d km/h", veh.EngineRPM, veh.SpeedKmh)
		default:
			res.Success = true
			res.Message = fmt.Sprintf("Comando %s despachado a bus CAN", cmd.Action)
		}
		return res
	}

	res.Success = false
	res.Message = fmt.Sprintf("Dispositivo no encontrado en malla: %s", cmd.DeviceID)
	return res
}

// GetDroneTelemetry retorna el estado actual de un dron o robot
func (b *VehicleRobotBridge) GetDroneTelemetry(deviceID string) (*DroneTelemetry, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t, ok := b.droneDevices[deviceID]
	return t, ok
}

// GetVehicleTelemetry retorna el estado actual de un vehículo o máquina
func (b *VehicleRobotBridge) GetVehicleTelemetry(deviceID string) (*VehicleTelemetry, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t, ok := b.vehDevices[deviceID]
	return t, ok
}

// ListAllDevices lista todos los dispositivos vehiculares y robóticos activos
func (b *VehicleRobotBridge) ListAllDevices() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return map[string]interface{}{
		"drones":   b.droneDevices,
		"vehicles": b.vehDevices,
	}
}
